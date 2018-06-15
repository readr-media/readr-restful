package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
)

var MemoStatus map[string]interface{}
var MemoPublishStatus map[string]interface{}

type Memo struct {
	ID            int        `json:"id" db:"memo_id"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
	CommentAmount NullInt    `json:"comment_amount" db:"comment_amount"`
	Title         NullString `json:"title" db:"title"`
	Content       NullString `json:"content" db:"content"`
	Link          NullString `json:"link" db:"link"`
	Author        NullInt    `json:"author" db:"author"`
	Project       NullInt    `json:"project_id" db:"project_id"`
	Active        NullInt    `json:"active" db:"active"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     NullInt    `json:"updated_by" db:"updated_by"`
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
	PublishStatus NullInt    `json:"publish_status" db:"publish_status"`
	Order         NullInt    `json:"memo_order" db:"memo_order"`
}

type MemoInterface interface {
	CountMemos(args *MemoGetArgs) (int, error)
	GetMemo(id int) (Memo, error)
	GetMemos(args *MemoGetArgs) ([]MemoDetail, error)
	InsertMemo(memo Memo) error
	UpdateMemo(memo Memo) error
	UpdateMemos(args MemoUpdateArgs) error
	SchedulePublish() error
}

type MemoGetArgs struct {
	MaxResult            int              `form:"max_result"`
	Page                 int              `form:"page"`
	Sorting              string           `form:"sort"`
	Author               []int64          `form:"author"`
	Project              []int64          `form:"project_id"`
	Slugs                []string         `form:"slugs"`
	Active               map[string][]int `form:"active"`
	MemoPublishStatus    map[string][]int `form:"memo_publish_status"`
	ProjectPublishStatus map[string][]int `form:"project_publish_status"`
	MemberID             int64            `form:"member_id"`
	AbstractLength       int64            `form:"abstract_length"`
	MemoID               int64
}

func (p *MemoGetArgs) Default() (result *MemoGetArgs) {
	return &MemoGetArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at", AbstractLength: 20}
}
func (p *MemoGetArgs) DefaultActive() {
	p.Active = map[string][]int{"$nin": []int{int(MemoStatus["deactive"].(float64))}}
}
func (p *MemoGetArgs) Validate() bool {
	if matched, err := regexp.MatchString("-?(updated_at|created_at|published_at|memo_id|author|project_id|memo_order)", p.Sorting); err != nil || !matched {
		return false
	}
	return true
}
func (p *MemoGetArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	// if p.MemoPublishStatus == nil && p.ProjectPublishStatus == nil && p.Active == nil && len(p.Author) == 0 && len(p.Project) == 0 {
	// 	return "", nil
	// }

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "memos.active", operatorHelper(k)))
			values = append(values, v)
		}
	}

	if p.MemoPublishStatus != nil {
		for k, v := range p.MemoPublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "memos.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.ProjectPublishStatus != nil {
		for k, v := range p.ProjectPublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "project.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}

	if len(p.Author) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "memos.author"))
		values = append(values, p.Author)
	}
	if len(p.Project) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "memos.project_id"))
		values = append(values, p.Project)
	}
	if len(p.Slugs) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "project.slug"))
		values = append(values, p.Slugs)
	}

	if p.MemoID != 0 {
		where = append(where, fmt.Sprintf("%s = ?", "memos.memo_id"))
		values = append(values, p.MemoID)
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
		restricts = append(restricts, fmt.Sprintf("ORDER BY %s%s", "memos.", orderByHelper(p.Sorting)))
		limit["order"] = fmt.Sprintf("ORDER BY %s", orderByHelper(p.Sorting))
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
	IDs         []int    `json:"ids"`
	UpdatedBy   int64    `json:"updated_by"`
	UpdatedAt   NullTime `json:"-"`
	PublishedAt NullTime `json:"-"`
	Active      NullInt  `json:"-"`
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
	query := fmt.Sprintf(`SELECT COUNT(memo_id) FROM memos LEFT JOIN projects AS project ON project.project_id = memos.project_id %s`, restricts)

	query, sqlArgs, err := sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = DB.Rebind(query)
	count, err := DB.Queryx(query, sqlArgs...)
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

	err = DB.Get(&memo, `SELECT * FROM memos WHERE memo_id = ?;`, id)

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
		LEFT JOIN (SELECT DISTINCT object_id, points FROM points WHERE points < 0) as points ON points.object_id = projects.object_id;`, roleProject)

	roleResult := []struct {
		ID         int64   `db:"id"`
		Role       NullInt `db:"role"`
		ObjectType *int    `db:"object_type"`
		ObjectID   *int    `db:"object_id"`
		Points     *int    `db:"points"`
	}{}
	roleQuery, roleArgs, err = sqlx.In(roleQuery, roleArgs...)
	roleQuery = DB.Rebind(roleQuery)
	if err = DB.Select(&roleResult, roleQuery, roleArgs...); err != nil {
		return []MemoDetail{}, err
	}

	if len(roleResult) > 0 && roleResult[0].Role.Valid {
		isAdmin = (roleResult[0].Role.Int == 9)
	}

	projectTags := getStructDBTags("full", Project{})
	projectField := makeFieldString("get", `project.%s "project.%s"`, projectTags)
	projectIDQuery := strings.Split(projectField[0], " ")
	projectPostQuery := strings.Split(projectField[5], " ")
	projectField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectIDQuery[0], projectIDQuery[1])
	projectField[5] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectPostQuery[0], projectPostQuery[1])

	memberTags := getStructDBTags("full", Member{})
	memberField := makeFieldString("get", `author.%s "author.%s"`, memberTags)
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
		SELECT memos.*, %s, %s FROM 
		(SELECT memos.* FROM memos LEFT JOIN projects AS project ON project.project_id = memos.project_id %s %s) AS memos 
		LEFT JOIN members AS author ON author.id = memos.author 
		LEFT JOIN projects AS project ON project.project_id = memos.project_id %s;`,
		strings.Join(projectField, ","), strings.Join(memberField, ","), restricts, limit["full"], limit["order"])

	query, sqlArgs, err := sqlx.In(rawQuery, values...)
	if err != nil {
		log.Println("GetMemos generate sql error", err)
		return nil, err
	}
	query = DB.Rebind(query)
	rows, err := DB.Queryx(query, sqlArgs...)
	if err != nil {
		log.Println("GetMemos query db error", err)
		return nil, err
	}

	memos = []MemoDetail{}

	cutAbstract := func(html string, length int64) (result string, err error) {
		// buf := bytes.NewBuffer(strings.NewReader(html))

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return "", err
		}
		content := doc.Find("p:not(:has(img))").First().Text()

		abstract := []rune(content)
		abstract = abstract[:int(math.Min(float64(len(content)), float64(length)))]
		result = fmt.Sprintf("<p>%s</p>", string(abstract))
		return result, nil
	}

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

			abstract, _ = cutAbstract(memo.Content.String, args.AbstractLength)
			// Default show abstract
			memo.Content.String = abstract
		}
		// Default show abstract
		// Show full content if
		// 1. User is admin
		// 2. The memo belongs to a finished project
		// 3. User paid for this project
		if isAdmin {
			memo.Content.String = fulltext
		} else if memo.Project.Status.Valid && memo.Project.PublishStatus.Int == 2 {
			memo.Content.String = fulltext
		} else {
			for _, project := range roleResult {
				if project.ObjectID != nil && memo.Project.ID == *project.ObjectID && project.Points != nil {
					memo.Project.Paid = true
					memo.Content.String = fulltext
					break
				}
			}
		}
		memos = append(memos, memo)
	}
	return memos, err
}
func (m *memoAPI) InsertMemo(memo Memo) (err error) {

	tags := getStructDBTags("full", Memo{})
	query := fmt.Sprintf(`INSERT INTO memos (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	result, err := DB.NamedExec(query, memo)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return errors.New("Duplicate entry")
		}
		return err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Post Not Found")
	}

	return err
}
func (m *memoAPI) UpdateMemo(memo Memo) (err error) {

	tags := getStructDBTags("partial", memo)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE memos SET %s WHERE memo_id = :memo_id`,
		strings.Join(fields, ", "))

	result, err := DB.NamedExec(query, memo)

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
	updateQuery = fmt.Sprintf("UPDATE memos SET %s ", updateQuery)

	restrictQuery, restrictArgs, err := sqlx.In(`WHERE memo_id IN (?)`, args.IDs)
	if err != nil {
		return err
	}
	restrictQuery = DB.Rebind(restrictQuery)
	updateArgs = append(updateArgs, restrictArgs...)

	result, err := DB.Exec(fmt.Sprintf("%s %s", updateQuery, restrictQuery), updateArgs...)
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

func (a *memoAPI) SchedulePublish() error {
	_, err := DB.Exec("UPDATE memos SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime);")
	if err != nil {
		return err
	}
	return nil
}

var MemoAPI MemoInterface = new(memoAPI)
