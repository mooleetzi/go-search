package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go-search/searcher/sorts"
	"go-search/searcher/utils"
	"log"
	"strconv"
	"sync/atomic"
)

var (
	RedisClient *redis.Client
	Hit         uint32 = 0
	Total       uint32 = 0
	IsDebug     bool
)

func Redis(_isDebug bool) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	ctx := context.TODO()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	RedisClient = client
	IsDebug = _isDebug
}

func minInt(args ...int) int {
	min := args[0]
	for _, arg := range args {
		if arg < min {
			min = arg
		}
	}
	return min
}
func Set(word string, scores *sorts.ScoreSlice, k int) {
	//var scores *sorts.ScoreSlice
	//var scores *[]model.SliceItem

	toAdded := (*scores)[0:minInt(k, len(*scores))]
	ctx := context.TODO()
	time := utils.ExecTime(func() {
		for _, item := range toAdded {
			_, err := RedisClient.HSet(ctx, word, item.Id, item.Score).Result()
			if err != nil {
				log.Println("HSet err", err)
			}
		}
	})
	if IsDebug {
		log.Println(word, "set, time is", time, "ms")
	}
}
func Get(word string, wanted *map[uint32]float64) (find bool, err error) {
	ctx := context.TODO()
	time := utils.ExecTime(func() {
		inverted, err := RedisClient.HGetAll(ctx, word).Result()
		if err != nil {
			log.Println("hgetall err", err)
			return
		}
		if len(inverted) != 0 {
			//	found
			find = true
			if IsDebug {
				atomic.AddUint32(&Hit, 1)
			}
			for idString, scoreString := range inverted {
				id, err := strconv.ParseUint(idString, 10, 64)
				if err != nil {
					log.Println("err in string parse", err)
				}
				score, err := strconv.ParseFloat(scoreString, 64)
				if err != nil {
					log.Println("err in string parse", err)
				}
				(*wanted)[uint32(id)] = score
			}
		}
		if IsDebug {
			log.Println(word, "get, len is", len(inverted))
			atomic.AddUint32(&Total, 1)
			log.Println("set hit record", atomic.LoadUint32(&Hit), Total)
		}
	})
	if IsDebug {
		log.Println(word, "get time is", time, "ms")
	}
	return
}
func GetRate() float64 {
	return float64(atomic.LoadUint32(&Hit)) / float64(atomic.LoadUint32(&Total))
}
