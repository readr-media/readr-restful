package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Member struct {
	ID       string     `json:"id" db:"user_id"`
	Name     NullString `json:"name" db:"name"`
	Nickname NullString `json:"nickname" db:"nick"`
	// Cannot parse Date format
	Birthday NullTime   `json:"birthday" db:"birthday"`
	Gender   NullString `json:"gender" db:"gender"`
	Work     NullString `json:"occupation" db:"work"`
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
	ProfileImage NullString `json:"profile_image" db:"profile_picture"`
	Identity     NullString `json:"identity" db:"identity"`

	Role   NullInt `json:"role" db:"role"`
	Active NullInt `json:"active" db:"active"`

	CustomEditor NullBool `json:"custom_editor" db:"c_editor"`
	HideProfile  NullBool `json:"hide_profile" db:"hide_profile"`
	ProfilePush  NullBool `json:"profile_push" db:"profile_push"`
	PostPush     NullBool `json:"post_push" db:"post_push"`
	CommentPush  NullBool `json:"comment_push" db:"comment_push"`
}

// Separate API and Member struct
type memberAPI struct{}

var MemberAPI MemberInterface = new(memberAPI)

type MemberInterface interface {
	GetMembers(maxResult uint8, page uint16, sortMethod string) ([]Member, error)
	GetMember(id string) (Member, error)
	InsertMember(m Member) error
	UpdateMember(m Member) error
	DeleteMember(id string) error
}

func (a *memberAPI) GetMembers(maxResult uint8, page uint16, sortMethod string) ([]Member, error) {
	var (
		result     []Member
		err        error
		sortString string
	)
	switch sortMethod {
	case "updated_at":
		sortString = "updated_at"
	case "-updated_at":
		sortString = "updated_at DESC"
	default:
		sortString = "updated_at DESC"
	}
	// limitBase := (page - 1) * uint16(maxResult)
	// limitIncrement := page * uint16(maxResult)
	query, _ := generateSQLStmt("get_all", "members", sortString)

	// fmt.Println("sortString: ", sortString, " , limitBase: ", limitBase, " , limitIncrement: ", limitIncrement)
	// fmt.Println(limitBase, limitIncrement)
	err = DB.Select(&result, query, (page-1)*uint16(maxResult), maxResult)
	if err != nil || len(result) == 0 {
		result = []Member{}
		err = errors.New("Members Not Found")
	}
	// fmt.Println(result)
	return result, err
}

func (a *memberAPI) GetMember(id string) (Member, error) {
	member := Member{}
	err := DB.QueryRowx("SELECT * FROM members where user_id = ?", id).StructScan(&member)
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
	query, _ := generateSQLStmt("insert", "members", m)

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
	query, _ := generateSQLStmt("partial_update", "members", m)
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
	result, err := DB.Exec("UPDATE members SET active = 0 WHERE user_id = ?", id)
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
