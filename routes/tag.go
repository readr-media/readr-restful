package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type tagHandler struct{}

func (r *tagHandler) bindGetQuery(c *gin.Context, args *models.GetTagsArgs) (err error) {
	err = c.Bind(args)
	if err != nil {
		return err
	}

	if c.Query("post_fields") != "" {
		if err = json.Unmarshal([]byte(c.Query("post_fields")), &args.PostFields); err != nil {
			log.Println(err.Error())
			return err
		}
		for key, field := range args.PostFields {
			if field == "id" {
				field = "post_id"
				args.PostFields[key] = "post_id"
			}
			if !r.validate(field, fmt.Sprintf("^(%s)$", strings.Join(args.FullPostTags(), "|"))) {
				return errors.New("Invalid Fields")
			}
		}
	} else {
		args.PostFields = args.FullPostTags()
	}

	if c.Query("project_fields") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_fields")), &args.ProjectFields); err != nil {
			log.Println(err.Error())
			return err
		}
		for key, field := range args.ProjectFields {
			if field == "id" {
				field = "project_id"
				args.PostFields[key] = "project_id"
			}
			if !r.validate(field, fmt.Sprintf("^(%s)$", strings.Join(args.FullProjectTags(), "|"))) {
				return errors.New("Invalid Fields")
			}
		}
	} else {
		args.ProjectFields = args.FullProjectTags()
	}
	return nil
}

func (r *tagHandler) Get(c *gin.Context) {
	args := models.DefaultGetTagsArgs()
	if err := r.bindGetQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if err := args.ValidateGet(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if matched, err := regexp.MatchString("-?text", args.Sorting); err == nil && matched {
		args.Sorting = strings.Replace(args.Sorting, "text", "tag_content", 1)
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
	args := models.DefaultGetTagsArgs()
	err := c.Bind(&args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	count, err := models.TagAPI.CountTags(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *tagHandler) Hot(c *gin.Context) {
	tags, err := models.TagAPI.GetHotTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": tags})
}

func (r *tagHandler) PutHot(c *gin.Context) {
	err := models.TagAPI.UpdateHotTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func bindGetPostReportArgs(c *gin.Context, args *models.GetPostReportArgs) (err error) {
	if err := c.ShouldBindQuery(args); err != nil {
		return err
	}
	if c.Param("tag_id") != "" {
		args.TagID, _ = strconv.Atoi(c.Param("tag_id"))
	}
	return err
}

func (r *tagHandler) GetPostReport(c *gin.Context) {

	args := models.NewGetPostReportArgs()
	if err := bindGetPostReportArgs(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result, err := models.TagAPI.GetPostReport(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *tagHandler) SetRoutes(router *gin.Engine) {

	tagRouter := router.Group("/tags")
	{
		tagRouter.GET("", r.Get)
		tagRouter.POST("", r.Post)
		tagRouter.PUT("", r.Put)
		tagRouter.DELETE("", r.Delete)

		tagRouter.GET("/count", r.Count)
		tagRouter.GET("/hot", r.Hot)
		tagRouter.PUT("/hot", r.PutHot)

		tagRouter.GET("/pnr/:tag_id", r.GetPostReport)
	}
}

func (r *tagHandler) validate(target string, paradigm string) bool {
	if matched, err := regexp.MatchString(paradigm, target); err != nil || !matched {
		return false
	}
	return true
}

var TagHandler tagHandler
