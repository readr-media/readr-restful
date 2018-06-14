package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Comment struct {
	ID            int64      `json:"id" db:"id"`
	Author        NullInt    `json:"author" db:"author"`
	Body          NullString `json:"body" db:"body"`
	OgTitle       NullString `json:"og_title" db:"og_title"`
	OgDescription NullString `json:"og_description" db:"og_description"`
	OgImage       NullString `json:"og_image" db:"og_image"`
	LikeAmount    NullInt    `json:"like_amount" db:"like_amount"`
	ParentID      NullInt    `json:"parent_id" db:"parent_id"`
	Resource      NullString `json:"resource" db:"resource"`
	Status        NullInt    `json:"status" db:"status"`
	Active        NullInt    `json:"active" db:"active"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
	IP            NullString `json:"ip" db:"ip"`
}

type CommentAuthor struct {
	Comment
	AuthorNickname NullString `json:"author_nickname" db:"author_nickname"`
	AuthorImage    NullString `json:"author_image" db:"author_image"`
	AuthorRole     NullInt    `json:"author_role" db:"author_role"`
	CommentAmount  NullInt    `json:"comment_amount" db:"comment_amount"`
}

type ReportedComment struct {
	ID        int64      `json:"id" db:"id"`
	CommentID int64      `json:"comment_id" db:"comment_id"`
	Reporter  NullInt    `json:"reporter" db:"reporter"`
	Reason    NullString `json:"reason" db:"reason"`
	Solved    NullInt    `json:"solved" db:"solved"`
	UpdatedAt NullTime   `json:"updated_at" db:"updated_at"`
	CreatedAt NullTime   `json:"created_at" db:"created_at"`
	IP        NullString `json:"ip" db:"ip"`
}

type ReportedCommentAuthor struct {
	Comment CommentAuthor   `json:"comments" db:"comments"`
	Report  ReportedComment `json:"reported" db:"reported"`
}

type GetCommentArgs struct {
	MaxResult int              `form:"max_result"`
	Page      int              `form:"page"`
	Sorting   string           `form:"sort"`
	Author    []int            `form:"author"`
	Resource  []string         `form:"resource"`
	Parent    []int            `form:"parent"`
	Status    map[string][]int `form:"status"`
}

func (p *GetCommentArgs) Default() (result *GetCommentArgs) {
	return &GetCommentArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (p *GetCommentArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Status != nil {
		for k, v := range p.Status {
			where = append(where, fmt.Sprintf("%s %s (?)", "comments.status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.Author) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "comments.author", operatorHelper("in")))
		values = append(values, p.Author)
	}
	if len(p.Resource) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?) %s", "comments.resource", operatorHelper("in"), " AND comments.parent_id IS NULL"))
		values = append(values, p.Resource)
	} else if len(p.Parent) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "comments.parent_id", operatorHelper("in")))
		values = append(values, p.Parent)
	}

	where = append(where, "comments.active=1")

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

type GetReportedCommentArgs struct {
	MaxResult int              `form:"max_result"`
	Page      int              `form:"page"`
	Sorting   string           `form:"sort"`
	Reporter  []int            `form:"reporter"`
	Parent    []int            `form:"parent"`
	Solved    map[string][]int `form:"solved"`
}

func (p *GetReportedCommentArgs) Default() (result *GetReportedCommentArgs) {
	return &GetReportedCommentArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (p *GetReportedCommentArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Solved != nil {
		for k, v := range p.Solved {
			where = append(where, fmt.Sprintf("%s %s (?)", "comments_reported.solved", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.Reporter) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "comments_reported.reporter", operatorHelper("in")))
		values = append(values, p.Reporter)
	}

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
		restricts = "WHERE " + restricts
	} else if len(where) == 1 {
		restricts = where[0]
		restricts = "WHERE " + restricts
	}
	return restricts, values
}

type CommentUpdateArgs struct {
	IDs       []int    `json:"ids"`
	UpdatedAt NullTime `json:"-" db:"updated_at"`
	Active    NullInt  `json:"active" db:"active"`
	Status    NullInt  `json:"status" db:"status"`
}

func (p *CommentUpdateArgs) parse() (updates string, values []interface{}) {
	setQuery := make([]string, 0)

	if p.Active.Valid {
		setQuery = append(setQuery, "active = ?")
		values = append(values, p.Active.Int)
	}
	if p.Status.Valid {
		setQuery = append(setQuery, "status = ?")
		values = append(values, p.Status.Int)
	}
	if p.UpdatedAt.Valid {
		setQuery = append(setQuery, "updated_at = ?")
		values = append(values, p.UpdatedAt.Time)
	}
	if len(setQuery) > 1 {
		updates = fmt.Sprintf(" %s", strings.Join(setQuery, " , "))
	} else if len(setQuery) == 1 {
		updates = fmt.Sprintf(" %s", setQuery[0])
	}

	return updates, values
}

type CommentInterface interface {
	GetComment(id int) (CommentAuthor, error)
	GetComments(args *GetCommentArgs) (result []CommentAuthor, err error)
	InsertComment(comment Comment) (id int64, err error)
	UpdateComment(comment Comment) (err error)
	UpdateComments(req CommentUpdateArgs) (err error)

	GetReportedComments(args *GetReportedCommentArgs) ([]ReportedCommentAuthor, error)
	InsertReportedComments(report ReportedComment) (id int64, err error)
	UpdateReportedComments(report ReportedComment) (err error)

	UpdateCommentAmountByResource(resource string, action string) (err error)
	UpdateCommentAmountByIDs(ids []int) (err error)
}

type commentAPI struct{}

func (a *commentAPI) GetComment(id int) (CommentAuthor, error) {
	comment := CommentAuthor{}
	err := DB.QueryRowx("SELECT comments.*, INET_NTOA(comments.ip) AS ip, members.nickname AS author_nickname, members.profile_image AS author_image, members.role AS author_role, IFNULL(count.count, 0) AS comment_amount FROM comments LEFT JOIN members ON comments.author = members.id LEFT JOIN (SELECT count(*) AS count, parent_id FROM comments GROUP BY parent_id) AS count ON comments.id = count.parent_id WHERE comments.id = ?", id).StructScan(&comment)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Comment Not Found")
		comment = CommentAuthor{}
	case err != nil:
		log.Println(err.Error())
		comment = CommentAuthor{}
	default:
		err = nil
	}
	return comment, err
}

func (c *commentAPI) GetComments(args *GetCommentArgs) (result []CommentAuthor, err error) {
	restricts, values := args.parse()

	query := fmt.Sprintf("SELECT comments.*, INET_NTOA(comments.ip) AS ip, members.nickname AS author_nickname, members.profile_image AS author_image, members.role AS author_role, IFNULL(count.count, 0) AS comment_amount FROM comments AS comments LEFT JOIN members AS members ON comments.author = members.id LEFT JOIN (SELECT count(*) AS count, parent_id FROM comments GROUP BY parent_id) AS count ON comments.id = count.parent_id WHERE %s ORDER BY %s LIMIT ? OFFSET ?;", restricts, orderByHelper(args.Sorting))
	values = append(values, args.MaxResult, (args.Page-1)*args.MaxResult)

	query, values, err = sqlx.In(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	for rows.Next() {
		var comment CommentAuthor
		if err = rows.StructScan(&comment); err != nil {
			result = []CommentAuthor{}
			return result, err
		}
		result = append(result, comment)
	}

	return result, err
}
func (c *commentAPI) InsertComment(comment Comment) (id int64, err error) {
	tags := getStructDBTags("full", Comment{})
	query := fmt.Sprintf(`INSERT INTO comments (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	query = strings.Replace(query, ":ip", "INET_ATON(:ip)", 1)
	result, err := DB.NamedExec(query, comment)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, errors.New("Duplicate entry")
		}
		return 0, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return 0, errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return 0, errors.New("Comment Not Found")
	}
	id, err = result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a comment: %v", err)
		return 0, err
	}

	c.generateCommentNotifications(int(id))

	return id, err
}

func (c *commentAPI) UpdateComment(comment Comment) (err error) {
	tags := getStructDBTags("partial", comment)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE comments SET %s WHERE id = :id`,
		strings.Join(fields, ", "))

	result, err := DB.NamedExec(query, comment)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Report Not Found")
	}

	return err
}
func (c *commentAPI) UpdateComments(args CommentUpdateArgs) (err error) {
	updateQuery, updateArgs := args.parse()
	updateQuery = fmt.Sprintf("UPDATE comments SET %s ", updateQuery)

	restrictQuery, restrictArgs, err := sqlx.In(`WHERE id IN (?)`, args.IDs)
	if err != nil {
		return err
	}

	restrictQuery = DB.Rebind(restrictQuery)
	updateArgs = append(updateArgs, restrictArgs...)
	_, err = DB.Exec(fmt.Sprintf("%s %s", updateQuery, restrictQuery), updateArgs...)
	if err != nil {
		return err
	}
	return err
}

func (c *commentAPI) GetReportedComments(args *GetReportedCommentArgs) (result []ReportedCommentAuthor, err error) {
	restricts, values := args.parse()
	commentTags := getStructDBTags("full", Comment{})
	reportTags := getStructDBTags("full", ReportedComment{})
	commentFields := makeFieldString("get", `comments.%s "comments.%s"`, commentTags)
	reportFields := makeFieldString("get", `comments_reported.%s "reported.%s"`, reportTags)

	query := fmt.Sprintf(`SELECT %s, %s, 
		members.nickname AS "comments.author_nickname", members.profile_image AS "comments.author_image", 
		members.role AS "comments.author_role", IFNULL(count.count, 0) AS "comments.comment_amount" 
			FROM comments AS comments LEFT JOIN members AS members ON comments.author = members.id 
				LEFT JOIN (SELECT count(*) AS count, parent_id FROM comments GROUP BY parent_id) AS count ON comments.id = count.parent_id 
				INNER JOIN comments_reported AS comments_reported ON comments_reported.comment_id = comments.id 
				%s ORDER BY %s LIMIT ? OFFSET ?;`,
		strings.Replace(strings.Join(commentFields, ","), `comments.ip "comments.ip"`, `INET_NTOA(comments.ip) "comments.ip"`, 1),
		strings.Replace(strings.Join(reportFields, ","), `comments_reported.ip "reported.ip"`, `INET_NTOA(comments_reported.ip) "reported.ip"`, 1),
		restricts, "comments_reported."+orderByHelper(args.Sorting))

	values = append(values, args.MaxResult, (args.Page-1)*args.MaxResult)

	query, values, err = sqlx.In(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	for rows.Next() {
		var comment ReportedCommentAuthor
		if err = rows.StructScan(&comment); err != nil {
			result = []ReportedCommentAuthor{}
			return result, err
		}
		result = append(result, comment)
	}

	return result, err
}
func (c *commentAPI) InsertReportedComments(report ReportedComment) (id int64, err error) {
	tags := getStructDBTags("full", ReportedComment{})
	query := fmt.Sprintf(`INSERT INTO comments_reported (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))
	query = strings.Replace(query, ":ip", "INET_ATON(:ip)", 1)

	result, err := DB.NamedExec(query, report)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, errors.New("Duplicate entry")
		}
		return 0, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return 0, errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return 0, errors.New("Report Not Found")
	}
	id, err = result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a report: %v", err)
		return 0, err
	}
	return id, err
}
func (c *commentAPI) UpdateReportedComments(report ReportedComment) (err error) {
	tags := getStructDBTags("partial", report)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE comments_reported SET %s WHERE id = :id`,
		strings.Join(fields, ", "))

	result, err := DB.NamedExec(query, report)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Report Not Found")
	}

	return err
}

func (c *commentAPI) UpdateCommentAmountByIDs(ids []int) (err error) {
	query, values, err := sqlx.In("SELECT DISTINCT resource FROM comments WHERE id IN (?);", ids)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	query = DB.Rebind(query)

	var resources []string
	err = DB.Select(&resources, query, values...)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, res := range resources {
		id, name, idname := c.parseCommentResource(res)
		if name != "" {
			_, err := DB.Exec(fmt.Sprintf(`UPDATE %s SET comment_amount=(SELECT count(id) FROM comments WHERE resource="%s" AND status=%d AND active=%d) WHERE %s="%s";`, name, res, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64)), idname, id))
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	}

	return err
}

func (c *commentAPI) UpdateCommentAmountByResource(resource string, action string) (err error) {
	id, resource_name, idname := c.parseCommentResource(resource)

	if resource_name != "" {
		var adjustment string
		switch action {
		case "+":
			adjustment = "+ 1"
		case "-":
			adjustment = "- 1"
		default:
			return errors.New("Unknown Action")
		}

		query := fmt.Sprintf(`UPDATE %s SET comment_amount= IF(comment_amount IS NULL, 1, comment_amount %s) WHERE %s="%s";`, resource_name, adjustment, idname, id)
		_, err = DB.Exec(query)
		if err != nil {
			return err
		}
	}
	return err
}

func (c *commentAPI) parseCommentResource(resource string) (string, string, string) {

	match_result := regexp.MustCompile("/post/([0-9]*)$").FindStringSubmatch(resource)
	r_id := ""
	r_name := ""
	r_idname := ""

	if len(match_result) > 1 {
		r_id = string(match_result[1])
		r_name = "post"
	}

	switch r_name {
	case "post":
		r_name = "posts"
		r_idname = "post_id"
	default:
		r_name = ""
		r_idname = ""
	}

	return r_id, r_name, r_idname
}

func (c *commentAPI) generateCommentNotifications(id int) (err error) {
	ns := Notifications{}

	commentDetail, err := c.GetComment(id)
	if err != nil {
		log.Println("Error get comment", id, err.Error())
	}

	r_id, r_name, _ := c.parseCommentResource(commentDetail.Resource.String)

	var parentCommentDetail CommentAuthor
	if commentDetail.ParentID.Valid {
		parentCommentDetail, err = c.GetComment(int(commentDetail.ParentID.Int))
		if err != nil {
			log.Println("Error get parent comment", commentDetail.ParentID.Int, err.Error())
		}
	}

	r_name = strings.TrimSuffix(r_name, "s")
	switch r_name {
	case "post":
		res_id, _ := strconv.Atoi(r_id)
		post, err := PostAPI.GetPost(uint32(res_id))
		if err != nil {
			log.Println("Error get post", res_id, err.Error())
		}

		postFollowers, err := FollowingAPI.GetFollowerMemberIDs("post", r_id)
		if err != nil {
			log.Println("Error get post followers", r_id, err.Error())
		}

		authorFollowers, err := FollowingAPI.GetFollowerMemberIDs("member", strconv.Itoa(int(post.Author.Int)))
		if err != nil {
			log.Println("Error get author followers", commentDetail.Author.Int, err.Error())
		}
		log.Println(authorFollowers)

		var commentors []int
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range postFollowers {
			if v != int(commentDetail.Author.Int) {
				ns[strconv.Itoa(v)] = NewNotification("follow_post_reply")
			}
		}

		for _, v := range authorFollowers {
			if v != int(commentDetail.Author.Int) {
				ns[strconv.Itoa(v)] = NewNotification("follow_member_reply")
			}
		}

		if commentDetail.Author.Int != post.Author.Int {
			ns[strconv.Itoa(int(post.Author.Int))] = NewNotification("post_reply")
		}

		if len(commentors) > 0 {
			for _, id := range commentors {
				if id != int(commentDetail.Author.Int) {
					ns[strconv.Itoa(id)] = NewNotification("comment_comment")
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			if commentDetail.Author.Int == post.Author.Int {
				ns[strconv.Itoa(int(parentCommentDetail.Author.Int))] = NewNotification("comment_reply_author")
			} else {
				ns[strconv.Itoa(int(parentCommentDetail.Author.Int))] = NewNotification("comment_reply")
			}
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = post.Member.Nickname.String
			v.ObjectType = r_name
			v.ObjectID = r_id
			v.PostType = strconv.Itoa(int(post.Type.Int))
			ns[k] = v
		}

		break

	case "project":
		res_id, _ := strconv.Atoi(r_id)
		project, err := ProjectAPI.GetProject(Project{ID: res_id})
		if err != nil {
			log.Println("Error get post", r_id, err.Error())
		}

		projectFollowers, err := FollowingAPI.GetFollowerMemberIDs("project", r_id)
		if err != nil {
			log.Println("Error get project followers", r_id, err.Error())
		}

		var commentors []int
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range projectFollowers {
			if v != int(commentDetail.Author.Int) {
				ns[strconv.Itoa(v)] = NewNotification("follow_project_reply")
			}
		}

		if len(commentors) > 0 {
			for _, id := range commentors {
				if id != int(commentDetail.Author.Int) {
					ns[strconv.Itoa(id)] = NewNotification("comment_comment")
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns[strconv.Itoa(int(parentCommentDetail.Author.Int))] = NewNotification("comment_reply")
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = project.Title.String
			v.ObjectType = r_name
			v.ObjectID = r_id
			v.PostType = "0"
			ns[k] = v
		}

		break

	case "memo":
		res_id, _ := strconv.Atoi(r_id)
		memo, err := MemoAPI.GetMemo(res_id)
		if err != nil {
			log.Println("Error get post", r_id, err.Error())
		}

		projectFollowers, err := FollowingAPI.GetFollowerMemberIDs("project", strconv.Itoa(int(memo.Project.Int)))
		if err != nil {
			log.Println("Error get project followers", memo.Project.Int, err.Error())
		}

		var commentors []int
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range projectFollowers {
			if v != int(commentDetail.Author.Int) {
				ns[strconv.Itoa(v)] = NewNotification("follow_memo_reply")
			}
		}

		if len(commentors) > 0 {
			for _, id := range commentors {
				if id != int(commentDetail.Author.Int) {
					ns[strconv.Itoa(id)] = NewNotification("comment_comment")
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns[strconv.Itoa(int(parentCommentDetail.Author.Int))] = NewNotification("comment_reply")
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = memo.Title.String
			v.ObjectType = r_name
			v.ObjectID = r_id
			v.PostType = "0"
			ns[k] = v
		}

		break

	case "report":
		break
	default:

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns[strconv.Itoa(int(parentCommentDetail.Author.Int))] = NewNotification("comment_reply")
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = commentDetail.Resource.String
			v.ObjectType = ""
			v.ObjectID = ""
			v.PostType = "0"
			ns[k] = v
		}

		break
	}
	ns.Send()
	return err
}

var CommentAPI CommentInterface = new(commentAPI)
var CommentActive map[string]interface{}
var CommentStatus map[string]interface{}
var ReportedCommentStatus map[string]interface{}
