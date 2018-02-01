package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type postHandler struct{}

func (r *postHandler) GetAll(c *gin.Context) {

	var params models.PostArgs = make(map[string]interface{})
	// Default query parameters
	// args := models.PostArgs{
	// 	BasicArgs: models.BasicArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"},
	// }
	args := struct {
		MaxResult uint8  `form:"max_result"`
		Page      uint16 `form:"page"`
		Sorting   string `form:"sort"`
		Active    string `form:"active"`
		Author    string `form:"author"`
	}{MaxResult: 20, Page: 1, Sorting: "-updated_at"}

	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Bind query error: %s", err.Error())})
		return
	}

	if args.Author != "" {
		authormap := map[string][]string{}
		if err := json.Unmarshal([]byte(args.Author), &authormap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid author list: %s", err.Error())})
			return
		}
		params["posts.author"] = authormap
	}
	if args.Active != "" {
		activemap := map[string][]int{}
		if err := json.Unmarshal([]byte(args.Active), &activemap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid active list: %s", err.Error())})
			return
		} else if err == nil {
			if err = models.ValidateActive(activemap, models.PostStatus); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid active list: %s", err.Error())})
				return
			}
			params["posts.active"] = activemap
		}
	} else {
		params["posts.active"] = map[string][]int{"$nin": []int{int(models.PostStatus["deactive"].(float64))}}
	}

	params["max_result"] = args.MaxResult
	params["page"] = args.Page
	params["sort"] = args.Sorting

	// fmt.Println(params)
	result, err := models.PostAPI.GetPosts(params)
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

	err := c.Bind(&post)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if post == (models.Post{}) {
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
	c.Status(http.StatusOK)
}

func (r *postHandler) Put(c *gin.Context) {

	post := models.Post{}

	err := c.ShouldBindJSON(&post)
	// Check if post struct was binded successfully
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if post == (models.Post{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	// Discard CreatedAt even if there is data
	if post.CreatedAt.Valid {
		post.CreatedAt.Time = time.Time{}
		post.CreatedAt.Valid = false
	}
	post.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	if !post.UpdatedBy.Valid {
		if post.Author.Valid {
			post.UpdatedBy.String = post.Author.String
			post.UpdatedBy.Valid = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		}

	}
	err = models.PostAPI.UpdatePost(post)
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

	var params models.PostArgs = make(map[string]interface{})

	args := struct {
		Active string `form:"active"`
		Author string `form:"author"`
	}{}
	if err := c.ShouldBindQuery(&args); err != nil {
		log.Printf(fmt.Sprintf("Binding query error: %s", err.Error()))
	}

	if args.Author != "" {
		authormap := map[string][]string{}
		if err := json.Unmarshal([]byte(args.Author), &authormap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid author list: %s", err.Error())})
			return
		}
		params["author"] = authormap
	}

	if args.Active != "" {
		activemap := map[string][]int{}
		if err := json.Unmarshal([]byte(args.Active), &activemap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid active list: %s", err.Error())})
			return
		} else if err == nil {
			if err = models.ValidateActive(activemap, models.PostStatus); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid active list: %s", err.Error())})
				return
			}
			params["active"] = activemap
		}
	} else {
		params["active"] = map[string][]int{"$nin": []int{int(models.PostStatus["deactive"].(float64))}}
	}

	count, err := models.PostAPI.Count(params)
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
