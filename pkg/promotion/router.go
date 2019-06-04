package promotion

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type promotionHandler struct{}

// bind parses query parameters from gin.Context, and save them in params
func bind(c *gin.Context, params *ListParams) (err error) {

	if err = c.ShouldBindQuery(params); err != nil {
		return err
	}
	// Validate query paramters
	if err := params.validate(); err != nil {
		return err
	}
	return nil
}

func (r *promotionHandler) List(c *gin.Context) {

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
	// fmt.Println(params)
	promos, err := DataAPI.Get(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": promos})
}

// func (r *promotionHandler) Get(c *gin.Context) {
// }

func (r *promotionHandler) Post(c *gin.Context) {

	var promo = Promotion{}

	if err := c.Bind(&promo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// Validate Promotion data
	// promo != empty, title != "", created_at = now, updated_at = now
	if err := (&promo).validate(
		validateNullBody,
		validateTitle,
		setCreatedAtNow,
		setUpdatedAtNow,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// Insert to db
	_, err := DataAPI.Insert(promo)
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

func (r *promotionHandler) Put(c *gin.Context) {

	var promo = Promotion{}

	if err := c.ShouldBindJSON(&promo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// Validate Promotion data
	// promo != empty, id != 0, updated_at = now
	if err := (&promo).validate(
		validateNullBody,
		validateID,
		setUpdatedAtNow,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// fmt.Printf("Promotion Put:%v\n", promo)
	err := DataAPI.Update(promo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	// return 204
	c.Status(http.StatusNoContent)
}

func (r *promotionHandler) Delete(c *gin.Context) {

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	err := DataAPI.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
		return
	}
	c.Status(http.StatusOK)
}

// SetRoutes is an exported API to register API routers in routes/routes.go
func (r *promotionHandler) SetRoutes(router *gin.Engine) {

	promotionRouter := router.Group("/promotions")
	{
		promotionRouter.GET("", r.List)
		// promotionRouter.GET("/:id", r.Get)
		promotionRouter.POST("", r.Post)
		promotionRouter.PUT("", r.Put)
		promotionRouter.DELETE("/:id", r.Delete)
	}
}

// Router is the interface for router layer
var Router promotionHandler
