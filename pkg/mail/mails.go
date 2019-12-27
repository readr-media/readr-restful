package mail

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"text/template"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
	"gopkg.in/gomail.v2"
)

type MailArgs struct {
	Receiver []string `json:"receiver"`
	CC       []string `json:"cc"`
	BCC      []string `json:"bcc"`
	Subject  string   `json:"subject"`
	Payload  string   `json:"content"`
	Type     string   `json:"type"`
	Bulk     bool
}

type MailInterface interface {
	Send(args MailArgs) (err error)
	SendUpdateNote(args models.GetFollowMapArgs) (err error)
	SendUpdateNoteAllResource(args models.GetFollowMapArgs) (err error)
	GenDailyDigest() (err error)
	SendDailyDigest(receiver []string) (err error)
	SendCECommentNotify(tmp models.TaggedPostMember) (err error)

	SendReportPublishMail(report models.ReportAuthors) (err error)
	SendMemoPublishMail(memo models.MemoDetail) (err error)
	SendFollowProjectMail(args models.FollowArgs) (err error)
}

type mailApi struct{}

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
	reqBody, _ := json.Marshal(map[string]interface{}{
		"user":     config.Config.Mail.User,
		"password": config.Config.Mail.Password,
		"from":     msg.FormatAddress(config.Config.Mail.User, config.Config.Mail.UserName),
		"to":       filteredReceivers,
		"cc":       args.CC,
		"bcc":      args.BCC,
		"subject":  args.Subject,
		"body":     args.Payload,
		"bulk":     args.Bulk,
	})

	resp, body, err := utils.HTTPRequest("POST", config.Config.Mail.Host,
		map[string]string{}, reqBody)

	if err != nil {
		log.Printf("Send mail error: %v\n", err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Printf("Send mail error: %v, status_code: %v", body, resp.StatusCode)
		return errors.New(string(body))
	}

	return err
}

func (m *mailApi) SendUpdateNote(args models.GetFollowMapArgs) (err error) {

	mapInterface, err := models.FollowingAPI.Get(&args)
	if err != nil {
		return err
	}
	followingMap, ok := mapInterface.([]models.FollowingMapItem)
	if !ok {
		log.Println("Error assert mapInterface @ SendUpdateNote ")
	}
	var (
		follower_index = make(map[string][]string)
		follower_info  = make(map[string]models.Member)
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

	members, err := models.MemberAPI.GetMembers(&models.GetMembersArgs{IDs: followers_list, Sorting: "member_id", MaxResult: uint8(len(followers_list)), Page: 1})
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

func (m *mailApi) SendUpdateNoteAllResource(args models.GetFollowMapArgs) (err error) {

	var (
		follower_reverse_index = make(map[string][]string)
		follower_index         = make(map[string]map[string][]string)
		follower_group_index   = make(map[string]map[string][]string)
		follower_info          = make(map[string]models.Member)
		//resource_list          = make(map[string][]string)
		followers_list []string
	)
	for _, t := range []string{"member", "post", "project"} {
		mapInterface, err := models.FollowingAPI.Get(&models.GetFollowMapArgs{Resource: models.Resource{ResourceName: t}, UpdateAfter: args.UpdateAfter})
		if err != nil {
			return err
		}
		followingMap, ok := mapInterface.([]models.FollowingMapItem)
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

	members, err := models.MemberAPI.GetMembers(&models.GetMembersArgs{IDs: followers_list, Sorting: "member_id", MaxResult: uint8(len(followers_list)), Page: 1})
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
	SettingLink  string
	MailID       string
	CampaignID   string
	APIDomain    string
	PreviewText  string
}

func (d *dailyDigest) hasContent() bool {
	return d.HasMemo || d.HasPost || d.HasReport || d.HasReadrPost
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
	Mail  string `db:"mail"`
	Role  int    `db:"role"`
	Subed bool   `db:"subed"`
}

func (m *mailApi) GetSettingLink() map[int]string {
	return map[int]string{
		9: "https://www.readr.tw/admin",
		3: "https://www.readr.tw/editor",
		2: "https://www.readr.tw/guesteditor",
		1: "https://www.readr.tw/member",
	}
}

func (m *mailApi) GenDailyDigest() (err error) {
	settingLink := m.GetSettingLink()

	conn := models.RedisHelper.WriteConn()
	defer conn.Close()

	// Delete daily mail entry in the begining.
	// Prevents from sending old mail if mail generting takes too long or blocked.
	for k, _ := range settingLink {
		conn.Send("DEL", fmt.Sprintf("dailydigest_%d", k))
	}

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
	t := template.Must(template.ParseGlob(fmt.Sprintf("%s/*.html", config.Config.Mail.TemplatePath)))

	data := dailyDigest{DateDay: date.Day(), DateMonth: int(date.Month()), DateYear: date.Year(), APIDomain: config.Config.DomainName}

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
		if p.AuthorID == 126 {
			data.ReadrPosts = append(data.ReadrPosts, p)
			continue OLP
		}
		for pk, pa := range data.Posts {
			if pa.ID == p.AuthorID {
				p.Order = len(pa.Posts) + 1
				data.Posts[pk].Posts = append(pa.Posts, p)
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
	data.MailID = fmt.Sprintf("DailyDigest_%s", date.Format("20060102"))
	data.CampaignID = "DailyDigest"

	for k, v := range settingLink {
		var s string
		if data.hasContent() {
			data.SettingLink = v
			buf := new(bytes.Buffer)
			err = t.ExecuteTemplate(buf, "newsletter.html", data)
			if err != nil {
				log.Println(fmt.Sprintf("Fail to execute template when generating daily digest: %v", err.Error()))
				return err
			}
			s = buf.String()
		} else {
			s = ""
		}
		conn.Send("SET", fmt.Sprintf("dailydigest_%d", k), s)
	}

	return err
}

func (m *mailApi) getDailyReport() (reports []dailyReport, err error) {

	query := fmt.Sprintf("SELECT post_id, title, hero_image, content, slug FROM posts WHERE published_at > (NOW() - INTERVAL 1 DAY) AND active = %d AND publish_status = %d AND type = %d;", config.Config.Models.Reports["active"], config.Config.Models.ReportsPublishStatus["publish"], config.Config.Models.PostType["report"])
	rows, err := rrsql.DB.Queryx(query)
	if err != nil {
		return nil, errors.New("Get Daily Report Error")
	}
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

	query := fmt.Sprintf("SELECT p.post_id AS id, p.title AS title, p.content AS content, p.slug AS slug, e.id AS author_id, e.nickname AS author, e.profile_image AS image FROM posts AS p LEFT JOIN members AS e ON p.author = e.id WHERE p.published_at > (NOW() - INTERVAL 1 DAY) AND p.active = %d AND p.publish_status = %d AND p.type = %d;", config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"], config.Config.Models.PostType["memo"])
	rows, err := rrsql.DB.Queryx(query)
	if err != nil {
		return nil, errors.New("Get Daily Memo Error")
	}
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

	query := fmt.Sprintf(`
		SELECT p.post_id AS id, p.title AS title, p.content AS content, 
			IFNULL(p.link, "") AS link, 
			IFNULL(p.link_title, "") AS link_title, 
			IFNULL(p.link_image, "") AS link_image, 
			m.id AS author_id, m.nickname AS author, 
			IFNULL(m.profile_image, "") AS image 
		FROM posts AS p 
		LEFT JOIN members AS m ON p.author = m.id 
		WHERE p.published_at > (NOW() - INTERVAL 1 DAY) 
			AND p.active = %d AND p.publish_status = %d 
			AND p.type IN (%d, %d);`,
		config.Config.Models.Posts["active"],
		config.Config.Models.PostPublishStatus["publish"],
		config.Config.Models.PostType["review"],
		config.Config.Models.PostType["news"])
	rows, err := rrsql.DB.Queryx(query)
	if err != nil {
		return nil, errors.New("Get Daily Post Error")
	}
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

		query = rrsql.DB.Rebind(query)
		rows, err = rrsql.DB.Queryx(query, args...)
	} else {
		query := fmt.Sprintf("SELECT mail, role FROM members WHERE active = %d AND daily_push = %d", config.Config.Models.Members["active"], config.Config.Models.MemberDailyPush["active"])
		rows, err = rrsql.DB.Queryx(query)
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
	if !config.Config.Mail.Enable {
		return errors.New("Mail Service Disabled")
	}
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
			Bulk:    true,
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
	conn := models.RedisHelper.ReadConn()
	defer conn.Close()

	settingLink := m.GetSettingLink()

	mailReceiverList, err := m.getMailingList(mailList)
	if err != nil {
		log.Println("Get mailing list error:", err.Error())
		return err
	}

	for k, _ := range settingLink {
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

		mail_body := string(s)
		// If no content to send, skip this sending process
		if mail_body == "" {
			continue
		}

		for _, receiver := range mailReceiverList {
			if receiver.Role == k {
				mails = append(mails, receiver.Mail)
			}
		}

		date := time.Now()
		err = m.sendToAll(fmt.Sprintf("Readr %d/%d 選文", int(date.Month()), date.Day()), mail_body, mails)
		if err != nil {
			log.Println("Fail to send group mail: ", err.Error())
			return err
		}
	}
	return err
}

type reportPublishData struct {
	ProjectTitle     string
	ProjectHeroImage string
	ProjectSlug      string
	Title            string
	Description      string
	Slug             string
	SettingLink      string
	MailID           string
}

func (m *mailApi) SendReportPublishMail(report models.ReportAuthors) (err error) {
	// newReport.html
	SettingLink := m.GetSettingLink()
	data := reportPublishData{
		ProjectTitle:     report.Project.Title.String,
		ProjectSlug:      report.Project.Slug.String,
		ProjectHeroImage: report.Project.HeroImage.String,
		Title:            report.Report.Title.String,
		Description:      report.Report.Content.String,
		Slug:             report.Report.Slug.String,
		MailID:           fmt.Sprintf("ProjectUpdate_%s", report.Project.Slug.String),
	}

	mailReceiverList, err := m.getProjectFollowerMailList(report.Project.ID)
	if err != nil {
		log.Print(err)
		return err
	}

	t := template.Must(template.ParseGlob(fmt.Sprintf("%s/newReport.html", config.Config.Mail.TemplatePath)))
	for k, v := range SettingLink {
		var mails []string
		for _, receiver := range mailReceiverList {
			if receiver.Role == k {
				mails = append(mails, receiver.Mail)
			}
		}

		if len(mails) > 0 {
			data.SettingLink = v

			buf := new(bytes.Buffer)
			err = t.ExecuteTemplate(buf, "newReport.html", data)
			s := buf.String()

			err = m.sendToAll(fmt.Sprintf("【%s】最新報導<%s>", data.ProjectTitle, data.Title), s, mails)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type memoPublishData struct {
	ProjectTitle       string
	ProjectSlug        string
	Title              string
	Content            string
	PartialContent     string
	CreatedAt          string // 2018/07/02
	AuthorNickname     string
	AuthorProfileImage string
	SettingLink        string
	UnfollowLink       string
	MailID             string
}

func (m *mailApi) SendMemoPublishMail(memo models.MemoDetail) (err error) {
	// newMemoPaid.html
	// newMemoUnPaid.html
	SettingLink := m.GetSettingLink()
	abstract, _ := utils.CutAbstract(memo.Memo.Content.String, 100, func(a string) string {
		return fmt.Sprintf(`<p>%s... <a href="https://www.readr.tw/series/%s/%d" target="_blank">閱讀完整內容</a><p>`, a, memo.Project.Slug.String, memo.Memo.ID)
	})

	data := memoPublishData{
		ProjectTitle:       memo.Project.Project.Title.String,
		ProjectSlug:        memo.Project.Project.Slug.String,
		Title:              memo.Memo.Title.String,
		Content:            memo.Memo.Content.String,
		PartialContent:     abstract,
		CreatedAt:          fmt.Sprintf("%d/%02d/%02d", memo.Memo.CreatedAt.Time.Year(), memo.Memo.CreatedAt.Time.Month(), memo.Memo.CreatedAt.Time.Day()),
		AuthorNickname:     memo.Authors.Nickname.String,
		AuthorProfileImage: fmt.Sprintf("https://www.readr.tw%s", memo.Authors.ProfileImage.String),
		MailID:             fmt.Sprintf("ProjectUpdate_%s", memo.Project.Project.Slug.String),
	}

	mailReceiverList, err := m.getProjectFollowerMailList(memo.Project.ID)
	if err != nil {
		log.Print(err)
		return err
	}

	t := template.Must(template.ParseGlob(fmt.Sprintf("%s/newMemo*.html", config.Config.Mail.TemplatePath)))
	for k, v := range SettingLink {
		var submails, unsubmails []string
		for _, receiver := range mailReceiverList {
			if receiver.Role == k {
				if receiver.Subed {
					submails = append(submails, receiver.Mail)
				} else {
					unsubmails = append(unsubmails, receiver.Mail)
				}

			}
		}

		data.SettingLink = v

		if len(submails) > 0 {
			buf := new(bytes.Buffer)
			_ = t.ExecuteTemplate(buf, "newMemoPaid.html", data)
			s := buf.String()

			err = m.sendToAll(fmt.Sprintf("【%s】筆記更新<%s>", data.ProjectTitle, data.Title), s, submails)
			if err != nil {
				return err
			}
		}

		if len(unsubmails) > 0 {
			buf := new(bytes.Buffer)
			_ = t.ExecuteTemplate(buf, "newMemoUnPaid.html", data)
			s := buf.String()

			err = m.sendToAll(fmt.Sprintf("【%s】筆記更新<%s>", data.ProjectTitle, data.Title), s, unsubmails)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *mailApi) SendFollowProjectMail(args models.FollowArgs) (err error) {
	project, err := models.ProjectAPI.GetProject(models.Project{ID: int(args.Object)})
	if err != nil {
		log.Println("Error get project when SendFollowProjectMail: ", err)
		return err
	}
	member, err := models.MemberAPI.GetMember(models.GetMemberArgs{
		ID:     strconv.Itoa(int(args.Subject)),
		IDType: "id",
	})
	if err != nil {
		log.Println("Error get member when SendFollowProjectMail: ", err)
		return err
	}
	if !member.PostPush.Bool {
		return nil
	}

	data := map[string]string{
		"ProjectTitle": project.Title.String,
		"ProjectSlug":  project.Slug.String,
		"SettingLink":  m.GetSettingLink()[int(member.Role.Int)],
		"MailID":       fmt.Sprintf("ProjectFollow_%s", project.Slug.String),
	}

	buf := new(bytes.Buffer)
	t := template.Must(template.ParseGlob(fmt.Sprintf("%s/toggleFollow.html", config.Config.Mail.TemplatePath)))
	_ = t.ExecuteTemplate(buf, "toggleFollow.html", data)
	s := buf.String()

	err = m.sendToAll(fmt.Sprintf("【%s】追蹤成功", data["ProjectTitle"]), s, []string{member.Mail.String})
	if err != nil {
		return err
	}

	return nil
}

func (m *mailApi) getProjectFollowerMailList(id int) (receiveres []mailReceiver, err error) {
	query := fmt.Sprintf(`
		SELECT mail, role, IF(p.id IS NOT NULL, true, false) as subed FROM members AS m 
		LEFT JOIN following AS f 
			ON m.id = f.member_id 
		LEFT JOIN (
			SELECT id, member_id FROM points WHERE object_type = %d AND object_id = %d
			) AS p 
			ON f.member_id = p.member_id 
		WHERE m.active = %d 
			AND m.post_push = %d 
			AND f.type = %d 
			AND f.emotion = %d 
			AND f.target_id = %d`,
		config.Config.Models.PointType["project_memo"], id,
		config.Config.Models.Members["active"],
		config.Config.Models.MemberPostPush["active"],
		config.Config.Models.FollowingType["project"], 0,
		id)

	rows, err := rrsql.DB.Queryx(query)
	if err != nil {
		log.Printf("Get followers of project %d error when SendProjectUpdateMail", id)
		return receiveres, err
	}
	for rows.Next() {
		var receiver mailReceiver
		if err = rows.StructScan(&receiver); err != nil {
			log.Printf("Scan followers of project %d error when SendProjectUpdateMail", id)
			return receiveres, err
		}
		receiveres = append(receiveres, receiver)
	}
	return receiveres, err
}

func (m *mailApi) SendCECommentNotify(tpm models.TaggedPostMember) (err error) {
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

// Mailer is the mail service interface
type Mailer interface {
	Send() (err error)
	Set(m Mail)
}

// Mail contains the manipulatable arguments for a mail
type Mail struct {
	UserName string
	To       []string `json:"to"`
	CC       []string `json:"cc"`
	BCC      []string `json:"bcc"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	Bulk     bool     `json:"bulk"`
}

// MailService holds all the information to request mail API
type MailService struct {
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`

	Mail
}

// Send will send letters with the information store in MailService
func (m *MailService) Send() (err error) {

	msg := gomail.NewMessage()
	// Populate secrets from config
	m.User = config.Config.Mail.User
	m.Password = config.Config.Mail.Password
	// default use user_name from config
	if m.Mail.UserName == "" {
		m.Mail.UserName = config.Config.Mail.UserName
	}
	m.From = msg.FormatAddress(config.Config.Mail.User, m.Mail.UserName)
	reqBody, _ := json.Marshal(m)

	resp, body, err := utils.HTTPRequest("POST", config.Config.Mail.Host,
		map[string]string{}, reqBody)

	if err != nil {
		log.Printf("Send mail error: %v\n", err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Printf("Send mail error: %v, status_code: %v", body, resp.StatusCode)
		return errors.New(string(body))
	}
	return err
}

// Set will modify the arguments in MailService with input arguments
func (m *MailService) Set(mail Mail) {
	m.Mail = mail
}
