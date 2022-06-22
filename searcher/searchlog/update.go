package searchlog

import (
	"encoding/csv"
	"go-search/searcher/model"
	"go-search/searcher/storage"
	"go-search/searcher/utils"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
)

type SearchLog []model.SearchLog

func (s SearchLog) Len() int { return len(s) }

func (s SearchLog) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s SearchLog) Less(i, j int) bool { return s[i].Time < s[j].Time }

func UpdatedRelatedSearch(isclear string, rs *storage.LeveldbStorage, logMem [][]string) {
	// rs, err := storage.NewStorage(fmt.Sprintf("%s_%d", "related_search", 0), 1000)
	// if err != nil {
	// 	log.Fatalf("can not open the realtedsearch db, err is %+v", err)
	// }
	//准备读取文件

	// logMem := global.Container.GetLogMem()

	fileName := "./searcher/searchlog.csv"

	fs, err := os.Open(fileName)

	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer fs.Close()

	users := make(map[string][]model.SearchLog)

	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	for {
		row, err := r.Read()
		//row0 row1 row2
		if err != nil && err != io.EOF {
			log.Fatalf("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}
		time, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			log.Fatalf("can not read timeunix from log, err is %+v", err)
		}
		//按用户ip进行分组
		users[row[0]] = append(users[row[0]], model.SearchLog{
			Query: row[1],
			Time:  time,
		})
		// fmt.Println(row)
	}
	for _, row := range logMem {
		time, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			log.Fatalf("can not read timeunix from log, err is %+v", err)
		}
		users[row[0]] = append(users[row[0]], model.SearchLog{
			Query: row[1],
			Time:  time,
		})
	}

	//每个用户ip按每5min分组
	addlog := new(model.IndexRelated)
	group := make([]model.IndexRelated, 0)

	//时间戳从小到大排序
	for _, user := range users {
		sort.Sort(SearchLog(user))
	}
	//
	// for _, user := range users {
	// 	for _, log := range user {
	// 		fmt.Println(log)
	// 	}
	// }

	for _, user := range users {
		length := len(user)
		begintime := user[0].Time
		addlog.KeyWord = user[0].Query
		temp := make(map[string]bool)
		temp[user[0].Query] = true

		for i := 1; i < length; i++ {
			if user[i].Time-begintime > 20 {
				// fmt.Println(user[i].Query)
				// fmt.Println(addlog.KeyWord, addlog.Success)

				group = append(group, *addlog)

				begintime = user[i].Time
				addlog.KeyWord = user[i].Query
				addlog.Success = []string{}
				for k := range temp {
					delete(temp, k)
				}
				temp[user[i].Query] = true
			}

			_, ok := temp[user[i].Query]
			if !ok {
				addlog.Success = append(addlog.Success, user[i].Query)
				temp[user[i].Query] = true
			}

		}
		group = append(group, *addlog)
		// fmt.Println(addlog.KeyWord)
	}

	//分组结果更新到后继词表
	for _, g := range group {
		// fmt.Println(g.KeyWord, g.Success)
		buf, found := rs.Get([]byte(g.KeyWord))
		if found {
			old := new(model.IndexRelated)
			utils.Decoder(buf, &old)
			//去重 更新？？
			temp := make(map[string]bool)
			for _, oldsuc := range old.Success {
				temp[oldsuc] = true
			}
			for _, gsuc := range g.Success {
				temp[gsuc] = true
			}
			g.Success = []string{}
			for newsuc := range temp {
				g.Success = append(g.Success, newsuc)
			}
			// fmt.Println(g.KeyWord, g.Success)
			err := rs.Delete([]byte(g.KeyWord))
			if err != nil {
				log.Fatalf("can not delete relatedsearch, err is %+v", err)

			}
			// fmt.Println(rs.Has([]byte(g.KeyWord)))
			rs.Set([]byte(g.KeyWord), utils.Encoder(g))

		} else {
			rs.Set([]byte(g.KeyWord), utils.Encoder(g))
		}

	}
	log.Println("整理搜索log并更新后继词表")

	if isclear == "clear" {
		log.Println("search log clear!!")
		//清空文件内容
		os.Create(fileName)
	}
	return
}
