package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

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
	GetMemos(args *MemoGetArgs) ([]Memo, error)
	InsertMemo(memo Memo) error
	UpdateMemo(memo Memo) error
	UpdateMemos(args MemoUpdateArgs) error
	SchedulePublish() error
}

type MemoGetArgs struct {
	MaxResult     int              `form:"max_result"`
	Page          int              `form:"page"`
	Sorting       string           `form:"sort"`
	Author        []int64          `form:"author"`
	Project       []int64          `form:"project_id"`
	Active        map[string][]int `form:"active"`
	PublishStatus map[string][]int `form:"publish_status"`
}

func (p *MemoGetArgs) Default() (result *MemoGetArgs) {
	return &MemoGetArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
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

	if p.PublishStatus == nil && p.Active == nil && len(p.Author) == 0 && len(p.Project) == 0 {
		return "", nil
	}

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "active", operatorHelper(k)))
			values = append(values, v)
		}
	}

	if p.PublishStatus != nil {
		for k, v := range p.PublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}

	if len(p.Author) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "author"))
		values = append(values, p.Author)
	}
	if len(p.Project) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "project_id"))
		values = append(values, p.Project)
	}
	if len(where) > 1 {
		restricts = fmt.Sprintf(" WHERE %s", strings.Join(where, " AND "))
	} else if len(where) == 1 {
		restricts = fmt.Sprintf(" WHERE %s", where[0])
	}
	return restricts, values
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

type memoAPI struct{}

func (m *memoAPI) CountMemos(args *MemoGetArgs) (result int, err error) {

	restricts, values := args.parse()
	query := fmt.Sprintf(`SELECT COUNT(memo_id) FROM memos %s`, restricts)

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
	result := Memo{}

	err = DB.Get(&result, `SELECT * FROM memos WHERE memo_id = ?;`, id)
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

	return result, err
}
func (m *memoAPI) GetMemos(args *MemoGetArgs) (memos []Memo, err error) {

	restricts, values := args.parse()

	rawQuery := fmt.Sprintf(`SELECT * FROM memos %s `, restricts)
	query, sqlArgs, err := sqlx.In(rawQuery, values...)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)

	query = query + fmt.Sprintf(`ORDER BY %s LIMIT ? OFFSET ?`, orderByHelper(args.Sorting))
	sqlArgs = append(sqlArgs, args.MaxResult, (args.Page-1)*args.MaxResult)

	rows, err := DB.Queryx(query, sqlArgs...)
	if err != nil {
		return nil, err
	}

	memos = []Memo{}
	for rows.Next() {
		var memo Memo
		if err = rows.StructScan(&memo); err != nil {
			return []Memo{}, err
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
