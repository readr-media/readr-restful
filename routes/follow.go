package routes

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type followingHandler struct{}

func bind(c *gin.Context) (result interface{}, err error) {

	var metadata = func(method, resource string) (table, key string, followtype int, active map[string][]int, err error) {

		switch resource {
		case "member":
			table = "members"
			key = "id"
			followtype = 1
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "post":
			table = "posts"
			key = "post_id"
			followtype = 2
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "project":
			table = "projects"
			key = "project_id"
			followtype = 3
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "memo":
			table = "memos"
			key = "memo_id"
			followtype = 4
			if method == "user" {
				active = map[string][]int{"$in": []int{1}}
			}
		case "report":
			table = "reports"
			key = "id"
			followtype = 5
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
		params.Table, params.PrimaryKey, params.FollowType, params.Active, err = metadata(c.Param("method"), params.ResourceName)
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
		params.Table, params.PrimaryKey, params.FollowType, _, err = metadata(c.Param("method"), params.ResourceName)
		if err != nil {
			return nil, err
		}
		if len(params.IDs) == 0 {
			return nil, errors.New("Bad Resource ID")
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
		params.Table, params.PrimaryKey, params.FollowType, _, err = metadata(c.Param("method"), params.ResourceName)
		if err != nil {
			return nil, err
		}
		result = params
	default:
		return nil, errors.New("Unsupported Resource")
	}
	return result, nil
}

func (r *followingHandler) Get(c *gin.Context) {

	var (
		input, result interface{}
		err           error
	)
	if input, err = bind(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result, err = models.FollowingAPI.Get(input)
	if err != nil {
		switch err.Error() {
		case "Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
			return
		case "Unsupported Resource":
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		default:
			fmt.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *followingHandler) SetRoutes(router *gin.Engine) {
	router.GET("/following/:method", r.Get)
}

var FollowingHandler followingHandler
