package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/mail"
)

type PostDesciption struct {
	models.Post
	Tags    models.NullIntSlice  `json:"tags" db:"tags"`
	Authors []models.AuthorInput `json:"authors" db:"authors"`
}

type postHandler struct{}

func (r *postHandler) bindQuery(c *gin.Context, args *models.PostArgs) (err error) {
	if err = c.ShouldBindQuery(args); err == nil {
		return nil
	}
	// Start parsing rest of request arguments
	if c.Query("author") != "" && args.Author == nil {
		if err = json.Unmarshal([]byte(c.Query("author")), &args.Author); err != nil {
			return err
		}
	}
	if c.Query("active") != "" && args.Active == nil {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, config.Config.Models.Posts); err != nil {
				return err
			}
		}
	}
	if c.Query("publish_status") != "" && args.PublishStatus == nil {
		if err = json.Unmarshal([]byte(c.Query("publish_status")), &args.PublishStatus); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.PublishStatus, config.Config.Models.PostPublishStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("type") != "" && args.Type == nil {
		if err = json.Unmarshal([]byte(c.Query("type")), &args.Type); err != nil {
			return err
		}
	}
	if c.Query("sort") != "" && r.validatePostSorting(c.Query("sort")) {
		args.Sorting = c.Query("sort")
	}
	if c.Query("ids") != "" {
		if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *postHandler) GetAll(c *gin.Context) {
	var args = models.NewPostArgs()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	result, err := models.PostAPI.GetPosts(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *postHandler) GetActivePosts(c *gin.Context) {
	var args = models.NewPostArgs()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	args.Active = map[string][]int{"$in": []int{config.Config.Models.Posts["active"]}}
	result, err := models.PostAPI.GetPosts(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *postHandler) Get(c *gin.Context) {

	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)

	var args = models.NewPostArgs()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}

	post, err := models.PostAPI.GetPost(id, args)

	if err != nil {
		switch err.Error() {
		case "Post Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Post Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": []models.TaggedPostMember{post}})
}

func (r *postHandler) Post(c *gin.Context) {

	var post PostDesciption
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if post.Post == (models.Post{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}

	if len(post.Authors) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Author"})
		return
	}

	// CreatedAt and UpdatedAt set default to now
	post.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	post.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	if !post.Active.Valid {
		post.Active = models.NullInt{int64(config.Config.Models.Posts["active"]), true}
	}
	if !post.PublishStatus.Valid {
		post.PublishStatus = models.NullInt{int64(config.Config.Models.PostPublishStatus["pending"]), true}
	}
	if !post.UpdatedBy.Valid {
		if len(post.Authors) > 0 && post.Authors[0].MemberID.Int != 0 {
			post.UpdatedBy.Int = post.Authors[0].MemberID.Int
			post.UpdatedBy.Valid = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		}
	}
	if post.PublishStatus.Valid && post.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"]) && !post.PublishedAt.Valid {
		post.PublishedAt = models.NullTime{Time: time.Now(), Valid: true}
	}

	postID, err := models.PostAPI.InsertPost(post.Post)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Post ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if post.Tags.Valid {
		err = models.TagAPI.UpdateTagging(config.Config.Models.TaggingType["post"], postID, post.Tags.Slice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if len(post.Authors) > 0 {
		post.Post.ID = uint32(postID)
		err = models.PostAPI.UpdateAuthors(post.Post, post.Authors)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	// Only do the pipeline when it's published
	if (!post.Active.Valid || post.Active.Int != int64(config.Config.Models.Posts["deactive"])) &&
		(post.PublishStatus.Valid && post.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"])) {
		r.PublishPipeline([]uint32{uint32(postID)})
	}
	c.Status(http.StatusOK)
}

func (r *postHandler) Put(c *gin.Context) {

	post := PostDesciption{}

	err := c.ShouldBindJSON(&post)
	// Check if post struct was binded successfully
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if post.Post == (models.Post{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Post"})
		return
	}
	// Discard CreatedAt even if there is data
	if post.CreatedAt.Valid {
		post.CreatedAt.Time = time.Time{}
		post.CreatedAt.Valid = false
	}
	if post.PublishStatus.Valid && post.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"]) && !post.PublishedAt.Valid {
		post.PublishedAt = models.NullTime{Time: time.Now(), Valid: true}
	}
	post.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	switch {
	case post.UpdatedBy.Valid:
	case len(post.Authors) > 0 && post.Authors[0].MemberID.Int != 0:
		post.UpdatedBy.Int = post.Authors[0].MemberID.Int
		post.UpdatedBy.Valid = true
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		return
	}

	err = models.PostAPI.UpdatePost(post.Post)
	if err != nil {
		switch err.Error() {
		case "Post Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Post Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if post.Tags.Valid {
		err = models.TagAPI.UpdateTagging(config.Config.Models.TaggingType["post"], int(post.ID), post.Tags.Slice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if len(post.Authors) > 0 {
		err = models.PostAPI.UpdateAuthors(post.Post, post.Authors)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if (post.PublishStatus.Valid && post.PublishStatus.Int != int64(config.Config.Models.PostPublishStatus["publish"])) ||
		(post.Active.Valid && post.Active.Int != int64(config.Config.Models.Posts["active"])) {
		// Case: Set a post to unpublished state, Delete the post from cache/searcher
		go models.SearchFeed.DeletePost([]int{int(post.ID)})
		go models.PostCache.Update(post.Post)

		if post.PublishStatus.Valid && post.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["pending"]) {
			m, err := models.PostAPI.GetPostAuthor(post.ID)
			if err != nil {
				log.Println("Get post author error: ", err.Error())
				return
			}
			if m.CustomEditor.Valid && m.CustomEditor.Bool == true {
				postDetail, err := models.PostAPI.GetPost(post.ID, &models.PostArgs{
					ShowAuthor: true,
				})
				if err != nil {
					log.Println("Fail to get post after updated: ", err.Error())
					return
				}
				go mail.MailAPI.SendCECommentNotify(postDetail)
				models.SlackHelper.SendCECommentNotify(postDetail)
			}

		}
	} else if post.PublishStatus.Valid || post.Active.Valid {
		// Case: Publish a post. Read whole post from database, then store to cache/searcher
		// Case: Update a post.
		err := r.PublishPipeline([]uint32{post.ID})
		if err != nil {
			log.Println("Handling publish pipeline fail when update all posts", err)
			return
		}
	} else {
		go models.PostCache.Update(post.Post)
	}

	c.Status(http.StatusOK)
}
func (r *postHandler) DeleteAll(c *gin.Context) {

	params := models.PostUpdateArgs{}
	// Bind params. If err return 400
	err := c.BindQuery(&params)
	// Unable to parse interface{} to []uint32
	// If change to []interface{} first, there's no way avoid for loop each time.
	// So here use individual parsing instead of function(interface{}) (interface{})
	if c.Query("ids") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	err = json.Unmarshal([]byte(c.Query("ids")), &params.IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if len(params.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	// if strings.HasPrefix(params.UpdatedBy, `"`) {
	// 	params.UpdatedBy = strings.TrimPrefix(params.UpdatedBy, `"`)
	// }
	// if strings.HasSuffix(params.UpdatedBy, `"`) {
	// 	params.UpdatedBy = strings.TrimSuffix(params.UpdatedBy, `"`)
	// }
	params.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	params.Active = models.NullInt{Int: int64(config.Config.Models.Posts["deactive"]), Valid: true}
	err = models.PostAPI.UpdateAll(params)
	if err != nil {
		switch err.Error() {
		case "Posts Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Posts Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	go models.SearchFeed.DeletePost(params.IDs)
	go models.PostCache.UpdateAll(params)

	c.Status(http.StatusOK)
}

func (r *postHandler) Delete(c *gin.Context) {

	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)
	err := models.PostAPI.DeletePost(id)
	if err != nil {
		switch err.Error() {
		case "Post Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Post Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *postHandler) PublishAll(c *gin.Context) {
	payload := models.PostUpdateArgs{}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if payload.IDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}
	payload.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}
	payload.PublishStatus = models.NullInt{Int: int64(config.Config.Models.PostPublishStatus["publish"]), Valid: true}
	err = models.PostAPI.UpdateAll(payload)
	if err != nil {
		switch err.Error() {
		case "Posts Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Posts Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	publishedIDs := make([]uint32, 0)
	for _, id := range payload.IDs {
		publishedIDs = append(publishedIDs, uint32(id))
	}
	err = r.PublishPipeline(publishedIDs)
	if err != nil {
		log.Println("Handling publish pipeline fail when update all posts", err)
		return
	}
}

func (r *postHandler) Count(c *gin.Context) {
	var args = &models.PostArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := models.PostAPI.Count(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

/*
func (r *postHandler) Hot(c *gin.Context) {
	result, err := models.PostAPI.Hot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}
*/

func (r *postHandler) PutCache(c *gin.Context) {
	models.PostCache.SyncFromDataStorage()
	c.Status(http.StatusOK)
}

func (r *postHandler) validatePostSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|created_at|published_at|post_id|author|comment_amount)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

func (r *postHandler) PublishPipeline(ids []uint32) error {
	// Insert to SearchFeed / Redis PostCache / Redis notification
	// Send notify mail / slack message

	if len(ids) == 0 {
		return nil
	}

	posts, err := models.PostAPI.GetPosts(models.NewPostArgs(func(arg *models.PostArgs) {
		arg.IDs = ids
		arg.ShowAuthor = true
		arg.ShowCommment = true
		arg.ShowTag = true
		arg.ShowUpdater = true
	}))
	if err != nil {
		log.Println("Getting posts info fail when running publish pipeline", err)
		return err
	}

	validPosts := make([]models.TaggedPostMember, 0)
	for _, post := range posts {
		if post.Active.Int == int64(config.Config.Models.Posts["active"]) &&
			post.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"]) {

			go models.NotificationGen.GeneratePostNotifications(post)
			validPosts = append(validPosts, post)
		}
	}
	go models.PostCache.SyncFromDataStorage()
	go models.SearchFeed.InsertPost(validPosts)

	return nil
}

func (r *postHandler) SetRoutes(router *gin.Engine) {

	postRouter := router.Group("/post")
	{
		postRouter.GET("/:id", r.Get)
		postRouter.POST("", r.Post)
		postRouter.PUT("", r.Put)
		postRouter.DELETE("/:id", r.Delete)
	}
	postsRouter := router.Group("/posts")
	{
		postsRouter.GET("", r.GetAll)
		postsRouter.GET("/active", r.GetActivePosts)
		postsRouter.DELETE("", r.DeleteAll)
		postsRouter.PUT("", r.PublishAll)

		postsRouter.GET("/count", r.Count)
		//postsRouter.GET("/hot", r.Hot)
		postsRouter.PUT("/cache", r.PutCache)
	}
}

var PostHandler postHandler
