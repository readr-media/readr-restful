package routes

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type tagHandler struct{}

func (r *tagHandler) Get(c *gin.Context) {
	args := models.DefaultGetTagsArgs()
	err := c.Bind(&args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if !validateSorting(args.Sorting) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Sorting Option"})
		return
	}

	result, err := models.TagAPI.GetTags(args)
	if err != nil {
		switch err.Error() {
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *tagHandler) Post(c *gin.Context) {
	args := models.Tag{}
	err := c.Bind(&args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if args.UpdatedBy.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Updater Not Sepcified"})
		return
	}
	args.UpdatedAt = models.NullTime{Valid: false}
	args.CreatedAt = models.NullTime{Valid: false}

	tag_id, err := models.TagAPI.InsertTag(args)
	if err != nil {
		switch err.Error() {
		case "Duplicate Entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"tag_id": tag_id})
}

func (r *tagHandler) Put(c *gin.Context) {
	args := models.Tag{}
	err := c.Bind(&args)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if args.UpdatedBy.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Updater Not Sepcified"})
		return
	}
	args.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	args.CreatedAt = models.NullTime{Valid: false}

	err = models.TagAPI.UpdateTag(args)
	if err != nil {
		switch err.Error() {
		case "Duplicate Entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		case models.ItemNotFoundError.Error():
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *tagHandler) Delete(c *gin.Context) {

	var IDs []int
	err := json.Unmarshal([]byte(c.Query("ids")), &IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Tag IDs"})
		return
	}

	if len(IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "No Tags To Be Operated"})
		return
	}

	updater := c.Query("updated_by")
	if updater == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Updater"})
		return
	}
	// args := models.UpdateMultipleTagsArgs{IDs: IDs, UpdatedBy: updater, Active: strconv.FormatFloat(models.TagStatus["deactive"].(float64), 'f', 6, 64)}
	args := models.UpdateMultipleTagsArgs{IDs: IDs, UpdatedBy: updater, Active: strconv.FormatFloat(float64(config.Config.Models.Tags["deactive"]), 'f', 6, 64)}

	err = models.TagAPI.ToggleTags(args)
	if err != nil {
		switch err.Error() {
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *tagHandler) Count(c *gin.Context) {
	count, err := models.TagAPI.CountTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *tagHandler) SetRoutes(router *gin.Engine) {

	tagRouter := router.Group("/tags")
	{
		tagRouter.GET("", r.Get)
		tagRouter.POST("", r.Post)
		tagRouter.PUT("", r.Put)
		tagRouter.DELETE("", r.Delete)

		tagRouter.GET("/count", r.Count)
	}
}

func validateSorting(s string) bool {
	if matched, err := regexp.MatchString("-?(text|updated_at|created_at|related_reviews|related_news)", s); err != nil || !matched {
		return false
	}
	return true
}

var TagHandler tagHandler
