package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/pkg/promotion"
	"github.com/readr-media/readr-restful/pkg/promotion/mysql"
)

// Handler comprises the controller function in promotion package
type Handler struct{}

// bind parses query parameters from gin.Context, and save them in params
func bind(c *gin.Context, params *ListParams) (err error) {

	if err = c.ShouldBindQuery(params); err != nil {
		return err
	}
	// Validate query paramters
	if err := params.validate(); err != nil {
		if err.Error() == "invalid sort" {
			log.Printf("warning: binding invalid sort:%s, using default: -created_at\n", params.Sort)
			params.Sort = "-created_at"
		}
	}
	return nil
}

func (h *Handler) List(c *gin.Context) {

	// Get a default filter pointer struct
	// max_result = 15, page = 1, sort = "created_at
	params, err := NewListParams(
		func(p *ListParams) error {
			p.MaxResult = 15
			p.Page = 1
			p.Sort = "-created_at"

			return nil
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Initializing Default Query Parameters"})
		return
	}
	// Bind query parameters
	if err = bind(c, params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	promos, err := mysql.DataAPI.Get(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": promos})
}

// func (h *Handler) Get(c *gin.Context) {
// }

func (h *Handler) Post(c *gin.Context) {

	var promo = promotion.Promotion{}

	if err := c.Bind(&promo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// promo != empty, title != "", created_at = now, updated_at = now
	if err := (&promo).Validate(
		promotion.ValidateNullBody,
		promotion.ValidateTitle,
		promotion.SetCreatedAtNow,
		promotion.SetUpdatedAtNow,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Insert to db
	_, err := mysql.DataAPI.Insert(promo)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Promotion ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	// return 201
	c.Status(http.StatusCreated)
}

func (h *Handler) Put(c *gin.Context) {

	var promo = promotion.Promotion{}

	if err := c.ShouldBindJSON(&promo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// Validate Promotion data
	// promo != empty, id != 0, updated_at = now
	if err := (&promo).Validate(
		promotion.ValidateNullBody,
		promotion.ValidateID,
		promotion.SetUpdatedAtNow,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// fmt.Printf("Promotion Put:%v\n", promo)
	err := mysql.DataAPI.Update(promo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// return 204
	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		// log.Printf("unable to parse id:%s\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse id:%s", c.Param("id"))})
		return
	}

	err = mysql.DataAPI.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// SetRoutes is an exported API to register API routers in routes/routes.go
func (h *Handler) SetRoutes(router *gin.Engine) {

	promotionRouter := router.Group("/promotions")
	{
		promotionRouter.GET("", h.List)
		// promotionRouter.GET("/:id", h.Get)
		promotionRouter.POST("", h.Post)
		promotionRouter.PUT("", h.Put)
		promotionRouter.DELETE("/:id", h.Delete)
	}
}

// Router is the interface for router layer
var Router Handler
