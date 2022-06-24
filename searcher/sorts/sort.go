package sorts

import (
	"go-search/searcher/model"
	"sort"
	"sync"
)

type ScoreSlice []model.SliceItem

func (x ScoreSlice) Len() int {
	return len(x)
}
func (x ScoreSlice) Less(i, j int) bool {
	if x[i].Score == x[j].Score {
		return x[i].Id < x[j].Id
	}
	return x[i].Score < x[j].Score
}
func (x ScoreSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

type SortResult struct {
	sync.Mutex

	IsDebug bool

	IdsAndScores []model.SliceItem

	Ids    []uint32
	Scores []float64
	//Words []string

	count int // 总数

	Order string // 排序方式
}

//func (f *SortResult) Add(ids *[]uint32) {
//	f.Ids = append(f.Ids, *ids...)
//}
func (f *SortResult) Add(idsToFreqs *map[uint32]float64) {
	f.Lock()
	defer f.Unlock()
	for id, score := range *idsToFreqs {
		f.Ids = append(f.Ids, id)
		f.Scores = append(f.Scores, score)
	}
}

//func (f *SortResult) find(target *uint32) (bool, int) {
//	low := 0
//	high := f.count - 1
//	for low <= high {
//		mid := (low + high) / 2
//		if f.IdsAndScores[mid].Id == *target {
//			return true, mid
//		} else if f.IdsAndScores[mid].Id < *target {
//			high = mid - 1
//		} else {
//			low = mid + 1
//		}
//	}
//	return false, -1
//}

func (f *SortResult) Process(block []uint32) {
	tmp := make(map[uint32]float64, len(f.Ids))
	if len(f.Ids) != len(f.Scores) {
		panic("!!!!!!!")
	}
	for pos, id := range f.Ids {
		if _, err := tmp[id]; err {
			tmp[id] = f.Scores[pos]
		} else {
			tmp[id] += f.Scores[pos]
		}
	}
	if len(block) != 0 {
		for _, blockId := range block {
			delete(tmp, blockId)
		}
	}

	for key, val := range tmp {
		f.IdsAndScores = append(f.IdsAndScores, model.SliceItem{
			Id:    key,
			Score: val,
		})
		f.count++
	}

	sort.Sort(sort.Reverse(ScoreSlice(f.IdsAndScores)))
}

func (f *SortResult) Count() int {
	return f.count
}

func (f *SortResult) GetAll(i *[]model.SliceItem, start int, end int) {
	*i = f.IdsAndScores[start:end]
}
