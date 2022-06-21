package cache

import (
	"context"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

var (
	once sync.Once
)

type InvertedIndexCache struct {
	myCache *cache.Cache
}

func (iic *InvertedIndexCache) autoInit() {

	once.Do(func() {
		if iic.myCache == nil {
			ring := redis.NewRing(&redis.RingOptions{
				Addrs: map[string]string{
					"localhost": ":6379",
				},
			})
			iic.myCache = cache.New(&cache.Options{
				Redis:      ring,
				LocalCache: cache.NewTinyLFU(5000, time.Second*5),
			})
		}
	})
	//ring := redis.NewRing(&redis.RingOptions{
	//	Addrs: map[string]string{
	//		"localhost": ":6379",
	//	},
	//})
	//iic.myCache = cache.New(&cache.Options{
	//	Redis:      ring,
	//	LocalCache: cache.NewTinyLFU(5000, time.Minute),
	//})
}
func (iic *InvertedIndexCache) Set(key string, val map[uint32]float64) {
	if iic.myCache == nil {
		iic.autoInit()
	}
	_myCache := iic.myCache
	ctx := context.TODO()
	if err := _myCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
		TTL:   time.Minute,
	}); err != nil {
		panic(err)
	}
}
func (iic *InvertedIndexCache) Get(key string, wanted *map[uint32]float64) error {
	if iic.myCache == nil {
		iic.autoInit()
	}
	_myCache := iic.myCache
	ctx := context.TODO()
	err := _myCache.Get(ctx, key, wanted)
	return err
}
