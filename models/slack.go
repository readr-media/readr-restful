package models

import (
	"bytes"
	"log"

	"net/http"
	"text/template"

	"github.com/readr-media/readr-restful/config"
)

type slackHelper struct{}

var SlackHelper = slackHelper{}

func (s *slackHelper) SendSlackMsg(msg []byte, hook string) error {
	client := &http.Client{}

	req, err := http.NewRequest("POST", hook, bytes.NewReader(msg))
	if err != nil {
		return err
	}

	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	log.Println("resp", resp)
	log.Println("err", err)

	if err != nil {
		return err
	}
	return nil
}

func (s *slackHelper) SendCECommentNotify(tpm TaggedPostMember) error {
	t, err := template.New("CEComment_slack_notify").Parse(`
	{
		"text": "客座總編 {{.Member.Nickname.String}} 已發布一篇新評論",
		"attachments":[
			{
				"title": "{{.Post.Title.String}}",
				"text": "{{.Post.Content.String}}"
			},
			{
				"title": "新聞連結",
				"text": "{{.Post.Link.String}}"
			}
		]
	}
	`)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, tpm)
	if err != nil {
		return err
	}

	msg := buf.Bytes()

	log.Println(msg, config.Config.Slack.NotifyWebhook)
	s.SendSlackMsg(msg, config.Config.Slack.NotifyWebhook)

	return nil
}
