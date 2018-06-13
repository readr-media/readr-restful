package models

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"io/ioutil"
	"text/template"

	"github.com/spf13/viper"
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
	SendUpdateNoteAllResource(args GetFollowMapArgs) (err error)
	GenDailyDigest() (err error)
	SendDailyDigest() (err error)
}

type mailApi struct {
	Dialer gomail.Dialer
}

func (m *mailApi) SetDialer(dialer gomail.Dialer) {
	m.Dialer = dialer
}

func (m *mailApi) Send(args MailArgs) (err error) {

	msg := gomail.NewMessage()
	msg.SetHeader("From", msg.FormatAddress(viper.Get("mail.user").(string), viper.Get("mail.user_name").(string)))
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
	DateDay    int
	DateMonth  int
	DateYear   int
	Reports    []dailyReport
	Memos      []dailyMemoAuthor
	Posts      []dailyPostAuthor
	ReadrPosts []dailyPost
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

func (m *mailApi) GenDailyDigest() (err error) {
	date := time.Now().AddDate(0, 0, -1)
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

	//buf := new(bytes.Buffer)
	t := template.New("newsletter.html")
	t = template.Must(t.ParseFiles("config/newsletter.html"))

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

	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		os.Mkdir("tmp", 0755)
	}
	f, err := os.OpenFile("tmp/newsletter.html", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		log.Println("err when open newsletter tmp file", err.Error())
		return err
	}
	fw := bufio.NewWriter(f)

	err = t.Execute(fw, data)
	if err = fw.Flush(); err != nil {
		log.Println("err when writing newsletter content", err.Error())
		return err
	}

	//s := buf.String()
	return err
}

func (m *mailApi) getDailyReport() (reports []dailyReport, err error) {
	query := fmt.Sprintf("SELECT id, title, hero_image, description, slug FROM reports WHERE DATE(created_at) = DATE(NOW() - INTERVAL 1 DAY) AND active = %d AND publish_status = %d;", int(ReportActive["active"].(float64)), int(ReportPublishStatus["publish"].(float64)))
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
	query := fmt.Sprintf("SELECT m.memo_id AS id, m.title AS title, m.content AS content, p.slug AS slug, e.id AS author_id, e.nickname AS author, e.profile_image AS image FROM memos AS m LEFT JOIN members AS e ON m.author = e.id LEFT JOIN projects AS p ON p.project_id = m.project_id WHERE DATE(m.updated_at) = DATE(NOW() - INTERVAL 1 DAY) AND m.active = %d AND m.publish_status = %d;", int(MemoStatus["active"].(float64)), int(MemoPublishStatus["publish"].(float64)))
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
	query := fmt.Sprintf(`SELECT p.post_id AS id, p.title AS title, p.content AS content, IFNULL(p.link, "") AS link, IFNULL(p.link_title, "") AS link_title, IFNULL(p.link_image, "") AS link_image, m.id AS author_id, m.nickname AS author, IFNULL(m.profile_image, "") AS image FROM posts AS p LEFT JOIN members AS m ON p.author = m.id WHERE DATE(p.updated_at) = DATE(NOW() - INTERVAL 1 DAY) AND p.active = %d AND p.publish_status = %d;`, int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)))
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

func (m *mailApi) getMailingList() (list []string, err error) {
	query := fmt.Sprintf("SELECT mail FROM members WHERE active = %d", int(MemberStatus["active"].(float64)))
	rows, err := DB.Queryx(query)
	for rows.Next() {
		var mail string
		if err = rows.Scan(&mail); err != nil {
			return list, err
		}
		list = append(list, mail)
	}
	return list, err
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
func (m *mailApi) SendDailyDigest() (err error) {
	s, err := ioutil.ReadFile("tmp/newsletter.html")
	if err != nil {
		log.Println("File readr error:", err.Error())
	}

	list, err := m.getMailingList()
	if err != nil {
		log.Println("Get mailing list error:", err.Error())
	}
	list = []string{"hcchien@mirrormedia.mg", "kaiwenhsiung@mirrormedia.mg", "cyu2197@gmail.com"} //for test

	args := MailArgs{
		Receiver: list,
		Subject:  "[Readr+] Daily Digest",
		Payload:  string(s),
		BCC:      []string{"yychen@mirrormedia.mg"},
	}

	err = m.Send(args)
	if err != nil {
		log.Println("Send mail error:", err.Error())
	}

	return err
}

var MailAPI MailInterface = new(mailApi)
