package http

import (
	"encoding/json"
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

type PaymentResponse struct {
	RecTradeId            string `json:"rec_trade_id"`
	AuthCode              string `json:"auth_code"`
	BankTransactionId     string `json:"bank_transaction_id"`
	OrderNumber           string `json:"order_number"`
	Amount                int    `json:"amount"`
	Status                int    `json:"status"`
	Msg                   string `json:"msg"`
	TransactionTimeMillis int64  `json:"transaction_time_millis"`
	//PayInfo struct{} //JSONObject	若此交易使用 LINE Pay, JKOPAY, 悠遊付 時回傳訊息 暫時不實作
	//e_invoice_carrier //JSONObject 電子發票載具資料 目前僅支援 : JKOPAY 暫時不實作
	Acquirer              string `json:"acquirer"`
	CardIdentifier        string `json:"card_identifier"`
	BankResultCode        string `json:"bank_result_code"`
	BankResultMsg         string `json:"bank_result_msg"`
	MerchantReferenceInfo struct {
		AffiliateCodes []interface{} `json:"affiliate_codes"`
	}
	// 暫時不實作: instalment_info, redeem_info, merchandise_details
	EventCode string `json:"event_code"` //與銀行或錢包合作之活動中，雙方協議的指定活動代碼。支援 : 悠遊付
}

func (h *Handler) Notify(c *gin.Context) {
	var p PaymentResponse
	bytes, err := c.GetRawData()
	if err != nil {
		c.JSON(500, gin.H{"msg": err.Error()})
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &p)

	// Not completed, should insert to database
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
		subscriptionRouter.POST("/notify", h.Notify)
	}
}

// Router is the instances for routing sets
var Router Handler
