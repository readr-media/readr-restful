package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

type miscHandler struct {
	Dialer gomail.Dialer
}

func (r *miscHandler) sendMail(c *gin.Context) {

	input := struct {
		Receiver []string `json:"receiver"`
		Subject  string   `json:"subject"`
		Payload  string   `json:"content"`
	}{}
	c.Bind(&input)

	if len(input.Receiver) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Receiver"})
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "yychen@mirrormedia.mg")
	m.SetHeader("To", input.Receiver...)
	m.SetHeader("Subject", input.Subject)
	m.SetBody("text/html", input.Payload)

	fmt.Println(input.Payload)

	// Send the email to Bob, Cora and Dan.
	if err := r.Dialer.DialAndSend(m); err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(200)
}

func (r *miscHandler) SetRoutes(router *gin.Engine, dialer gomail.Dialer) {
	router.POST("/mail", r.sendMail)
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.Dialer = dialer
}

var MiscHandler miscHandler
