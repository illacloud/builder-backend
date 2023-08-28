package recovery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type any = interface{}

func CorsHandleRecovery(c *gin.Context, err any) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, "+
		"Access-Control-Allow-Headers, Authorization, Cache-Control, Content-Language, Content-Type, illa-token")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
	c.Header("Content-Type", "application/json")
	c.AbortWithStatus(http.StatusInternalServerError)
}
