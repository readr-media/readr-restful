package models

import (
	"time"

	"database/sql"
	"database/sql/driver"
	"encoding/json"

	// For NewDB() usage
	_ "github.com/go-sql-driver/mysql"
)

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

// Before is wrap of time.Time.Before, used in test
func (nt *NullTime) Before(value NullTime) bool {
	return nt.Time.Before(value.Time)
}

// After is wrap of time.Time.After, used in test
func (nt *NullTime) After(value NullTime) bool {
	return nt.Time.After(value.Time)
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

// Create our own null string type for prettier marshal JSON format
type NullInt struct {
	Int   int64
	Valid bool // Valid is true if Int is not NULL
}

func (ns *NullInt) Scan(value interface{}) error {
	if value == nil {
		ns.Int, ns.Valid = 0, false
		return nil
	}
	x := sql.NullInt64{}
	err := x.Scan(value)
	ns.Int, ns.Valid = x.Int64, x.Valid
	return err
}

func (ns NullInt) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Int, nil
}

func (ns NullInt) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.Int)
	}
	return json.Marshal(nil)
}

func (ns *NullInt) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	if string(text) == "null" {
		return nil
	}
	if err := json.Unmarshal(text, &ns.Int); err == nil {
		ns.Valid = true
	}
	return nil
}

// ----------------------------- END OF NULLABLE TYPE DEFINITION -----------------------------
