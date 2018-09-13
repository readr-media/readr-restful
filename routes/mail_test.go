package routes

import (
	//"log"
	"testing"
	//"time"

	"net/http"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"gopkg.in/gomail.v2"
)

func InitMailDialer() gomail.Dialer {
	// dialer := gomail.NewDialer(
	// 	viper.Get("mail.host").(string),
	// 	int(viper.Get("mail.port").(float64)),
	// 	viper.Get("mail.user").(string),
	// 	viper.Get("mail.password").(string),
	// )
	dialer := gomail.NewDialer(
		config.Config.Mail.Host,
		config.Config.Mail.Port,
		config.Config.Mail.User,
		config.Config.Mail.Password,
	)
	return *dialer
}

type mockMailAPI struct{}

func (m *mockMailAPI) SetDialer(dialer gomail.Dialer)                                     { return }
func (m *mockMailAPI) Send(args models.MailArgs) (err error)                              { return nil }
func (m *mockMailAPI) SendUpdateNote(args models.GetFollowMapArgs) (err error)            { return nil }
func (m *mockMailAPI) SendUpdateNoteAllResource(args models.GetFollowMapArgs) (err error) { return nil }
func (m *mockMailAPI) GenDailyDigest() (err error)                                        { return err }
func (m *mockMailAPI) SendDailyDigest(s []string) (err error)                             { return err }
func (m *mockMailAPI) SendProjectUpdateMail(resource interface{}, resourceTyep string) (err error) {
	return err
}
func (m *mockMailAPI) SendCECommentNotify(tmp models.TaggedPostMember) (err error) { return nil }

func TestRouteEmail(t *testing.T) {

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
			genericTestcase{"SendMailOK", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			//genericTestcase{"SendMailCC", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, CC: []string{"cyu2197@gmail.com"}, Subject: "CCTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			//genericTestcase{"SendMailBCC", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"cyu2197@gmail.com"}, BCC: []string{"yychen@mirrormedia.mg"}, Subject: "BCCTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			//genericTestcase{"SendMail2RecvOK", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
			//genericTestcase{"SendMailNoRecv", "POST", "/mail", SendEmailCaseIn{Receiver: []string{"yychen@mirrormedia.mg"}, Subject: "RecvTestSuccess", Payload: "<b>HTML</b> payload"}, http.StatusOK, ``},
		} {
			genericDoTest(testcase, t, asserter)
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
			genericDoTest(testcase, t, asserter)
		}
	})
}
