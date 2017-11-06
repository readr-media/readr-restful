package models

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
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

func (db *DB) Get(id string) (interface{}, error) {

	member := Member{}
	// err := db.QueryRowx("SELECT work, birthday, description, register_mode, social_id, c_editor, hide_profile, profile_push, post_push, comment_push, user_id, name, nick, create_time, updated_at, gender, mail, updated_by, password, active, profile_picture, identity FROM members where user_id = ?", userID).StructScan(&member)
	err := db.QueryRowx("SELECT * FROM members where user_id = ?", id).StructScan(&member)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("User Not Found")
		err = errors.New("User Not Found")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Successful get user:%s\n", id)
	}
	return member, err
}

func (db *DB) Create(item interface{}) (interface{}, error) {

	member := item.(Member)
	// query := `INSERT INTO members (
	//   user_id,
	//   name,
	//   nick,
	//   birthday,
	//   gender,
	//   work,
	//   mail,
	//   register_mode,
	//   social_id,
	//   updated_by,
	//   password,
	//   description,
	//   profile_picture,
	//   identity,
	//   c_editor,
	//   hide_profile,
	//   profile_push,
	//   post_push,
	//   comment_push,
	//   active)
	// VALUES(
	//   :user_id,
	//   :name,
	//   :nick,
	//   :birthday,
	//   :gender,
	//   :work,
	//   :mail,
	//   :register_mode,
	//   :social_id,
	//   :updated_by,
	//   :password,
	//   :description,
	//   :profile_picture,
	//   :identity,
	//   :c_editor,
	//   :hide_profile,
	//   :profile_push,
	//   :post_push,
	//   :comment_push,
	//   :active)`
	query, _ := makeSQL(&member, "insert")
	result, err := db.NamedExec(query, member)
	// Cannot handle duplicate insert, crash
	if err != nil {
		log.Fatal(err)
		return Member{}, err
	}
	return result, nil
}

func (db *DB) Update(item interface{}) (interface{}, error) {

	member := item.(Member)
	query, _ := makeSQL(&member, "partial_update")
	result, err := db.NamedExec(query, member)

	if err != nil {
		log.Fatal(err)
		return Member{}, err
	}
	return result, nil
}

func (db *DB) Delete(id string) (interface{}, error) {

	result, err := db.Exec("UPDATE members SET active = 0 WHERE user_id = ?", id)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return result, nil
}

func makeSQL(m *Member, mode string) (query string, err error) {

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
