package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/utils"
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

type InsertCommentArgs struct {
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
	ResourceName  NullString `json:"resource_name"`
	ResourceID    NullInt    `json:"resource_id"`
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
	InsertComment(comment InsertCommentArgs) (id int64, err error)
	UpdateComment(comment Comment) (err error)
	UpdateComments(req CommentUpdateArgs) (err error)

	GetReportedComments(args *GetReportedCommentArgs) ([]ReportedCommentAuthor, error)
	InsertReportedComments(report ReportedComment) (id int64, err error)
	UpdateReportedComments(report ReportedComment) (err error)

	UpdateCommentAmountByResource(resource string, resourceID int, action string) (err error)
	UpdateAllCommentAmount() (err error)
}

type commentAPI struct{}

func (c *commentAPI) GetComment(id int) (CommentAuthor, error) {
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

func (c *commentAPI) InsertComment(comment InsertCommentArgs) (id int64, err error) {
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

	comment.ID = id
	NotificationGen.GenerateCommentNotifications(comment)

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

type CommentCount struct {
	Count    int    `db:"count"`
	Resource string `db:"resource"`
}

func (c *commentAPI) GetCommentCountByResources(resources []string) (commentCounts []CommentCount, err error) {
	var values []interface{}
	var query string = "SELECT count(id) AS count, resource FROM comments GROUP BY resource"

	if len(resources) > 0 {
		query = query + " WHERE resource in (?)"
		query, values, err = sqlx.In(query, resources)
	} else {
		query, values, err = sqlx.In(query)
	}
	if err != nil {
		log.Println(err.Error())
		return commentCounts, err
	}
	query = DB.Rebind(query)

	err = DB.Select(&commentCounts, query, values...)
	if err != nil {
		log.Println(err.Error())
		return commentCounts, err
	}
	return commentCounts, err
}

func (c *commentAPI) UpdateCommentAmountByResource(resourceName string, resourceID int, action string) (err error) {
	tableName, idName := utils.GetResourceTableInfo(resourceName)

	if resourceName != "" {
		var adjustment string
		switch action {
		case "+":
			adjustment = "+ 1"
		case "-":
			adjustment = "- 1"
		default:
			return errors.New("Unknown Action")
		}

		query := fmt.Sprintf(`UPDATE %s SET comment_amount= IF(comment_amount IS NULL, 1, comment_amount %s) WHERE %s="%d";`, tableName, adjustment, idName, resourceID)
		_, err = DB.Exec(query)
		if err != nil {
			return err
		}
	}
	return err
}

func (c *commentAPI) UpdateAllCommentAmount() (err error) {
	resources, err := c.GetCommentCountByResources(make([]string, 0))
	if err != nil {
		log.Println("Error when getting comment count by resources, %v")
		return err
	}

	tx, err := DB.Begin()
	stmPost, _ := tx.Prepare(`UPDATE posts SET comment_amount=? WHERE post_id=? AND (comment_amount!=? OR comment_amount IS NULL);`)
	stmProject, _ := tx.Prepare(`UPDATE projects SET comment_amount=? WHERE slug=? AND (comment_amount!=? OR comment_amount IS NULL);`)
	stmMemo, _ := tx.Prepare(`UPDATE memos SET comment_amount=? WHERE memo_id=? AND (comment_amount!=? OR comment_amount IS NULL);`)
	stmReport, _ := tx.Prepare(`UPDATE reports SET comment_amount=? WHERE slug=? AND (comment_amount!=? OR comment_amount IS NULL);`)

	for _, v := range resources {
		resourceType, resourceID := utils.ParseResourceInfo(v.Resource)
		switch resourceType {
		case "post":
			_, err := stmPost.Exec(v.Count, resourceID, v.Count)
			if err != nil {
				log.Println("Error update comment counts: ", err.Error())
			}
		case "project":
			_, err := stmProject.Exec(v.Count, resourceID, v.Count)
			if err != nil {
				log.Println("Error update comment counts: ", err.Error())
			}
		case "memo":
			_, err := stmMemo.Exec(v.Count, resourceID, v.Count)
			if err != nil {
				log.Println("Error update comment counts: ", err.Error())
			}
		case "report":
			_, err := stmReport.Exec(v.Count, resourceID, v.Count)
			if err != nil {
				log.Println("Error update comment counts: ", err.Error())
			}
		}
	}
	tx.Commit()

	return err
}

var CommentAPI CommentInterface = new(commentAPI)

// var CommentActive map[string]interface{}
// var CommentStatus map[string]interface{}
// var ReportedCommentStatus map[string]interface{}
