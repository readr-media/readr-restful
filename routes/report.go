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
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/mail"
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
			// if err = models.ValidateActive(args.Active, models.ReportActive); err != nil {
			if err = models.ValidateActive(args.Active, config.Config.Models.Reports); err != nil {
				return err
			}
		}
	}
	if c.Query("report_publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("report_publish_status")), &args.ReportPublishStatus); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			// if err = models.ValidateActive(args.PublishStatus, models.ProjectPublishStatus); err != nil {
			if err = models.ValidateActive(args.ReportPublishStatus, config.Config.Models.ReportsPublishStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("project_publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_publish_status")), &args.ProjectPublishStatus); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			// if err = models.ValidateActive(args.PublishStatus, models.ProjectPublishStatus); err != nil {
			if err = models.ValidateActive(args.ProjectPublishStatus, config.Config.Models.ProjectsPublishStatus); err != nil {
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
	if c.Query("report_slugs") != "" {
		if err = json.Unmarshal([]byte(c.Query("report_slugs")), &args.Slugs); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("project_slugs") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_slugs")), &args.ProjectSlugs); err != nil {
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
		sortingString := c.Query("sort")
		for _, v := range strings.Split(sortingString, ",") {
			if v == "id" {
				sortingString = strings.Replace(sortingString, "id", "post_id", -1)
			}
		}
		args.Sorting = sortingString
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
	if !report.Title.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report"})
		return
	}

	if report.PublishStatus.Valid == true && report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) && report.Slug.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
		return
	}

	if !report.CreatedAt.Valid {
		report.CreatedAt = models.NullTime{time.Now(), true}
	}
	report.UpdatedAt = models.NullTime{time.Now(), true}
	report.Active = models.NullInt{int64(config.Config.Models.Reports["active"]), true}

	args := models.GetReportArgs{
		MaxResult: 1,
		Slugs:     []string{report.Slug.String},
		Active:    map[string][]int{"in": []int{1}},
	}
	args.Fields = args.FullAuthorTags()
	result, err := models.ReportAPI.GetReports(args)
	if err != nil {
		if err.Error() != "Not Found" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	} else if err == nil && len(result) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Duplicate Slug"})
		return
	}

	lastID, err := models.ReportAPI.InsertReport(report)
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

	if report.PublishStatus.Valid && report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) {
		r.PublishHandler([]int{lastID})
		if report.UpdatedBy.Valid {
			r.UpdateHandler([]int{lastID}, report.UpdatedBy.Int)
		} else {
			r.UpdateHandler([]int{lastID})
		}
	}

	resp := map[string]int{"last_id": lastID}
	c.JSON(http.StatusOK, gin.H{"_items": resp})
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

	if report.PublishStatus.Valid == true && !r.validateReportPublishStatus(report.PublishStatus.Int) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
		return
	}

	// if report.PublishStatus.Valid == true && (report.PublishStatus.Int == int64(models.ReportPublishStatus["publish"].(float64)) || report.PublishStatus.Int == int64(models.ReportPublishStatus["schedule"].(float64))) {
	if report.PublishStatus.Valid == true && (report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) || report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["schedule"])) {
		p, err := models.ReportAPI.GetReport(report)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Not Found"})
			return
		} else if p.Slug.Valid == false {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
			return
		}

		switch p.PublishStatus.Int {
		// case int64(models.ReportPublishStatus["schedule"].(float64)):
		case int64(config.Config.Models.ReportsPublishStatus["schedule"]):
			if !report.PublishedAt.Valid && !p.PublishedAt.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Publish Time"})
				return
			}
			fallthrough
		// case int64(models.ReportPublishStatus["publish"].(float64)):
		case int64(config.Config.Models.ReportsPublishStatus["publish"]):
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
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	if (report.PublishStatus.Valid && report.PublishStatus.Int != int64(config.Config.Models.ReportsPublishStatus["publish"])) ||
		(report.Active.Valid && report.Active.Int != int64(config.Config.Models.Reports["active"])) {
		// Case: Set a report to unpublished state, Delete the report from cache/searcher
		// go models.Algolia.DeleteReport([]int{int(report.ID)})
	} else if report.PublishStatus.Valid || report.Active.Valid {
		// Case: Publish a report or update a report.
		if report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) ||
			report.Active.Int == int64(config.Config.Models.Reports["active"]) {
			r.PublishHandler([]int{int(report.ID)})
			if report.UpdatedBy.Valid {
				r.UpdateHandler([]int{int(report.ID)}, report.UpdatedBy.Int)
			} else {
				r.UpdateHandler([]int{int(report.ID)})
			}
		}
	} else {
		r.UpdateHandler([]int{int(report.ID)})
	}
	c.Status(http.StatusOK)
}

func (r *reportHandler) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID Must Be Integer"})
		return
	}
	input := models.Report{ID: uint32(id)}
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
	// go models.Algolia.DeleteReport([]int{id})

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

func (r *reportHandler) PublishHandler(ids []int) error {
	// Redis notification
	// Mail notification

	if len(ids) == 0 {
		return nil
	}

	args := models.GetReportArgs{
		IDs:                 ids,
		Active:              map[string][]int{"IN": []int{int(config.Config.Models.Reports["active"])}},
		ReportPublishStatus: map[string][]int{"IN": []int{int(config.Config.Models.ReportsPublishStatus["publish"])}},
	}
	args.Fields = args.FullAuthorTags()

	reports, err := models.ReportAPI.GetReports(args)
	if err != nil {
		log.Println("Getting reports info fail when running publish handling process", err)
		return err
	}
	if len(reports) == 0 {
		return nil
	}

	for _, report := range reports {
		p := models.Project{ID: int(report.ProjectID.Int), UpdatedAt: models.NullTime{Time: time.Now(), Valid: true}}
		if report.UpdatedBy.Valid {
			p.UpdatedBy = report.UpdatedBy
		}
		err := models.ProjectAPI.UpdateProjects(p)
		if err != nil {
			return err
		}
	}
	// go models.Algolia.InsertReport(reports)
	for _, report := range reports {
		go models.NotificationGen.GenerateProjectNotifications(report, "report")
		go mail.MailAPI.SendReportPublishMail(report)
	}

	return nil
}

func (r *reportHandler) UpdateHandler(ids []int, params ...int64) error {
	// update update time for projects

	if len(ids) == 0 {
		return nil
	}

	args := models.GetReportArgs{
		IDs:                 ids,
		Active:              map[string][]int{"IN": []int{int(config.Config.Models.Reports["active"])}},
		ReportPublishStatus: map[string][]int{"IN": []int{int(config.Config.Models.ReportsPublishStatus["publish"])}},
	}
	args.Fields = args.FullAuthorTags()

	reports, err := models.ReportAPI.GetReports(args)
	if err != nil {
		log.Println("Getting reports info fail when running publish handling process", err)
		return err
	}
	if len(reports) == 0 {
		return nil
	}

	for _, report := range reports {
		p := models.Project{ID: report.Project.ID, UpdatedAt: models.NullTime{Time: time.Now(), Valid: true}}
		if len(params) > 0 {
			p.UpdatedBy = models.NullInt{Int: params[0], Valid: true}
		}
		go models.ProjectAPI.UpdateProjects(p)
	}
	return nil
}

func (r *reportHandler) SetRoutes(router *gin.Engine) {
	reportRouter := router.Group("/report")
	{
		reportRouter.GET("/count", r.Count)
		reportRouter.GET("/list", r.Get)
		reportRouter.POST("", r.Post)
		reportRouter.PUT("", r.Put)
		reportRouter.DELETE("/:id", r.Delete)

		authorRouter := reportRouter.Group("/author")
		{
			authorRouter.POST("", r.PostAuthors)
			authorRouter.PUT("", r.PutAuthors)
		}
	}
}

func (r *reportHandler) validateReportStatus(i int64) bool {
	// for _, v := range models.ReportActive {
	for _, v := range config.Config.Models.Reports {
		// if i == int64(v.(float64)) {
		if i == int64(v) {
			return true
		}
	}
	return false
}
func (r *reportHandler) validateReportPublishStatus(i int64) bool {
	// for _, v := range models.ReportActive {
	for _, v := range config.Config.Models.ReportsPublishStatus {
		// if i == int64(v.(float64)) {
		if i == int64(v) {
			return true
		}
	}
	return false
}
func (r *reportHandler) validateReportSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|published_at|id|slug|comment_amount)", v); err != nil || !matched {
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
