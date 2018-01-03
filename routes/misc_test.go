package routes

import (
	"bytes"
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteSendEmail(t *testing.T) {

	type SendEmailCaseIn struct {
		Receiver []string `receiver:id,omitempty`
		Subject  string   `json:"subject,omitempty"`
		Payload  string   `json:"content,omitempty"`
	}

	var TestRouteRegisterCases = []struct {
		name     string
		in       SendEmailCaseIn
		httpcode int
	}{
		{"SendMailOK", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "TestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK},
		{"SendMail2RecvOK", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "TestSuccess", Payload: "<b>HTML</b> 2 recvs"}, http.StatusOK},
		{"SendMailNoRecv", SendEmailCaseIn{Receiver: []string{}, Subject: "TestFail", Payload: "<b>HTML</b> payload"}, http.StatusBadRequest},
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
