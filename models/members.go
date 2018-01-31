package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

var MemberStatus map[string]interface{}

type Member struct {
	ID       string     `json:"id" db:"member_id"`
	Name     NullString `json:"name" db:"name"`
	Nickname NullString `json:"nickname" db:"nickname"`
	// Cannot parse Date format
	Birthday NullTime   `json:"birthday" db:"birthday"`
	Gender   NullString `json:"gender" db:"gender"`
	Work     NullString `json:"work" db:"work"`
	Mail     NullString `json:"mail" db:"mail"`

	RegisterMode NullString `json:"register_mode" db:"register_mode"`
	SocialID     NullString `json:"social_id,omitempty" db:"social_id"`
	CreatedAt    NullTime   `json:"created_at" db:"created_at"`
	UpdatedAt    NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy    NullString `json:"updated_by" db:"updated_by"`
	Password     NullString `json:"-" db:"password"`
	Salt         NullString `json:"-" db:"salt"`
	// Ignore password JSON marshall for now

	Description  NullString `json:"description" db:"description"`
	ProfileImage NullString `json:"profile_image" db:"profile_image"`
	Identity     NullString `json:"identity" db:"identity"`

	Role   NullInt `json:"role" db:"role"`
	Active NullInt `json:"active" db:"active"`

	CustomEditor NullBool `json:"custom_editor" db:"custom_editor"`
	HideProfile  NullBool `json:"hide_profile" db:"hide_profile"`
	ProfilePush  NullBool `json:"profile_push" db:"profile_push"`
	PostPush     NullBool `json:"post_push" db:"post_push"`
	CommentPush  NullBool `json:"comment_push" db:"comment_push"`
}

// Separate API and Member struct
type memberAPI struct{}

var MemberAPI MemberInterface = new(memberAPI)

type MemberInterface interface {
	DeleteMember(id string) error
	GetMember(id string) (Member, error)
	GetMembers(req MemberArgs) ([]Member, error)
	InsertMember(m Member) error
	SetMultipleActive(ids []string, active int) error
	UpdateMember(m Member) error
	Count(args MemberMap) (result int, err error)
}

type MemberMap map[string]interface{}

func (m *MemberMap) parse() (restricts string, values []interface{}) {
	whereString := make([]string, 0)
	for arg, value := range *m {
		switch arg {
		case "custom_editor":
			whereString = append(whereString, "custom_editor = ?")
			values = append(values, value)
		case "active":
			for k, v := range value.(map[string][]int) {
				whereString = append(whereString, fmt.Sprintf("%s %s (?)", arg, operatorHelper(k)))
				values = append(values, v)
			}
		}
	}

	if len(whereString) > 1 {
		restricts = strings.Join(whereString, " AND ")
	} else if len(whereString) == 1 {
		restricts = whereString[0]
	}
	return restricts, values
}

func (m *MemberMap) ValidateActive(args map[string][]int) (err error) {

	if len(args) > 1 {
		return errors.New("Too many active lists")
	}
	valid := make([]int, 0)
	result := make([]int, 0)
	for _, v := range MemberStatus {
		valid = append(valid, int(v.(float64)))
	}
	activeCount := 0
	// Extract active slice from map
	for _, activeSlice := range args {
		activeCount = len(activeSlice)
		for _, target := range activeSlice {
			for _, value := range valid {
				if target == value {
					result = append(result, target)
				}
			}
		}
	}
	if len(result) != activeCount {
		err = errors.New("Not all active elements are valid")
	}
	if len(result) == 0 {
		err = errors.New("No valid active request")
	}
	return err
}

// MemberArgs is used to hold url/payload parameters while querys
// NullBool types should be put last, so that default type like string,
// could be parsed during form parsing
type MemberArgs struct {
	BasicArgs
	Active       map[string][]int `form:"active" db:"active"`
	CustomEditor NullBool         `form:"custom_editor" db:"custom_editor"`
}

func (args *MemberArgs) parse() (restricts string, values []interface{}) {
	input := reflect.ValueOf(args).Elem()
	whereString := make([]string, 0)
	for i := 0; i < input.NumField(); i++ {
		tag := input.Type().Field(i).Tag.Get("form")
		switch tag {
		case "custom_editor":
			if args.CustomEditor.Valid {
				whereString = append(whereString, "custom_editor = ?")
				values = append(values, args.CustomEditor.Bool)
			}
		case "active":
			if args.Active != nil {
				for k, v := range args.Active {
					whereString = append(whereString, fmt.Sprintf("%s %s (?)", tag, operatorHelper(k)))
					values = append(values, v)
				}
			}
		}
	}

	if len(whereString) > 1 {
		restricts = strings.Join(whereString, " AND ")
	} else if len(whereString) == 1 {
		restricts = whereString[0]
	}
	return restricts, values
}

func (args MemberArgs) ValidateActive() (err error) {

	if len(args.Active) > 1 {
		return errors.New("Too many active lists")
	}
	valid := make([]int, 0)
	result := make([]int, 0)
	for _, v := range MemberStatus {
		valid = append(valid, int(v.(float64)))
	}
	activeCount := 0
	// Extract active slice from map
	for _, activeSlice := range args.Active {
		activeCount = len(activeSlice)
		for _, target := range activeSlice {
			for _, value := range valid {
				if target == value {
					result = append(result, target)
				}
			}
		}
	}
	if len(result) != activeCount {
		err = errors.New("Not all active elements are valid")
	}
	if len(result) == 0 {
		err = errors.New("No valid active request")
	}
	return err
}

func (a *memberAPI) GetMembers(req MemberArgs) (result []Member, err error) {

	restricts, values := req.parse()
	query := fmt.Sprintf(`SELECT * FROM members where %s `, restricts)
	// query := fmt.Sprintf(`SELECT * FROM members where active != %d ORDER BY %s LIMIT ? OFFSET ?`, int(MemberStatus["delete"].(float64)), orderByHelper(req.Sorting))

	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)
	query = query + fmt.Sprintf(`ORDER BY %s LIMIT ? OFFSET ?`, orderByHelper(req.Sorting))
	args = append(args, req.MaxResult, (req.Page-1)*uint16(req.MaxResult))

	err = DB.Select(&result, query, args...)
	if err != nil || len(result) == 0 {
		result = []Member{}
		err = errors.New("Members Not Found")
	}

	return result, err
}

func (a *memberAPI) GetMember(id string) (Member, error) {
	member := Member{}
	err := DB.QueryRowx("SELECT * FROM members where member_id = ?", id).StructScan(&member)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("User Not Found")
		member = Member{}
	case err != nil:
		log.Fatal(err)
		member = Member{}
	default:
		fmt.Printf("Successful get user: %s\n", id)
		err = nil
	}
	return member, err
}

func (a *memberAPI) InsertMember(m Member) error {

	tags := getStructDBTags("full", Member{})
	query := fmt.Sprintf(`INSERT INTO members (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))
	// query, _ := generateSQLStmt("insert", "members", m)
	result, err := DB.NamedExec(query, m)

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
		return errors.New("No Row Inserted")
	}
	return nil
}

func (a *memberAPI) UpdateMember(m Member) error {
	// query, _ := generateSQLStmt("partial_update", "members", m)
	tags := getStructDBTags("partial", m)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE members SET %s WHERE member_id = :member_id`, strings.Join(fields, ", "))
	result, err := DB.NamedExec(query, m)

	if err != nil {
		log.Fatal(err)
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("User Not Found")
	}
	return nil
}

func (a *memberAPI) DeleteMember(id string) error {

	// result := Member{}
	result, err := DB.Exec(fmt.Sprintf("UPDATE members SET active = %d WHERE member_id = ?", int(MemberStatus["delete"].(float64))), id)
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

func (a *memberAPI) SetMultipleActive(ids []string, active int) (err error) {
	prep := fmt.Sprintf("UPDATE members SET active = %d WHERE member_id IN (?);", active)
	query, args, err := sqlx.In(prep, ids)
	if err != nil {
		return err
	}
	query = DB.Rebind(query)
	result, err := DB.Exec(query, args...)
	if err != nil {
		fmt.Println(err)
		return err
	}
	rowCnt, err := result.RowsAffected()
	fmt.Println(rowCnt)
	if rowCnt > int64(len(ids)) {
		return errors.New("More Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Members Not Found")
	}
	return err
}

func (a *memberAPI) Count(args MemberMap) (result int, err error) {

	if len(args) == 0 {

		rows, err := DB.Queryx(`SELECT COUNT(*) FROM (SELECT member_id FROM members) AS subquery`)
		if err != nil {
			return 0, err
		}
		for rows.Next() {
			err = rows.Scan(&result)
		}

	} else if len(args) > 0 {

		restricts, values := args.parse()
		query := fmt.Sprintf(`SELECT COUNT(*) FROM (SELECT member_id FROM members WHERE %s) AS subquery`, restricts)

		query, args, err := sqlx.In(query, values...)
		if err != nil {
			return 0, err
		}
		query = DB.Rebind(query)
		// fmt.Println(query, args)
		count, err := DB.Queryx(query, args...)
		if err != nil {
			return 0, err
		}
		for count.Next() {
			if err = count.Scan(&result); err != nil {
				return 0, err
			}
		}
	}
	return result, err
}
