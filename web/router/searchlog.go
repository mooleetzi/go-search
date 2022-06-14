package router

import (
	"go-search/web/controller"

	"github.com/gin-gonic/gin"
)

// searchlogRouter 搜索log更新后继词表路由
func InitSearchLogRouter(Router *gin.RouterGroup) {

	searchlogRouter := Router.Group("searchlog")
	{
		searchlogRouter.GET("update", controller.UpdatedRelatedSearch)
	}
}
