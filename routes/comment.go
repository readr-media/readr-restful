package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type commentsHandler struct{}

func (r *commentsHandler) Post(c *gin.Context) {

	payload := models.CommentEvent{}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	log.Println(payload)
	models.CommentHandler.CreateNotifications(payload)
	if err != nil {
		switch err.Error() {
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *commentsHandler) SetRoutes(router *gin.Engine) {
	commentsRouter := router.Group("/comments")
	{
		commentsRouter.POST("", r.Post)
	}
}

var CommentsHandler commentsHandler
