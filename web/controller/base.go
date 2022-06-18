package controller

import (
	"github.com/gin-gonic/gin/binding"
	"go-search/searcher/model"

	"github.com/gin-gonic/gin"
)

func Welcome(c *gin.Context) {
	ResponseSuccessWithData(c, "Welcome to go-search")
}

// Query 查询
func Query(c *gin.Context) {
	var request = &model.SearchRequest{
		Database: c.Query("database"),
		ClientIP: c.ClientIP(),
	}

	if err := c.ShouldBindBodyWith(request, binding.JSON); err != nil {
		ResponseErrorWithMsg(c, err.Error())
		return
	}
	// 调用搜索
	r := srv.Base.Query(request)
	ResponseSuccessWithData(c, r)
}

// GC 释放GC
func GC(c *gin.Context) {
	srv.Base.GC()
	ResponseSuccess(c)
}
