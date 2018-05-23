package routes

import (
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type followingHandler struct{}

func (r *followingHandler) GetByUser(c *gin.Context) {
	var input models.GetFollowingArgs
	c.ShouldBindJSON(&input)

	result, err := models.FollowingAPI.GetFollowing(input)

	if err != nil {
		switch err.Error() {
		case "Not Found":
			c.JSON(http.StatusOK, make([]string, 0))
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	c.JSON(http.StatusOK, result)
}

func (r *followingHandler) GetByResource(c *gin.Context) {
	var input models.GetFollowedArgs
	c.ShouldBindJSON(&input)
	if len(input.Ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Resource ID"})
		return
	}
	for _, v := range input.Ids {
		_, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Resource ID"})
			return
		}
	}

	switch input.Resource {
	case "member", "post", "project":
		break
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unsupported Resource"})
		return
	}

	result, err := models.FollowingAPI.GetFollowed(input)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *followingHandler) GetFollowMap(c *gin.Context) {
	var input models.GetFollowMapArgs
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Parameter Format"})
		return
	}

	result, err := models.FollowingAPI.GetFollowMap(input)

	if err != nil {
		switch err.Error() {
		case "Resource Not Supported":
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	c.JSON(http.StatusOK, struct {
		Map      []models.FollowingMapItem `json:"list"`
		Resource string                    `json:"resource"`
	}{result, input.Resource})
}

func (r *followingHandler) SetRoutes(router *gin.Engine) {
	followRouter := router.Group("following")
	{
		followRouter.GET("/byuser", r.GetByUser)
		followRouter.GET("/byresource", r.GetByResource)
		followRouter.GET("/map", r.GetFollowMap)
	}
}

var FollowingHandler followingHandler
