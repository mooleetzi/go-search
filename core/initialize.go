package core

import (
	"context"
	"fmt"
	"go-search/global"
	"go-search/searcher"
	"go-search/searcher/words"
	"go-search/web/controller"
	"go-search/web/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jasonlvhit/gocron"
)

func NewContainer(tokenizer *words.Tokenizer) *searcher.Container {
	container := &searcher.Container{
		Dir:       global.CONFIG.Data,
		Tokenizer: tokenizer,
		Shard:     global.CONFIG.Shard,
		Timeout:   global.CONFIG.Timeout,
	}
	go container.Init()

	return container
}

func NewTokenizer() *words.Tokenizer {
	return words.NewTokenizer(global.CONFIG.Dictionary)
}

// Initialize 初始化
func Initialize() {

	global.CONFIG = Parser()

	defer func() {

		if r := recover(); r != nil {
			fmt.Printf("panic: %s\n", r)
		}
	}()

	// 初始化分词器
	tokenizer := NewTokenizer()
	global.Container = NewContainer(tokenizer)

	// 初始化业务逻辑
	controller.NewServices()

	// // 初始化定时器
	// clock()

	// 注册路由
	r := router.SetupRouter()
	// 启动服务
	srv := &http.Server{
		Addr:    global.CONFIG.Addr,
		Handler: r,
	}
	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("listen:", err)
		}
	}()

	// 优雅关机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	global.Container.MustWriteLog()
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}

	log.Println("Server exiting")
}

//定时器任务，执行读取log更新后继词表
func clock() {
	gocron.Every(10).Second().DoSafely(taskWithParams, 1, "hello")
	// gocron.Every(1).Day().At("10:30").Do(task)
	<-gocron.Start()
}

func taskWithParams(a int, b string) {

	fmt.Println(a, b)
}
