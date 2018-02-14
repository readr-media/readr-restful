package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type postHandler struct{}

type postArgs struct {
	MaxResult uint8  `form:"max_result"`
	Page      uint16 `form:"page"`
	Sorting   string `form:"sort"`
	Active    string `form:"active"`
	Author    string `form:"author"`
	Type      string `form:"type"`
}

func (r *postHandler) bindQuery(c *gin.Context, args *models.PostArgs) (err error) {
	if err = c.ShouldBindQuery(args); err == nil {
		return nil
	}
	// Start parsing rest of request arguments
	if c.Query("author") != "" && args.Author == nil {
		if err = json.Unmarshal([]byte(c.Query("author")), &args.Author); err != nil {
			return err
		}
	}
	if c.Query("active") != "" && args.Active == nil {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, models.PostStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("type") != "" && args.Type == nil {
		if err = json.Unmarshal([]byte(c.Query("type")), &args.Type); err != nil {
			return err
		}
	}
	return nil
}

func (r *postHandler) GetAll(c *gin.Context) {
	var args = &models.PostArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	result, err := models.PostAPI.GetPosts(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *postHandler) Get(c *gin.Context) {

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
	c.JSON(http.StatusOK, gin.H{"_items": []models.TaggedPostMember{post}})
}

func (r *postHandler) Post(c *gin.Context) {

	post := models.TaggedPost{}

	err := c.Bind(&post)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if post.Post == (models.Post{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	if !post.Author.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Author"})
		return
	}
	// CreatedAt and UpdatedAt set default to now
	post.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	post.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	if !post.Active.Valid {
		post.Active.Int = 3
		post.Active.Valid = true
	}
	if !post.UpdatedBy.Valid {
		if post.Author.Valid {
			post.UpdatedBy.String = post.Author.String
			post.UpdatedBy.Valid = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		}

	}
	post_id, err := models.PostAPI.InsertPost(post.Post)
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
	if len(post.Tags) > 0 {
		err = models.TagAPI.UpdatePostTags(post_id, post.Tags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *postHandler) Put(c *gin.Context) {

	post := models.TaggedPost{}

	err := c.ShouldBindJSON(&post)
	// Check if post struct was binded successfully
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if post.Post == (models.Post{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	// Discard CreatedAt even if there is data
	if post.CreatedAt.Valid {
		post.CreatedAt.Time = time.Time{}
		post.CreatedAt.Valid = false
	}
	post.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	switch {
	case post.UpdatedBy.Valid:
		fallthrough
	case len(post.Tags) > 0:
		fallthrough
	case post.Author.Valid:
		post.UpdatedBy.String = post.Author.String
		post.UpdatedBy.Valid = true
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		return
	}

	err = models.PostAPI.UpdatePost(post.Post)
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

	if len(post.Tags) > 0 {
		err = models.TagAPI.UpdatePostTags(int(post.ID), post.Tags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}
func (r *postHandler) DeleteAll(c *gin.Context) {

	params := models.PostUpdateArgs{}
	// Bind params. If err return 400
	err := c.BindQuery(&params)
	// Unable to parse interface{} to []uint32
	// If change to []interface{} first, there's no way avoid for loop each time.
	// So here use individual parsing instead of function(interface{}) (interface{})
	if c.Query("ids") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	err = json.Unmarshal([]byte(c.Query("ids")), &params.IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if len(params.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	if strings.HasPrefix(params.UpdatedBy, `"`) {
		params.UpdatedBy = strings.TrimPrefix(params.UpdatedBy, `"`)
	}
	if strings.HasSuffix(params.UpdatedBy, `"`) {
		params.UpdatedBy = strings.TrimSuffix(params.UpdatedBy, `"`)
	}
	params.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	params.Active = models.NullInt{Int: int64(models.PostStatus["deactive"].(float64)), Valid: true}
	err = models.PostAPI.UpdateAll(params)
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

	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)
	err := models.PostAPI.DeletePost(id)
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
	c.Status(http.StatusOK)
}

func (r *postHandler) PublishAll(c *gin.Context) {
	payload := models.PostUpdateArgs{}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if payload.IDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}
	payload.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	payload.Active = models.NullInt{Int: int64(models.PostStatus["active"].(float64)), Valid: true}
	err = models.PostAPI.UpdateAll(payload)
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

func (r *postHandler) Count(c *gin.Context) {
	var args = &models.PostArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	count, err := models.PostAPI.Count(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *postHandler) SetRoutes(router *gin.Engine) {

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

		postsRouter.GET("/count", r.Count)
	}
}

var PostHandler postHandler
