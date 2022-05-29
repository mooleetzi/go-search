package searcher

import (
	"fmt"
	"go-search/pagination"
	"go-search/searcher/arrays"
	"go-search/searcher/model"
	"go-search/searcher/sorts"
	"go-search/searcher/storage"
	"go-search/searcher/utils"
	"go-search/searcher/words"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	IndexPath string  // 索引文件存储目录
	Option    *Option // 配置各种数据库名称

	invertedIndexStorages []*storage.LeveldbStorage // 关键字和Id映射，倒排索引,key=id,value=[]words
	positiveIndexStorages []*storage.LeveldbStorage // ID和key映射，用于计算相关度，一个id 对应多个key，正排索引
	docStorages           []*storage.LeveldbStorage // 文档仓

	sync.WaitGroup
	sync.Mutex
	Tokenizer             *words.Tokenizer // 分词器
	addDocumentWorkerChan []chan *model.IndexDoc
	DatabaseName          string // 数据库名

	Shard   int   // 分片数
	Timeout int64 // 超时时间,单位秒
	IsDebug bool  // 是否调试模式
}
type Option struct {
	InvertedIndexName string // 倒排索引
	PositiveIndexName string // 正排索引
	DocIndexName      string // 文档存储
}

func (e *Engine) Init() {
	e.Add(1)
	defer e.Done()

	e.addDocumentWorkerChan = make([]chan *model.IndexDoc, e.Shard)
	for shard := 0; shard < e.Shard; shard++ {
		worker := make(chan *model.IndexDoc, 1000)
		e.addDocumentWorkerChan[shard] = worker
		go e.DocumentWorkerExec(worker)
		s, err := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.DocIndexName, shard)), e.Timeout)
		if err != nil {
			panic(err)
		}
		e.docStorages = append(e.docStorages, s)

		// 初始化Keys存储
		ks, kerr := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.InvertedIndexName, shard)), e.Timeout)
		if kerr != nil {
			panic(err)
		}
		e.invertedIndexStorages = append(e.invertedIndexStorages, ks)

		// id和keys映射
		iks, ikerr := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.PositiveIndexName, shard)), e.Timeout)
		if ikerr != nil {
			panic(ikerr)
		}
		e.positiveIndexStorages = append(e.positiveIndexStorages, iks)

	}
	go e.automaticGC()
	log.Println("初始化完成")
}

// 自动保存索引，10秒钟检测一次
func (e *Engine) automaticGC() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		<-ticker.C
		// 定时GC
		runtime.GC()
	}
}
func (e *Engine) getFilePath(fileName string) string {
	return e.IndexPath + string(os.PathSeparator) + fileName
}
func (e *Engine) DocumentWorkerExec(worker chan *model.IndexDoc) {
	for {
		doc := <-worker
		e.AddDocument(doc)
	}
}
func (e *Engine) AddDocument(index *model.IndexDoc) {
	// 等待初始化完成
	e.Wait()
	text := index.Text

	wordsToFreqs, splitwords := e.Tokenizer.Cut(text)

	id := index.Id
	isUpdate := e.deleteInvalidDocId(id, splitwords)
	// 没有更新
	if !isUpdate {
		return
	}

	for word, freq := range *wordsToFreqs {
		e.addInvertedIndex(word, freq, id)
	}
	// 添加id索引
	e.addPositiveIndex(index, splitwords)
}

//	移除没有的词
func (e *Engine) deleteInvalidDocId(id uint32, newWords []string) bool {
	// 判断id是否存在
	e.Lock()
	defer e.Unlock()

	// 从id的正向索引中读取并和当前分词结果比较
	removes, found := e.getDifference(id, newWords)
	if found && len(removes) > 0 {
		// 从倒序索引的存储中删除无效docId
		for _, word := range removes {
			e.delInInvertedStorage(id, word)
		}
	}

	// 有没有更新
	return !found || len(removes) > 0
}

func (e *Engine) getDifference(id uint32, newWords []string) ([]string, bool) {

	shard := e.getShard(id)
	wordStorage := e.positiveIndexStorages[shard]
	key := utils.Uint32ToBytes(id)
	buf, found := wordStorage.Get(key)
	if found {
		oldWords := make([]string, 0)
		utils.Decoder(buf, &oldWords)

		// 计算需要移除的
		removes := make([]string, 0)
		for _, word := range oldWords {

			// 旧的在新的里面不存在，就是需要移除的
			if !arrays.ArrayStringExists(newWords, word) {
				removes = append(removes, word)
			}
		}
		return removes, true
	}
	return nil, false
}

// getShard 计算索引分布在哪个文件块
func (e *Engine) getShard(id uint32) int {
	return int(id % uint32(e.Shard))
}

// 添加倒排索引
func (e *Engine) addInvertedIndex(word string, frequency int, id uint32) {
	e.Lock()
	defer e.Unlock()
	// 找到dB
	shard := e.getShardByWord(word)
	s := e.invertedIndexStorages[shard]
	// word作为key在反向索引dB中查找
	key := []byte(word)
	buf, find := s.Get(key)
	docIdsToFreqs := make(map[uint32]int)
	if find {
		utils.Decoder(buf, &docIdsToFreqs)
	}
	// map增加和访问
	docIdsToFreqs[id] = frequency
	s.Set(key, utils.Encoder(docIdsToFreqs))
}

func (e *Engine) getShardByWord(word string) int {

	return int(utils.StringToInt(word) % uint32(e.Shard))
}

func (e *Engine) delInInvertedStorage(id uint32, word string) {
	// word为倒序索引,id为待删除doc
	shard := e.getShardByWord(word)
	wordStorage := e.invertedIndexStorages[shard]

	// string作为key
	key := []byte(word)

	buf, found := wordStorage.Get(key)
	if found {
		//	读取文件存储
		wordsToFreqs := make(map[uint32]int)
		utils.Decoder(buf, &wordsToFreqs)
		//	移除map中的无效id
		delete(wordsToFreqs, id)

		if len(wordsToFreqs) == 0 {
			err := wordStorage.Delete(key)
			if err != nil {
				panic(err)
			}
		} else {
			wordStorage.Set(key, utils.Encoder(wordsToFreqs))
		}
	}

}

func (e *Engine) addPositiveIndex(index *model.IndexDoc, keys []string) {
	e.Lock()
	defer e.Unlock()

	key := utils.Uint32ToBytes(index.Id)
	shard := e.getShard(index.Id)
	docStorage := e.docStorages[shard]

	// id和key的映射
	positiveIndexStorage := e.positiveIndexStorages[shard]

	doc := &model.StorageIndexDoc{
		IndexDoc: index,
		Keys:     keys,
	}

	// 存储id和key以及文档的映射
	docStorage.Set(key, utils.Encoder(doc))

	// 设置到id和key的映射中
	positiveIndexStorage.Set(key, utils.Encoder(keys))
}

// GetIndexCount 获取索引数量
func (e *Engine) GetIndexCount() int64 {
	var size int64
	for i := 0; i < e.Shard; i++ {
		size += e.invertedIndexStorages[i].GetCount()
	}
	return size
}

// GetDocumentCount 获取文档数量
func (e *Engine) GetDocumentCount() int64 {
	var count int64
	for i := 0; i < e.Shard; i++ {
		count += e.docStorages[i].GetCount()
	}
	return count
}
func (e *Engine) InitOption(option *Option) {

	if option == nil {
		// 默认值
		option = e.GetOptions()
	}
	e.Option = option
	// shard默认值
	if e.Shard <= 0 {
		e.Shard = 10
	}
	// 初始化其他的
	e.Init()

}

func (e *Engine) GetOptions() *Option {
	return &Option{
		DocIndexName:      "docs",
		InvertedIndexName: "inverted_index",
		PositiveIndexName: "positive_index",
	}
}

func (e *Engine) MultiSearch(request *model.SearchRequest) *model.SearchResult {

	// 等待搜索初始化完成
	e.Wait()
	// 分词搜索
	_, splitWords := e.Tokenizer.Cut(request.Query)
	_, blockWords := e.Tokenizer.Cut(request.Block)
	splitWords = utils.SliceDiffStr(splitWords, blockWords)

	totalTime := float64(0)

	sortResult := &sorts.SortResult{
		IsDebug: e.IsDebug,
		Order:   request.Order,
		//Words:   splitWords,
	}
	blockSortResult := &sorts.SortResult{
		IsDebug: e.IsDebug,
		Order:   request.Order,
	}

	_time := e.search(splitWords, sortResult)
	_time += e.search(blockWords, blockSortResult)
	if e.IsDebug {
		log.Println("数组查找耗时：", totalTime, "ms")
		log.Println("搜索时间:", _time, "ms")
	}

	sortResult.Ids = utils.SliceDiffI32(sortResult.Ids, blockSortResult.Ids)

	// 处理分页
	request = request.GetAndSetDefault()

	// 计算得分
	sortResult.Process()

	wordMap := make(map[string]bool)
	for _, word := range splitWords {
		wordMap[word] = true
	}

	// 读取文档
	var result = &model.SearchResult{
		Total: sortResult.Count(),
		Page:  request.Page,
		Limit: request.Limit,
		Words: splitWords,
	}

	_time += utils.ExecTime(func() {

		pager := new(pagination.Pagination)

		pager.Init(request.Limit, sortResult.Count())
		// 设置总页数
		result.PageCount = pager.PageCount

		// 读取单页的id
		if pager.PageCount != 0 {

			start, end := pager.GetPage(request.Page)

			var resultItems = make([]model.SliceItem, 0)
			sortResult.GetAll(&resultItems, start, end)

			count := len(resultItems)

			result.Documents = make([]model.ResponseDoc, count)
			// 只读取前面100个
			wg := new(sync.WaitGroup)
			wg.Add(count)
			for index, item := range resultItems {
				go e.getDocument(item, &result.Documents[index], request, &wordMap, wg)
			}
			wg.Wait()
		}
	})
	if e.IsDebug {
		log.Println("处理数据耗时：", _time, "ms")
	}
	result.Time = _time

	return result
}

func (e *Engine) search(words []string, sortResult *sorts.SortResult) (_time float64) {
	_time = utils.ExecTime(func() {
		base := len(words)
		wg := &sync.WaitGroup{}
		wg.Add(base)

		for _, word := range words {
			go e.processKeySearch(word, sortResult, wg)
		}
		wg.Wait()
	})
	return
}

func (e *Engine) processKeySearch(word string, sortResult *sorts.SortResult, wg *sync.WaitGroup) {
	defer wg.Done()

	shard := e.getShardByWord(word)
	// 读取id
	invertedIndexStorage := e.invertedIndexStorages[shard]
	key := []byte(word)

	buf, find := invertedIndexStorage.Get(key)
	if find {
		idsToFreqs := make(map[uint32]int)
		// 解码
		utils.Decoder(buf, &idsToFreqs)

		scores := make(map[uint32]float64)
		for id, freq := range idsToFreqs {
			docFreq := float64(len(idsToFreqs))
			docCount := float64(e.GetCountById(id))
			idf := math.Log(docCount) - math.Log(docFreq+1) + 1
			tf := math.Sqrt(float64(freq))
			scores[id] = idf * tf
		}

		sortResult.Add(&scores)
	}

}

func (e *Engine) getDocument(item model.SliceItem, doc *model.ResponseDoc, request *model.SearchRequest, wordMap *map[string]bool, wg *sync.WaitGroup) {
	buf := e.GetDocById(item.Id)
	defer wg.Done()
	doc.Score = item.Score
	if buf != nil {
		// gob解析
		storageDoc := new(model.StorageIndexDoc)
		utils.Decoder(buf, &storageDoc)
		doc.Url = storageDoc.Url
		doc.Keys = storageDoc.Keys
		text := storageDoc.Text
		// 处理关键词高亮
		highlight := request.Highlight
		if highlight != nil {
			// 全部小写
			text = strings.ToLower(text)
			// 还可以优化，只替换击中的词
			for _, key := range storageDoc.Keys {
				if ok := (*wordMap)[key]; ok {
					text = strings.ReplaceAll(text, key, fmt.Sprintf("%s%s%s", highlight.PreTag, key, highlight.PostTag))
				}
			}
			// 放置原始文本
			doc.OriginalText = storageDoc.Text
		}
		doc.Text = text
		doc.Id = item.Id

	}
}

// GetDocById 通过id获取文档
func (e *Engine) GetDocById(id uint32) []byte {
	shard := e.getShard(id)
	key := utils.Uint32ToBytes(id)
	buf, found := e.docStorages[shard].Get(key)
	if found {
		return buf
	}

	return nil
}

func (e *Engine) GetCountById(id uint32) int64 {
	shard := e.getShard(id)
	return e.docStorages[shard].GetCount()
}
