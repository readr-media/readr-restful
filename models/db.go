package models

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"log"

	// For NewDB() usage
	"github.com/jmoiron/sqlx"
)

var DB database = database{nil}

type database struct {
	*sqlx.DB
}

type Datastore interface {
	Get(item TableStruct) (TableStruct, error)
	Create(item TableStruct) (interface{}, error)
	Update(item TableStruct) (interface{}, error)
	Delete(item TableStruct) (interface{}, error)
}

type TableStruct interface {
	GetFromDatabase() (TableStruct, error)
	InsertIntoDatabase() error
	UpdateDatabase() error
	DeleteFromDatabase() error
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

func Connect(dbURI string) {
	d, err := sqlx.Open("mysql", dbURI)
	if err != nil {
		log.Panic(err)
	}
	if err = d.Ping(); err != nil {
		log.Panic(err)
	}
	DB = database{d}
}

// Get implemented for Datastore interface below
func (db *database) Get(item TableStruct) (TableStruct, error) {

	// Declaration of return set
	var (
		result TableStruct
		err    error
	)

	switch item := item.(type) {
	case Member:
		result, err = item.GetFromDatabase()
		if err != nil {
			result = Member{}
		}
	case Article:
		result, err = item.GetFromDatabase()
		if err != nil {
			result = Article{}
		}
	}
	return result, err
}

func (db *database) Create(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.InsertIntoDatabase()
	case Article:
		err = item.InsertIntoDatabase()
	default:
		err = errors.New("Insert fail")
	}
	return result, err
}

func (db *database) Update(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.UpdateDatabase()
	case Article:
		err = item.UpdateDatabase()
	default:
		err = errors.New("Update Fail")
	}
	return result, err
}

func (db *database) Delete(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.DeleteFromDatabase()
		if err != nil {
			result = Member{}
		} else {
			result = item
		}
	case Article:
		err = item.DeleteFromDatabase()
		if err != nil {
			result = Article{}
		} else {
			result = item
		}
	}
	return result, err
}

func generateSQLStmt(input interface{}, mode string, tableName string) (query string, err error) {

	columns := make([]string, 0)
	// u := reflect.ValueOf(input).Elem()
	u := reflect.ValueOf(input)

	bytequery := &bytes.Buffer{}

	switch mode {
	case "insert":
		fmt.Println("insert")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		bytequery.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))
		bytequery.WriteString(strings.Join(columns, ","))
		bytequery.WriteString(") VALUES ( :")
		bytequery.WriteString(strings.Join(columns, ",:"))
		bytequery.WriteString(");")

		query = bytequery.String()
		err = nil

	case "full_update":

		fmt.Println("full_update")
		var idName string
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			columns = append(columns, tag.Get("db"))

			if tag.Get("json") == "id" {
				idName = tag.Get("db")
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}

		bytequery.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(fmt.Sprintf(" WHERE %s = :%s", idName, idName))

		query = bytequery.String()
		err = nil

	case "partial_update":

		var idName string
		fmt.Println("partial")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			field := u.Field(i).Interface()

			switch field := field.(type) {
			case string:
				if field != "" {
					if tag.Get("json") == "id" {
						fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
						idName = tag.Get("db")
					}
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

			case bool, int:
				columns = append(columns, tag.Get("db"))
			default:
				fmt.Println("unrecognised format: ", u.Field(i).Type())
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(fmt.Sprintf(" WHERE %s = :%s;", idName, idName))

		query = bytequery.String()
		err = nil
	}
	return
}
