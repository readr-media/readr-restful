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

type Article struct {
	ID            string     `json:"id" db:"post_id"`
	Author        NullString `json:"author" db:"author"`
	CreateTime    NullTime   `json:"created_at" db:"create_time"`
	LikeAmount    int        `json:"liked" db:"like_amount"`
	CommentAmount int        `json:"comment_amount" db:"comment_amount"`
	Title         NullString `json:"title" db:"title"`
	Content       NullString `json:"content" db:"content"`
	Link          NullString `json:"link" db:"link"`
	OgTitle       NullString `json:"og_title" db:"og_title"`
	OgDescription NullString `json:"og_description" db:"og_description"`
	OgImage       NullString `json:"og_image" db:"og_image"`
	Active        int        `json:"active" db:"active"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     NullString `json:"updated_by" db:"updated_by"`
}

func (a *Article) generateSQLStmt(mode string) (query string, err error) {

	columns := make([]string, 0)
	u := reflect.ValueOf(a).Elem()

	bytequery := &bytes.Buffer{}

	switch mode {
	case "insert":
		// fmt.Println("insert")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		bytequery.WriteString("INSERT INTO article_infos ( ")
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
		bytequery.WriteString("UPDATE article_infos SET ")
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(" WHERE post_id = :post_id")

		query = bytequery.String()
		err = nil

	case "partial_update":

		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			field := u.Field(i).Interface()

			switch field := field.(type) {
			case string:
				if field != "" && tag.Get("db") == "post_id" {
					// fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
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
			case int:
				// log.Println("int type ")
				columns = append(columns, tag.Get("db"))
			default:
				fmt.Println("unrecognised format: ", u.Field(i).Type())
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString("UPDATE article_infos SET ")
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(" WHERE post_id = :post_id")

		query = bytequery.String()
		err = nil
	default:
		query = ""
		err = errors.New("No statement was generated")
	}
	// fmt.Println(columns)
	return query, err
}

func (a Article) GetFromDatabase(db *DB) (TableStruct, error) {
	article := Article{}
	err := db.QueryRowx("SELECT * FROM article_infos WHERE post_id = ?", a.ID).StructScan(&article)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Article Not Found")
		article = Article{}
	case err != nil:
		log.Fatal(err)
		article = Article{}
	default:
		fmt.Printf("Successfully get article: %v\n", a.ID)
		err = nil
	}
	return article, err
}

func (a Article) InsertIntoDatabase(db *DB) error {

	query, _ := a.generateSQLStmt("insert")
	result, err := db.NamedExec(query, a)
	if err != nil {
		// fmt.Println(err)
		if strings.Contains(err.Error(), "Duplicate entry") {
			return errors.New("Duplicate entry")
		}
		return err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Article Not Found")
	}
	return nil
}

func (a Article) UpdateDatabase(db *DB) error {

	query, err := a.generateSQLStmt("partial_update")
	if err != nil {
		return errors.New("Generate SQL statement failed")
	}
	result, err := db.NamedExec(query, a)

	if err != nil {
		fmt.Println(err)
		return err
	}
	rowCnt, err := result.RowsAffected()
	// fmt.Println(rowCnt)
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Article Not Found")
	}
	return nil
}

func (a Article) DeleteFromDatabase(db *DB) error {

	_, err := db.Exec("UPDATE article_infos SET active = 0 WHERE post_id = ?", a.ID)
	if err != nil {
		log.Println(err)
	} else {
		err = nil
	}
	return err
}
