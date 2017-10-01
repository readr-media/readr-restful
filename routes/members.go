package routes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time)
	}
	return json.Marshal(nil)
}

func (nt *NullTime) UnmarshalJSON(text []byte) error {
	nt.Valid = false
	txt := string(text)
	if txt == "null" || txt == "" {
		return nil
	}

	t := time.Time{}
	err := t.UnmarshalJSON(text)
	if err == nil {
		nt.Time = t
		nt.Valid = true
	}

	return err
}

// Create our own null string type for prettier marshal JSON format
type NullString sql.NullString

// Currently it's a wrap of sql.NullString.Scan()
func (ns *NullString) Scan(value interface{}) error {
	// ns.String, ns.Valid = value.(string)
	// fmt.Printf("string:%s\n, valid:%s\n", ns.String, ns.Valid)
	// return nil
	x := sql.NullString{}
	err := x.Scan(value)
	ns.String, ns.Valid = x.String, x.Valid
	return err
}
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

func (ns *NullString) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	if string(text) == "null" {
		return nil
	}
	if err := json.Unmarshal(text, &ns.String); err == nil {
		ns.Valid = true
	}
	return nil
}

type member struct {
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
	Password     NullString `db:"password"`

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

func GetSingleMember(c *gin.Context) {
	userID := c.Param("id")
	db := c.MustGet("DB").(*sqlx.DB)
	member := member{}
	// err := db.QueryRowx("SELECT work, birthday, description, register_mode, social_id, c_editor, hide_profile, profile_push, post_push, comment_push, user_id, name, nick, create_time, updated_at, gender, mail, updated_by, password, active, profile_picture, identity FROM members where user_id = ?", userID).StructScan(&member)
	err := db.QueryRowx("SELECT * FROM members where user_id = ?", userID).StructScan(&member)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user found.")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Successful get user:%s\n", userID)
	}

	c.JSON(200, member)
}

func InsertNewMember(c *gin.Context) {
	member := member{}
	c.Bind(&member)

	// Need to implement checking for empty string user_id

	// Need to manually parse null before insert
	// Do not insert create_time and updated_by,
	// left them to the mySQL default
	db := c.MustGet("DB").(*sqlx.DB)
	query := `INSERT INTO members (
		user_id,
		name,
		nick,
		birthday,
		gender,
		work,
		mail,
		register_mode,
		social_id,
		updated_by,
		password,
		description,
		profile_picture,
		identity,
		c_editor,
		hide_profile,
		profile_push,
		post_push,
		comment_push,
		active)
	VALUES(
		:user_id,
		:name,
		:nick,
		:birthday,
		:gender,
		:work,
		:mail,
		:register_mode,
		:social_id,
		:updated_by,
		:password,
		:description,
		:profile_picture,
		:identity,
		:c_editor,
		:hide_profile,
		:profile_push,
		:post_push,
		:comment_push,
		:active)`
	_, err := db.NamedExec(query, member)
	if err != nil {
		panic(err)
	}
	c.JSON(200, member)
}
