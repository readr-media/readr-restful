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
	CreateTime   NullTime   `json:"created_at" db:"create_time"`
	UpdatedAt    NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy    NullString `json:"updated_by" db:"updated_by"`
	Password     NullString `json:"-" db:"password"`
	// Ignore password JSON marshall for now

	Description  NullString `json:"description" db:"description"`
	ProfileImage NullString `json:"profile_image" db:"profile_picture"`
	Identity     NullString `json:"identity" db:"identity"`

	CustomEditor bool `json:"custom_editor" db:"c_editor"`
	HideProfile  bool `json:"hide_profile" db:"hide_profile"`
	ProfilePush  bool `json:"profile_push" db:"profile_push"`
	PostPush     bool `json:"post_push" db:"post_push"`
	CommentPush  bool `json:"comment_push" db:"comment_push"`
	Active       bool `json:"active" db:"active"`
}

// Separate API and Member struct
type memberAPI struct{}

var MemberAPI MemberInterface = new(memberAPI)

type MemberInterface interface {
	GetMember(id string) (Member, error)
	InsertMember(m Member) error
	UpdateMember(m Member) error
	DeleteMember(id string) (Member, error)
}

func (api *memberAPI) GetMember(id string) (Member, error) {
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

func (api *memberAPI) InsertMember(m Member) error {
	query, _ := generateSQLStmt(m, "insert", "members")

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

func (api *memberAPI) UpdateMember(m Member) error {
	query, _ := generateSQLStmt(m, "partial_update", "members")
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

func (api *memberAPI) DeleteMember(id string) (Member, error) {

	result := Member{}
	_, err := DB.Exec("UPDATE members SET active = 0 WHERE user_id = ?", id)
	if err != nil {
		log.Fatal(err)
	} else {
		err = nil
	}
	return result, err
}
