package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/pkg/subscription"
	"github.com/readr-media/readr-restful/pkg/subscription/mysql"
)

// Handler is the object with routing methods and database service
type Handler struct {
	Service subscription.Subscriber
}

type ListRequest struct {
	Status     int       `form:"status"`
	LastPaidAt time.Time `form:"last_paid_at" time_format:"2006-01-02"`
}

// Get handles the list-all request
func (h *Handler) Get(c *gin.Context) {

	var err error
	var req = ListRequest{}
	if err = c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	fmt.Printf("ListArgs:\n%v\n", req)

	// Formatting JSON here
	var results struct {
		Items []subscription.Subscription `json:"_itmes"`
	}
	results.Items, err = h.Service.GetSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// Post handles create requests
func (h *Handler) Post(c *gin.Context) {

	var sub = subscription.Subscription{}

	if err := c.BindJSON(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	err := h.Service.CreateSubscription(sub)
	if err != nil {
		switch err.Error() {
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}
	// Should return 201
	c.Status(http.StatusCreated)
}

// Put handles update requests
func (h *Handler) Put(c *gin.Context) {
	var sub = subscription.Subscription{}
	if err := c.BindJSON(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	err := h.Service.UpdateSubscriptions(sub)
	if err != nil {
		switch err.Error() {
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}
	// return 204
	c.Status(http.StatusNoContent)
}

// SetRoutes provides a public function to set gin router
func (h *Handler) SetRoutes(router *gin.Engine) {

	// Create a mySQL subscription service, and make handlers use it to process subscriptions
	s := &mysql.SubscriptionService{DB: rrsql.DB.DB}
	h.Service = s
	// Register subscriptions endpoints
	subscriptionRouter := router.Group("/subscriptions")
	{
		subscriptionRouter.GET("", h.Get)
		subscriptionRouter.POST("", h.Post)
		subscriptionRouter.PUT("/:id", h.Put)
	}
}

// Router is the instances for routing sets
var Router Handler
