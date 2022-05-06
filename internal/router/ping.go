package router

import (
	"github.com/gin-gonic/gin"
)

func Ping() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}
