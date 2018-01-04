package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type postHandler struct{}

func (r *postHandler) PostsGetHandler(c *gin.Context) {

	mr := c.DefaultQuery("max_result", "20")
	u64MaxResult, _ := strconv.ParseUint(mr, 10, 8)
	maxResult := uint8(u64MaxResult)

	pg := c.DefaultQuery("page", "1")
	u64Page, _ := strconv.ParseUint(pg, 10, 16)
	page := uint16(u64Page)

	sorting := c.DefaultQuery("sort", "-updated_at")

	result, err := models.PostAPI.GetPosts(maxResult, page, sorting)
	if err != nil {
		switch err.Error() {
		case "Posts Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Posts Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

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
	if post.Active != 3 {
		post.Active = 3
	}
	// if !post.UpdatedBy.Valid {
	// 	post.UpdatedBy.String = post.Author.String
	// 	post.UpdatedBy.Valid = true
	// }
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
	if post.Active == 0 {
		post.Active = 4
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

	router.GET("posts", r.PostsGetHandler)

	postRouter := router.Group("/post")
	{
		postRouter.GET("/:id", r.PostGetHandler)
		postRouter.POST("", r.PostPostHandler)
		postRouter.PUT("", r.PostPutHandler)
		postRouter.DELETE("/:id", r.PostDeleteHandler)
	}
}

var PostHandler postHandler
