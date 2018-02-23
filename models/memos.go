package models

import (
	"database/sql"
	//"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

var MemoStatus map[string]interface{}

type Memo struct {
	ID            int        `json:"id" db:"memo_id"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
	CommentAmount NullInt    `json:"comment_amount" db:"comment_amount"`
	Title         NullString `json:"title" db:"title"`
	Content       NullString `json:"content" db:"content"`
	Link          NullString `json:"link" db:"link"`
	Author        NullString `json:"author" db:"author"`
	Project       NullInt    `json:"project_id" db:"project_id"`
	Active        NullInt    `json:"active" db:"active"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     NullString `json:"updated_by" db:"updated_by"`
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
}

type MemoInterface interface {
	CountMemos(args *MemoGetArgs) (int, error)
	GetMemo(id int) (Memo, error)
	GetMemos(args *MemoGetArgs) ([]Memo, error)
	InsertMemo(memo Memo) error
	UpdateMemo(memo Memo) error
	UpdateMemos(args MemoUpdateArgs) error
}

type MemoGetArgs struct {
	MaxResult int              `form:"max_result"`
	Page      int              `form:"page"`
	Sorting   string           `form:"sort"`
	Author    []string         `form:"author"`
	Project   []int            `form:"project_id"`
	Active    map[string][]int `form:"active"`
}

func (p *MemoGetArgs) Default() (result *MemoGetArgs) {
	return &MemoGetArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}
func (p *MemoGetArgs) DefaultActive() {
	p.Active = map[string][]int{"$nin": []int{int(MemoStatus["deactive"].(float64))}}
}
func (p *MemoGetArgs) Validate() bool {
	if matched, err := regexp.MatchString("-?(updated_at|created_at|published_at|memo_id|author|project_id)", p.Sorting); err != nil || !matched {
		return false
	}
	return true
}
func (p *MemoGetArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Active == nil && len(p.Author) == 0 && len(p.Project) == 0 {
		return "", nil
	}

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.Author) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "author"))
		values = append(values, p.Author)
		log.Println(p.Author)
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
	UpdatedBy   string   `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt   NullTime `json:"-" db:"updated_at"`
	PublishedAt NullTime `json:"-" db:"published_at"`
	Active      NullInt  `json:"-" db:"active"`
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
	fmt.Println(query, sqlArgs)
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
			err = errors.New("Post Not Found")
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

	log.Println(query, sqlArgs)

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

	//log.Println(query)
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
		return errors.New("Post Not Found")
	}

	return err
}
func (m *memoAPI) UpdateMemos(args MemoUpdateArgs) (err error) {

	query, sqlArgs, err := sqlx.In(`UPDATE memos SET updated_by = ?, updated_at = ?, active = ? WHERE memo_id IN (?)`, args.UpdatedBy, args.UpdatedAt, args.Active, args.IDs)
	if err != nil {
		return err
	}
	query = DB.Rebind(query)
	result, err := DB.Exec(query, sqlArgs...)
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

var MemoAPI MemoInterface = new(memoAPI)
