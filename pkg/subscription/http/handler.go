package http

import (
	"log"
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

// Get handles the list-all request
func (h *Handler) Get(c *gin.Context) {

	var err error
	params := NewListRequest()
	if err = params.bind(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Formatting JSON
	var results struct {
		Items []subscription.Subscription `json:"_itmes"`
	}
	results.Items, err = h.Service.GetSubscriptions(params)
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
	// return 201
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

// RecurringPay handles today's interval, get the subscription info list, and Pay with RoutinePay().
func (h *Handler) RecurringPay(c *gin.Context) {

	// Get subscription list on today's interval
	start, end, _ := payInterval(time.Now())
	params := NewListRequest(func(p *ListRequest) {
		p.LastPaidAt = map[string]time.Time{
			"$gte": start,
			"$lt":  end,
		}
		p.Status = subscription.StatusOK
	})

	list, err := h.Service.GetSubscriptions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	err = h.Service.RoutinePay(list)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// Return 202
	c.Status(http.StatusAccepted)
}

// SetRoutes provides a public function to set gin router
func (h *Handler) SetRoutes(router *gin.Engine) {

	// Create a mySQL subscription service, and make handlers use it to process subscriptions if there is no other service
	if h.Service == nil {
		// Set default service to MySQL
		log.Println("Set subscription service to default MySQL")

		s := &mysql.SubscriptionService{DB: rrsql.DB.DB}
		h.Service = s
	}
	// Register subscriptions endpoints
	subscriptionRouter := router.Group("/subscriptions")
	{
		// subscriptionRouter.GET("", h.Get)
		subscriptionRouter.POST("", h.Post)
		subscriptionRouter.PUT("/:id", h.Put)

		subscriptionRouter.POST("/recurring", h.RecurringPay)
	}
}

// Router is the instances for routing sets
var Router Handler
