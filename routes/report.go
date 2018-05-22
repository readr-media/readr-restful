package routes

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type reportHandler struct {
}

func (r *reportHandler) bindQuery(c *gin.Context, args *models.GetReportArgs) (err error) {

	// Start parsing rest of request arguments
	if c.Query("active") != "" {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, models.ReportActive); err != nil {
				return err
			}
		}
	}
	if c.Query("publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("publish_status")), &args.PublishStatus); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.PublishStatus, models.ReportPublishStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("ids") != "" {
		if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("slugs") != "" {
		if err = json.Unmarshal([]byte(c.Query("slugs")), &args.Slugs); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("project_id") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_id")), &args.Project); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("max_result") != "" {
		if err = json.Unmarshal([]byte(c.Query("max_result")), &args.MaxResult); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("page") != "" {
		if err = json.Unmarshal([]byte(c.Query("page")), &args.Page); err != nil {
			log.Println(err.Error())
			return err
		}
	}

	if c.Query("sort") != "" && r.validateReportSorting(c.Query("sort")) {
		args.Sorting = c.Query("sort")
	}

	if c.Query("keyword") != "" {
		args.Keyword = c.Query("keyword")
	}
	if c.Query("fields") != "" {
		if err = json.Unmarshal([]byte(c.Query("fields")), &args.Fields); err != nil {
			log.Println(err.Error())
			return err
		}
		for _, field := range args.Fields {
			if !r.validate(field, fmt.Sprintf("^(%s)$", strings.Join(args.FullAuthorTags(), "|"))) {
				return errors.New("Invalid Fields")
			}
		}
	} else {
		switch c.Query("mode") {
		case "full":
			args.Fields = args.FullAuthorTags()
		default:
			args.Fields = []string{"nickname"}
		}
	}
	return nil
}

func (r *reportHandler) Count(c *gin.Context) {
	var args = models.GetReportArgs{}
	args.Default()
	if err := r.bindQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := models.ReportAPI.CountReports(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *reportHandler) Get(c *gin.Context) {
	var args = models.GetReportArgs{}
	args.Default()

	if err := r.bindQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	reports, err := models.ReportAPI.GetReports(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": reports})
}

func (r *reportHandler) Post(c *gin.Context) {

	report := models.Report{}
	err := c.ShouldBind(&report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report"})
		return
	}

	// Pre-request test
	if report.Title.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report"})
		return
	}

	if report.PublishStatus.Valid == true && report.PublishStatus.Int == int64(models.ReportPublishStatus["publish"].(float64)) && report.Slug.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
		return
	}

	if !report.CreatedAt.Valid {
		report.CreatedAt = models.NullTime{time.Now(), true}
	}
	report.UpdatedAt = models.NullTime{time.Now(), true}
	report.Active = models.NullInt{int64(models.ReportActive["active"].(float64)), true}

	err = models.ReportAPI.InsertReport(report)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *reportHandler) Put(c *gin.Context) {

	report := models.Report{}
	err := c.ShouldBind(&report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Data"})
		return
	}

	if report.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Data"})
		return
	}

	if report.Active.Valid == true && !r.validateReportStatus(report.Active.Int) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
		return
	}

	if report.PublishStatus.Valid == true && (report.PublishStatus.Int == int64(models.ReportPublishStatus["publish"].(float64)) || report.PublishStatus.Int == int64(models.ReportPublishStatus["schedule"].(float64))) {
		p, err := models.ReportAPI.GetReport(report)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Not Found"})
			return
		} else if p.Slug.Valid == false {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
			return
		}

		switch p.PublishStatus.Int {
		case int64(models.ReportPublishStatus["schedule"].(float64)):
			if !report.PublishedAt.Valid && !p.PublishedAt.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Publish Time"})
				return
			}
			fallthrough
		case int64(models.ReportPublishStatus["publish"].(float64)):
			if !report.Title.Valid && !p.Title.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Title"})
				return
			}
			if !report.Slug.Valid && !p.Slug.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Content"})
				return
			}
			if !report.PublishedAt.Valid {
				report.PublishedAt = models.NullTime{Time: time.Now(), Valid: true}
			}
			break
		}
	}

	if report.CreatedAt.Valid {
		report.CreatedAt.Valid = false
	}
	report.UpdatedAt = models.NullTime{time.Now(), true}

	err = models.ReportAPI.UpdateReport(report)
	if err != nil {
		switch err.Error() {
		case "Report Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *reportHandler) SchedulePublish(c *gin.Context) {
	models.ReportAPI.SchedulePublish()
	c.Status(http.StatusOK)
}

func (r *reportHandler) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID Must Be Integer"})
		return
	}
	input := models.Report{ID: id}
	err = models.ReportAPI.DeleteReport(input)

	if err != nil {
		switch err.Error() {
		case "Report Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Report Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

// func (r *projectHandler) GetAuthors(c *gin.Context) {
// 	//project/authors?ids=[1000010,1000013]&mode=[full]&fields=["id","member_id"]
// 	args := models.GetProjectArgs{}
// 	if err := r.bindQuery(c, &args); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	fmt.Println(args)
// 	authors, err := models.ProjectAPI.GetAuthors(args)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"_items": authors})
// }

func (r *reportHandler) PostAuthors(c *gin.Context) {
	params := struct {
		ReportID  *int  `json:"report_id"`
		AuthorIDs []int `json:"author_ids"`
	}{}
	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	if params.ReportID == nil || params.AuthorIDs == nil || len(params.AuthorIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Insufficient Parameters"})
		return
	}
	if err := models.ReportAPI.InsertAuthors(*params.ReportID, params.AuthorIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
}

func (r *reportHandler) PutAuthors(c *gin.Context) {
	params := struct {
		ReportID  *int  `json:"report_id"`
		AuthorIDs []int `json:"author_ids"`
	}{}
	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	if params.ReportID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Insufficient Parameters"})
		return
	}
	if err := models.ReportAPI.UpdateAuthors(*params.ReportID, params.AuthorIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
}

func (r *reportHandler) SetRoutes(router *gin.Engine) {
	reportRouter := router.Group("/report")
	{
		reportRouter.GET("/count", r.Count)
		reportRouter.GET("/list", r.Get)
		reportRouter.POST("", r.Post)
		reportRouter.PUT("", r.Put)
		reportRouter.PUT("/schedule/publish", r.SchedulePublish)
		reportRouter.DELETE("/:id", r.Delete)

		authorRouter := reportRouter.Group("/author")
		{
			authorRouter.POST("", r.PostAuthors)
			authorRouter.PUT("", r.PutAuthors)
		}
	}
}

func (r *reportHandler) validateReportStatus(i int64) bool {
	for _, v := range models.ReportActive {
		if i == int64(v.(float64)) {
			return true
		}
	}
	return false
}
func (r *reportHandler) validateReportSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|published_at|id|slug)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

func (r *reportHandler) validate(target string, paradigm string) bool {
	if matched, err := regexp.MatchString(paradigm, target); err != nil || !matched {
		return false
	}
	return true
}

var ReportHandler reportHandler
