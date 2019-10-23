package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/mail"
)

type memoHandler struct{}

func (r *memoHandler) bindQuery(c *gin.Context, args *models.MemoGetArgs) (err error) {
	_ = c.ShouldBindQuery(args)

	if c.Query("active") != "" && args.Active == nil {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			return err
		} else if err == nil {
			// if err = rrsql.ValidateActive(args.Active, models.MemoStatus); err != nil {
			if err = rrsql.ValidateActive(args.Active, config.Config.Models.Memos); err != nil {
				return err
			}
		}
	}
	if c.Query("memo_publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("memo_publish_status")), &args.MemoPublishStatus); err != nil {
			return err
		} else if err == nil {
			// if err = rrsql.ValidateActive(args.MemoPublishStatus, models.MemoPublishStatus); err != nil {
			if err = rrsql.ValidateActive(args.MemoPublishStatus, config.Config.Models.MemosPublishStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("project_publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_publish_status")), &args.ProjectPublishStatus); err != nil {
			return err
		} else if err == nil {
			// if err = rrsql.ValidateActive(args.ProjectPublishStatus, models.ProjectPublishStatus); err != nil {
			if err = rrsql.ValidateActive(args.ProjectPublishStatus, config.Config.Models.ProjectsPublishStatus); err != nil {
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
	if c.Query("slugs") != "" {
		if err = json.Unmarshal([]byte(c.Query("slugs")), &args.Slugs); err != nil {
			return err
		}
	}

	if c.Query("sort") != "" && r.validateMemoSorting(c.Query("sort")) {
		sortingString := c.Query("sort")
		for _, v := range strings.Split(sortingString, ",") {
			if v == "memo_id" {
				sortingString = strings.Replace(sortingString, "memo_id", "post_id", -1)
			} else if v == "id" {
				sortingString = strings.Replace(sortingString, "id", "post_id", -1)
			}
		}
		args.Sorting = sortingString
	}

	if c.Query("keyword") != "" {
		args.Keyword = c.Query("keyword")
	}

	if c.Query("abstract_length") != "" {
		if args.AbstractLength, err = strconv.ParseInt(c.Query("abstract_length"), 10, 64); err != nil {
			return err
		}
	}
	if c.Query("member_id") != "" {
		if args.MemberID, err = strconv.ParseInt(c.Query("member_id"), 10, 64); err != nil {
			return err
		}
	}
	if c.Param("id") != "" {
		id, _ := strconv.Atoi(c.Param("id"))
		args.IDs = []int64{int64(id)}
	}
	if c.Param("ids") != "" {
		if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
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
	var args = &models.MemoGetArgs{}
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.AbstractLength == 0 {
		args.AbstractLength = 20
	}
	result, err := models.MemoAPI.GetMemos(args)
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

	if !memo.ProjectID.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project"})
		return
	}

	memo.CreatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	memo.UpdatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}

	if !memo.Active.Valid {
		// memo.Active.Int = int64(models.MemoStatus["active"].(float64))
		memo.Active.Int = int64(config.Config.Models.Memos["active"])
		memo.Active.Valid = true
	}
	if !memo.PublishStatus.Valid {
		// memo.PublishStatus.Int = int64(models.MemoPublishStatus["draft"].(float64))
		memo.PublishStatus.Int = int64(config.Config.Models.MemosPublishStatus["draft"])
		memo.PublishStatus.Valid = true
	}
	if !memo.UpdatedBy.Valid {
		if memo.Author.Valid {
			memo.UpdatedBy = rrsql.NullInt{Int: memo.Author.Int, Valid: true}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Updator"})
		}
	}
	lastID, err := models.MemoAPI.InsertMemo(memo)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Memo ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if memo.PublishStatus.Valid && memo.PublishStatus.Int == int64(config.Config.Models.MemosPublishStatus["publish"]) &&
		memo.Active.Valid && memo.Active.Int == int64(config.Config.Models.Memos["active"]) {

		r.PublishHandler([]int{lastID})
	}
	if memo.UpdatedBy.Valid {
		r.UpdateHandler([]int{lastID}, memo.UpdatedBy.Int)
	} else {
		r.UpdateHandler([]int{lastID})
	}

	c.Status(http.StatusOK)
}

func (r *memoHandler) Put(c *gin.Context) {

	memo := models.Memo{}

	err := c.ShouldBindJSON(&memo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Memo Data"})
		return
	}
	if memo.PublishStatus.Valid {
		result, err := models.MemoAPI.GetMemo(int(memo.ID))
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

		switch memo.PublishStatus.Int {
		// case int64(models.MemoPublishStatus["schedule"].(float64)):
		case int64(config.Config.Models.MemosPublishStatus["schedule"]):
			if !memo.PublishedAt.Valid && !result.PublishedAt.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Publish Time"})
				return
			}
			fallthrough
		// case int64(models.MemoPublishStatus["publish"].(float64)):
		case int64(config.Config.Models.MemosPublishStatus["publish"]):
			if !memo.Title.Valid && !result.Title.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Memo Title"})
				return
			}
			if !memo.Content.Valid && !result.Content.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Memo Content"})
				return
			}
			if !memo.PublishedAt.Valid {
				memo.PublishedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
			}
			break
		}
	}

	if memo.CreatedAt.Valid {
		memo.CreatedAt.Valid = false
	}
	memo.UpdatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}

	switch {
	case memo.UpdatedBy.Valid:
		break
	case memo.Author.Valid:
		memo.UpdatedBy.Int = memo.Author.Int
		memo.UpdatedBy.Valid = true
		break
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

	if memo.PublishStatus.Valid || memo.Active.Valid {
		r.PublishHandler([]int{int(memo.ID)})
	}
	if memo.UpdatedBy.Valid {
		r.UpdateHandler([]int{int(memo.ID)}, memo.UpdatedBy.Int)
	} else {
		r.UpdateHandler([]int{int(memo.ID)})
	}

	c.Status(http.StatusOK)
}

func (r *memoHandler) Delete(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))
	err := models.MemoAPI.UpdateMemo(models.Memo{ID: uint32(id), Active: rrsql.NullInt{0, true}})

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
	r.UpdateHandler([]int{id})

	c.Status(http.StatusOK)
}

func (r *memoHandler) DeleteMany(c *gin.Context) {

	params := models.MemoUpdateArgs{}
	err := c.ShouldBindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Memo"})
		return
	}
	if len(params.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}

	if params.UpdatedBy == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Updater Not Specified"})
		return
	}

	params.UpdatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	params.Active = rrsql.NullInt{Int: int64(config.Config.Models.Posts["deactive"]), Valid: true}

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

	if params.UpdatedBy != 0 {
		r.UpdateHandler(params.IDs, params.UpdatedBy)
	} else {
		r.UpdateHandler(params.IDs)
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
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := models.MemoAPI.CountMemos(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *memoHandler) validateMemoSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(created_at|updated_at|published_at|id|memo_id|comment_amount|author|project_id|memo_order)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

func (r *memoHandler) PublishHandler(ids []int) error {
	// Redis notification
	// Mail notification

	if len(ids) == 0 {
		return nil
	}

	memoIDs := make([]int64, 0)
	for _, id := range ids {
		memoIDs = append(memoIDs, int64(id))
	}

	memos, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{
		IDs:               memoIDs,
		Active:            map[string][]int{"IN": []int{int(config.Config.Models.Memos["active"])}},
		MemoPublishStatus: map[string][]int{"IN": []int{int(config.Config.Models.MemosPublishStatus["publish"])}},
		AbstractLength:    20,
	})
	if err != nil {
		log.Println("Getting memos info fail when running publish handling process", err)
		return err
	}
	if len(memos) == 0 {
		return nil
	}

	for _, memo := range memos {
		p := models.Project{ID: memo.Project.ID, UpdatedAt: rrsql.NullTime{Time: time.Now(), Valid: true}}
		err := models.ProjectAPI.UpdateProjects(p)
		if err != nil {
			return err
		}
	}

	for _, memo := range memos {
		go models.NotificationGen.GenerateProjectNotifications(memo, "memo")
		go mail.MailAPI.SendMemoPublishMail(memo)
	}

	return nil
}

func (r *memoHandler) UpdateHandler(ids []int, params ...int64) error {
	// update update time for projects

	if len(ids) == 0 {
		return nil
	}

	memoIDs := make([]int64, 0)
	for _, id := range ids {
		memoIDs = append(memoIDs, int64(id))
	}
	memos, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{
		IDs:               memoIDs,
		Active:            map[string][]int{"IN": []int{int(config.Config.Models.Memos["active"])}},
		MemoPublishStatus: map[string][]int{"IN": []int{int(config.Config.Models.MemosPublishStatus["publish"])}},
		AbstractLength:    20,
	})
	if err != nil {
		log.Println("Getting memos info fail when running publish handling process", err)
		return err
	}
	if len(memos) == 0 {
		return nil
	}

	for _, memo := range memos {
		p := models.Project{ID: memo.Project.ID, UpdatedAt: rrsql.NullTime{Time: time.Now(), Valid: true}}
		if len(params) > 0 {
			p.UpdatedBy = rrsql.NullInt{Int: params[0], Valid: true}
		}
		go models.ProjectAPI.UpdateProjects(p)
	}

	return nil
}

func (r *memoHandler) SetRoutes(router *gin.Engine) {

	memoRouter := router.Group("/memo")
	{
		memoRouter.GET("/:id", r.Get)
		memoRouter.POST("", r.Post)
		memoRouter.PUT("", r.Put)
		memoRouter.DELETE("/:id", r.Delete)
	}
	memosRouter := router.Group("/memos")
	{
		memosRouter.GET("", r.GetMany)
		memosRouter.GET("/count", r.Count)
		memosRouter.DELETE("", r.DeleteMany)
	}
}

var MemoHandler memoHandler
