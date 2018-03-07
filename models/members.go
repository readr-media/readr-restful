package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	TalkID       NullString `json:"talk_id" db:"talk_id"`

	CreatedAt NullTime   `json:"created_at" db:"created_at"`
	UpdatedAt NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy NullString `json:"updated_by" db:"updated_by"`
	Password  NullString `json:"-" db:"password"`
	Salt      NullString `json:"-" db:"salt"`
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
	GetMembers(req *MemberArgs) ([]Member, error)
	InsertMember(m Member) error
	UpdateAll(ids []string, active int) error
	UpdateMember(m Member) error
	Count(req *MemberArgs) (result int, err error)
}

// type MemberArgs map[string]interface{}
type MemberArgs struct {
	MaxResult    uint8            `form:"max_result"`
	Page         uint16           `form:"page"`
	Sorting      string           `form:"sort"`
	CustomEditor bool             `form:"custom_editor"`
	Active       map[string][]int `form:"active"`
	Role         *int64           `form:"role"`
	IDs          []string         `form:"ids"`
}

func (m *MemberArgs) Default() (result *MemberArgs) {
	return &MemberArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (m *MemberArgs) DefaultActive() {
	m.Active = map[string][]int{"$nin": []int{int(MemberStatus["delete"].(float64))}}
}

func (m *MemberArgs) anyFilter() bool {
	return m.Active != nil || m.CustomEditor == true
}

func (m *MemberArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if m.CustomEditor {
		where = append(where, "custom_editor = ?")
		values = append(values, m.CustomEditor)
	}
	if m.Active != nil {
		for k, v := range m.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "members.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if m.Role != nil {
		where = append(where, "role = ?")
		values = append(values, *m.Role)
	}
	if len(m.IDs) > 0 {
		a := make([]string, len(m.IDs))
		for i := range a {
			a[i] = "?"
		}
		where = append(where, fmt.Sprintf("members.member_id IN (%s)", strings.Join(a, ", ")))
		for i := range m.IDs {
			values = append(values, m.IDs[i])
		}
	}
	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (a *memberAPI) GetMembers(req *MemberArgs) (result []Member, err error) {

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
	// fmt.Println(query, args)
	err = DB.Select(&result, query, args...)
	if err != nil {
		return []Member{}, err
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

func (a *memberAPI) UpdateAll(ids []string, active int) (err error) {
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

func (a *memberAPI) Count(req *MemberArgs) (result int, err error) {

	if !req.anyFilter() {

		rows, err := DB.Queryx(`SELECT COUNT(*) FROM members`)
		if err != nil {
			return 0, err
		}
		for rows.Next() {
			err = rows.Scan(&result)
		}

	} else {

		restricts, values := req.parse()
		query := fmt.Sprintf(`SELECT COUNT(*) FROM (SELECT member_id FROM members WHERE %s) AS subquery`, restricts)

		query, args, err := sqlx.In(query, values...)
		if err != nil {
			return 0, err
		}
		query = DB.Rebind(query)
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
