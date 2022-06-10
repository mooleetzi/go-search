package controller

import (
	"github.com/gin-gonic/gin"
	"go-search/searcher/model"
)

func Welcome(c *gin.Context) {
	ResponseSuccessWithData(c, "Welcome to go-search")
}

// Query 查询
func Query(c *gin.Context) {
	var request = &model.SearchRequest{
		Database: c.Query("database"),
	}
	if err := c.ShouldBind(request); err != nil {
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
