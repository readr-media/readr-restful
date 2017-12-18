package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type projectHandler struct{}

func (r *projectHandler) projectGet(c *gin.Context){
	fmt.Println(c)
	//input := models.Project{ID: c.Param("id")}
	result, ok := models.ProjectAPI.GetProjects(models.Project{})
	fmt.Println(result, ok)
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
	projectRouter := router.Group("/project")
	{
		projectRouter.GET("/:id", r.projectGet)
		projectRouter.POST("", r.projectPost)
		projectRouter.PUT("", r.projectPut)
		projectRouter.DELETE("/:id", r.projectDelete)
	}
}

var ProjectHandler projectHandler