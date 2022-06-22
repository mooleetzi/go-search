package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

var RedisClient *redis.Client

func Redis() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	ctx := context.TODO()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	RedisClient = client
}
func Set(word string, invertedIndexEntry map[uint32]float64) {
	ctx := context.TODO()

	for docId, score := range invertedIndexEntry {
		_, err := RedisClient.HSet(ctx, word, docId, score).Result()
		RedisClient.Expire(ctx, word, time.Minute)
		if err != nil {
			log.Println("Hset err", err)
		}
	}
}
func Get(word string, wanted *map[uint32]float64) (find bool, err error) {
	ctx := context.TODO()
	inverted, err := RedisClient.HGetAll(ctx, word).Result()
	if err != nil {
		log.Println("hgetall err", err)
		return
	}
	if len(inverted) != 0 {
		//	notfound
		find = true
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
	return
}
