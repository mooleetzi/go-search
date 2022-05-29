package seacher

import (
	"goSearch/seacher/arrays"
	"goSearch/seacher/model"
	"goSearch/seacher/storage"
	"goSearch/seacher/utils"
	"goSearch/seacher/words"
	"sync"
)

type Engine struct {
	invertedIndexStorages []*storage.LeveldbStorage //关键字和Id映射，倒排索引,key=id,value=[]words
	positiveIndexStorages []*storage.LeveldbStorage //ID和key映射，用于计算相关度，一个id 对应多个key，正排索引
	docStorages           []*storage.LeveldbStorage //文档仓
	sync.WaitGroup
	sync.Mutex
	addDocumentWorkerChan []chan *model.IndexDoc
	Shard                 int              //分片数
	Tokenizer             *words.Tokenizer //分词器
}

func (e *Engine) Init() {
	e.Add(1)
	defer e.Done()

	e.addDocumentWorkerChan = make([]chan *model.IndexDoc, e.Shard)
	for shard := 0; shard < e.Shard; shard++ {
		worker := make(chan *model.IndexDoc, 1000)
		e.addDocumentWorkerChan[shard] = worker
		go e.DocumentWorkerExec(worker)
	}
}
func (e *Engine) DocumentWorkerExec(worker chan *model.IndexDoc) {
	for {
		doc := <-worker
		e.AddDocument(doc)
	}
}
func (e *Engine) AddDocument(index *model.IndexDoc) {
	//等待初始化完成
	e.Wait()
	text := index.Text

	wordsToFreqs, words := e.Tokenizer.Cut(text)

	id := index.Id
	isUpdate := e.deleteInvalidDocId(id, words)
	//没有更新
	if !isUpdate {
		return
	}

	for word, freq := range *wordsToFreqs {
		e.addInvertedIndex(word, freq, id)
	}
	//添加id索引
	e.addPositiveIndex(index, words)
}

//	移除没有的词
func (e *Engine) deleteInvalidDocId(id uint32, newWords []string) bool {
	//判断id是否存在
	e.Lock()
	defer e.Unlock()

	//从id的正向索引中读取并和当前分词结果比较
	removes, found := e.getDifference(id, newWords)
	if found && len(removes) > 0 {
		//从倒序索引的存储中删除无效docId
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

		//计算需要移除的
		removes := make([]string, 0)
		for _, word := range oldWords {

			//旧的在新的里面不存在，就是需要移除的
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
	//找到dB
	shard := e.getShardByWord(word)
	s := e.invertedIndexStorages[shard]
	//word作为key在反向索引dB中查找
	key := []byte(word)
	buf, find := s.Get(key)
	docIdsToFreqs := make(map[uint32]int)
	if find {
		utils.Decoder(buf, &docIdsToFreqs)
	}
	//map增加和访问
	docIdsToFreqs[id] = frequency
	s.Set(key, utils.Encoder(docIdsToFreqs))
}

func (e *Engine) getShardByWord(word string) int {

	return int(utils.StringToInt(word) % uint32(e.Shard))
}

func (e *Engine) delInInvertedStorage(id uint32, word string) {
	//word为倒序索引,id为待删除doc
	shard := e.getShardByWord(word)
	wordStorage := e.invertedIndexStorages[shard]

	//string作为key
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

	//id和key的映射
	positiveIndexStorage := e.positiveIndexStorages[shard]

	doc := &model.StorageIndexDoc{
		IndexDoc: index,
		Keys:     keys,
	}

	//存储id和key以及文档的映射
	docStorage.Set(key, utils.Encoder(doc))

	//设置到id和key的映射中
	positiveIndexStorage.Set(key, utils.Encoder(keys))
}
