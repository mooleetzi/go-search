package service

import (
	"go-search/global"
	"go-search/searcher"
)

type Word struct {
	Container *searcher.Container
}

func NewWord() *Word {
	return &Word{
		Container: global.Container,
	}
}

// WordCut 分词
func (w *Word) WordCut(keyword string) (ss []string) {
	_, ss = w.Container.Tokenizer.Cut(keyword)
	return
}
