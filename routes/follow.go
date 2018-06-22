package routes

import (
	"encoding/json"
	"errors"
	"fmt"
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

func bind(c *gin.Context, args *models.GetFollowArgs) (err error) {

	// if there is no url paramter, bind json
	if c.Query("resource") == "" && c.Query("ids") == "" {
		if err = c.ShouldBindJSON(args); err != nil {
			return err
		}
	} else {
		if err = c.Bind(args); err != nil {
			return err
		}
		if c.Query("ids") != "" {
			if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
				return err
			}
		}
	}

	switch args.ResourceName {
	case "member":
		args.TableName = "members"
		args.KeyName = "id"
		args.Active = map[string][]int{"$in": []int{1}}
		args.Type = 1
	case "post":
		args.TableName = "posts"
		args.KeyName = "post_id"
		args.Active = map[string][]int{"$in": []int{1}}
		args.Type = 2
	case "project":
		args.TableName = "projects"
		args.KeyName = "project_id"
		args.Active = map[string][]int{"$in": []int{1}}
		args.Type = 3
	case "memo":
		args.TableName = "memos"
		args.KeyName = "memo_id"
		args.Active = map[string][]int{"$in": []int{1}}
		args.Type = 4
	case "report":
		args.TableName = "reports"
		args.KeyName = "id"
		args.Active = map[string][]int{"$in": []int{1}}
		args.Type = 5
	default:
		return errors.New("Unsupported Resource")
	}
	args.Method = c.Param("method")

	if args.Method == "user" {
		args.IDs = args.IDs[:1]
	} else {
		if len(args.IDs) == 0 {
			return errors.New("Bad Resource ID")
		}
	}
	return nil
}

func (r *followingHandler) Get(c *gin.Context) {

	var input = &models.GetFollowArgs{}
	if err := bind(c, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	var (
		result interface{}
		err    error
	)

	switch input.Method {
	case "user":
		result, err = models.FollowingAPI.PseudoGetFollowing(*input)
	case "resource":
		result, err = models.FollowingAPI.PseudoGetFollowed(*input)
	case "map":
		fmt.Println(input)
	}
	if err != nil {
		switch err.Error() {
		case "Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *followingHandler) SetRoutes(router *gin.Engine) {
	followRouter := router.Group("following")
	{
		followRouter.GET("/byuser", r.GetByUser)
		followRouter.GET("/byresource", r.GetByResource)
		followRouter.GET("/map", r.GetFollowMap)

	}
	router.GET("/follows/:method", r.Get)
}

var FollowingHandler followingHandler
