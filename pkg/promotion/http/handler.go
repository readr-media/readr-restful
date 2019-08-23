package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	rt "github.com/readr-media/readr-restful/internal/router"
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
	if c.Query("active") != "" {
		actives := strings.Split(c.Query("active"), "::")
		if len(actives) > 0 {
			for _, statement := range actives {
				splitStatement := strings.Split(statement, ":")
				if len(splitStatement) != 2 {
					continue
				}
				operator := splitStatement[0]
				splitValues := strings.Split(splitStatement[1], ",")
				var values []int
				for _, v := range splitValues {
					value, err := strconv.ParseInt(v, 10, 64)
					if err != nil {
						continue
					}
					// If value is not valid active, skip it
					if isValidActive(int(value)) {
						values = append(values, int(value))
					}
				}
				params.Active[operator] = values
			}
		}
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

// List is the controller returning promotion list to clients
func (h *Handler) List(c *gin.Context) {

	// Get a default filter pointer struct
	// max_result = 15, page = 1, sort = "created_at
	params, err := NewListParams(
		func(p *ListParams) error {
			p.MaxResult = 15
			p.Page = 1
			p.Sort = "-created_at"

			p.Active = make(map[string][]int)
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
	var results struct {
		Items []promotion.Promotion `json:"_items"`
		Meta  *rt.ResponseMeta      `json:"_meta,omitempty"`
	}
	results.Items, err = mysql.DataAPI.Get(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	if params.Total {
		totalPromotions, err := mysql.DataAPI.Count(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		var promotionMeta = rt.ResponseMeta{
			Total: &totalPromotions,
		}
		results.Meta = &promotionMeta
	}
	c.JSON(http.StatusOK, results)
}

// func (h *Handler) Get(c *gin.Context) {
// }

// Post is the controller handling POST method
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

// Put is the controller handling UPDATE with PUT method
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

// Delete handles DELETE method by setting active to 0
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
