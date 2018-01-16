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
		log.Fatalf("%v", err)
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

func (r *followingHandler) Get(c *gin.Context) {
	var (
		user_id  = c.Param("user_id")
		resource = c.Param("resource")
	)

	result, err := models.FollowingAPI.GetFollowing(map[string]string{
		"subject":  user_id,
		"resource": resource,
	})

	if err != nil {
		switch err.Error() {
		case "Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	c.JSON(http.StatusOK, result)
}

func (r *followingHandler) SetRoutes(router *gin.Engine) {
	followRouter := router.Group("following")
	{
		followRouter.GET("/:user_id/:resource", r.Get)
	}

	router.POST("/api/pubsub", r.Push)
}

var FollowingHandler followingHandler
