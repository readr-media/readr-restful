package mail

import (
	"bytes"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

func newTestServer() (r *gin.Engine) {
	gin.SetMode(gin.TestMode)
	r = gin.New()
	Router.SetRoutes(r)
	MailAPI = new(mockMailAPI)
	return r
}

type genericTestcase struct {
	name     string
	method   string
	url      string
	body     interface{}
	httpcode int
	resp     interface{}
}

func genericDoTest(r *gin.Engine, tc genericTestcase, t *testing.T, function interface{}) {
	t.Run(tc.name, func(t *testing.T) {
		w := httptest.NewRecorder()
		jsonStr := []byte{}
		if s, ok := tc.body.(string); ok {
			jsonStr = []byte(s)
		} else {
			p, err := json.Marshal(tc.body)
			if err != nil {
				t.Errorf("%s, Error when marshaling input parameters", tc.name)
			}
			jsonStr = p
		}
		req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(jsonStr))
		if tc.method == "GET" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}

		r.ServeHTTP(w, req)

		if w.Code != tc.httpcode {
			t.Errorf("%s want %d but get %d", tc.name, tc.httpcode, w.Code)
		}
		switch tc.resp.(type) {
		case string:
			if w.Body.String() != tc.resp {
				t.Errorf("%s expect (error) message %v but get %v", tc.name, tc.resp, w.Body.String())
			}
		default:
			if fn, ok := function.(func(resp string, tc genericTestcase, t *testing.T)); ok {
				fn(w.Body.String(), tc, t)
			}
		}
	})
}

type mockMailAPI struct{}

func (m *mockMailAPI) Send(args MailArgs) (err error)                                     { return nil }
func (m *mockMailAPI) SendUpdateNote(args models.GetFollowMapArgs) (err error)            { return nil }
func (m *mockMailAPI) SendUpdateNoteAllResource(args models.GetFollowMapArgs) (err error) { return nil }
func (m *mockMailAPI) GenDailyDigest() (err error)                                        { return err }
func (m *mockMailAPI) SendDailyDigest(s []string) (err error)                             { return err }
func (m *mockMailAPI) SendProjectUpdateMail(resource interface{}, resourceTyep string) (err error) {
	return err
}
func (m *mockMailAPI) SendCECommentNotify(tmp models.TaggedPostMember) (err error)   { return nil }
func (m *mockMailAPI) SendReportPublishMail(report models.ReportAuthors) (err error) { return nil }
func (m *mockMailAPI) SendMemoPublishMail(memo models.MemoDetail) (err error)        { return nil }
func (m *mockMailAPI) SendFollowProjectMail(args models.FollowArgs) (err error)      { return nil }

func TestRouteEmail(t *testing.T) {

	server := newTestServer()

	type SendEmailCaseIn struct {
		Receiver []string `json:receiver,omitempty`
		CC       []string `json:cc,omitempty`
		BCC      []string `json:bcc,omitempty`
		Subject  string   `json:"subject,omitempty"`
		Payload  string   `json:"content,omitempty"`
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) { return }

	t.Run("SendEmail", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
		/*
			genericTestcase{"SendMailOK", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			genericTestcase{"SendMailCC", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, CC: []string{"cyu2197@gmail.com"}, Subject: "CCTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			genericTestcase{"SendMailBCC", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"cyu2197@gmail.com"}, BCC: []string{"yychen@mirrormedia.mg"}, Subject: "BCCTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			genericTestcase{"SendMail2RecvOK", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			genericTestcase{"SendMailNoRecv", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
		*/
		} {
			genericDoTest(server, testcase, t, asserter)
		}
	})
	t.Run("SendUpdateNoteMail", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
		/*
			genericTestcase{"SendUNMailOK", "POST", "/mail/updatenote", `{"resource":"post","updated_after":"` + time.Now().AddDate(-1, 0, 0).Format(time.RFC3339) + `"}`, http.StatusOK, ``},
			genericTestcase{"SendUNMailJSTimeFormat", "POST", "/mail/updatenote", `{"resource":"blah","updated_after":"2015-03-06T03:01:39.385Z"}`, http.StatusBadRequest, `{"Error":"Invalid Resource"}`},
			genericTestcase{"SendUNMailAll", "POST", "/mail/updatenote", `{"updated_after":"` + time.Now().AddDate(-1, 0, 0).Format(time.RFC3339) + `"}`, http.StatusOK, ``},
		*/
		} {
			genericDoTest(server, testcase, t, asserter)
		}
	})
}
