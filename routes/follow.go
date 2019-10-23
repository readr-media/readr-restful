package routes

import (
	"encoding/json"
	"errors"
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

type followingHandler struct{}

func bindFollow(c *gin.Context) (result interface{}, err error) {

	switch c.Param("method") {
	case "user":

		// TODO: check resource name parameter

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

		params.Active = map[string][]int{"$in": []int{1}}

		err = json.Unmarshal([]byte(params.ResourceName), &params.Resources)
		if err != nil {
			params.Resources = []string{params.ResourceName}
		}

		for _, resName := range params.Resources {
			if !validateFollowType(resName) {
				return nil, errors.New("Bad Following Type")
			}
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
		params.Table, params.PrimaryKey, params.FollowType, err = rrsql.GetResourceMetadata(params.ResourceName)
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
	/*
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
			params.Table, params.PrimaryKey, params.FollowType, err = models.GetResourceMetadata(params.ResourceName)
			if err != nil {
				return nil, err
			}
			result = params
	*/
	default:
		return nil, errors.New("Unsupported Method")
	}
	return result, nil
}

func validateFollowType(resourceName string) bool {
	if _, ok := config.Config.Models.FollowingType[resourceName]; !ok {
		return false
	}
	return true
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
	/*
		case *models.GetFollowMapArgs:
			result, err = models.FollowingAPI.Get(input)
	*/
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
