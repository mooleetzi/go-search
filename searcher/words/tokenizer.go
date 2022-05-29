package words

import (
	"fmt"
	"github.com/wangbin/jiebago"
	"go-search/searcher/utils"
	"strings"
)

type Tokenizer struct {
	seg jiebago.Segmenter
}

func NewTokenizer(filename string) *Tokenizer {
	tokenizer := &Tokenizer{}
	fmt.Println("filename:", filename)
	err := tokenizer.seg.LoadDictionary(filename)
	if err != nil {
		panic(err)
	}
	return tokenizer
}

// 安全返回liangge切片
func (t *Tokenizer) Cut(text string) (*map[string]int, []string) {
	// 不区分大小写
	text = strings.ToLower(text)
	// 移除所有的标点符号
	text = utils.RemovePunctuation(text)
	// 移除所有的空格
	text = utils.RemoveSpace(text)

	var wordMap = make(map[string]int)

	resultChan := t.seg.CutForSearch(text, true)
	for {
		w, ok := <-resultChan
		if !ok {
			break
		}
		_, found := wordMap[w]
		if !found {
			// 去除重复的词
			wordMap[w] = 1
		} else {
			wordMap[w] = wordMap[w] + 1
		}
	}
	num := len(wordMap)
	words := make([]string, num)
	// freqs := make([]int, num)
	for k := range wordMap {
		words = append(words, k)
		// freqs = append(freqs, v)
	}
	return &wordMap, words
}
