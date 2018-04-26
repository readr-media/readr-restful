package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type pointsHandler struct{}

func (r *pointsHandler) Get(c *gin.Context) {

	id := c.Param("id")
	// Convert id to uint32
	uid, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Parsing id error. %v", err)})
		return
	}
	typestr := c.Param("type")
	var objtype *int64
	if typestr != "" && strings.HasPrefix(typestr, "/") {
		if len(typestr) > 1 {
			typestr = typestr[1:]
			type64, err := strconv.ParseInt(typestr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return
			}
			objtype = &type64
		} else if len(typestr) == 1 {
			objtype = nil
		}
	}
	points, err := models.PointsAPI.Get(uint32(uid), objtype)
	if err != nil {
		switch err.Error() {
		case "Points Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": points})
}

func (r *pointsHandler) Post(c *gin.Context) {
	pts := models.Points{}
	if err := c.ShouldBindJSON(&pts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if !pts.CreatedAt.Valid {
		pts.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	}
	p, err := models.PointsAPI.Insert(pts)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Already exists"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"points": p})
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
