package routes

import (
	"strconv"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type projectHandler struct {
}

func (r *projectHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID Must Be Integer"})
		return
	}
	input := models.Project{ID: id}
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

func (r *projectHandler) Post(c *gin.Context) {

	project := models.Project{}
	c.Bind(&project)

	// Pre-request test
	if project.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project"})
		return
	}
	if !project.CreatedAt.Valid {
		project.CreatedAt.Time = time.Now()
		project.CreatedAt.Valid = true
	}
	if !project.UpdatedAt.Valid {
		project.UpdatedAt.Time = time.Now()
		project.UpdatedAt.Valid = true
	}

	err := models.ProjectAPI.InsertProject(project)
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

func (r *projectHandler) Put(c *gin.Context) {

	project := models.Project{}
	c.Bind(&project)
	if project.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project Data"})
		return
	}
	if project.CreatedAt.Valid {
		project.CreatedAt.Time = time.Time{}
		project.CreatedAt.Valid = false
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

func (r *projectHandler) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID Must Be Integer"})
		return
	}
	input := models.Project{ID: id}
	err = models.ProjectAPI.DeleteProjects(input)

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
		projectRouter.GET("/:id", r.Get)
		projectRouter.POST("", r.Post)
		projectRouter.PUT("", r.Put)
		projectRouter.DELETE("/:id", r.Delete)
	}
}

var ProjectHandler projectHandler
