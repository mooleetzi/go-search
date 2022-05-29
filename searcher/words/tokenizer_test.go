package words

import (
	"fmt"
	"testing"
)

func TestNewTokenizer(t *testing.T) {

	tokenizer := NewTokenizer("./data/dict.txt")
	resChan := tokenizer.seg.CutForSearch("我想要实习！！", true)
	for {
		word, ok := <-resChan
		if !ok {
			break
		}
		fmt.Println(word)
	}
}
