package routes

import (
	"fmt"
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

var supportedAction = map[string]bool{
	"follow":         true,
	"unfollow":       true,
	"insert_emotion": true,
	"update_emotion": true,
	"delete_emotion": true,
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
	Emotion  string `json:"emotion"`
	Subject  int    `json:"subject"`
	Object   int    `json:"object"`
}

type pubsubHandler struct{}

func (r *pubsubHandler) Push(c *gin.Context) {
	var (
		input PubsubMessageMeta
		err   error
	)
	c.ShouldBindJSON(&input)

	msgType := input.Message.Attr["type"]
	actionType := input.Message.Attr["action"]

	switch msgType {
	case "follow", "emotion":

		var body PubsubFollowMsgBody

		err = json.Unmarshal(input.Message.Body, &body)
		if err != nil {
			log.Printf("Parse msg body fail: %v \n", err.Error())
			c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
			return
		}
		params := models.FollowArgs{Resource: body.Resource, Subject: int64(body.Subject), Object: int64(body.Object)}
		if val, ok := config.Config.Models.FollowingType[body.Resource]; ok {
			params.Type = val
		} else {
			c.JSON(http.StatusOK, gin.H{"Error": "Unsupported Resource"})
			return
		}

		if msgType == "follow" {

			// Follow situation set Emotion to none.
			if params.Emotion != 0 {
				params.Emotion = 0
			}

			switch actionType {
			case "follow":
				err = models.FollowingAPI.Insert(params)
				if err == nil {
					// Send Follow mail and Follow notification
					// Case: If user followed a project
					if params.Type == config.Config.Models.FollowingType["project"] {
						go models.MailAPI.SendFollowProjectMail(params)
					}
				}
			case "unfollow":
				err = models.FollowingAPI.Delete(params)
			default:
				log.Println("Follow action Type Not Support", actionType)
				c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
				return
			}

		} else if msgType == "emotion" {

			// Rule out member
			if params.Resource == "member" {
				c.JSON(http.StatusOK, gin.H{"Error": "Emotion Not Available For Member"})
				return
			}
			if val, ok := config.Config.Models.Emotions[body.Emotion]; ok {
				params.Emotion = val
			} else {
				c.JSON(http.StatusOK, gin.H{"Error": "Unsupported Emotion"})
				return
			}

			switch actionType {
			case "insert":
				err = models.FollowingAPI.Insert(params)
			case "update":
				err = models.FollowingAPI.Update(params)
			case "delete":
				err = models.FollowingAPI.Delete(params)
			default:
				log.Printf("Emotion action Type %s Not Support", actionType)
				c.JSON(http.StatusOK, gin.H{"Error": "Bad Request"})
				return
			}
		}

		if err != nil {
			log.Printf("%s fail: %v\n", actionType, err.Error())
			c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
			return
		}

		go models.FollowCache.Revoke(actionType, params.Resource, params.Emotion, params.Object)

		c.Status(http.StatusOK)

	case "comment":
		switch actionType {

		case "post":
			comment := models.InsertCommentArgs{}
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

			comment.Body.String = strings.Trim(html.EscapeString(comment.Body.String), " \n")
			escapedBody := url.PathEscape(comment.Body.String)
			escapedBody = strings.Replace(escapedBody, `%2F`, "/", -1)
			escapedBody = strings.Replace(escapedBody, `%20`, " ", -1)
			commentUrls := r.parseUrl(escapedBody)
			if len(commentUrls) > 0 {
				for _, v := range commentUrls {
					if !comment.OgTitle.Valid {
						ogInfo, err := models.OGParser.GetOGInfoFromUrl(v)
						if err != nil {
							log.Printf("%s %s parse embeded url fail: %v \n", msgType, actionType, err.Error())
							c.JSON(http.StatusOK, gin.H{"Error": fmt.Sprintf("%s %s parse embeded url fail: %v \n", msgType, actionType, err.Error())})
							break
						}
						comment.OgTitle = models.NullString{String: ogInfo.Title, Valid: true}
						if ogInfo.Description != "" {
							comment.OgDescription = models.NullString{String: ogInfo.Description, Valid: true}
						}
						if ogInfo.Image != "" {
							comment.OgImage = models.NullString{String: ogInfo.Image, Valid: true}
						}
					}
					escapedBody = strings.Replace(escapedBody, v, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, v, v), -1)
				}
				comment.Body.String, _ = url.PathUnescape(escapedBody)
			}

			comment.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
			// comment.Active = models.NullInt{Int: int64(models.CommentActive["active"].(float64)), Valid: true}
			// comment.Status = models.NullInt{Int: int64(models.CommentStatus["show"].(float64)), Valid: true}
			comment.Active = models.NullInt{Int: int64(config.Config.Models.Comment["active"]), Valid: true}
			comment.Status = models.NullInt{Int: int64(config.Config.Models.CommentStatus["show"]), Valid: true}

			commentID, err := models.CommentAPI.InsertComment(comment)
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, actionType, err.Error())
				c.Status(http.StatusOK)
				return
			}

			err = models.CommentAPI.UpdateCommentAmountByResource(comment.ResourceName.String, int(comment.ResourceID.Int), "+")
			if err != nil {
				log.Printf("%s %s fail: %v \n", msgType, "update comment amount", err.Error())
				c.Status(http.StatusOK)
				return
			}

			commentResType, commentResID := utils.ParseResourceInfo(comment.Resource.String)
			if commentResType == "memo" {
				commentMemoID, err := strconv.Atoi(commentResID)
				if err != nil {
					log.Println("Error parsing memoID when parsing comment resources")
				}
				memos, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{IDs: []int64{int64(commentMemoID)}})
				if err != nil || len(memos) > 1 {
					log.Println("Error getting memo info when insert comment")
				}
				if memos[0].Project.Project.Status.Int != int64(config.Config.Models.ProjectsStatus["done"]) {
					return
				}
			}

			commentAuthor, err := models.CommentAPI.GetComment(int(commentID))
			if err != nil {
				log.Printf("get comment fail when handling comment insertion: %v \n", err.Error())
				c.Status(http.StatusOK)
				return
			}
			go models.CommentCache.Insert(commentAuthor)

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

			if comment.Body.Valid {
				comment.Body.String = strings.Trim(html.EscapeString(comment.Body.String), " \n")
				escapedBody := url.PathEscape(comment.Body.String)
				escapedBody = strings.Replace(escapedBody, `%2F`, "/", -1)
				escapedBody = strings.Replace(escapedBody, `%20`, " ", -1)
				commentUrls := r.parseUrl(escapedBody)
				if len(commentUrls) > 0 {
					for _, v := range commentUrls {
						if !comment.OgTitle.Valid {
							ogInfo, err := models.OGParser.GetOGInfoFromUrl(v)
							if err != nil {
								log.Printf("%s %s parse embeded url fail: %v \n", msgType, actionType, err.Error())
								c.JSON(http.StatusOK, gin.H{"Error": fmt.Sprintf("%s %s parse embeded url fail: %v \n", msgType, actionType, err.Error())})
								break
							}
							comment.OgTitle = models.NullString{String: ogInfo.Title, Valid: true}
							if ogInfo.Description != "" {
								comment.OgDescription = models.NullString{String: ogInfo.Description, Valid: true}
							}
							if ogInfo.Image != "" {
								comment.OgImage = models.NullString{String: ogInfo.Image, Valid: true}
							}
						}
						escapedBody = strings.Replace(escapedBody, v, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, v, v), -1)
					}
					comment.Body.String, _ = url.PathUnescape(escapedBody)
				}
			}

			comment.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

			err = models.CommentAPI.UpdateComment(comment)
			if err != nil {
				log.Printf("%s %s UpdateComment fail: %v \n", msgType, actionType, err.Error())
			}

			if comment.Status.Valid || comment.Active.Valid {
				err = models.CommentAPI.UpdateAllCommentAmount()
				if err != nil {
					log.Printf("%s %s UpdateAllCommentAmount fail: %v \n", msgType, actionType, err.Error())
				}
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
					// Active:    models.NullInt{int64(models.CommentActive["deactive"].(float64)), true},
					Active: models.NullInt{Int: int64(config.Config.Models.Comment["deactive"]), Valid: true},
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

			if args.Status.Valid || args.Active.Valid {
				err = models.CommentAPI.UpdateAllCommentAmount()
				if err != nil {
					log.Printf("%s %s UpdateAllCommentAmount fail: %v \n", msgType, actionType, err.Error())
				}
			}

			c.Status(http.StatusOK)
		}

	default:
		log.Println("Pubsub Message Type Not Support", actionType)
		fmt.Println(msgType)
		c.Status(http.StatusOK)
		return
	}
}

func (r *pubsubHandler) parseUrl(body string) []string {
	matchResult := regexp.MustCompile("https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{2,256}\\.[a-z]{2,6}([-a-zA-Z0-9@:%_\\+.~#?&\\/\\/=]*)").FindAllString(body, -1)
	return matchResult
}

func (r *pubsubHandler) SetRoutes(router *gin.Engine) {
	router.POST("/restful/pubsub", r.Push)
}

var PubsubHandler pubsubHandler
