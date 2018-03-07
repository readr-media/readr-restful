package routes

import (
	"log"
	"regexp"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
	"gopkg.in/gomail.v2"
)

type mailHandler struct{}

func (r *mailHandler) sendMail(c *gin.Context) {

	input := models.MailArgs{}
	c.Bind(&input)

	if len(input.Receiver) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Receiver"})
		return
	}

	if err := models.MailAPI.Send(input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
	}

	c.Status(200)
}

func (r *mailHandler) updateNote(c *gin.Context) {

	args := models.GetFollowMapArgs{}
	err := c.Bind(&args)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	if ok, err := regexp.MatchString("(post|project|member)", args.Resource); args.Resource != "" && (err != nil || !ok) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Resource"})
		return
	}
	if args.Resource == "" {
		if err := models.MailAPI.SendUpdateNoteAll(args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	} else {
		if err := models.MailAPI.SendUpdateNote(args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(200)
}

func (r *mailHandler) SetRoutes(router *gin.Engine, dialer gomail.Dialer) {
	router.POST("/mail", r.sendMail)
	router.POST("/mail/updatenote", r.updateNote)

	models.MailAPI.SetDialer(dialer)
}

var MailHandler mailHandler
