package models

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
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

// Scan is currently a wrap of sql.NullString.Scan()
func (ns *NullString) Scan(value interface{}) error {
	// ns.String, ns.Valid = value.(string)
	// fmt.Printf("string:%s\n, valid:%s\n", ns.String, ns.Valid)
	// return nil
	x := sql.NullString{}
	err := x.Scan(value)
	ns.String, ns.Valid = x.String, x.Valid
	return err
}

// Value validate the value
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

func makeSQL(m *member, mode string) (query string, err error) {

	columns := make([]string, 0)
	u := reflect.ValueOf(m).Elem()

	bytequery := &bytes.Buffer{}

	switch mode {
	case "insert":
		fmt.Println("insert")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		bytequery.WriteString("INSERT INTO members ( ")
		bytequery.WriteString(strings.Join(columns, ","))
		bytequery.WriteString(") VALUES ( :")
		bytequery.WriteString(strings.Join(columns, ",:"))
		bytequery.WriteString(")")

		// fmt.Println(query)
		query = bytequery.String()
		err = nil

	case "full_update":

		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString("UPDATE members SET ")
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(" WHERE user_id = :user_id")

		query = bytequery.String()
		err = nil

	case "partial_update":

		fmt.Println("partial")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			field := u.Field(i).Interface()

			switch field := field.(type) {
			case string:
				if field != "" {
					fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
					columns = append(columns, tag.Get("db"))
				}

			case NullString:
				if field.Valid {
					fmt.Println("valid NullString : ", field.String)
					columns = append(columns, tag.Get("db"))
				}
			case NullTime:
				if field.Valid {
					fmt.Println("valid NullTime : ", field.Time)
					columns = append(columns, tag.Get("db"))
				}

			case bool:
				fmt.Println("bool type", field)
				columns = append(columns, tag.Get("db"))
			default:
				fmt.Println("unrecognised format: ", u.Field(i).Type())
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString("UPDATE members SET ")
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(" WHERE user_id = :user_id")

		query = bytequery.String()
		err = nil
	}
	fmt.Println(columns)
	return
}

func GetMember(c *gin.Context) {
	userID := c.Param("id")
	db := c.MustGet("DB").(*sqlx.DB)
	member := member{}
	// err := db.QueryRowx("SELECT work, birthday, description, register_mode, social_id, c_editor, hide_profile, profile_push, post_push, comment_push, user_id, name, nick, create_time, updated_at, gender, mail, updated_by, password, active, profile_picture, identity FROM members where user_id = ?", userID).StructScan(&member)
	err := db.QueryRowx("SELECT * FROM members where user_id = ?", userID).StructScan(&member)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user found.")
		c.JSON(404, "User Not Found")
		return
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Successful get user:%s\n", userID)
	}

	c.JSON(200, member)
}

func InsertMember(c *gin.Context) {
	member := member{}
	c.Bind(&member)

	// Need to implement checking for empty string user_id

	if !member.CreateTime.Valid {
		member.CreateTime.Time = time.Now()
		member.CreateTime.Valid = true
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}
	// Need to manually parse null before insert
	// Do not insert create_time and updated_by,
	// left them to the mySQL default
	db := c.MustGet("DB").(*sqlx.DB)
	query, _ := makeSQL(&member, "insert")
	result, err := db.NamedExec(query, member)
	if err != nil {
		panic(err)
	}
	c.JSON(200, result)
	// query := `INSERT INTO members (
	// 	user_id,
	// 	name,
	// 	nick,
	// 	birthday,
	// 	gender,
	// 	work,
	// 	mail,
	// 	register_mode,
	// 	social_id,
	// 	updated_by,
	// 	password,
	// 	description,
	// 	profile_picture,
	// 	identity,
	// 	c_editor,
	// 	hide_profile,
	// 	profile_push,
	// 	post_push,
	// 	comment_push,
	// 	active)
	// VALUES(
	// 	:user_id,
	// 	:name,
	// 	:nick,
	// 	:birthday,
	// 	:gender,
	// 	:work,
	// 	:mail,
	// 	:register_mode,
	// 	:social_id,
	// 	:updated_by,
	// 	:password,
	// 	:description,
	// 	:profile_picture,
	// 	:identity,
	// 	:c_editor,
	// 	:hide_profile,
	// 	:profile_push,
	// 	:post_push,
	// 	:comment_push,
	// 	:active)`
}

func UpdateMember(c *gin.Context) {
	member := member{}
	c.Bind(&member)

	if member.CreateTime.Valid {
		// member.CreateTime.Time = nil
		member.CreateTime.Valid = false
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}
	db := c.MustGet("DB").(*sqlx.DB)
	query, _ := makeSQL(&member, "partial_update")
	result, err := db.NamedExec(query, member)
	// fmt.Print(query)
	if err != nil {
		panic(err)
	}
	c.JSON(200, result)
}

func DeleteMember(c *gin.Context) {
	userID := c.Param("id")

	db := c.MustGet("DB").(*sqlx.DB)
	result, err := db.Exec("UPDATE members SET active = 0 WHERE user_id = ?", userID)
	if err != nil {
		panic(err)
	}
	c.JSON(200, result)
}
