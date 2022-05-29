package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"go-search/core"
	"go-search/searcher/model"
	"go-search/searcher/storage"
	"go-search/searcher/utils"
	"go-search/searcher/words"
	"io"
	"log"
	"os"
)

const (
	INIT = false
)

func main() {
	if INIT {
		initDB()
	}
	// 初始化容器和参数解析
	core.Initialize()
}

//初始化db
func initDB() {
	path := "/Users/mool/Downloads/wukong50k_release.csv"
	csvFile, _ := os.Open(path)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	isTitle := true
	tokenizer := words.NewTokenizer("./searcher/words/data/dict.txt")
	db, _ := storage.NewStorage("./wukong.db", 1000)
	id := (uint32)(0)
	for {
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
		_, keys := tokenizer.Cut(line[1])
		doc := model.StorageIndexDoc{
			IndexDoc: &model.IndexDoc{
				Id:   id + 1,
				Text: line[1],
				Url:  line[0],
			},
			Keys: keys,
		}
		db.Set([]byte(fmt.Sprint(id)), utils.Encoder(doc))
		id += 1
	}
}
