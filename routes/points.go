package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type pointsHandler struct{}

func (r *pointsHandler) bindPointsQuery(c *gin.Context, args *models.PointsArgs) (err error) {

	// Fill in MaxResult, Page, OrderBy first to avoid custom parsing result overwritten
	if err = c.ShouldBindQuery(args); err != nil {
		return err
	}
	// Parse id
	if c.Param("id") != "" {
		id := c.Param("id")
		// Convert id to uint32
		args.ID, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return fmt.Errorf("Parsing id error. %v", err)
		}
	}
	// Parse ObjectType
	typestr := c.Param("type")
	// var objtype *int64
	if typestr != "" && strings.HasPrefix(typestr, "/") {
		if len(typestr) > 1 {
			typestr = typestr[1:]
			type64, err := strconv.ParseInt(typestr, 10, 32)
			if err != nil {
				return fmt.Errorf("Parsing object type error. %v", err)
			}
			args.ObjectType = &type64
		} else if len(typestr) == 1 {
			args.ObjectType = nil
		}
	} else {
		// typestr == "" or string has not prefix "/"
		args.ObjectType = nil
	}
	if c.Query("object_ids") != "" && args.ObjectIDs == nil {
		if err = json.Unmarshal([]byte(c.Query("object_ids")), &args.ObjectIDs); err != nil {
			return err
		}
	}
	return err
}

func (r *pointsHandler) Get(c *gin.Context) {

	var args = &models.PointsArgs{}
	args.Set(map[string]interface{}{
		"max_result": 15,
		"page":       1,
		"sort":       "-created_at",
	})
	if err := r.bindPointsQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	points, err := models.PointsAPI.Get(args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": points})
}

func (r *pointsHandler) Post(c *gin.Context) {
	pts := models.PointsToken{}
	if err := c.ShouldBindJSON(&pts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if !pts.CreatedAt.Valid {
		pts.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	}
	if !pts.UpdatedAt.Valid {
		pts.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	}

	if pts.Points.ObjectType == config.Config.Models.PointType["topup"] ||
		pts.Points.ObjectType == config.Config.Models.PointType["project"] {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ObjectType Deprecated"})
		return
	}

	// user can only gain currency
	if pts.Points.Currency < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Payment Amount"})
		return
	} else if pts.Points.Currency > 0 {
		if pts.Points.ObjectType == config.Config.Models.PointType["donate"] || pts.Points.ObjectType == config.Config.Models.PointType["project_memo"] {
			if pts.Token == nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Token"})
				return
			}
			if !pts.MemberName.Valid || pts.MemberPhone == nil || !pts.MemberMail.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Payment Info"})
				return
			}
		} else {
			// currency can only use in project_memo and donate type
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Currency Not Supported By ObjectType"})
			return
		}
	}

	if pts.Points.ObjectID == 0 && (pts.Points.ObjectType == config.Config.Models.PointType["project_memo"] || pts.Points.ObjectType == config.Config.Models.PointType["project"]) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Object ID"})
		return
	}

	// user can not donate point
	if pts.Points.Points != 0 && pts.Points.ObjectType == config.Config.Models.PointType["donate"] {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Cannot Donate Point"})
		return
	}

	points, id, err := models.PointsAPI.Insert(pts)
	if err != nil {
		switch err.Error() {
		case "Less than minimum points":
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points, "id": id})
}

func (r *pointsHandler) SetRoutes(router *gin.Engine) {
	pointsRouter := router.Group("/points")
	{
		// Redirect path ended without / to use r.Get as well
		pointsRouter.GET("/:id", r.Get)
		pointsRouter.GET("/:id/*type", r.Get)
		pointsRouter.POST("", r.Post)
	}
}

var PointsHandler pointsHandler
