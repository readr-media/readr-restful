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

	// For NewDB() usage
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// var db *sqlx.DB

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

type Datastore interface {
	Get(item TableStruct) (TableStruct, error)
	Create(item interface{}) (interface{}, error)
	Update(item interface{}) (interface{}, error)
	Delete(id string) (interface{}, error)
}

type DB struct {
	*sqlx.DB
}

// func InitDB(dataURI string) {
// 	var err error
// 	db, err = sqlx.Open("mysql", dataURI)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	if err = db.Ping(); err != nil {
// 		log.Panic(err)
// 	}
// }
type TableStruct interface {
	GetFromDatabase(*DB) (TableStruct, error)
}

func NewDB(dbURI string) (*DB, error) {
	db, err := sqlx.Open("mysql", dbURI)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Get implemented for Datastore interface below
func (db *DB) Get(item TableStruct) (TableStruct, error) {

	// Declaration of return set
	var (
		result TableStruct
		err    error
	)

	switch item := item.(type) {
	case Member:
		result, err = item.GetFromDatabase(db)
		if err != nil {
			// log.Fatal(err)
			result = Member{}
		}
		// err = db.QueryRowx("SELECT * FROM members where user_id = ?", item.ID).StructScan(&member)

		// switch {
		// case err == sql.ErrNoRows:
		// 	log.Printf("User Not Found")
		// 	err = errors.New("User Not Found")
		// 	result = nil
		// case err != nil:
		// 	log.Fatal(err)
		// 	result = nil
		// default:
		// 	fmt.Printf("Successful get user:%s\n", item.ID)
		// 	result = member
		// }

	}
	return result, err
	// err := db.QueryRowx("SELECT work, birthday, description, register_mode, social_id, c_editor, hide_profile, profile_push, post_push, comment_push, user_id, name, nick, create_time, updated_at, gender, mail, updated_by, password, active, profile_picture, identity FROM members where user_id = ?", userID).StructScan(&member)

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
