package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
