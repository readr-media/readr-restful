package routes

import (
	"log"
	"time"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

var supportedAction = map[string]bool{
	"follow":         true,
	"unfollow":       true,
	"post_comment":   true,
	"edit_comment":   true,
	"delete_comment": true,
}

type PubsubMessageMetaBody struct {
	ID   string            `json:"messageId"`
	Body []byte            `json:"data"`
	Attr map[string]string `json:"attributes"`
}

type PubsubMessageMeta struct {
	Subscription string `json:"subscription"`
	Message      PubsubMessageMetaBody
}

type PubsubFollowMsgBody struct {
	Resource string `json:"resource"`
	Subject  int    `json:"subject"`
	Object   int    `json:"object"`
}

type pubsubHandler struct{}

func (r *pubsubHandler) Push(c *gin.Context) {
	var input PubsubMessageMeta
	c.ShouldBindJSON(&input)

	msgType := input.Message.Attr["type"]
	actionType := input.Message.Attr["action"]

	switch msgType {
	case "follow":

		var body PubsubFollowMsgBody

		err := json.Unmarshal(input.Message.Body, &body)
		if err != nil {
			log.Printf("Parse msg body fail: %v \n", err.Error())
			c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
			return
		}

		params := models.FollowArgs{Resource: body.Resource, Subject: int64(body.Subject), Object: int64(body.Object)}

		switch actionType {
		case "follow":
			if err = models.FollowingAPI.AddFollowing(params); err != nil {
				log.Printf("%s fail: %v \n", actionType, err.Error())
				c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
				return
			}
			c.Status(http.StatusOK)
		case "unfollow":
			if err = models.FollowingAPI.DeleteFollowing(params); err != nil {
				log.Printf("%s fail: %v \n", actionType, err.Error())
				c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
				return
			}
		default:
			log.Println("Action Type Not Support", actionType)
			c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
		}
	case "comment":
		switch actionType {

		case "post":
			comment := models.Comment{}
			err := json.Unmarshal(input.Message.Body, &comment)
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
				c.Status(http.StatusOK)
				return
			}

			if !comment.Body.Valid || !comment.Author.Valid || !comment.Resource.Valid {
				log.Printf("%s %s fail: %v \n", msgType, actionType, "Missing Required Parameters")
				c.JSON(http.StatusOK, gin.H{"Error": "Missing Required Parameters"})
				return
			}

			comment.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
			comment.Active = models.NullInt{Int: int64(models.CommentActive["active"].(float64)), Valid: true}
			comment.Status = models.NullInt{Int: int64(models.CommentStatus["show"].(float64)), Valid: true}

			_, err = models.CommentAPI.InsertComment(comment)
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
				c.Status(http.StatusOK)
				return
			}
			c.Status(http.StatusOK)

		case "put":
			comment := models.Comment{}
			err := json.Unmarshal(input.Message.Body, &comment)
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
				c.Status(http.StatusOK)
				return
			}

			if comment.ID == 0 || comment.ParentID.Valid || comment.Resource.Valid || comment.CreatedAt.Valid || comment.Author.Valid {
				log.Printf("%s %s fail: %v \n", msgType, actionType, "Invalid Parameters")
				c.JSON(http.StatusOK, gin.H{"Error": "Invalid Parameters"})
				return
			}

			comment.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

			err = models.CommentAPI.UpdateComment(comment)
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
			}
			c.Status(http.StatusOK)

		case "putstatus", "delete":
			args := models.CommentUpdateArgs{}
			err := json.Unmarshal(input.Message.Body, &args)
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
				c.Status(http.StatusOK)
				return
			}

			if len(args.IDs) == 0 {
				log.Printf("%s %s fail: %v \n", msgType, actionType, "ID List Empty")
				c.JSON(http.StatusOK, gin.H{"Error": "ID List Empty"})
				return
			}

			if actionType == "delete" {
				args = models.CommentUpdateArgs{
					IDs:       args.IDs,
					UpdatedAt: models.NullTime{Time: time.Now(), Valid: true},
					Active:    models.NullInt{int64(models.CommentActive["deactive"].(float64)), true},
				}
			} else {
				args.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
			}

			err = models.CommentAPI.UpdateComments(args)
			if err != nil {
				switch err.Error() {
				case "Posts Not Found":
					log.Printf("%s %s fail: %v \n", msgType, actionType, "Comments Not Found")
				default:
					log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
				}
			}
			c.Status(http.StatusOK)
		}

	default:
		log.Println("Pubsub Message Type Not Support", actionType)
		c.Status(http.StatusOK)
		return
	}
}

func (r *pubsubHandler) SetRoutes(router *gin.Engine) {
	router.POST("/restful/pubsub", r.Push)
}

var PubsubHandler pubsubHandler
