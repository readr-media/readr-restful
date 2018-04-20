package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type notificationHandler struct{}

func (r *notificationHandler) Read(c *gin.Context) {

	payload := models.UpdateNotificationArgs{}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if payload.IDs == nil || payload.MemberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}

	models.CommentHandler.ReadNotifications(payload)
	c.Status(http.StatusOK)
}

func (r *notificationHandler) SetRoutes(router *gin.Engine) {
	notificationRouter := router.Group("/notify")
	{
		notificationRouter.PUT("/read", r.Read)
	}
}

var NotificationHandler notificationHandler
