package main

import (
	"go-search/core"
	"go-search/log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func main() {

	go func() {
		runtime.SetBlockProfileRate(1) // 开启对阻塞操作的跟踪，block
		runtime.SetMutexProfileFraction(1)
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
	}()
	// 初始化容器和参数解析
	core.Initialize()
}
