package searcher

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"go-search/pagination"
	"go-search/searcher/arrays"
	"go-search/searcher/model"
	"go-search/searcher/searchlog"
	"go-search/searcher/sorts"
	"go-search/searcher/storage"
	"go-search/searcher/utils"
	"go-search/searcher/words"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jasonlvhit/gocron"
)

type Engine struct {
	IndexPath string  // 索引文件存储目录
	Option    *Option // 配置各种数据库名称

	invertedIndexStorages []*storage.LeveldbStorage // 关键字和Id映射，倒排索引,key=id,value=[]words
	positiveIndexStorages []*storage.LeveldbStorage // ID和key映射，用于计算相关度，一个id 对应多个key，正排索引
	docStorages           []*storage.LeveldbStorage // 文档仓
	relatedStorages       []*storage.LeveldbStorage // 后继词表

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
	RelatedIndexName  string // 后继表存储
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

		rs, err := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.RelatedIndexName, shard)), e.Timeout)
		if err != nil {
			panic(err)
		}
		e.relatedStorages = append(e.relatedStorages, rs)

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
	go e.automaticUpdate()
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

// 自动更新后继词表
func (e *Engine) automaticUpdate() {
	//等待初始化完成
	e.Wait()
	//debug用，每10s触发一次更新
	// gocron.Every(10).Second().DoSafely(e.addSearchLogToRelatedStorage, "")
	//
	gocron.Every(1).Day().At("0:25").DoSafely(e.addSearchLogToRelatedStorage, "")
	<-gocron.Start()
}

func (e *Engine) getFilePath(fileName string) string {
	return e.IndexPath + string(os.PathSeparator) + fileName
}
func (e *Engine) DocumentWorkerExec(worker chan *model.IndexDoc) {
	//启用一个无限循环的进程等待worker传递东西进来
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
	log.Println("开始添加悟空数据集")
	e.InitWuKong()
	log.Println("开始添加初始后继词数据集")
	e.InitRelatedSearch()

}

func (e *Engine) GetOptions() *Option {
	return &Option{
		DocIndexName:      "docs",
		InvertedIndexName: "inverted_index",
		PositiveIndexName: "positive_index",
		RelatedIndexName:  "related_search",
	}
}

func (e *Engine) MultiSearch(request *model.SearchRequest) *model.SearchResult {
	// 等待搜索初始化完成
	e.Wait()

	//记录搜索log
	e.addSearchLog(request)

	// 分词搜索
	_, splitWords := e.Tokenizer.Cut(request.Query)
	_, blockWords := e.Tokenizer.Cut(request.Block)
	splitWords = utils.SliceDiffStr(splitWords, blockWords)

	totalTime := float64(0)

	sortResult := &sorts.SortResult{
		IsDebug: e.IsDebug,
		Order:   request.Order,
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

	//检索 相关搜索词
	relatedResult := make([]string, 0)
	// relatedResult, _timerelated := e.relatedSearch(splitWords, relatedResult)
	// searchword := splitWords
	searchword := append(splitWords, request.Query)

	_timerelated := e.relatedSearch(searchword, &relatedResult) //分词or全
	if e.IsDebug {
		log.Println("相关搜索时间:", _timerelated, "ms")
	}

	wordMap := make(map[string]bool)
	for _, word := range splitWords {
		wordMap[word] = true
	}

	// 读取文档
	var result = &model.SearchResult{
		Total:     sortResult.Count(),
		Page:      request.Page,
		Limit:     request.Limit,
		Words:     splitWords,
		RelatedSc: relatedResult,
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

func (e *Engine) relatedSearch(words []string, Result *[]string) (_time float64) {
	temp := make(map[string]bool)
	newwords := make([]string, 0)
	for _, word := range words { //去重
		_, ok := temp[word]
		if !ok { //去重
			temp[word] = true
			newwords = append(newwords, word)
		}
	}

	_time = utils.ExecTime(func() {
		base := len(newwords)
		wg := &sync.WaitGroup{}
		wg.Add(base)

		for _, word := range newwords {
			go e.processKeyRelatedSearch(word, Result, wg)
		}

		wg.Wait()
	})

	return

}

func (e *Engine) processKeyRelatedSearch(word string, Result *[]string, wg *sync.WaitGroup) {
	defer wg.Done()
	buf, found := e.relatedStorages[0].Get([]byte(word))

	if found {
		storageDoc := new(model.IndexRelated)
		utils.Decoder(buf, &storageDoc)
		suc := storageDoc.Success
		// keyword := storageDoc.KeyWord
		// fmt.Println(word)

		for _, r := range suc {
			if r != "" {
				*Result = append(*Result, r)
			}
		}
		// fmt.Println(Result, found)

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
func (e *Engine) InitWuKong() {
	path := "./searcher/wukong50k_release.csv"
	csvFile, _ := os.Open(path)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	isTitle := true
	id := (uint32)(0)
	//wg := new(sync.WaitGroup)
	exectime := utils.ExecTime(func() {
		for {
			fmt.Printf("%v \n", id)
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			if isTitle {
				isTitle = false
				continue
			}
			doc := model.IndexDoc{
				Id:   id + 1,
				Text: line[1],
				Url:  line[0],
			}
			e.IndexDocument(&doc)
			id += 1
		}
	})
	fmt.Println(exectime/1e3, "s add wukong_5k into workchan")
}

func (e *Engine) InitRelatedSearch() { //初始化后继词表
	path := "./searcher/related_searchs100.csv"
	csvFile, _ := os.Open(path)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	isTitle := true
	id := (uint32)(0)
	//wg := new(sync.WaitGroup)
	exectime := utils.ExecTime(func() {
		for {
			fmt.Printf("%v \n", id)
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			if isTitle {
				isTitle = false
				continue
			}
			res := strings.Split(line[1], ",")
			doc := model.IndexRelated{
				Id:      id + 1,
				Success: res,
				KeyWord: line[0],
			}
			e.relatedStorages[0].Set([]byte(line[0]), utils.Encoder(doc))
			id += 1
		}
	})
	fmt.Println(exectime/1e3, "s add related_search")
}

func (e *Engine) IndexDocument(doc *model.IndexDoc) {
	//将一个doc放入到数据库当中
	e.addDocumentWorkerChan[e.getShard(doc.Id)] <- doc
}

// GetQueue 获取队列剩余
func (e *Engine) GetQueue() int {
	total := 0
	for _, v := range e.addDocumentWorkerChan {
		total += len(v)
	}
	return total
}

// 添加日志
func (e *Engine) addSearchLog(request *model.SearchRequest) {
	e.Wait()

	e.Lock()
	defer e.Unlock()
	// fmt.Println("添加日志")

	//创建一个新文件
	newFileName := "./searcher/searchlog.csv"
	//这样打开，每次都会清空文件内容
	//nfs, err := os.Create(newFileName)

	//这样可以追加写
	nfs, err := os.OpenFile(newFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("can not create file, err is %+v", err)
	}
	defer nfs.Close()
	nfs.Seek(0, io.SeekEnd)

	w := csv.NewWriter(nfs)
	//设置属性
	w.Comma = ','
	w.UseCRLF = true

	row := []string{request.ClientIP, request.Query, strconv.FormatInt(request.Time, 10)}
	err = w.Write(row)
	if err != nil {
		log.Fatalf("can not write, err is %+v", err)
	}
	//这里必须刷新，才能将数据写入文件。
	w.Flush()

}

// 读取日志
func (e *Engine) addSearchLogToRelatedStorage(isclean string) {
	e.Wait()
	// e.Lock()

	searchlog.UpdatedRelatedSearch(isclean, e.relatedStorages[0])
	// defer e.Unlock()

}

func (e *Engine) GetRelatedStorage() *storage.LeveldbStorage {
	//等待初始化完成
	e.Wait()
	return e.relatedStorages[0]
}
