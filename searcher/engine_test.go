package searcher

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"go-search/searcher/utils"
	"io"
	"os"
	"sync"
	"testing"
)

func TestWuKong(t *testing.T) {
	path := "./wukong50k_release.csv"
	csvFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	wg := sync.WaitGroup{}
	id := uint32(0)
	time := utils.ExecTime(func() {
		for {
			wg.Add(1)
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println("!!")
			}
			fmt.Printf("%v %v", line[0], line[1])
			id += 1
		}
		wg.Wait()
	})
	fmt.Println(time)
}
