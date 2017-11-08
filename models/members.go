package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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

func (m Member) GetFromDatabase(db *DB) (TableStruct, error) {

	member := Member{}
	err := db.QueryRowx("SELECT * FROM members where user_id = ?", m.ID).StructScan(&member)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("User Not Found")
		member = Member{}
	case err != nil:
		log.Fatal(err)
		member = Member{}
	default:
		fmt.Printf("Successful get user: %s\n", m.ID)
		err = nil
	}
	return member, err
}

func (m Member) InsertIntoDatabase(db *DB) error {

	query, _ := makeSQL(&m, "insert")
	_, err := db.NamedExec(query, m)
	// Cannot handle duplicate insert, crash
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (m Member) UpdateDatabase(db *DB) error {

	query, _ := makeSQL(&m, "partial_update")
	fmt.Println(query)
	_, err := db.NamedExec(query, m)

	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (m Member) DeleteFromDatabase(db *DB) error {

	_, err := db.Exec("UPDATE members SET active = 0 WHERE user_id = ?", m.ID)
	if err != nil {
		log.Fatal(err)
	} else {
		err = nil
	}
	return err
}
