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

func makeFieldString(mode string, pattern string, tags []string) (result []string) {
	switch mode {
	case "get":
		for _, field := range tags {
			result = append(result, fmt.Sprintf(pattern, field, field))
		}
	case "update":
		for _, value := range tags {
			result = append(result, fmt.Sprintf(pattern, value, value))
		}
	}
	return result
}

func orderByHelper(sortMethod string) (result string) {
	// if strings.Contains(sortMethod, )
	if strings.HasPrefix(sortMethod, "-") {
		result = sortMethod[1:] + " DESC"
	} else {
		result = sortMethod
	}
	return result
}

func getStructDBTags(mode string, input interface{}) []string {
	columns := make([]string, 0)
	u := reflect.ValueOf(input)
	for i := 0; i < u.NumField(); i++ {
		tag := u.Type().Field(i).Tag
		if mode == "full" {
			columns = append(columns, tag.Get("db"))
		} else if mode == "partial" {
			field := u.Field(i).Interface()

			switch field := field.(type) {
			case string:
				if field != "" {
					columns = append(columns, tag.Get("db"))
				}
			// Could not put NullString, NullTime in one case
			case NullString:
				if field.Valid {
					columns = append(columns, tag.Get("db"))
				}
			case NullTime:
				if field.Valid {
					columns = append(columns, tag.Get("db"))
				}
			case NullInt:
				if field.Valid {
					fmt.Println("valid NullInt : ", field.Int)
					columns = append(columns, tag.Get("db"))
				}
			case NullBool:
				if field.Valid {
					fmt.Println("valid NullBool : ", field.Bool)
					columns = append(columns, tag.Get("db"))
				}
			case bool, int, uint32:
				columns = append(columns, tag.Get("db"))
			default:
				fmt.Println("unrecognised format: ", u.Field(i).Type())
			}
		}
	}
	return columns
}

// Use ... operator to encompass the potential for variadic input in the future
func generateSQLStmt(mode string, tableName string, input ...interface{}) (query string, err error) {

	columns := make([]string, 0)
	// u := reflect.ValueOf(input).Elem()

	bytequery := &bytes.Buffer{}

	switch mode {
	case "get_all":
		bytequery.WriteString(fmt.Sprintf("SELECT * FROM %s ORDER BY %s LIMIT ?, ?", tableName, input[0].(string)))
		query = bytequery.String()
		err = nil
	case "insert":
		// Parse first input
		u := reflect.ValueOf(input[0])
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

		u := reflect.ValueOf(input[0])
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
		u := reflect.ValueOf(input[0])
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
			case NullInt:
				if field.Valid {
					fmt.Println("valid NullInt : ", field.Int)
					columns = append(columns, tag.Get("db"))
				}
			case NullBool:
				if field.Valid {
					fmt.Println("valid NullBool : ", field.Bool)
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
