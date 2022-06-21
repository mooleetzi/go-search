package searcher

import (
	"encoding/csv"
	"fmt"
	"go-search/searcher/words"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"unsafe"
)

type Container struct {
	Dir       string             // 文件夹
	engines   map[string]*Engine // 引擎
	Debug     bool               // 调试
	Tokenizer *words.Tokenizer   // 分词器
	Shard     int                // 分片
	Timeout   int64              // 超时关闭数据库
	logMem    [][]string
	sync.Mutex
}

func (c *Container) Init() error {

	c.logMem = make([][]string, 0, 1024)
	c.engines = make(map[string]*Engine)

	// 读取当前路径下的所有目录，就是数据库名称
	dirs, err := ioutil.ReadDir(c.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建
			err := os.MkdirAll(c.Dir, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// 初始化数据库
	for _, dir := range dirs {
		if dir.IsDir() {
			c.engines[dir.Name()] = c.GetDataBase(dir.Name())
			log.Println("db:", dir.Name())
		}
	}

	return nil
}

//创建一个引擎
func (c *Container) NewEngine(name string) *Engine {
	var engine = &Engine{
		IndexPath:    fmt.Sprintf("%s%c%s", c.Dir, os.PathSeparator, name),
		DatabaseName: name,
		Tokenizer:    c.Tokenizer,
		Shard:        c.Shard,
		Timeout:      c.Timeout,
		container:    c,
	}
	option := engine.GetOptions()

	engine.InitOption(option)
	engine.IsDebug = c.Debug
	return engine
}

// GetDataBase 获取或创建引擎
func (c *Container) GetDataBase(name string) *Engine {

	// 默认数据库名为default
	if name == "" {
		name = "default"
	}

	// log.Println("Get DataBase:", name)
	engine, ok := c.engines[name]
	if !ok {
		// 创建引擎
		engine = c.NewEngine(name)
		c.engines[name] = engine
		// 释放引擎
	}

	return engine
}

// GetDataBases 获取数据库列表
func (c *Container) GetDataBases() map[string]*Engine {
	for _, engine := range c.engines {
		size := unsafe.Sizeof(&engine)
		fmt.Printf("%s:%d\n", engine.DatabaseName, size)
	}
	return c.engines
}

func (c *Container) GetDataBaseNumber() int {
	return len(c.engines)
}

func (c *Container) GetIndexCount() int64 {
	var count int64
	for _, engine := range c.engines {
		count += engine.GetIndexCount()
	}
	return count
}

func (c *Container) GetDocumentCount() int64 {
	var count int64
	for _, engine := range c.engines {
		count += engine.GetDocumentCount()
	}
	return count
}
func (c *Container) GetLogMem() [][]string {
	c.Lock()
	defer c.Unlock()
	return c.logMem
}
func (c *Container) AddLogMem(row []string) {
	c.Lock()

	c.logMem = append(c.logMem, row)
	if len(c.logMem) > 1024 {
		c.Unlock()
		c.MustWriteLog()
		return
	}
	c.Unlock()
}
func (c *Container) MustWriteLog() {
	c.Lock()
	defer c.Unlock()
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
	_, err = nfs.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatalf("can not create file, err is %+v", err)
	}
	w := csv.NewWriter(nfs)
	//设置属性
	w.Comma = ','
	w.UseCRLF = true

	err = w.WriteAll(c.logMem)
	if err != nil {
		log.Fatalf("can not write, err is %+v", err)
	}
	c.clearLog()
	//这里必须刷新，才能将数据写入文件。
	w.Flush()
}

func (c *Container) clearLog() { //必须在持有锁的情况调用
	c.logMem = make([][]string, 0, 1024)
}
