package routes

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
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
						ogInfo, err := OGParser.GetOGInfoFromUrl(v)
						if err != nil {
							log.Printf("%s %s parse embeded url fail: %v \n", msgType, actionType, err.Error())
							return
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

			_, err = models.CommentAPI.InsertComment(comment)
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
							ogInfo, err := OGParser.GetOGInfoFromUrl(v)
							if err != nil {
								log.Printf("%s %s parse embeded url fail: %v \n", msgType, actionType, err.Error())
								return
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

//** OG PArser **//

type OGInfo struct {
	Title       string `meta:"og:title"`
	Description string `meta:"og:description"`
	Image       string `meta:"og:image,og:image:url"`
}

type ogParser struct{}

func (o *ogParser) GetOGInfoFromUrl(urlStr string) (*OGInfo, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", urlStr, nil)
	// for k, v := range OGParserHeaders {
	for k, v := range config.Config.Crawler.Headers {
		req.Header.Add(k, v)
	}
	if !regexp.MustCompile("\\.readr\\.tw\\/").MatchString(urlStr) {
		req.Header.Del("Cookie")
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return o.GetPageInfoFromResponse(resp)
}

func (o *ogParser) GetPageInfoFromResponse(response *http.Response) (*OGInfo, error) {
	info := OGInfo{}
	html, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	err = o.GetPageDataFromHtml(html, &info)

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (o *ogParser) GetPageDataFromHtml(html []byte, data interface{}) error {
	buf := bytes.NewBuffer(html)
	doc, err := goquery.NewDocumentFromReader(buf)

	if err != nil {
		return err
	}

	return o.getPageData(doc, data)
}

func (o *ogParser) getPageData(doc *goquery.Document, data interface{}) error {
	var rv reflect.Value
	var ok bool
	if rv, ok = data.(reflect.Value); !ok {
		rv = reflect.ValueOf(data)
	}

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Should not be non-ptr or nil")
	}

	rt := rv.Type()

	for i := 0; i < rv.Elem().NumField(); i++ {
		fv := rv.Elem().Field(i)
		field := rt.Elem().Field(i)

		switch fv.Type().Kind() {
		case reflect.Ptr:
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			e := o.getPageData(doc, fv)

			if e != nil {
				return e
			}
		case reflect.Struct:
			e := o.getPageData(doc, fv.Addr())

			if e != nil {
				return e
			}
		case reflect.Slice:
			if fv.IsNil() {
				fv.Set(reflect.MakeSlice(fv.Type(), 0, 0))
			}

			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				last := reflect.New(field.Type.Elem())
				for {
					data := reflect.New(field.Type.Elem())
					e := o.getPageData(doc, data.Interface())

					if e != nil {
						return e
					}

					//Ugly solution (I can't remove nodes. Why?)
					if !reflect.DeepEqual(last.Elem().Interface(), data.Elem().Interface()) {
						fv.Set(reflect.Append(fv, data.Elem()))
						last.Elem().Set(data.Elem())

					} else {
						break
					}
				}
			case reflect.Ptr:
				last := reflect.New(field.Type.Elem().Elem())
				for {
					data := reflect.New(field.Type.Elem().Elem())
					e := o.getPageData(doc, data.Interface())

					if e != nil {
						return e
					}

					//Ugly solution (I can't remove nodes. Why?)
					if !reflect.DeepEqual(last.Elem().Interface(), data.Elem().Interface()) {
						fv.Set(reflect.Append(fv, data))
						last.Elem().Set(data.Elem())

					} else {
						break
					}
				}
			default:
				if tag, ok := field.Tag.Lookup("meta"); ok {
					tags := strings.Split(tag, ",")

					for _, t := range tags {
						contents := []reflect.Value{}

						processMeta := func(idx int, sel *goquery.Selection) {
							if c, existed := sel.Attr("content"); existed {
								if field.Type.Elem().Kind() == reflect.String {
									contents = append(contents, reflect.ValueOf(c))
								} else {
									i, e := strconv.Atoi(c)

									if e == nil {
										contents = append(contents, reflect.ValueOf(i))
									}
								}

								fv.Set(reflect.Append(fv, contents...))
							}
						}

						doc.Find(fmt.Sprintf("meta[property=\"%s\"]", t)).Each(processMeta)

						doc.Find(fmt.Sprintf("meta[name=\"%s\"]", t)).Each(processMeta)

						fv = reflect.Append(fv, contents...)
					}
				}
			}
		default:
			if tag, ok := field.Tag.Lookup("meta"); ok {

				tags := strings.Split(tag, ",")

				content := ""
				existed := false
				sel := (*goquery.Selection)(nil)
				for _, t := range tags {
					if sel = doc.Find(fmt.Sprintf("meta[property=\"%s\"]", t)).First(); sel.Size() > 0 {
						content, existed = sel.Attr("content")
					}

					if !existed {
						if sel = doc.Find(fmt.Sprintf("meta[name=\"%s\"]", t)).First(); sel.Size() > 0 {
							content, existed = sel.Attr("content")
						}
					}

					if existed {
						if fv.Type().Kind() == reflect.String {
							fv.Set(reflect.ValueOf(content))
						} else if fv.Type().Kind() == reflect.Int {
							if i, e := strconv.Atoi(content); e == nil {
								fv.Set(reflect.ValueOf(i))
							}
						}
						break
					}
				}
			}
		}
	}
	return nil
}

var OGParser ogParser

// var OGParserHeaders map[string]string
