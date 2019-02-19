package asset

import (
	"regexp"
	"strings"
	"time"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type taggedAsset struct {
	Asset
	Tags models.NullIntSlice `json:"tags" db:"tags"`
}

type router struct{}

func (r *router) bindQuery(c *gin.Context, args *GetAssetArgs) (err error) {
	_ = c.ShouldBindQuery(args)

	if c.Query("active") != "" && args.Active == nil {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, config.Config.Models.Assets); err != nil {
				return err
			}
		}
	}
	if c.Query("file_type") != "" {
		if err = json.Unmarshal([]byte(c.Query("file_type")), &args.FileType); err != nil {
			return err
		}
	}
	if c.Query("asset_type") != "" {
		if err = json.Unmarshal([]byte(c.Query("asset_type")), &args.AssetType); err != nil {
			return err
		}
	}
	if c.Query("ids") != "" && args.IDs == nil {
		if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
			return err
		}
	}
	if c.Query("sort") != "" && r.validateSorting(c.Query("sort")) {
		args.Sorting = c.Query("sort")
	}

	return nil
}

func (r *router) Count(c *gin.Context) {
	var args = NewAssetArgs()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := AssetAPI.Count(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *router) Delete(c *gin.Context) {

	IDs := []int{}
	if c.Query("ids") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	err := json.Unmarshal([]byte(c.Query("ids")), &IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if len(IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}

	err = AssetAPI.Delete(IDs)
	if err != nil {
		switch err.Error() {
		case "Assets Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Assets Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *router) Get(c *gin.Context) {
	var args = NewAssetArgs()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	result, err := AssetAPI.GetAssets(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *router) Post(c *gin.Context) {
	asset := taggedAsset{}

	err := c.Bind(&asset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// CreatedAt and UpdatedAt set default to now
	asset.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	asset.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	if !asset.Active.Valid {
		asset.Active = models.NullInt{int64(config.Config.Models.Assets["active"]), true}
	}

	if !asset.Destination.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Missing Destination"})
		return
	}

	if asset.AssetType.Valid {
		if !r.validateEnums(int(asset.AssetType.Int), config.Config.Models.AssetType) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Missing AssetType"})
		return
	}

	if asset.Copyright.Valid {
		if !r.validateEnums(int(asset.Copyright.Int), config.Config.Models.AssetCopyright) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
			return
		}
	}

	assetID, err := AssetAPI.Insert(asset.Asset)
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

	if asset.Tags.Valid {
		err = models.TagAPI.UpdateTagging(config.Config.Models.TaggingType["asset"], int(assetID), asset.Tags.Slice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *router) Put(c *gin.Context) {
	asset := taggedAsset{}

	err := c.ShouldBindJSON(&asset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Discard CreatedAt even if there is data
	if asset.CreatedAt.Valid {
		asset.CreatedAt = models.NullTime{time.Time{}, false}
	}
	// CreatedAt and UpdatedAt set default to now
	asset.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	if !asset.UpdatedBy.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Neither updated_by or author is valid"})
		return
	}

	if asset.AssetType.Valid {
		if !r.validateEnums(int(asset.AssetType.Int), config.Config.Models.AssetType) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
			return
		}
	}

	if asset.Copyright.Valid {
		if !r.validateEnums(int(asset.Copyright.Int), config.Config.Models.AssetCopyright) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
			return
		}
	}

	err = AssetAPI.Update(asset.Asset)
	if err != nil {
		switch err.Error() {
		case "Assets Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Assets Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	if asset.Tags.Valid {
		err = models.TagAPI.UpdateTagging(config.Config.Models.TaggingType["asset"], int(asset.ID), asset.Tags.Slice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *router) validateSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|created_at|file_type)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

func (r *router) validateEnums(value int, enums map[string]int) bool {
	for _, v := range enums {
		if value == v {
			return true
			break
		}
	}
	return false
}

func (r *router) SetRoutes(router *gin.Engine) {

	postRouter := router.Group("/asset")
	{
		postRouter.GET("/count", r.Count)
		postRouter.DELETE("", r.Delete)
		postRouter.GET("", r.Get)
		postRouter.POST("", r.Post)
		postRouter.PUT("", r.Put)
	}
}

var Router router
