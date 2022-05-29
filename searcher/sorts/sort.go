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
	return x[i].Score < x[j].Score
}
func (x ScoreSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

type SortResult struct {
	sync.Mutex

	IsDebug bool

	IdsAndScores []model.SliceItem

	Ids []uint32

	count int // 总数

	Order string // 排序方式
}

func (f *SortResult) Add(ids *[]uint32) {
	f.Ids = append(f.Ids, *ids...)
}

func (f *SortResult) find(target *uint32) (bool, int) {

	low := 0
	high := f.count - 1
	for low <= high {
		mid := (low + high) / 2
		if f.IdsAndScores[mid].Id == *target {
			return true, mid
		} else if f.IdsAndScores[mid].Id < *target {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return false, -1
}

func (f *SortResult) Process() {
	// TODO: 计算得分
	for _, id := range f.Ids {
		if found, index := f.find(&id); found {
			f.IdsAndScores[index].Score += 1
		} else {
			f.IdsAndScores = append(f.IdsAndScores, model.SliceItem{
				Id:    id,
				Score: 1,
			})
			f.count++
		}
	}
	// 对分数进行排序
	sort.Sort(sort.Reverse(ScoreSlice(f.IdsAndScores)))
}

func (f *SortResult) Count() int {
	return f.count
}

func (f *SortResult) GetAll(i *[]model.SliceItem, start int, end int) {
	*i = f.IdsAndScores[start:end]
}
