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
	CommentPush  NullBool `json:"comment_push" db:"comment_push"`
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
	GetUUIDsByNickname(key string, roles map[string][]int) (result []NicknameUUID, err error)
}

type NicknameUUID struct {
	UUID     string     `json:"uuid"`
	Nickname NullString `json:"nickname"`
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

	result, err := DB.Exec(fmt.Sprintf("UPDATE members SET active = %d WHERE %s = ?", int(MemberStatus["delete"].(float64)), idType), id)
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
func (a *memberAPI) GetUUIDsByNickname(key string, roles map[string][]int) (result []NicknameUUID, err error) {
	query := `SELECT uuid, nickname FROM members WHERE active = ? AND nickname LIKE ?`
	if len(roles) != 0 {
		values := []interface{}{int(MemberStatus["active"].(float64)), key + "%"}
		for k, v := range roles {
			query = fmt.Sprintf("%s %s", query, fmt.Sprintf(" AND %s %s (?)", "members.role", operatorHelper(k)))
			values = append(values, v)
		}

		query, values, err := sqlx.In(query, values...)
		if err != nil {
			log.Println(err)
			return []NicknameUUID{}, err
		}

		query = DB.Rebind(query)

		err = DB.Select(&result, query, values...)
		if err != nil {
			return []NicknameUUID{}, err
		}
	} else {
		err = DB.Select(&result, query, int(MemberStatus["active"].(float64)), key+"%")
		if err != nil {
			return []NicknameUUID{}, err
		}
	}
	return result, err
}
