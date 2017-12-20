package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type projectHandler struct {
}

func (r *projectHandler) projectGet(c *gin.Context) {
	input := models.Project{ID: c.Param("id")}
	project, err := models.ProjectAPI.GetProject(input)

	if err != nil {
		switch err.Error() {
		case "Project Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Project Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, project)
}

func (r *projectHandler) projectPost(c *gin.Context) {

	project := models.Project{}
	c.Bind(&project)

	// Pre-request test
	if project.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project"})
		return
	}
	if !project.CreateTime.Valid {
		project.CreateTime.Time = time.Now()
		project.CreateTime.Valid = true
	}
	if !project.UpdatedAt.Valid {
		project.UpdatedAt.Time = time.Now()
		project.UpdatedAt.Valid = true
	}

	err := models.ProjectAPI.PostProject(project)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Project Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, nil)
}

func (r *projectHandler) projectPut(c *gin.Context) {

	project := models.Project{}
	c.Bind(&project)
	if project.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project Data"})
		return
	}
	if project.CreateTime.Valid {
		project.CreateTime.Time = time.Time{}
		project.CreateTime.Valid = false
	}
	if !project.UpdatedAt.Valid {
		project.UpdatedAt.Time = time.Now()
		project.UpdatedAt.Valid = true
	}
	err := models.ProjectAPI.UpdateProjects(project)
	if err != nil {
		switch err.Error() {
		case "Project Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Project Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.String(http.StatusOK, "ok")
}

func (r *projectHandler) projectDelete(c *gin.Context) {

	input := models.Project{ID: c.Param("id")}
	err := models.ProjectAPI.DeleteProjects(input)

	if err != nil {
		switch err.Error() {
		case "Project Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Project Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.String(http.StatusOK, "ok")
}

func (r *projectHandler) SetRoutes(router *gin.Engine) {
	projectRouter := router.Group("/project")
	{
		projectRouter.GET("/:id", r.projectGet)
		projectRouter.POST("", r.projectPost)
		projectRouter.PUT("", r.projectPut)
		projectRouter.DELETE("/:id", r.projectDelete)
	}
}

var ProjectHandler projectHandler
