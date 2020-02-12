package routes

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	rt "github.com/readr-media/readr-restful/internal/router"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/asset"
)

type filterHandler struct{}

func (r *filterHandler) bindQuery(c *gin.Context, args *models.FilterArgs) (err error) {
	_ = c.ShouldBindQuery(args)

	if c.Query("id") != "" {
		args.ID, err = strconv.ParseInt(c.Query("id"), 10, 64)
		if err != nil {
			return err
		}
	}
	if c.Query("slug") != "" {
		args.Slug = c.Query("slug")
	}
	if c.Query("mail") != "" {
		args.Mail = c.Query("mail")
	}
	if c.Query("nickname") != "" {
		args.Nickname = c.Query("nickname")
	}

	if c.Query("title") != "" {
		args.Title = strings.Split(c.Query("title"), ",")
	}
	if c.Query("description") != "" {
		args.Description = strings.Split(c.Query("description"), ",")
	}
	if c.Query("content") != "" {
		args.Content = strings.Split(c.Query("content"), ",")
	}
	if c.Query("author") != "" {
		args.Author = strings.Split(c.Query("author"), ",")
	}
	if c.Query("tag") != "" {
		args.Tag = strings.Split(c.Query("tag"), ",")
	}

	if c.Query("published_at") != "" {
		timearg, err := r.bindTimeOp(c.Query("published_at"))
		if err != nil {
			return err
		}
		args.PublishedAt = timearg
	}
	if c.Query("updated_at") != "" {
		timearg, err := r.bindTimeOp(c.Query("updated_at"))
		if err != nil {
			return err
		}
		args.UpdatedAt = timearg
	}
	return nil
}

func (r *filterHandler) bindTimeOp(param string) (timeOps map[string]time.Time, err error) {
	timeOps = make(map[string]time.Time)
	statements := strings.Split(param, "::")
	if len(statements) > 0 {
		for _, statement := range statements {
			statementTokens := strings.Split(statement, ":")
			if len(statementTokens) != 2 {
				return timeOps, errors.New("Malformed Time Argument")
			}

			if len(statementTokens[1]) > 10 {
				statementTokens[1] = statementTokens[1][0:10]
			}

			timeStamp, err := strconv.ParseInt(statementTokens[1], 10, 64)
			if err != nil {
				return timeOps, errors.New("Malformed Time Argument")
			}

			tm := time.Unix(timeStamp, 0)
			timeOps[statementTokens[0]] = tm
		}
	}
	return timeOps, err
}

func (r *filterHandler) Get(c *gin.Context) {

	var resource = strings.Split(c.Request.URL.Path, "/")[1]
	var args = &models.FilterArgs{}

	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if args.MaxResult == 0 {
		args.MaxResult = 15
	}
	if args.Page == 0 {
		args.Page = 1
	}
	if args.Sorting == "" {
		args.Sorting = "-created_at"
	}

	switch resource {
	case "project":
		getArg := models.FilterProjectArgs{
			FilterArgs: *args,
			Fields:     []string{"project_id", "title", "slug", "progress", "status", "publish_status", "published_at"},
		}
		projects, err := models.ProjectAPI.FilterProjects(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		total, err := models.ProjectAPI.CountProjects(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"_items": projects,
			"_meta":  rt.ResponseMeta{Total: &total},
		})
	case "posts":
		getArg := models.FilterPostArgs{
			FilterArgs: *args,
		}
		posts, err := models.PostAPI.FilterPosts(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		total, err := models.PostAPI.Count(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"_items": posts,
			"_meta":  rt.ResponseMeta{Total: &total},
		})
	case "asset":
		getArg := asset.FilterAssetArgs{
			FilterArgs: *args,
		}
		assets, err := asset.AssetAPI.FilterAssets(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		total, err := asset.AssetAPI.Count(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"_items": assets,
			"_meta":  rt.ResponseMeta{Total: &total},
		})
	case "members":
		getArg := models.FilterMemberArgs{
			FilterArgs: *args,
			Fields:     []string{"id", "mail", "nickname", "role", "custom_editor", "updated_at"},
		}
		members, err := models.MemberAPI.FilterMembers(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		total, err := models.MemberAPI.Count(&getArg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"_items": members,
			"_meta":  rt.ResponseMeta{Total: &total},
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Resource Not Supported"})
	}

}

func (r *filterHandler) SetRoutes(router *gin.Engine) {

	projectRouter := router.Group("/project")
	projectRouter.GET("/filter", r.Get)
	postRouter := router.Group("/posts")
	postRouter.GET("/filter", r.Get)
	assetRouter := router.Group("/asset")
	assetRouter.GET("/filter", r.Get)
	memberRouter := router.Group("/members")
	memberRouter.GET("/filter", r.Get)

}

var FilterHandler filterHandler
