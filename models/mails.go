package models

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"gopkg.in/gomail.v2"
)

type MailArgs struct {
	Receiver []string `json:"receiver"`
	CC       []string `json:"cc"`
	BCC      []string `json:"bcc"`
	Subject  string   `json:"subject"`
	Payload  string   `json:"content"`
}

type MailInterface interface {
	SetDialer(dialer gomail.Dialer)
	Send(args MailArgs) (err error)
	SendUpdateNote(args GetFollowMapArgs) (err error)
	SendUpdateNoteAll(args GetFollowMapArgs) (err error)
}

type mailApi struct {
	Dialer gomail.Dialer
}

func (m *mailApi) SetDialer(dialer gomail.Dialer) {
	m.Dialer = dialer
}

func (m *mailApi) Send(args MailArgs) (err error) {

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.Dialer.Username)
	msg.SetHeader("To", args.Receiver...)
	msg.SetHeader("Cc", args.CC...)
	msg.SetHeader("Bcc", args.BCC...)
	msg.SetHeader("Subject", args.Subject)
	msg.SetBody("text/html", args.Payload)

	if err = m.Dialer.DialAndSend(msg); err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func (m *mailApi) SendUpdateNote(args GetFollowMapArgs) (err error) {

	followingMap, err := FollowingAPI.GetFollowMap(args)
	if err != nil {
		return err
	}

	var (
		follower_index = make(map[string][]string)
		follower_info  = make(map[string]Member)
		followers_list []string
	)

	for _, m := range followingMap {
		for _, v := range m.Followers {
			followers_list = append(followers_list, v)
			follower_index[v] = m.ResourceIDs
		}
	}

	members, err := MemberAPI.GetMembers(&MemberArgs{IDs: followers_list, Sorting: "member_id", MaxResult: uint8(len(followers_list)), Page: 1})
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, v := range members {
		follower_info[v.ID] = v
	}

	for _, m := range followingMap {
		var mails []string
		for _, f := range m.Followers {
			mails = append(mails, follower_info[f].Mail.String)
		}
		/*m.Send(MailArgs{
			Receiver: mails,
			Subject:  fmt.Sprintf("更新通知：%s %s", args.Resource, args.Type),
			Payload:  fmt.Sprintf("%s %s 更新項目：%s", args.Resource, args.Type, strings.Join(m.ResourceIDs, ",")),
		})*/
	}
	return nil
}

func (m *mailApi) SendUpdateNoteAll(args GetFollowMapArgs) (err error) {

	var (
		follower_reverse_index = make(map[string][]string)
		follower_index         = make(map[string]map[string][]string)
		follower_group_index   = make(map[string]map[string][]string)
		follower_info          = make(map[string]Member)
		//resource_list          = make(map[string][]string)
		followers_list []string
	)
	for _, t := range []string{"member", "post", "project"} {
		followingMap, err := FollowingAPI.GetFollowMap(GetFollowMapArgs{Resource: t, UpdateAfter: args.UpdateAfter})
		if err != nil {
			return err
		}
		for _, m := range followingMap {
			for _, v := range m.Followers {
				followers_list = append(followers_list, v)
				if follower_index[v] == nil {
					follower_index[v] = map[string][]string{}
				}
				follower_index[v][t] = m.ResourceIDs
			}
		}
	}

	members, err := MemberAPI.GetMembers(&MemberArgs{IDs: followers_list, Sorting: "member_id", MaxResult: uint8(len(followers_list)), Page: 1})
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, v := range members {
		follower_info[v.ID] = v
	}

	for follower, followings := range follower_index {
		var keyBuf bytes.Buffer
		for k, v := range followings {
			keyBuf.WriteString(k)
			keyBuf.WriteString(":")
			keyBuf.WriteString(strings.Join(v, ","))
			keyBuf.WriteString("\n")
		}
		key := keyBuf.String()
		follower_reverse_index[key] = append(follower_reverse_index[key], follower)
	}

	for k, followers := range follower_reverse_index {
		follower_group_index[k] = follower_index[followers[0]]
		m.Send(MailArgs{
			Receiver: followers,
			Subject:  fmt.Sprintf("[Readr]更新通知"),
			Payload:  fmt.Sprintf("更新項目：%s", k),
		})
	}

	return nil
}

func mailGenerator(resource string, items interface{}) string {
	return fmt.Sprintf("Updated %s: %s", resource, fmt.Sprint(resource))
}

var MailAPI MailInterface = new(mailApi)
