package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type postHandler struct{}

func (r *postHandler) GetAll(c *gin.Context) {

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
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *postHandler) Get(c *gin.Context) {

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
	c.JSON(http.StatusOK, gin.H{"_items": []models.PostMember{post}})
}

func (r *postHandler) Post(c *gin.Context) {

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
	if !post.Author.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Author"})
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
	// When insert post set default active to 3
	if !post.Active.Valid {
		post.Active.Int = 3
	}
	if !post.UpdatedBy.Valid {
		post.UpdatedBy = post.Author
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
	// c.JSON(http.StatusOK, models.Post{})
	c.Status(http.StatusOK)
}

func (r *postHandler) Put(c *gin.Context) {

	post := models.Post{}
	emptyPost := models.Post{}

	c.Bind(&post)
	// Check if post struct was binded successfully
	if post == emptyPost {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	// CreatedAt would be list in the updating tag list
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
	// c.JSON(http.StatusOK, models.Post{})
	c.Status(http.StatusOK)
}
func (r *postHandler) DeleteAll(c *gin.Context) {

	ids := []uint32{}
	// Disable to parse interface{} to []uint32
	// If change to []interface{} first, there's no way avoid for loop each time.
	// So here use individual parsing instead of function(interface{}) (interface{})
	err := json.Unmarshal([]byte(c.Query("ids")), &ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	err = models.PostAPI.SetMultipleActive(ids, int(models.PostStatus["deactive"].(float64)))
	if err != nil {
		switch err.Error() {
		case "Posts Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Posts Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *postHandler) Delete(c *gin.Context) {

	// input := models.Post{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	// post, err := models.DS.Delete(input)
	// id := c.Param("id")
	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)
	err := models.PostAPI.DeletePost(id)
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
	// c.JSON(http.StatusOK, gin.H{"result": fmt.Sprintf("%d row affected", post)})
	c.Status(http.StatusOK)
}

func (r *postHandler) PublishAll(c *gin.Context) {
	payload := struct {
		IDs []uint32 `json:"ids"`
	}{}
	err := c.Bind(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if payload.IDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}
	err = models.PostAPI.SetMultipleActive(payload.IDs, int(models.PostStatus["active"].(float64)))
	if err != nil {
		switch err.Error() {
		case "Posts Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Posts Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *postHandler) SetRoutes(router *gin.Engine) {

	// router.GET("posts", r.PostsGetHandler)

	postRouter := router.Group("/post")
	{
		postRouter.GET("/:id", r.Get)
		postRouter.POST("", r.Post)
		postRouter.PUT("", r.Put)
		postRouter.DELETE("/:id", r.Delete)
	}
	postsRouter := router.Group("/posts")
	{
		postsRouter.GET("", r.GetAll)
		postsRouter.DELETE("", r.DeleteAll)
		postsRouter.PUT("", r.PublishAll)
	}
}

var PostHandler postHandler
