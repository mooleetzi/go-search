package router

import (
	"github.com/gin-gonic/gin"
	"go-search/web/controller"
	"go-search/web/middleware"
)

var (
	brokers = []string{"127.0.0.1:9092"}
	topic   = "Test"
)

// InitBaseRouter 基础管理路由
func InitBaseRouter(Router *gin.RouterGroup) {
	var singletonProducer = &middleware.SingletonProducer{
		BrokerList: brokers,
		Topic:      topic,
	}
	Router.Use(middleware.PostBySync(singletonProducer))
	BaseRouter := Router.Group("")
	{
		BaseRouter.GET("/", controller.Welcome)
		BaseRouter.POST("query", controller.Query)
		BaseRouter.GET("gc", controller.GC)
	}
}
