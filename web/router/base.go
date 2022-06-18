package router

import (
	"github.com/gin-gonic/gin"
	"go-search/web/controller"
	"go-search/web/middleware"
)

// InitBaseRouter 基础管理路由
func InitBaseRouter(Router *gin.RouterGroup) {
	Router.Use(middleware.PostBySync())
	BaseRouter := Router.Group("")
	{
		BaseRouter.GET("/", controller.Welcome)
		BaseRouter.POST("query", controller.Query)
		BaseRouter.GET("gc", controller.GC)
	}
}
