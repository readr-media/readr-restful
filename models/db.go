package models

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"

	// For NewDB() usage
	"github.com/jmoiron/sqlx"
)

var DB database = database{nil}

type database struct {
	*sqlx.DB
}

type dataStore struct{}

type DatastoreInterface interface {
	Get(item interface{}) (result interface{}, err error)
	Create(item interface{}) (result interface{}, err error)
	Update(item interface{}) (result interface{}, err error)
	Delete(item interface{}) (result interface{}, err error)
}

// var DS DatastoreInterface = new(dataStore)

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
			// Get table id and set it to idName
			if tag.Get("json") == "id" {
				fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
				idName = tag.Get("db")
			}

			switch field := field.(type) {
			case string:
				if field != "" {
					// if tag.Get("json") == "id" {
					// 	fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
					// 	idName = tag.Get("db")
					// }
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
			case bool, int, uint32:
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
