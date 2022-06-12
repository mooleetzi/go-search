package service

import (
	"go-search/global"
	"go-search/searcher"
	"go-search/searcher/searchlog"
)

type Database struct {
	Container *searcher.Container
}
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
	rs := s.Container.GetDataBase("default")
	searchlog.UpdatedRelatedSearch(isclear, rs.GetRelatedStorage())
	ss = string("update related search db")
	// fmt.Println(ss)
	return ss
}
