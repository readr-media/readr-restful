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

type projectHandler struct {
}

func (r *projectHandler) bindQuery(c *gin.Context, args *models.GetProjectArgs) (err error) {

	// Start parsing rest of request arguments
	if c.Query("active") != "" {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, models.ProjectActive); err != nil {
				return err
			}
		}
	}
	if c.Query("status") != "" {
		if err = json.Unmarshal([]byte(c.Query("status")), &args.Status); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Status, models.ProjectStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("publish_status")), &args.PublishStatus); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.PublishStatus, models.ProjectPublishStatus); err != nil {
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

	if c.Query("sort") != "" && r.validateProjectSorting(c.Query("sort")) {
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

func (r *projectHandler) Count(c *gin.Context) {
	var args = models.GetProjectArgs{}
	args.Default()
	if err := r.bindQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := models.ProjectAPI.CountProjects(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *projectHandler) Get(c *gin.Context) {
	var args = models.GetProjectArgs{}
	args.Default()

	if err := r.bindQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	projects, err := models.ProjectAPI.GetProjects(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": projects})
}

func (r *projectHandler) Post(c *gin.Context) {

	project := models.Project{}
	c.Bind(&project)

	// Pre-request test
	if project.Title.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project"})
		return
	}
	if project.Active.Valid == true && !r.validateProjectStatus(project.Active.Int) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
		return
	}

	if project.Status.Valid == true && project.Status.Int == int64(models.ProjectStatus["done"].(float64)) && project.Slug.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
		return
	}

	if !project.CreatedAt.Valid {
		project.CreatedAt = models.NullTime{time.Now(), true}
	}
	project.UpdatedAt = models.NullTime{time.Now(), true}

	err := models.ProjectAPI.InsertProject(project)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Project Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *projectHandler) Put(c *gin.Context) {

	project := models.Project{}
	c.Bind(&project)
	if project.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Project Data"})
		return
	}

	if project.Active.Valid == true && !r.validateProjectStatus(project.Active.Int) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
		return
	}

	if project.Status.Valid == true && project.Status.Int == int64(models.ProjectStatus["done"].(float64)) {
		p, err := models.ProjectAPI.GetProject(project)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Project Not Found"})
			return
		} else if p.Slug.Valid == false {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
			return
		}
	}

	if project.CreatedAt.Valid {
		project.CreatedAt.Valid = false
	}
	project.UpdatedAt = models.NullTime{time.Now(), true}

	err := models.ProjectAPI.UpdateProjects(project)
	if err != nil {
		switch err.Error() {
		case "Project Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Project Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *projectHandler) SchedulePublish(c *gin.Context) {
	models.ProjectAPI.SchedulePublish()
	c.Status(http.StatusOK)
}

func (r *projectHandler) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID Must Be Integer"})
		return
	}
	input := models.Project{ID: id}
	err = models.ProjectAPI.DeleteProjects(input)

	if err != nil {
		switch err.Error() {
		case "Project Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Project Not Found"})
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

func (r *projectHandler) PostAuthors(c *gin.Context) {
	params := struct {
		ProjectID *int  `json:"project_id"`
		AuthorIDs []int `json:"author_ids"`
	}{}
	c.Bind(&params)
	if params.ProjectID == nil || params.AuthorIDs == nil || len(params.AuthorIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Insufficient Parameters"})
		return
	}
	if err := models.ProjectAPI.InsertAuthors(*params.ProjectID, params.AuthorIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
}

func (r *projectHandler) PutAuthors(c *gin.Context) {
	params := struct {
		ProjectID *int  `json:"project_id"`
		AuthorIDs []int `json:"author_ids"`
	}{}
	c.Bind(&params)
	if params.ProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Insufficient Parameters"})
		return
	}
	if err := models.ProjectAPI.UpdateAuthors(*params.ProjectID, params.AuthorIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
}

func (r *projectHandler) SetRoutes(router *gin.Engine) {
	projectRouter := router.Group("/project")
	{
		projectRouter.GET("/count", r.Count)
		projectRouter.GET("/list", r.Get)
		projectRouter.POST("", r.Post)
		projectRouter.PUT("", r.Put)
		projectRouter.PUT("/schedule/publish", r.SchedulePublish)
		projectRouter.DELETE("/:id", r.Delete)

		authorRouter := projectRouter.Group("/author")
		{
			// authorRouter.GET("", r.GetAuthors)
			authorRouter.POST("", r.PostAuthors)
			authorRouter.PUT("", r.PutAuthors)
		}
	}
}

func (r *projectHandler) validateProjectStatus(i int64) bool {
	for _, v := range models.ProjectStatus {
		if i == int64(v.(float64)) {
			return true
		}
	}
	return false
}
func (r *projectHandler) validateProjectSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|published_at|project_id|project_order|status|slug)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

func (r *projectHandler) validate(target string, paradigm string) bool {
	if matched, err := regexp.MatchString(paradigm, target); err != nil || !matched {
		return false
	}
	return true
}

var ProjectHandler projectHandler
