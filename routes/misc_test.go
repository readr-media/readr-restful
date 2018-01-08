package routes

import (
	"bytes"
	"log"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func initMailDialer() gomail.Dialer {
	viper.AddConfigPath("../config")
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	dialer := gomail.NewDialer(
		viper.Get("mail.host").(string),
		int(viper.Get("mail.port").(float64)),
		viper.Get("mail.user").(string),
		viper.Get("mail.password").(string),
	)

	return *dialer
}

func TestRouteSendEmail(t *testing.T) {

	type SendEmailCaseIn struct {
		Receiver []string `json:receiver,omitempty`
		CC       []string `json:cc,omitempty`
		BCC      []string `json:bcc,omitempty`
		Subject  string   `json:"subject,omitempty"`
		Payload  string   `json:"content,omitempty"`
	}

	var TestRouteRegisterCases = []struct {
		name     string
		in       SendEmailCaseIn
		httpcode int
	}{
		{"SendMailOK", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK},
		//{"SendMailCC", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, CC: []string{"cyu2197@gmail.com"}, Subject: "CCTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK},
		//{"SendMailBCC", SendEmailCaseIn{Receiver: []string{"cyu2197@gmail.com"}, BCC: []string{"yychen@mirrormedia.mg"}, Subject: "BCCTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK},
		//{"SendMail2RecvOK", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "TestSuccess", Payload: "<b>HTML</b> 2 recvs"}, http.StatusOK},
		//{"SendMailNoRecv", SendEmailCaseIn{Receiver: []string{}, Subject: "TestFail", Payload: "<b>HTML</b> payload"}, http.StatusBadRequest},
	}

	for _, testcase := range TestRouteRegisterCases {
		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest("POST", "/mail", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.httpcode, w.Code, testcase.name)
			t.Fail()
		}
	}
}
