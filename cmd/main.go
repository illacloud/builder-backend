package main

import (
	"build-backend/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	pingRouter := r.Group("/ping")
	{
		pingRouter.GET("", router.Ping())
	}
	realtimeRouter := r.Group("/realtime")
	{
		realtimeRouter.GET("/ping", router.WsPing())
	}
	_ = r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
