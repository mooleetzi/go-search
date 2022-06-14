package controller

import "github.com/gin-gonic/gin"

func UpdatedRelatedSearch(c *gin.Context) {
	ic := c.Query("ic")
	r := srv.ScLog.UpdatedRelatedSearch(ic)
	ResponseSuccessWithData(c, r)

}
