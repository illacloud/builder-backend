package router

import (
	"build-backend/api"
	"github.com/gin-gonic/gin"
)

func WsPing() func(c *gin.Context) {
	return func(c *gin.Context) {
		ws, err := api.UpGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				break
			}
			if string(message) == "ping" {
				message = []byte("pong")
			}
			err = ws.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}
}
