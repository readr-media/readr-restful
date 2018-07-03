package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

// var MemberStatus map[string]interface{}

type Member struct {
	ID       int64      `json:"id" db:"id"`
	MemberID string     `json:"member_id" db:"member_id"`
	UUID     string     `json:"uuid" db:"uuid"`
	Points   NullInt    `json:"points" db:"points"`
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
	UpdatedBy NullInt    `json:"updated_by" db:"updated_by"`
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
	DailyPush    NullBool `json:"daily_push" db:"daily_push"`
	CommentPush  NullBool `json:"comment_push" db:"comment_push"`
}

// Stunt could be regarded as an experimental, pre-transitional wrap of Member, which provide omitempty tag for json
// and Use *Null type instead of Null type to made omitempty work
// In this way we could control the fields returned by update SQL select fields
type Stunt struct {
	// Make ID, MemberID, UUID pointer to avoid situation we have to use IFNULL
	ID       *int64      `json:"id,omitempty" db:"id"`
	MemberID *string     `json:"member_id,omitempty" db:"member_id"`
	UUID     *string     `json:"uuid,omitempty" db:"uuid"`
	Points   *NullInt    `json:"points,omitempty" db:"points"`
	Name     *NullString `json:"name,omitempty" db:"name"`
	Nickname *NullString `json:"nickname,omitempty" db:"nickname"`

	Birthday *NullTime   `json:"birthday,omitempty" db:"birthday"`
	Gender   *NullString `json:"gender,omitempty" db:"gender"`
	Work     *NullString `json:"work,omitempty" db:"work"`
	Mail     *NullString `json:"mail,omitempty" db:"mail"`

	RegisterMode *NullString `json:"register_mode,omitempty" db:"register_mode"`
	SocialID     *NullString `json:"social_id,omitempty,omitempty" db:"social_id"`
	TalkID       *NullString `json:"talk_id,omitempty" db:"talk_id"`

	CreatedAt *NullTime  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *NullTime  `json:"updated_at,omitempty" db:"updated_at"`
	UpdatedBy *NullInt   `json:"updated_by,omitempty" db:"updated_by"`
	Password  NullString `json:"-" db:"password"`
	Salt      NullString `json:"-" db:"salt"`

	Description  *NullString `json:"description,omitempty" db:"description"`
	ProfileImage *NullString `json:"profile_image,omitempty" db:"profile_image"`
	Identity     *NullString `json:"identity,omitempty" db:"identity"`

	Role   *NullInt `json:"role,omitempty" db:"role"`
	Active *NullInt `json:"active,omitempty" db:"active"`

	CustomEditor *NullBool `json:"custom_editor,omitempty" db:"custom_editor"`
	HideProfile  *NullBool `json:"hide_profile,omitempty" db:"hide_profile"`
	ProfilePush  *NullBool `json:"profile_push,omitempty" db:"profile_push"`
	PostPush     *NullBool `json:"post_push,omitempty" db:"post_push"`
	DailyPush    *NullBool `json:"daily_push,omitempty" db:"daily_push"`
	CommentPush  *NullBool `json:"comment_push,omitempty" db:"comment_push"`
}

// Separate API and Member struct
type memberAPI struct{}

var MemberAPI MemberInterface = new(memberAPI)

type MemberInterface interface {
	DeleteMember(idType string, id string) error
	GetMember(idType string, id string) (Member, error)
	GetMembers(req *MemberArgs) ([]Member, error)
	InsertMember(m Member) (id int, err error)
	UpdateAll(ids []int64, active int) error
	UpdateMember(m Member) error
	Count(req *MemberArgs) (result int, err error)
	GetIDsByNickname(params GetMembersKeywordsArgs) (result []Stunt, err error)
}

// type MemberArgs map[string]interface{}
type MemberArgs struct {
	MaxResult    uint8            `form:"max_result"`
	Page         uint16           `form:"page"`
	Sorting      string           `form:"sort"`
	CustomEditor bool             `form:"custom_editor"`
	Role         *int64           `form:"role"`
	Active       map[string][]int `form:"active"`
	IDs          []string         `form:"ids"`
	UUIDs        []string         `form:"uuids"`
}

func (m *MemberArgs) SetDefault() {
	m.MaxResult = 20
	m.Page = 1
	m.Sorting = "-updated_at"
}

func (m *MemberArgs) DefaultActive() {
	// m.Active = map[string][]int{"$nin": []int{int(MemberStatus["delete"].(float64))}}
	m.Active = map[string][]int{"$nin": []int{config.Config.Models.Members["delete"]}}
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
		where = append(where, fmt.Sprintf("members.id IN (%s)", strings.Join(a, ", ")))
		for i := range m.IDs {
			values = append(values, m.IDs[i])
		}
	}
	if len(m.UUIDs) > 0 {
		a := make([]string, len(m.UUIDs))
		for i := range a {
			a[i] = "?"
		}
		where = append(where, fmt.Sprintf("members.uuid IN (%s)", strings.Join(a, ", ")))
		for i := range m.UUIDs {
			values = append(values, m.UUIDs[i])
		}
	}
	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

type GetMembersKeywordsArgs struct {
	Keywords string `form:"keyword"`
	Roles    map[string][]int
	Fields   sqlfields
}

func (a *GetMembersKeywordsArgs) Validate() (err error) {
	// Validate keyword
	if a.Keywords == "" {
		return errors.New("Invalid keyword")
	}
	// Validate field
	validFields := getStructDBTags("full", Stunt{})

CheckEachFieldLoop:
	for _, f := range a.Fields {
		for _, F := range validFields {
			if f == F {
				continue CheckEachFieldLoop
			}
		}
		return fmt.Errorf("Invalid fields: %s", f)
	}
	var containfield = func(field string) bool {
		for _, f := range a.Fields {
			if f == field {
				return true
			}
		}
		return false
	}
	// Set default fields id & nickname
	for _, fs := range []string{"id", "nickname"} {
		if !containfield(fs) {
			a.Fields = append(a.Fields, fs)
		}
	}
	return err
}

func (a *memberAPI) GetMembers(req *MemberArgs) (result []Member, err error) {

	restricts, values := req.parse()
	query := fmt.Sprintf(`SELECT * FROM members where %s `, restricts)

	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return []Member{}, err
	}
	query = DB.Rebind(query)
	query = query + fmt.Sprintf(`ORDER BY %s LIMIT ? OFFSET ?`, orderByHelper(req.Sorting))
	args = append(args, req.MaxResult, (req.Page-1)*uint16(req.MaxResult))
	err = DB.Select(&result, query, args...)
	if err != nil {
		return []Member{}, err
	}
	if len(result) == 0 {
		return []Member{}, nil
	}
	return result, err
}

func (a *memberAPI) GetMember(idType string, id string) (Member, error) {
	member := Member{}
	err := DB.QueryRowx(fmt.Sprintf("SELECT * FROM members where %s = ?", idType), id).StructScan(&member)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("User Not Found")
		member = Member{}
	case err != nil:
		log.Fatal(err)
		member = Member{}
	default:
		err = nil
	}
	return member, err
}

func (a *memberAPI) InsertMember(m Member) (id int, err error) {
	existedID := 0
	err = DB.Get(&existedID, `SELECT id FROM members WHERE id=? OR member_id=? LIMIT 1;`, m.ID, m.MemberID)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}
	if existedID != 0 {
		return 0, errors.New("Duplicate entry")
	}

	tags := getStructDBTags("full", Member{})
	query := fmt.Sprintf(`INSERT INTO members (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))
	result, err := DB.NamedExec(query, m)
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
		return 0, errors.New("No Row Inserted")
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last inserted ID when insert a member: %v", err)
		return 0, err
	}
	return int(lastID), nil
}

func (a *memberAPI) UpdateMember(m Member) error {
	// query, _ := generateSQLStmt("partial_update", "members", m)
	tags := getStructDBTags("partial", m)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE members SET %s WHERE id = :id`, strings.Join(fields, ", "))
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

func (a *memberAPI) DeleteMember(idType string, id string) error {

	// result, err := DB.Exec(fmt.Sprintf("UPDATE members SET active = %d WHERE %s = ?", int(MemberStatus["delete"].(float64)), idType), id)
	result, err := DB.Exec(fmt.Sprintf("UPDATE members SET active = %d WHERE %s = ?", config.Config.Models.Members["delete"], idType), id)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("User Not Found")
	}
	return err
}

func (a *memberAPI) UpdateAll(ids []int64, active int) (err error) {
	prep := fmt.Sprintf("UPDATE members SET active = %d WHERE id IN (?);", active)
	query, args, err := sqlx.In(prep, ids)
	if err != nil {
		return err
	}
	query = DB.Rebind(query)
	result, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
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
		query := fmt.Sprintf(`SELECT COUNT(*) FROM (SELECT id FROM members WHERE %s) AS subquery`, restricts)

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

// GetMembersByNickname select nickname and uuid from active members only
// when their nickname fits certain keyword
func (a *memberAPI) GetIDsByNickname(params GetMembersKeywordsArgs) (result []Stunt, err error) {

	query := fmt.Sprintf(`SELECT %s FROM members WHERE active = ? AND nickname LIKE ?`, strings.Join(params.Fields, ", "))
	values := []interface{}{}
	values = append(values, config.Config.Models.Members["active"], params.Keywords+"%")

	if len(params.Roles) != 0 {
		for k, v := range params.Roles {
			query = fmt.Sprintf("%s AND %s %s (?)", query, "members.role", operatorHelper(k))
			values = append(values, v)
		}
	}
	query, values, err = sqlx.In(query, values...)
	query = DB.Rebind(query)
	if err = DB.Select(&result, query, values...); err != nil {
		return []Stunt{}, err
	}
	return result, err
}
