package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/utils"
)

// var MemoStatus map[string]interface{}
// var MemoPublishStatus map[string]interface{}
/*
type Memo struct {
	ID            int        `json:"id" db:"memo_id"`
	CreatedAt     rrsql.NullTime   `json:"created_at" db:"created_at"`
	CommentAmount rrsql.NullInt    `json:"comment_amount" db:"comment_amount"`
	Title         rrsql.NullString `json:"title" db:"title"`
	Content       rrsql.NullString `json:"content" db:"content"`
	Link          rrsql.NullString `json:"link" db:"link"`
	Author        rrsql.NullInt    `json:"author" db:"author"`
	Project       rrsql.NullInt    `json:"project_id" db:"project_id"`
	Active        rrsql.NullInt    `json:"active" db:"active"`
	UpdatedAt     rrsql.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     rrsql.NullInt    `json:"updated_by" db:"updated_by"`
	PublishedAt   rrsql.NullTime   `json:"published_at" db:"published_at"`
	PublishStatus rrsql.NullInt    `json:"publish_status" db:"publish_status"`
	Order         rrsql.NullInt    `json:"memo_order" db:"memo_order"`
}
*/
type Memo struct {
	ID              uint32           `json:"id" db:"post_id" redis:"post_id"`
	Author          rrsql.NullInt    `json:"author" db:"author" redis:"author"`
	CreatedAt       rrsql.NullTime   `json:"created_at" db:"created_at" redis:"created_at"`
	LikeAmount      rrsql.NullInt    `json:"like_amount" db:"like_amount" redis:"like_amount"`
	CommentAmount   rrsql.NullInt    `json:"comment_amount" db:"comment_amount" redis:"comment_amount"`
	Title           rrsql.NullString `json:"title" db:"title" redis:"title"`
	Subtitle        rrsql.NullString `json:"subtitle" db:"subtitle" redis:"subtitle"`
	Content         rrsql.NullString `json:"content" db:"content" redis:"content"`
	Type            rrsql.NullInt    `json:"type" db:"type" redis:"type"`
	Link            rrsql.NullString `json:"link" db:"link" redis:"link"`
	OgTitle         rrsql.NullString `json:"og_title" db:"og_title" redis:"og_title"`
	OgDescription   rrsql.NullString `json:"og_description" db:"og_description" redis:"og_description"`
	OgImage         rrsql.NullString `json:"og_image" db:"og_image" redis:"og_image"`
	Active          rrsql.NullInt    `json:"active" db:"active" redis:"active"`
	UpdatedAt       rrsql.NullTime   `json:"updated_at" db:"updated_at" redis:"updated_at"`
	UpdatedBy       rrsql.NullInt    `json:"updated_by" db:"updated_by" redis:"updated_by"`
	PublishedAt     rrsql.NullTime   `json:"published_at" db:"published_at" redis:"published_at"`
	LinkTitle       rrsql.NullString `json:"link_title" db:"link_title" redis:"link_title"`
	LinkDescription rrsql.NullString `json:"link_description" db:"link_description" redis:"link_description"`
	LinkImage       rrsql.NullString `json:"link_image" db:"link_image" redis:"link_image"`
	LinkName        rrsql.NullString `json:"link_name" db:"link_name" redis:"link_name"`
	VideoID         rrsql.NullString `json:"video_id" db:"video_id" redis:"video_id"`
	VideoViews      rrsql.NullInt    `json:"video_views" db:"video_views" redis:"video_views"`
	PublishStatus   rrsql.NullInt    `json:"publish_status" db:"publish_status" redis:"publish_status"`
	ProjectID       rrsql.NullInt    `json:"project_id" db:"project_id" redis:"project_id"`
	Order           rrsql.NullInt    `json:"post_order" db:"post_order" redis:"post_order"`
	HeroImage       rrsql.NullString `json:"hero_image" db:"hero_image" redis:"hero_image"`
	Slug            rrsql.NullString `json:"slug" db:"slug" redis:"slug"`
	CSS             rrsql.NullString `json:"css" db:"css" redis:"css"`
	JS              rrsql.NullString `json:"javascript" db:"javascript" redis:"javascript"`
}

type MemoInterface interface {
	CountMemos(args *MemoGetArgs) (int, error)
	GetMemo(id int) (Memo, error)
	GetMemos(args *MemoGetArgs) ([]MemoDetail, error)
	InsertMemo(memo Memo) (int, error)
	UpdateMemo(memo Memo) error
	UpdateMemos(args MemoUpdateArgs) error
	SchedulePublish() (ids []int, err error)
}

type MemoGetArgs struct {
	MaxResult            int              `form:"max_result"`
	Page                 int              `form:"page"`
	Sorting              string           `form:"sort"`
	Keyword              string           `form:"keyword"`
	Author               []int64          `form:"author"`
	Project              []int64          `form:"project_id"`
	Slugs                []string         `form:"slugs"`
	Active               map[string][]int `form:"active"`
	MemoPublishStatus    map[string][]int `form:"memo_publish_status"`
	ProjectPublishStatus map[string][]int `form:"project_publish_status"`
	MemberID             int64            `form:"member_id"`
	AbstractLength       int64            `form:"abstract_length"`
	IDs                  []int64
	IsAdmin              bool
}

func (p *MemoGetArgs) Default() (result *MemoGetArgs) {
	return &MemoGetArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at", AbstractLength: 20}
}

func (p *MemoGetArgs) DefaultActive() {
	// p.Active = map[string][]int{"$nin": []int{int(MemoStatus["deactive"].(float64))}}
	p.Active = map[string][]int{"$nin": []int{config.Config.Models.Memos["deactive"]}}
}

func (p *MemoGetArgs) Validate() bool {
	if matched, err := regexp.MatchString("-?(updated_at|created_at|published_at|post_id|author|project_id|post_order)", p.Sorting); err != nil || !matched {
		return false
	}
	return true
}

func (p *MemoGetArgs) parse() (restricts string, values []interface{}) {

	where := make([]string, 0)
	where = append(where, fmt.Sprintf("%s %s %d", "posts.type", "=", config.Config.Models.PostType["memo"]))
	// if p.MemoPublishStatus == nil && p.ProjectPublishStatus == nil && p.Active == nil && len(p.Author) == 0 && len(p.Project) == 0 {
	// 	return "", nil
	// }
	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.active", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.MemoPublishStatus != nil {
		for k, v := range p.MemoPublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.publish_status", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.ProjectPublishStatus != nil {
		for k, v := range p.ProjectPublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "project.publish_status", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.Author) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "posts.author"))
		values = append(values, p.Author)
	}
	if len(p.Project) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "posts.project_id"))
		values = append(values, p.Project)
	}
	if len(p.Slugs) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "posts.slug"))
		values = append(values, p.Slugs)
	}
	if len(p.IDs) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "posts.post_id"))
		values = append(values, p.IDs)
	}
	if p.Keyword != "" {
		p.Keyword = fmt.Sprintf("%s%s%s", "%", p.Keyword, "%")
		where = append(where, "(posts.title LIKE ? OR posts.post_id LIKE ?)")
		values = append(values, p.Keyword, p.Keyword)
	}

	if len(where) > 1 {
		restricts = fmt.Sprintf(" WHERE %s", strings.Join(where, " AND "))
	} else if len(where) == 1 {
		restricts = fmt.Sprintf(" WHERE %s", where[0])
	}
	return restricts, values
}

func (p *MemoGetArgs) parseLimit() (limit map[string]string, values []interface{}) {

	restricts := make([]string, 0)
	limit = make(map[string]string, 2)
	if p.Sorting != "" {
		restricts = append(restricts, fmt.Sprintf("ORDER BY %s%s", "posts.", rrsql.OrderByHelper(p.Sorting)))
		limit["order"] = fmt.Sprintf("ORDER BY %s", rrsql.OrderByHelper(p.Sorting))
	}
	if p.MaxResult != 0 {
		restricts = append(restricts, "LIMIT ?")
		values = append(values, p.MaxResult)
	}
	if p.Page != 0 {
		restricts = append(restricts, "OFFSET ?")
		values = append(values, (p.Page-1)*(p.MaxResult))
	}
	if len(restricts) > 0 {
		limit["full"] = fmt.Sprintf(" %s", strings.Join(restricts, " "))
	}
	return limit, values
}

type MemoUpdateArgs struct {
	IDs         []int          `json:"ids"`
	UpdatedBy   int64          `json:"updated_by"`
	UpdatedAt   rrsql.NullTime `json:"-"`
	PublishedAt rrsql.NullTime `json:"-"`
	Active      rrsql.NullInt  `json:"-"`
}

func (p *MemoUpdateArgs) parse() (updates string, values []interface{}) {

	setQuery := make([]string, 0)

	if p.Active.Valid {
		setQuery = append(setQuery, "active = ?")
		values = append(values, p.Active.Int)
	}
	if p.PublishedAt.Valid {
		setQuery = append(setQuery, "published_at = ?")
		values = append(values, p.PublishedAt.Time)
	}
	if p.UpdatedAt.Valid {
		setQuery = append(setQuery, "updated_at = ?")
		values = append(values, p.UpdatedAt.Time)
	}
	if p.UpdatedBy != 0 {
		setQuery = append(setQuery, "updated_by = ?")
		values = append(values, p.UpdatedBy)
	}
	if len(setQuery) > 1 {
		updates = fmt.Sprintf(" %s", strings.Join(setQuery, " , "))
	} else if len(setQuery) == 1 {
		updates = fmt.Sprintf(" %s", setQuery[0])
	}

	return updates, values
}

type MemoDetail struct {
	Memo
	Authors Member `json:"author" db:"author"`
	// Project Project `json:"project" db:"project"`
	Project struct {
		Project
		Paid bool `json:"paid"`
	} `json:"project" db:"project"`
}

type memoAPI struct{}

func (m *memoAPI) CountMemos(args *MemoGetArgs) (result int, err error) {

	restricts, values := args.parse()
	query := fmt.Sprintf(`SELECT COUNT(posts.post_id) FROM posts LEFT JOIN projects AS project ON project.project_id = posts.project_id %s`, restricts)

	query, sqlArgs, err := sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = rrsql.DB.Rebind(query)

	count, err := rrsql.DB.Queryx(query, sqlArgs...)
	if err != nil {
		return 0, err
	}
	for count.Next() {
		if err = count.Scan(&result); err != nil {
			return 0, err
		}
	}
	return result, err
}

func (m *memoAPI) GetMemo(id int) (memo Memo, err error) {

	err = rrsql.DB.Get(&memo, `SELECT * FROM posts WHERE post_id = ?;`, id)

	if err != nil {
		log.Println(err.Error())
		switch {
		case err == sql.ErrNoRows:
			err = errors.New("Not Found")
			return Memo{}, err
		case err != nil:
			log.Fatal(err)
			return Memo{}, err
		default:
			err = nil
		}
	}

	return memo, err
}

func (m *memoAPI) GetMemos(args *MemoGetArgs) (memos []MemoDetail, err error) {

	// Implementation of business logic
	// Get paid data for projects and roles data for member
	var (
		roleProject string
		roleArgs    []interface{}
		isAdmin     bool
	)
	roleArgs = append(roleArgs, args.MemberID, args.MemberID)
	if len(args.Project) > 0 {
		roleProject = fmt.Sprintf("AND %s IN (?)", "object_id")
		roleArgs = append(roleArgs, args.Project)
	}

	// User payment in points would be negative values
	roleQuery := fmt.Sprintf(`SELECT user.id, user.role, projects.object_type, projects.object_id, points.points FROM
		(SELECT id, role FROM members WHERE id = ?) AS user
		LEFT JOIN (SELECT DISTINCT member_id, object_type, object_id FROM points WHERE member_id = ? AND object_type = 2 %s) AS projects ON projects.member_id = user.id
		LEFT JOIN (SELECT DISTINCT object_id, points FROM points WHERE points >= 0) as points ON points.object_id = projects.object_id;`, roleProject)

	roleResult := []struct {
		ID         int64         `db:"id"`
		Role       rrsql.NullInt `db:"role"`
		ObjectType *int          `db:"object_type"`
		ObjectID   *int          `db:"object_id"`
		Points     *int          `db:"points"`
	}{}
	roleQuery, roleArgs, err = sqlx.In(roleQuery, roleArgs...)
	roleQuery = rrsql.DB.Rebind(roleQuery)
	if err = rrsql.DB.Select(&roleResult, roleQuery, roleArgs...); err != nil {
		return []MemoDetail{}, err
	}

	if len(roleResult) > 0 && roleResult[0].Role.Valid {
		isAdmin = (roleResult[0].Role.Int == 9)
	}

	if args.IsAdmin {
		isAdmin = true
	}

	projectTags := rrsql.GetStructDBTags("full", Project{})
	projectField := rrsql.MakeFieldString("get", `project.%s "project.%s"`, projectTags)
	projectIDQuery := strings.Split(projectField[0], " ")
	projectPostQuery := strings.Split(projectField[5], " ")
	projectField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectIDQuery[0], projectIDQuery[1])
	projectField[5] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectPostQuery[0], projectPostQuery[1])

	memberTags := rrsql.GetStructDBTags("full", Member{})
	memberField := rrsql.MakeFieldString("get", `author.%s "author.%s"`, memberTags)
	memberIDQuery := strings.Split(memberField[0], " ")
	memberMemberIDQuery := strings.Split(memberField[1], " ")
	memberUUIDQuery := strings.Split(memberField[2], " ")
	memberField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, memberIDQuery[0], memberIDQuery[1])
	memberField[1] = fmt.Sprintf(`IFNULL(%s, "") %s`, memberMemberIDQuery[0], memberMemberIDQuery[1])
	memberField[2] = fmt.Sprintf(`IFNULL(%s, "") %s`, memberUUIDQuery[0], memberUUIDQuery[1])

	restricts, values := args.parse()

	limit, largs := args.parseLimit()
	values = append(values, largs...)

	rawQuery := fmt.Sprintf(`
		SELECT posts.*, %s, %s FROM 
		(SELECT posts.* FROM posts LEFT JOIN projects AS project ON project.project_id = posts.project_id %s %s) AS posts 
		LEFT JOIN members AS author ON author.id = posts.author 
		LEFT JOIN projects AS project ON project.project_id = posts.project_id %s;`,
		strings.Join(projectField, ","), strings.Join(memberField, ","), restricts, limit["full"], limit["order"])

	query, sqlArgs, err := sqlx.In(rawQuery, values...)
	if err != nil {
		log.Println("GetMemos generate sql error", err)
		return nil, err
	}
	query = rrsql.DB.Rebind(query)
	rows, err := rrsql.DB.Queryx(query, sqlArgs...)
	if err != nil {
		log.Println("GetMemos query db error", err)
		return nil, err
	}

	memos = []MemoDetail{}

	for rows.Next() {
		var (
			abstract string
			fulltext string
			memo     MemoDetail
		)
		if err = rows.StructScan(&memo); err != nil {
			log.Println("GetMemos scan result error", err)
			return []MemoDetail{}, err
		}

		if memo.Content.Valid {
			fulltext = memo.Content.String

			abstract, _ = utils.CutAbstract(memo.Content.String, args.AbstractLength, func(abstact string) string {
				return fmt.Sprintf("<p>%s</p>", abstact)
			})
			// Default show abstract
			memo.Content.String = abstract
		}
		// Default show abstract
		// Show full content if
		// 1. User is admin
		// 2. The memo belongs to a finished project
		// 3. User paid for this project
		if isAdmin {
			// fmt.Println("first process")
			memo.Content.String = fulltext
		} else if memo.Project.Status.Valid && memo.Project.Status.Int == 2 {
			// fmt.Println("second process")
			memo.Content.String = fulltext
		}

		for _, project := range roleResult {
			if project.ObjectID != nil && memo.Project.ID == *project.ObjectID && project.Points != nil {
				// fmt.Println("paid!")
				memo.Project.Paid = true
				memo.Content.String = fulltext
				break
			}
		}
		memos = append(memos, memo)
	}
	return memos, err
}

func (m *memoAPI) InsertMemo(memo Memo) (lastID int, err error) {

	memo.Type = rrsql.NullInt{int64(config.Config.Models.PostType["memo"]), true}

	tags := rrsql.GetStructDBTags("full", Memo{})
	query := fmt.Sprintf(`INSERT INTO posts (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	result, err := rrsql.DB.NamedExec(query, memo)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return lastID, errors.New("Duplicate entry")
		}
		return lastID, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return lastID, errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return lastID, errors.New("Post Not Found")
	}
	lastid, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a memo: %v", err)
		return lastID, err
	}
	lastID = int(lastid)
	if memo.ID == 0 {
		memo.ID = uint32(lastID)
	}

	return lastID, err
}

func (m *memoAPI) UpdateMemo(memo Memo) (err error) {

	tags := rrsql.GetStructDBTags("partial", memo)
	fields := rrsql.MakeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE posts SET %s WHERE post_id = :post_id`,
		strings.Join(fields, ", "))

	result, err := rrsql.DB.NamedExec(query, memo)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Not Found")
	}

	return err
}

func (m *memoAPI) UpdateMemos(args MemoUpdateArgs) (err error) {

	updateQuery, updateArgs := args.parse()
	updateQuery = fmt.Sprintf("UPDATE posts SET %s ", updateQuery)

	restrictQuery, restrictArgs, err := sqlx.In(`WHERE post_id IN (?)`, args.IDs)
	if err != nil {
		return err
	}
	restrictQuery = rrsql.DB.Rebind(restrictQuery)
	updateArgs = append(updateArgs, restrictArgs...)

	result, err := rrsql.DB.Exec(fmt.Sprintf("%s %s", updateQuery, restrictQuery), updateArgs...)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > int64(len(args.IDs)) {
		return errors.New("More Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Posts Not Found")
	}

	return nil
}

func (m *memoAPI) SchedulePublish() (ids []int, err error) {

	rows, err := rrsql.DB.Queryx(fmt.Sprintf("SELECT post_id FROM posts WHERE publish_status=3 AND published_at <= cast(now() as datetime) AND type = %d;;", config.Config.Models.PostType["memo"]))
	if err != nil {
		log.Println("Getting posts error when schedule publishing memos", err)
		return ids, err
	}

	for rows.Next() {
		var i int
		if err = rows.Scan(&i); err != nil {
			continue
		}
		ids = append(ids, i)
	}

	_, err = rrsql.DB.Exec(fmt.Sprintf("UPDATE posts SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime) AND type = %d;", config.Config.Models.PostType["memo"]))
	if err != nil {
		return ids, err
	}
	return ids, nil
}

var MemoAPI MemoInterface = new(memoAPI)
