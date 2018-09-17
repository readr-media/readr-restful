package models

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"text/template"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"gopkg.in/gomail.v2"
)

type MailArgs struct {
	Receiver []string `json:"receiver"`
	CC       []string `json:"cc"`
	BCC      []string `json:"bcc"`
	Subject  string   `json:"subject"`
	Payload  string   `json:"content"`
	Type     string   `json:"type"`
}

type MailInterface interface {
	SetDialer(dialer gomail.Dialer)
	Send(args MailArgs) (err error)
	SendUpdateNote(args GetFollowMapArgs) (err error)
	SendUpdateNoteAllResource(args GetFollowMapArgs) (err error)
	GenDailyDigest() (err error)
	SendDailyDigest(receiver []string) (err error)
	SendProjectUpdateMail(resource interface{}, resourceTyep string) (err error)
	SendCECommentNotify(tmp TaggedPostMember) (err error)
}

type mailApi struct {
	Dialer gomail.Dialer
}

func (m *mailApi) SetDialer(dialer gomail.Dialer) {
	m.Dialer = dialer
}

func (m *mailApi) Send(args MailArgs) (err error) {
	receivers := make([]mailReceiver, 0)
	if len(args.Receiver) > 0 {
		receivers, err = m.getMailingList(args.Receiver, args.Type)
		if err != nil {
			fmt.Println("Error get mail list when send mail: ", err)
			return err
		}
	}

	var filteredReceivers []string
	for _, receiver := range receivers {
		filteredReceivers = append(filteredReceivers, receiver.Mail)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", msg.FormatAddress(config.Config.Mail.User, config.Config.Mail.UserName))
	msg.SetHeader("To", filteredReceivers...)
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

	mapInterface, err := FollowingAPI.Get(&args)
	if err != nil {
		return err
	}
	followingMap, ok := mapInterface.([]FollowingMapItem)
	if !ok {
		log.Println("Error assert mapInterface @ SendIpdateNote ")
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

	if len(followers_list) == 0 {
		return nil
	}

	members, err := MemberAPI.GetMembers(&MemberArgs{IDs: followers_list, Sorting: "member_id", MaxResult: uint8(len(followers_list)), Page: 1})
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, v := range members {
		follower_info[v.MemberID] = v
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

func (m *mailApi) SendUpdateNoteAllResource(args GetFollowMapArgs) (err error) {

	var (
		follower_reverse_index = make(map[string][]string)
		follower_index         = make(map[string]map[string][]string)
		follower_group_index   = make(map[string]map[string][]string)
		follower_info          = make(map[string]Member)
		//resource_list          = make(map[string][]string)
		followers_list []string
	)
	for _, t := range []string{"member", "post", "project"} {
		mapInterface, err := FollowingAPI.Get(&GetFollowMapArgs{Resource: Resource{ResourceName: t}, UpdateAfter: args.UpdateAfter})
		if err != nil {
			return err
		}
		followingMap, ok := mapInterface.([]FollowingMapItem)
		if !ok {
			log.Println("Error assert mapInterface @ SendIpdateNote ")
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

	if len(followers_list) == 0 {
		return nil
	}

	members, err := MemberAPI.GetMembers(&MemberArgs{IDs: followers_list, Sorting: "member_id", MaxResult: uint8(len(followers_list)), Page: 1})
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, v := range members {
		follower_info[v.MemberID] = v
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

type dailyDigest struct {
	DateDay      int
	DateMonth    int
	DateYear     int
	Reports      []dailyReport
	Memos        []dailyMemoAuthor
	Posts        []dailyPostAuthor
	ReadrPosts   []dailyPost
	HasReport    bool
	HasMemo      bool
	HasPost      bool
	HasReadrPost bool
	SubLink      string
}
type dailyReport struct {
	ID          int    `db:"id"`
	Title       string `db:"title"`
	Slug        string `db:"slug"`
	Image       string `db:"hero_image"`
	Description string `db:"description"`
}
type dailyMemoAuthor struct {
	ID    int    `db:"id"`
	Name  string `db:"author"`
	Image string `db:"image"`
	Memos []dailyMemo
}
type dailyMemo struct {
	ID         int    `db:"id"`
	Title      string `db:"title"`
	Content    string `db:"content"`
	Slug       string `db:"slug"`
	AuthorID   int    `db:"author_id"`
	AuthorName string `db:"author"`
	Image      string `db:"image"`
}
type dailyPostAuthor struct {
	ID    int    `db:"id"`
	Name  string `db:"author"`
	Image string `db:"image"`
	Posts []dailyPost
}
type dailyPost struct {
	ID         int    `db:"id"`
	Title      string `db:"title"`
	Content    string `db:"content"`
	Link       string `db:"link"`
	LinkTitle  string `db:"link_title"`
	LinkImage  string `db:"link_image"`
	AuthorID   int    `db:"author_id"`
	AuthorName string `db:"author"`
	Image      string `db:"image"`
	Order      int
}
type mailReceiver struct {
	Mail string `db:"mail"`
	Role int    `db:"role"`
}

func (m *mailApi) GetSubLink() map[int]string {
	return map[int]string{
		9: "https://www.readr.tw/admin",
		3: "https://www.readr.tw/editor",
		2: "https://www.readr.tw/guesteditor",
		1: "https://www.readr.tw/member",
	}
}

func (m *mailApi) GenDailyDigest() (err error) {
	subLink := m.GetSubLink()
	date := time.Now() //.AddDate(0, 0, -1)
	reports, err := m.getDailyReport()
	if err != nil {
		log.Println("getDailyReport", err)
		return err
	}
	memos, err := m.getDailyMemo()
	if err != nil {
		log.Println("getDailyMemo", err)
		return err
	}
	posts, err := m.getDailyPost()
	if err != nil {
		log.Println("getDailyPost", err)
		return err
	}

	//t := template.New("newsletter.html")
	//t = template.Must(t.ParseFiles("config/newsletter.html"))
	t := template.Must(template.ParseGlob("config/*.html"))

	data := dailyDigest{DateDay: date.Day(), DateMonth: int(date.Month()), DateYear: date.Year()}

	data.Reports = reports

OLM:
	for _, m := range memos {
		for mk, ma := range data.Memos {
			if ma.ID == m.AuthorID {
				data.Memos[mk].Memos = append(ma.Memos, m)
				continue OLM
			}
		}
		data.Memos = append(data.Memos, dailyMemoAuthor{ID: m.AuthorID, Name: m.AuthorName, Image: m.Image, Memos: []dailyMemo{m}})
	}

OLP:
	for _, p := range posts {
		for pk, pa := range data.Posts {
			if pa.ID == p.AuthorID {
				p.Order = len(pa.Posts) + 1
				data.Posts[pk].Posts = append(pa.Posts, p)
				continue OLP
			} else if p.AuthorID == 126 {
				data.ReadrPosts = append(data.ReadrPosts, p)
				continue OLP
			}
		}
		p.Order = 1
		data.Posts = append(data.Posts, dailyPostAuthor{ID: p.AuthorID, Name: p.AuthorName, Image: p.Image, Posts: []dailyPost{p}})
	}

	data.HasReport = len(data.Reports) > 0
	data.HasMemo = len(data.Memos) > 0
	data.HasPost = len(data.Posts) > 0
	data.HasReadrPost = len(data.ReadrPosts) > 0

	for k, v := range subLink {
		data.SubLink = v
		buf := new(bytes.Buffer)
		err = t.ExecuteTemplate(buf, "newsletter.html", data)
		s := buf.String()

		conn := RedisHelper.Conn()
		defer conn.Close()
		conn.Send("SET", fmt.Sprintf("dailydigest_%d", k), s)
	}

	return err
}

func (m *mailApi) getDailyReport() (reports []dailyReport, err error) {

	query := fmt.Sprintf("SELECT id, title, hero_image, description, slug FROM reports WHERE DATE(created_at) = DATE(NOW() - INTERVAL 1 DAY) AND active = %d AND publish_status = %d;", config.Config.Models.Reports["active"], config.Config.Models.ReportsPublishStatus["publish"])
	rows, err := DB.Queryx(query)
	for rows.Next() {
		var report dailyReport
		if err = rows.StructScan(&report); err != nil {
			reports = []dailyReport{}
			return reports, err
		}
		report.Description = m.htmlEscape(report.Description, 100)
		reports = append(reports, report)
	}
	return reports, err
}

func (m *mailApi) getDailyMemo() (memos []dailyMemo, err error) {

	query := fmt.Sprintf("SELECT m.memo_id AS id, m.title AS title, m.content AS content, p.slug AS slug, e.id AS author_id, e.nickname AS author, e.profile_image AS image FROM memos AS m LEFT JOIN members AS e ON m.author = e.id LEFT JOIN projects AS p ON p.project_id = m.project_id WHERE DATE(m.updated_at) = DATE(NOW() - INTERVAL 1 DAY) AND m.active = %d AND m.publish_status = %d;", config.Config.Models.Memos["active"], config.Config.Models.MemosPublishStatus["publish"])
	rows, err := DB.Queryx(query)
	for rows.Next() {
		var memo dailyMemo
		if err = rows.StructScan(&memo); err != nil {
			memos = []dailyMemo{}
			return memos, err
		}
		memo.Content = m.htmlEscape(memo.Content, 100)
		memos = append(memos, memo)
	}
	return memos, err
}

func (m *mailApi) getDailyPost() (posts []dailyPost, err error) {

	query := fmt.Sprintf(`SELECT p.post_id AS id, p.title AS title, p.content AS content, IFNULL(p.link, "") AS link, IFNULL(p.link_title, "") AS link_title, IFNULL(p.link_image, "") AS link_image, m.id AS author_id, m.nickname AS author, IFNULL(m.profile_image, "") AS image FROM posts AS p LEFT JOIN members AS m ON p.author = m.id WHERE DATE(p.updated_at) = DATE(NOW() - INTERVAL 1 DAY) AND p.active = %d AND p.publish_status = %d;`, config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"])
	rows, err := DB.Queryx(query)
	for rows.Next() {
		var post dailyPost
		if err = rows.StructScan(&post); err != nil {
			posts = []dailyPost{}
			return posts, err
		}
		post.Content = m.htmlEscape(post.Content, 100)
		post.LinkTitle = m.htmlEscape(post.LinkTitle, 40)
		posts = append(posts, post)
	}
	return posts, err
}

func (m *mailApi) getMailingList(listArgs ...interface{}) (list []mailReceiver, err error) {
	var (
		receiverList []string
		mailType     string
	)
	// if there are 2 arguments, set the first argument as the receivers of mails
	// and set the second argument as the type of mail, which will only be passed in new user registration
	switch len(listArgs) {
	case 2:
		receiverList = listArgs[0].([]string)
		mailType = listArgs[1].(string)
	default:
		receiverList = listArgs[0].([]string)
	}

	var rows *sqlx.Rows
	if len(receiverList) > 0 {

		var activeRestrict int

		// new user condition, active is 0(deacitve)
		if mailType == "init" {
			activeRestrict = config.Config.Models.Members["deactive"]
		} else {
			activeRestrict = config.Config.Models.Members["active"]
		}
		query, args, err := sqlx.In(fmt.Sprintf("SELECT mail, role FROM members WHERE active = %d AND mail IN (?);", activeRestrict), receiverList)
		if err != nil {
			return list, err
		}

		query = DB.Rebind(query)
		rows, err = DB.Queryx(query, args...)
	} else {
		query := fmt.Sprintf("SELECT mail, role FROM members WHERE active = %d AND daily_push = %d", config.Config.Models.Members["active"], config.Config.Models.MemberDailyPush["active"])
		rows, err = DB.Queryx(query)
	}
	for rows.Next() {
		var receiver mailReceiver
		if err = rows.StructScan(&receiver); err != nil {
			return list, err
		}
		list = append(list, receiver)
	}
	return list, nil
}

func (m *mailApi) htmlEscape(s string, length int) string {
	s = strings.Replace(s, "&lt;", "<", -1)
	s = strings.Replace(s, "&gt;", ">", -1)
	r := []rune(s)
	if len(r) > length {
		return string(r[:length]) + " ..."
	} else {
		return string(r)
	}
}

func (m *mailApi) sendToAll(t string, s string, mailList []string) (err error) {
	for len(mailList) > 0 {
		receiver := mailList
		if len(mailList) > 100 {
			receiver = mailList[:100]
			mailList = mailList[100:]
		} else {
			mailList = []string{}
		}

		args := MailArgs{
			Subject: t,
			Payload: s,
			BCC:     receiver,
		}

		err = m.Send(args)
		if err != nil {
			log.Println("Send mail error:", err.Error())
			return err
		}
	}

	return err

}

func (m *mailApi) SendDailyDigest(mailList []string) (err error) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	subLink := m.GetSubLink()

	mailReceiverList, err := m.getMailingList(mailList)
	if err != nil {
		log.Println("Get mailing list error:", err.Error())
	}

	for k, _ := range subLink {
		var mails []string

		s, err := redis.Bytes(conn.Do("GET", fmt.Sprintf("dailydigest_%d", k)))
		if err != nil {
			log.Println("Redis commend error: ", err.Error())
			if err.Error() == "redigo: nil returned" {
				return errors.New("Not Found")
			} else {
				return err
			}
		}

		for _, receiver := range mailReceiverList {
			if receiver.Role == k {
				mails = append(mails, receiver.Mail)
			}
		}

		date := time.Now()
		err = m.sendToAll(fmt.Sprintf("Readr %d/%d 選文", int(date.Month()), date.Day()), string(s), mails)
	}
	return err
}

type projectUpdateNotify struct {
	ProjectID    int
	ProjectTitle string
	ProjectSlug  string
	AuthorID     int
	AuthorName   string
	AuthorImage  string
	Slug         string
	Title        string
	Content      string
	Image        string
	HasReport    bool
	HasMemo      bool
	SubLink      string
}

func (m *mailApi) SendProjectUpdateMail(resource interface{}, resourceTyep string) (err error) {
	var mailData projectUpdateNotify
	switch resourceTyep {
	case "report":
		p := resource.(ReportAuthors)
		mailData.ProjectID = p.ProjectID
		mailData.Slug = p.Slug.String
		mailData.Title = p.Title.String
		mailData.Content = p.Description.String
		mailData.Image = p.OgImage.String
		mailData.HasReport = true

		mailData.ProjectSlug = p.Project.Slug.String
		mailData.ProjectTitle = p.Project.Title.String

	case "memo":
		m := resource.(MemoDetail)
		mailData.ProjectID = int(m.Project.ID)
		mailData.Slug = strconv.Itoa(m.ID)
		mailData.Title = m.Title.String
		mailData.Content = m.Content.String
		mailData.HasMemo = true

		mailData.AuthorID = int(m.Author.Int)
		mailData.AuthorImage = m.Authors.ProfileImage.String
		mailData.AuthorName = m.Authors.Nickname.String

		mailData.ProjectSlug = m.Project.Slug.String
		mailData.ProjectTitle = m.Project.Title.String
	}

	subLink := m.GetSubLink()
	templatesForRoles := make(map[int]string, 0)
	var mailReceiverList []mailReceiver

	//t := template.New("project_notifyletter.html")
	t := template.Must(template.ParseGlob("config/*.html"))
	for k, v := range m.GetSubLink() {
		mailData.SubLink = v
		buf := new(bytes.Buffer)
		err = t.ExecuteTemplate(buf, "project_notifyletter.html", mailData)
		s := buf.String()
		templatesForRoles[k] = s
	}

	query := fmt.Sprintf(`
		SELECT mail, role FROM members AS m 
		LEFT JOIN following AS f 
			ON m.id = f.member_id 
		WHERE m.active = %d 
			AND m.daily_push = %d 
			AND f.type = %d 
			AND f.emotion = %d 
			AND f.target_id = %d`,
		config.Config.Models.Members["active"],
		config.Config.Models.MemberDailyPush["active"],
		config.Config.Models.FollowingType["project"], 0,
		mailData.ProjectID)

	rows, err := DB.Queryx(query)
	if err != nil {
		log.Println("Get followers of project %d error when SendProjectUpdateMail", mailData.ProjectID)
		return err
	}
	for rows.Next() {
		var receiver mailReceiver
		if err = rows.StructScan(&receiver); err != nil {
			log.Println("Scan followers of project %d error when SendProjectUpdateMail", mailData.ProjectID)
			return err
		}
		mailReceiverList = append(mailReceiverList, receiver)
	}

	for k, _ := range subLink {
		var mails []string
		s := templatesForRoles[k]

		for _, receiver := range mailReceiverList {
			if receiver.Role == k {
				mails = append(mails, receiver.Mail)
			}
		}
		err = m.sendToAll("Readr 專題內容更新", s, mails)
	}
	return err

}

func (m *mailApi) SendCECommentNotify(tpm TaggedPostMember) (err error) {
	t, err := template.New("CEComment_notify").Parse(`
		客座總編 <strong>{{.Member.Nickname.String}}</strong> 已發布一篇新評論：<strong>{{.Post.Title.String}}</strong> <br>
		新聞連結： <strong>{{.Post.Link.String}}</strong> <br>
		內文： <strong>{{.Post.Content.String}}</strong>
		`)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, tpm)
	if err != nil {
		return err
	}

	s := buf.String()

	err = m.sendToAll("[READr] 客座編輯發文通知", s, []string{config.Config.Mail.DevTeam})
	return nil
}

var MailAPI MailInterface = new(mailApi)
