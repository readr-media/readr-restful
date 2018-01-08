package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

type miscHandler struct {
	Dialer gomail.Dialer
}

func queryArrayParser(i string) (result []string, err error) {
	err = json.Unmarshal([]byte(i), &result)
	if err != nil {
		return []string{}, err
	}
	return result, nil
}

func (r *miscHandler) sendMail(c *gin.Context) {

	input := struct {
		Receiver []string `json:"receiver"`
		CC       []string `json:"cc"`
		BCC      []string `json:"bcc"`
		Subject  string   `json:"subject"`
		Payload  string   `json:"content"`
	}{}
	c.Bind(&input)

	if len(input.Receiver) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Receiver"})
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", r.Dialer.Username)
	m.SetHeader("To", input.Receiver...)
	m.SetHeader("Cc", input.CC...)
	m.SetHeader("Bcc", input.BCC...)
	m.SetHeader("Subject", input.Subject)
	m.SetBody("text/html", input.Payload)

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
