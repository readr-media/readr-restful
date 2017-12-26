package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type postHandler struct{}

func (r *postHandler) PostGetHandler(c *gin.Context) {

	// input := models.Post{ID: c.Param("id")}
	// post, err := models.DS.Get(input)
	// id := c.Param("id")
	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)
	post, err := models.PostAPI.GetPost(id)

	if err != nil {
		switch err.Error() {
		case "Post Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Post Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, post)
}

func (r *postHandler) PostPostHandler(c *gin.Context) {

	post := models.Post{}
	emptyPost := models.Post{}

	err := c.Bind(&post)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if post == emptyPost {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	if !post.CreatedAt.Valid {
		post.CreatedAt.Time = time.Now()
		post.CreatedAt.Valid = true
	}
	if !post.UpdatedAt.Valid {
		post.UpdatedAt.Time = time.Now()
		post.UpdatedAt.Valid = true
	}
	if post.Active != 1 {
		post.Active = 1
	}
	if !post.UpdatedBy.Valid {
		post.UpdatedBy.String = post.Author.String
		post.UpdatedBy.Valid = true
	}
	// result, err := models.DS.Create(post)
	err = models.PostAPI.InsertPost(post)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Post ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, models.Post{})
}

func (r *postHandler) PostPutHandler(c *gin.Context) {

	post := models.Post{}
	emptyPost := models.Post{}

	c.Bind(&post)
	// Check if post struct was binded successfully
	if post == emptyPost {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	if post.CreatedAt.Valid {
		post.CreatedAt.Time = time.Time{}
		post.CreatedAt.Valid = false
	}
	if !post.UpdatedAt.Valid {
		post.UpdatedAt.Time = time.Now()
		post.UpdatedAt.Valid = true
	}
	// result, err := models.DS.Update(post)
	err := models.PostAPI.UpdatePost(post)
	if err != nil {
		switch err.Error() {
		case "Post Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Post Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, models.Post{})
}

func (r *postHandler) PostDeleteHandler(c *gin.Context) {

	// input := models.Post{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	// post, err := models.DS.Delete(input)
	// id := c.Param("id")
	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)
	post, err := models.PostAPI.DeletePost(id)
	// member, err := req.Delete()
	if err != nil {
		switch err.Error() {
		case "Post Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Post Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, post)
}

func (r *postHandler) SetRoutes(router *gin.Engine) {
	postRouter := router.Group("/post")
	{
		postRouter.GET("/:id", r.PostGetHandler)
		postRouter.POST("", r.PostPostHandler)
		postRouter.PUT("", r.PostPutHandler)
		postRouter.DELETE("/:id", r.PostDeleteHandler)
	}
}

var PostHandler postHandler
