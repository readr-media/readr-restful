package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type memoHandler struct{}

func (r *memoHandler) bindQuery(c *gin.Context, args *models.MemoGetArgs) (err error) {
	_ = c.ShouldBindQuery(args)

	if c.Query("active") != "" && args.Active == nil {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			log.Println("active")
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, models.MemoStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("author") != "" {
		if err = json.Unmarshal([]byte(c.Query("author")), &args.Author); err != nil {
			return err
		}
	}
	if c.Query("project_id") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_id")), &args.Project); err != nil {
			return err
		}
	}

	return nil
}

func (r *memoHandler) GetMany(c *gin.Context) {
	var args = &models.MemoGetArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if !args.Validate() {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	result, err := models.MemoAPI.GetMemos(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *memoHandler) Get(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))
	result, err := models.MemoAPI.GetMemo(id)

	if err != nil {
		switch err.Error() {
		case "Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *memoHandler) Post(c *gin.Context) {

	memo := models.Memo{}

	err := c.Bind(&memo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if !memo.Author.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Author"})
		return
	}
	if !memo.Project.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project"})
		return
	}

	memo.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	memo.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	if !memo.Active.Valid {
		memo.Active.Int = int64(models.MemoStatus["pending"].(float64))
		memo.Active.Valid = true
	}
	if !memo.UpdatedBy.Valid {
		if memo.Author.Valid {
			memo.UpdatedBy = models.NullString{String: memo.Author.String, Valid: true}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Updater"})
		}
	}
	err = models.MemoAPI.InsertMemo(memo)
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

func (r *memoHandler) Put(c *gin.Context) {

	memo := models.Memo{}

	err := c.ShouldBindJSON(&memo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if memo.CreatedAt.Valid {
		memo.CreatedAt.Valid = false
	}
	memo.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	switch {
	case memo.UpdatedBy.Valid:
	case memo.Author.Valid:
		memo.UpdatedBy.String = memo.Author.String
		memo.UpdatedBy.Valid = true
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		return
	}

	err = models.MemoAPI.UpdateMemo(memo)
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
func (r *memoHandler) DeleteMany(c *gin.Context) {

	params := models.MemoUpdateArgs{}
	// Bind params. If err return 400
	err := c.ShouldBindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if len(params.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}

	params.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	params.Active = models.NullInt{Int: int64(models.PostStatus["deactive"].(float64)), Valid: true}

	err = models.MemoAPI.UpdateMemos(params)
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

func (r *memoHandler) PublishMany(c *gin.Context) {
	params := models.MemoUpdateArgs{}

	err := c.ShouldBindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if params.IDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Empty Memo ID"})
		return
	}
	if params.UpdatedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Updater Not Specified"})
		return
	}
	params.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	params.PublishedAt = models.NullTime{Time: time.Now(), Valid: true}
	params.Active = models.NullInt{Int: int64(models.MemoStatus["active"].(float64)), Valid: true}
	err = models.MemoAPI.UpdateMemos(params)
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

func (r *memoHandler) Count(c *gin.Context) {
	var args = &models.MemoGetArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	count, err := models.MemoAPI.CountMemos(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *memoHandler) SetRoutes(router *gin.Engine) {

	memoRouter := router.Group("/memo")
	{
		memoRouter.GET("/:id", r.Get)
		memoRouter.POST("", r.Post)
		memoRouter.PUT("", r.Put)
	}
	memosRouter := router.Group("/memos")
	{
		memosRouter.GET("", r.GetMany)
		memosRouter.GET("/count", r.Count)

		memosRouter.DELETE("", r.DeleteMany)
		memosRouter.PUT("", r.PublishMany)
	}
}

var MemoHandler memoHandler
