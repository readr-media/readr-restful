package routes

import (
	"log"
	"strconv"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

var supportedAction = map[string]bool{
	"follow":   true,
	"unfollow": true,
}

type PubsubMessageStruct struct {
	Subscription string `json:"subscription"`
	Message      struct {
		ID   string   `json:"messageId"`
		Attr []string `json:"attributes"`
		Body []byte   `json:"data"`
	}
}

type PubsubMessageBody struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Subject  string `json:"subject"`
	Object   string `json:"object"`
}

type followingHandler struct{}

func (r *followingHandler) Push(c *gin.Context) {
	var input PubsubMessageStruct
	c.ShouldBindJSON(&input)

	var body PubsubMessageBody
	err := json.Unmarshal(input.Message.Body, &body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
		return
	}

	//log.Printf("%v", body)

	switch body.Action {
	case "follow", "unfollow":

		if _, err := strconv.Atoi(body.Object); body.Resource != "member" && err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
			return
		}

		params := map[string]string{
			"resource": body.Resource,
			"subject":  body.Subject,
			"object":   body.Object,
		}

		switch body.Action {
		case "follow":
			if err = models.FollowingAPI.AddFollowing(params); err != nil {
				log.Printf("%s fail: %v \n", body.Action, err.Error())
				c.JSON(http.StatusOK, err.Error())
				return
			}
			c.Status(http.StatusOK)
		case "unfollow":
			if err = models.FollowingAPI.DeleteFollowing(params); err != nil {
				log.Printf("%s fail: %v \n", body.Action, err.Error())
				c.JSON(http.StatusOK, err.Error())
				return
			}
			c.Status(http.StatusOK)
		}
	default:
		c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
		return
	}
}

func (r *followingHandler) GetByUser(c *gin.Context) {
	input := struct {
		MemberId string `json:"subject"`
		Resource string `json:"resource"`
		Type     string `json:"resource_type,omitempty"`
	}{}
	c.ShouldBindJSON(&input)

	result, err := models.FollowingAPI.GetFollowing(map[string]string{
		"subject":       input.MemberId,
		"resource":      input.Resource,
		"resource_type": input.Type,
	})

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

	switch input.Resource {
	case "member":
		break
	case "post", "project":
		for _, value := range input.Ids {
			_, err := strconv.Atoi(string(value))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Resource ID"})
				return
			}
		}
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

	router.POST("/restful/pubsub", r.Push)
}

var FollowingHandler followingHandler
