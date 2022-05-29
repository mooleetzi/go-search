package router

import (
	"go-search/global"
	"go-search/web/middleware"
	"log"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// SetupRouter 路由管理
func SetupRouter() *gin.Engine {
	if global.CONFIG.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	// 启用GZIP压缩
	if global.CONFIG.EnableGzip {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	var handlers []gin.HandlerFunc

	// 分组管理 中间件管理
	router.Use(middleware.Cors(), middleware.Exception())
	group := router.Group("/api", handlers...)
	{
		InitBaseRouter(group) // 基础管理
		InitWordRouter(group) // 分词管理
	}
	log.Printf("API Url: \t http://%v/api", global.CONFIG.Addr)
	return router
}
