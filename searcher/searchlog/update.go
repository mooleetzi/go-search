package searchlog

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

func UpdatedRelatedSearch(isclear string) {
	//准备读取文件
	fileName := "./searcher/searchlog.csv"

	if isclear == "clear" {
		fmt.Println("clear!!")
		//这样打开，每次都会清空文件内容
		// os.Create(fileName)
	}

	fs, err := os.Open(fileName)

	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer fs.Close()

	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	for {
		// row, err := r.Read()
		_, err := r.Read()
		if err != nil && err != io.EOF {
			log.Fatalf("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}
		// fmt.Println(row)
	}
	return
}
