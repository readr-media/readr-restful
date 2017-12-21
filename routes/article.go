package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type articleHandler struct{}

func (r *articleHandler) ArticleGetHandler(c *gin.Context) {

	input := models.Article{ID: c.Param("id")}
	article, err := models.DS.Get(input)

	if err != nil {
		switch err.Error() {
		case "Article Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Article Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, article)
}

func (r *articleHandler) ArticlePostHandler(c *gin.Context) {

	article := models.Article{}
	err := c.Bind(&article)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if article.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Article ID"})
		return
	}
	if !article.CreateTime.Valid {
		article.CreateTime.Time = time.Now()
		article.CreateTime.Valid = true
	}
	if !article.UpdatedAt.Valid {
		article.UpdatedAt.Time = time.Now()
		article.UpdatedAt.Valid = true
	}
	if article.Active != 1 {
		article.Active = 1
	}
	result, err := models.DS.Create(article)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Article ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (r *articleHandler) ArticlePutHandler(c *gin.Context) {

	article := models.Article{}
	c.Bind(&article)
	// Check if article struct was binded successfully
	if article.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Article Data"})
		return
	}
	if article.CreateTime.Valid {
		article.CreateTime.Time = time.Time{}
		article.CreateTime.Valid = false
	}
	if !article.UpdatedAt.Valid {
		article.UpdatedAt.Time = time.Now()
		article.UpdatedAt.Valid = true
	}
	result, err := models.DS.Update(article)
	if err != nil {
		switch err.Error() {
		case "Article Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Article Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (r *articleHandler) ArticleDeleteHandler(c *gin.Context) {

	input := models.Article{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	article, err := models.DS.Delete(input)

	// member, err := req.Delete()
	if err != nil {
		switch err.Error() {
		case "Article Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Article Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, article)
}

func (r *articleHandler) SetRoutes(router *gin.Engine) {
	articleRouter := router.Group("/article")
	{
		articleRouter.GET("/:id", r.ArticleGetHandler)
		articleRouter.POST("", r.ArticlePostHandler)
		articleRouter.PUT("", r.ArticlePutHandler)
		articleRouter.DELETE("/:id", r.ArticleDeleteHandler)
	}
}

var ArticleHandler articleHandler
