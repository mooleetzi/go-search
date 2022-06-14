package service

import (
	"go-search/global"
	"go-search/searcher"
	"go-search/searcher/model"
	"runtime"
	"strings"
	"time"
)

// Base 基础管理
type Base struct {
	Container *searcher.Container
	Callback  func() map[string]interface{}
}

func NewBase() *Base {
	return &Base{
		Container: global.Container,
		Callback:  Callback,
	}
}

// Query 查询
func (b *Base) Query(request *model.SearchRequest) *model.SearchResult {
	//是否匹配到“- ”？第一段为原本的查询，第二段为阻塞
	ss := strings.Split(request.Query, " -")
	request.Query = ss[0]
	if len(ss) > 1 {
		request.Block = ss[1]
	}
	//获取调用search query的时间戳
	timeUnix := time.Now().Unix() //单位s,打印结果:1491888244
	request.Time = timeUnix
	return b.Container.GetDataBase(request.Database).MultiSearch(request)
}

// GC 释放GC
func (b *Base) GC() {
	runtime.GC()
}
