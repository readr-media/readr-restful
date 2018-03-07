package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type miscHandler struct{}

func (r *miscHandler) SetRoutes(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})
}

var MiscHandler miscHandler
