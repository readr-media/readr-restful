package routes

import (
	"log"
	"regexp"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"gopkg.in/gomail.v2"
)

type mailHandler struct{}

func (r *mailHandler) sendMail(c *gin.Context) {

	input := models.MailArgs{}
	c.Bind(&input)

	/*  Send to all user if receiver not specified
	if len(input.Receiver) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Receiver"})
		return
	}
	*/
	if err := models.MailAPI.Send(input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (r *mailHandler) updateNote(c *gin.Context) {

	args := models.GetFollowMapArgs{}
	err := c.Bind(&args)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	if ok, err := regexp.MatchString("(post|project|member)", args.ResourceName); args.ResourceName != "" && (err != nil || !ok) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Resource"})
		return
	}
	if args.ResourceName == "" {
		if err := models.MailAPI.SendUpdateNoteAllResource(args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	} else {
		if err := models.MailAPI.SendUpdateNote(args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *mailHandler) GenDailyDigest(c *gin.Context) {

	err := models.MailAPI.GenDailyDigest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
	//c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(md))
}

func (r *mailHandler) SendDailyDigest(c *gin.Context) {
	receiver := []string{}
	if c.Query("receiver") != "" {
		if err := json.Unmarshal([]byte(c.Query("receiver")), &receiver); err != nil {
			log.Println(err.Error())
			return
		}
	}
	err := models.MailAPI.SendDailyDigest(receiver)
	if err != nil {
		if err.Error() == "Not Found" {
			c.JSON(http.StatusNotFound, gin.H{"Error": "DailyDigest File Not Found, Please Generate a File First."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}
	c.Status(http.StatusOK)
}

func (r *mailHandler) SetRoutes(router *gin.Engine) {
	router.POST("/mail", r.sendMail)
	router.POST("/mail/updatenote", r.updateNote)
	router.POST("/mail/gendailydigest", r.GenDailyDigest)
	router.POST("/mail/senddailydigest", r.SendDailyDigest)

	// init mail sender
	dialer := gomail.NewDialer(
		config.Config.Mail.Host,
		config.Config.Mail.Port,
		config.Config.Mail.User,
		config.Config.Mail.Password,
	)
	models.MailAPI.SetDialer(*dialer)
}

var MailHandler mailHandler
