package http

import (
	"fmt"
	"net/http"
	"strings"
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
	Status     int                  `form:"status"`
	LastPaidAt map[string]time.Time `form:"last_paid_at"`
}

func (r *ListRequest) Select() (query string, values []interface{}, err error) {
	fields := []string{"subscriptions.*"}

	// ws = where string, wv = where values
	ws := make([]string, 0)
	wv := make([]interface{}, 0)
	var where string
	// r.Status
	ws = append(ws, fmt.Sprintf("%s %s (?)", "subscriptions.status", "="))
	wv = append(wv, r.Status)
	// r.LastPaidAt
	if r.LastPaidAt != nil {
		fmt.Printf("LastPayAt:%v\n", r.LastPaidAt)
		for o, v := range r.LastPaidAt {
			ops, err := rrsql.OperatorCoverter(o)
			if err != nil {
				return "", nil, err
			}
			ws = append(ws, fmt.Sprintf("%s %s (?)", "subscriptions.last_paid_at", ops))
			wv = append(wv, v)
		}
	}
	if len(ws) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(ws, " AND "))
	} else if len(ws) == 0 {
		where = ""
	}
	query = fmt.Sprintf(`
	SELECT %s FROM subscriptions %s `,
		strings.Join(fields, ","),
		where,
	)
	values = append(values, wv...)
	return query, values, nil
}

// NewListRequest accepts option to modify the initial content of ListRequest, and return a pointer of ListRequest
func NewListRequest(options ...func(*ListRequest)) *ListRequest {
	var params ListRequest

	for _, f := range options {
		f(&params)
	}
	return &params
}

// Get handles the list-all request
func (h *Handler) Get(c *gin.Context) {

	var err error
	params := NewListRequest()
	if err = c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	fmt.Printf("ListArgs:\n%v\n", params)

	// Formatting JSON here
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

func payInterval(t time.Time) (start time.Time, end time.Time, err error) {
	truncatedT := t.UTC().Truncate(24 * time.Hour)
	year, month, day := truncatedT.Date()
	fmt.Println(year, month, day)
	switch month {
	case time.May, time.July, time.October, time.December:
		if day == 30 {
			return time.Time{}, time.Time{}, nil
		} else if day == 31 {
			return truncatedT.AddDate(0, -1, -1), truncatedT.AddDate(0, -1, 0), nil
		}
	case time.April, time.June, time.September, time.November:
		if day == 30 {
			return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 2), nil
		}
	case time.February:
		if day == 28 {
			return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 4), nil
		}
	case time.March:
		if day > 28 && day < 31 {
			return time.Time{}, time.Time{}, nil
		} else if day == 31 {
			return truncatedT.AddDate(0, -1, -3), truncatedT.AddDate(0, -1, -2), nil
		}
	default:
		return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 1), nil
	}
	return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 1), nil
}

// RecurringPay handles today's interval, get the subscription info list, and Pay with RoutinePay().
func (h *Handler) RecurringPay(c *gin.Context) {

	// Get subscription list on today's interval
	start, end, _ := payInterval(time.Now())

	params := NewListRequest(func(p *ListRequest) {
		p.LastPaidAt = map[string]time.Time{
			"gte": start,
			"lt":  end,
		}
		p.Status = 1
	})

	list, err := h.Service.GetSubscriptions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	err = h.Service.RoutinePay(list)

	// Return 201
	c.Status(http.StatusCreated)
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

		subscriptionRouter.POST("/recurring", h.RecurringPay)
	}
}

// Router is the instances for routing sets
var Router Handler
