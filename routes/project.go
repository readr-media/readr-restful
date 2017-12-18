package routes

import (
	"github.com/gin-gonic/gin"
	//"github.com/readr-media/readr-restful/models"
)

type projectHandler struct{}

func (r *projectHandler) projectGet(c *gin.Context){
	//implements...
}

func (r *projectHandler) projectPost(c *gin.Context){
	//implements...
}

func (r *projectHandler) projectPut(c *gin.Context){
	//implements...
}

func (r *projectHandler) projectDelete(c *gin.Context){
	//implements...
}

func (r *projectHandler) SetRoutes(router *gin.Engine){
	router.Group("/project")
	{
		router.GET("/:id", r.projectGet)
		router.POST("", r.projectPost)
		router.PUT("", r.projectPut)
		router.DELETE("/:id", r.projectDelete)
	}
}

var ProjectHandler projectHandler