package system

import (
	"net/http"
	"sun-panel/api/api_v1"

	"github.com/gin-gonic/gin"
)

func PingCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		c.Header("Access-Control-Allow-Private-Network", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func InitPingRouter(router *gin.RouterGroup) {
	ping := api_v1.ApiGroupApp.ApiSystem.Ping
	router.GET("ping", PingCors(), ping.Get)
	router.POST("ping", PingCors(), ping.Get)
	router.OPTIONS("ping", PingCors(), func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNoContent)
	})
}
