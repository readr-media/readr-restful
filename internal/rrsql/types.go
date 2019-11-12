package rrsql

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

var (
	DuplicateError           = errors.New("Duplicate Entry")
	InternalServerError      = errors.New("Internal Server Error")
	ItemNotFoundError        = errors.New("Item Not Found")
	MultipleRowAffectedError = errors.New("More Than One Rows Affected")

	SQLInsertionFail = errors.New("SQL Insertion Fail")
	SQLUpdateFail    = errors.New("SQL Update Fail")
)

type Sqlfields []string

func (s *Sqlfields) GetFields(template string) (result string) {
	return strings.Join(MakeFieldString("get", template, *s), ", ")
}

// ------------------------------  NULLABLE TYPE DEFINITION -----------------------------

type Nullable interface {
	Value() (driver.Value, error)
}

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

// RedisScan implements Scanner interface in redis for NullTime
func (nt *NullTime) RedisScan(src interface{}) (err error) {
	// Handle null input
	if src == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return errors.New("RedisScan assert error: string")
	}

	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return errors.New("RedisScan validate curly bracket error")
	}

	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	if strings.HasSuffix(s, " true") {
		s = strings.TrimSuffix(s, " true")
		nt.Time, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", s)
		if err != nil {
			return err
		}
		nt.Valid = true
	} else if strings.HasSuffix(s, " false") {
		nt.Time, nt.Valid = time.Time{}, false
	}
	return nil
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
	if err := json.Unmarshal(text, &ns.String); err != nil {
		return err
	}
	ns.Valid = true
	return nil
}

// RedisScan implement Scanner interface in redis package for NullString
func (ns *NullString) RedisScan(src interface{}) (err error) {

	if src == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	s, ok := src.(string)
	if !ok {
		return errors.New("RedisScan assert error: string")
	}

	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return errors.New("RedisScan validate curly bracket error")
	}

	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	if strings.HasSuffix(s, " true") {

		s = strings.TrimSuffix(s, " true")
		ns.String, ns.Valid = s, true

	} else if strings.HasSuffix(s, " false") {
		ns.String, ns.Valid = "", false
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

// RedisScan implement Scanner interface in redis for NullInt
func (ns *NullInt) RedisScan(src interface{}) (err error) {

	if src == nil {
		ns.Int, ns.Valid = 0, false
		return nil
	}
	s, ok := src.(string)
	if !ok {
		return errors.New("RedisScan assert error: int")
	}

	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return errors.New("RedisScan validate curly bracket error")
	}

	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	if strings.HasSuffix(s, " true") {

		s = strings.TrimSuffix(s, " true")
		ns.Int, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		ns.Valid = true

	} else if strings.HasSuffix(s, " false") {
		ns.Int, ns.Valid = 0, false
	}
	return nil
}

// Create our own null boolean type for prettier marshal JSON format
type NullBool struct {
	Bool  bool
	Valid bool // Valid is true if Int is not NULL
}

func (ns *NullBool) Scan(value interface{}) error {
	if value == nil {
		ns.Bool, ns.Valid = false, false
		return nil
	}
	x := sql.NullBool{}
	err := x.Scan(value)
	ns.Bool, ns.Valid = x.Bool, x.Valid
	return err
}

func (ns NullBool) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Bool, nil
}

func (ns NullBool) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.Bool)
	}
	return json.Marshal(nil)
}

func (ns *NullBool) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	if string(text) == "null" {
		return nil
	}
	if err := json.Unmarshal(text, &ns.Bool); err == nil {
		ns.Valid = true
	}
	return nil
}

// Create our own null float type for prettier marshal JSON format
type NullFloat struct {
	Float float64
	Valid bool // Valid is true if Int is not NULL
}

func (ns *NullFloat) Scan(value interface{}) error {
	if value == nil {
		ns.Float, ns.Valid = 0, false
		return nil
	}
	x := sql.NullFloat64{}
	err := x.Scan(value)
	ns.Float, ns.Valid = x.Float64, x.Valid
	return err
}

func (ns NullFloat) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Float, nil
}

func (ns NullFloat) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.Float)
	}
	return json.Marshal(nil)
}

func (ns *NullFloat) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	if string(text) == "null" {
		return nil
	}
	if err := json.Unmarshal(text, &ns.Float); err == nil {
		ns.Valid = true
	}
	return nil
}

type NullIntSlice struct {
	Slice []int
	Valid bool // Valid is true if Int is not NULL
}

func (ns NullIntSlice) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Slice, nil
}

func (ns NullIntSlice) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.Slice)
	}
	return json.Marshal(nil)
}

func (ns *NullIntSlice) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	if string(text) == "null" {
		return nil
	}
	if err := json.Unmarshal(text, &ns.Slice); err == nil {
		ns.Valid = true
	}
	return nil
}

// ----------------------------- END OF NULLABLE TYPE DEFINITION -----------------------------

// // Since the logic of this function is deeply coupled with sql types, and nullable types are now belongs to sql package.
// // To prevent circular dependency, remove this function from redis package and add it here.
// func convertRedisAssign(dest, src interface{}) error {
// 	var err error
// 	b, ok := src.([]byte)
// 	if !ok {
// 		return errors.New("RedisScan error assert byte array")
// 	}
// 	s := string(b)
// 	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
// 		fmt.Println(string(b), " failed")
// 	} else {
// 		s = strings.TrimPrefix(s, "{")
// 		s = strings.TrimSuffix(s, "}")

// 		if strings.HasSuffix(s, " true") {
// 			s = strings.TrimSuffix(s, " true")

// 			switch d := dest.(type) {
// 			case *NullBool:
// 				d.Bool, err = strconv.ParseBool(s)
// 				if err != nil {
// 					fmt.Println(err)
// 					return err
// 				}
// 				d.Valid = true
// 			default:
// 				fmt.Println(s, " non case ", d)
// 				return errors.New("Cannot parse non-nil nullable type")
// 			}

// 		} else if strings.HasSuffix(s, " false") {
// 			s = strings.TrimSuffix(s, " false")

// 			switch d := dest.(type) {
// 			case *NullBool:
// 				d.Valid, d.Valid = false, false
// 			default:
// 				fmt.Println(s, " FALSE non case ", d)
// 				return errors.New("redis conversion error: invalid null* valid field")
// 			}
// 		}
// 	}
// 	return nil
// }
