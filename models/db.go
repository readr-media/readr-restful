package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	// For NewDB() usage
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// var db *sqlx.DB

// ------------------------------  NULLABLE TYPE DEFINITION -----------------------------

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

// ----------------------------- END OF NULLABLE TYPE DEFINITION -----------------------------

type Datastore interface {
	Get(item TableStruct) (TableStruct, error)
	Create(item TableStruct) (interface{}, error)
	Update(item TableStruct) (interface{}, error)
	Delete(item TableStruct) (interface{}, error)
}

type DB struct {
	*sqlx.DB
}

type TableStruct interface {
	GetFromDatabase(*DB) (TableStruct, error)
	InsertIntoDatabase(*DB) error
	UpdateDatabase(*DB) error
	DeleteFromDatabase(*DB) error
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
	case Article:
		result, err = item.GetFromDatabase(db)
		if err != nil {
			result = Article{}
		}
	}
	return result, err
}

func (db *DB) Create(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.InsertIntoDatabase(db)

	}
	return result, err
}

func (db *DB) Update(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.UpdateDatabase(db)
	}
	return result, err
}

func (db *DB) Delete(item TableStruct) (interface{}, error) {

	var (
		result Member
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.DeleteFromDatabase(db)
		if err != nil {
			result = Member{}
		} else {
			result = item
		}
	}
	return result, err
}
