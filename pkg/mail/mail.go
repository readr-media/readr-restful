package mail

import (
	"log"
	"regexp"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type router struct{}

func (r *router) sendMail(c *gin.Context) {

	input := MailArgs{}
	c.Bind(&input)

	/*  Send to all user if receiver not specified
	if len(input.Receiver) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Receiver"})
		return
	}
	*/
	if err := MailAPI.Send(input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (r *router) updateNote(c *gin.Context) {

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
		if err := MailAPI.SendUpdateNoteAllResource(args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	} else {
		if err := MailAPI.SendUpdateNote(args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *router) GenDailyDigest(c *gin.Context) {

	err := MailAPI.GenDailyDigest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
	//c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(md))
}

func (r *router) SendDailyDigest(c *gin.Context) {
	receiver := []string{}
	if c.Query("receiver") != "" {
		if err := json.Unmarshal([]byte(c.Query("receiver")), &receiver); err != nil {
			log.Println(err.Error())
			return
		}
	}
	err := MailAPI.SendDailyDigest(receiver)
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

func (r *router) SetRoutes(router *gin.Engine) {
	router.POST("/mail", r.sendMail)
	// router.POST("/mail/updatenote", r.updateNote)
	router.POST("/mail/gendailydigest", r.GenDailyDigest)
	router.POST("/mail/senddailydigest", r.SendDailyDigest)
}

var Router router
