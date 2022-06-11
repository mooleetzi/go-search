package service

import (
	"fmt"
	"go-search/global"
	"go-search/searcher"
	"go-search/searcher/searchlog"
)

type ScLog struct {
	Container *searcher.Container
}

func NewScLog() *ScLog {
	return &ScLog{
		Container: global.Container,
	}
}

// 更新后继词表
func (s *ScLog) UpdatedRelatedSearch(isclear string) (ss string) {
	searchlog.UpdatedRelatedSearch(isclear)
	ss = string("success")
	fmt.Println(ss)
	return ss
}
