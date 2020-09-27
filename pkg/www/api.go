package www

import (
	"github.com/gin-gonic/gin"
)

func v1API(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		v1.POST("google", GsuiteWebhook)
	}
}
