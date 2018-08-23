package routes

import (
	"encoding/json"
	"errors"
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type followingHandler struct{}

func bindFollow(c *gin.Context) (result interface{}, err error) {

	var metadata = func(resource string) (table, key string, followtype int, active map[string][]int, err error) {

		method := c.Param("method")
		switch resource {
		case "member":
			table = "members"
			key = "id"
			followtype = config.Config.Models.FollowingType["member"]
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "post":
			table = "posts"
			key = "post_id"
			followtype = config.Config.Models.FollowingType["post"]
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "project":
			table = "projects"
			key = "project_id"
			followtype = config.Config.Models.FollowingType["project"]
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "memo":
			table = "memos"
			key = "memo_id"
			followtype = config.Config.Models.FollowingType["memo"]
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "report":
			table = "reports"
			key = "id"
			followtype = config.Config.Models.FollowingType["report"]
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "tag":
			table = "tags"
			key = "tag_id"
			if val, ok := config.Config.Models.FollowingType["tag"]; ok {
				followtype = val
			} else {
				return "", "", 0, nil, errors.New("Invalid following_type: tag")
			}
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		default:
			return "", "", 0, nil, errors.New("Unsupported Resource")
		}
		return table, key, followtype, active, nil
	}

	switch c.Param("method") {
	case "user":

		var params = &models.GetFollowingArgs{}
		// if there is no url paramter, bind json
		if c.Query("resource") == "" && c.Query("id") == "" {
			if err = c.ShouldBindJSON(params); err != nil {
				return nil, err
			}
		} else {
			if err = c.Bind(params); err != nil {
				return nil, err
			}
		}
		if c.Query("target_ids") != "" {
			err = json.Unmarshal([]byte(c.Query("target_ids")), &params.TargetIDs)
			if err != nil {
				return nil, err
			}
		}
		params.Table, params.PrimaryKey, params.FollowType, params.Active, err = metadata(params.ResourceName)
		if err != nil {
			return nil, err
		}
		if params.MemberID == 0 {
			return nil, errors.New("Bad Resource ID")
		}
		result = params

	case "resource":

		var params = &models.GetFollowedArgs{}
		if c.Query("resource") == "" && c.Query("ids") == "" {
			if err = c.ShouldBindJSON(params); err != nil {
				return nil, err
			}
		} else {
			if err = c.Bind(params); err != nil {
				return nil, err
			}
			if c.Query("ids") != "" {
				if err = json.Unmarshal([]byte(c.Query("ids")), &params.IDs); err != nil {
					return nil, errors.New("Bad Resource ID")
				}
			}
		}
		params.Table, params.PrimaryKey, params.FollowType, _, err = metadata(params.ResourceName)
		if err != nil {
			return nil, err
		}
		if len(params.IDs) == 0 {
			return nil, errors.New("Bad Resource ID")
		}
		// Only parse emotion parameter in resource
		if c.Query("emotion") != "" {

			if c.Query("resource") != "member" {
				if val, ok := config.Config.Models.Emotions[c.Query("emotion")]; ok {
					params.Emotion = val
				} else {
					return nil, errors.New("Unsupported Emotion")
				}
			} else {
				return nil, errors.New("Emotion Not Available For Member")
			}
		}
		result = params
	case "map":
		var params = &models.GetFollowMapArgs{}
		if c.Query("resource") == "" && c.Query("id") == "" {
			if err = c.ShouldBindJSON(params); err != nil {
				return nil, err
			}
		} else {
			if err = c.Bind(params); err != nil {
				return nil, err
			}
		}
		params.Table, params.PrimaryKey, params.FollowType, _, err = metadata(params.ResourceName)
		if err != nil {
			return nil, err
		}
		result = params
	default:
		return nil, errors.New("Unsupported Method")
	}
	return result, nil
}

func (r *followingHandler) Get(c *gin.Context) {

	var (
		input, result interface{}
		err           error
	)
	if input, err = bindFollow(c); err != nil {
		switch err.Error() {
		case "Unsupported Method":
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		}
		return
	}

	switch input := input.(type) {
	case *models.GetFollowingArgs:
		result, err = models.FollowingAPI.Get(input)
	case *models.GetFollowedArgs:
		result, err = models.FollowingAPI.Get(input)
		if err == nil {
			models.FollowCache.Update(*input, result.([]models.FollowedCount))
		}
	case *models.GetFollowMapArgs:
		result, err = models.FollowingAPI.Get(input)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Cannot Found Proper API"})
		return
	}

	if err != nil {
		switch err.Error() {
		case "Unsupported Resource", "Invalid Post Type":
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		default:
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *followingHandler) SetRoutes(router *gin.Engine) {
	router.GET("/following/:method", r.Get)
}

var FollowingHandler followingHandler
