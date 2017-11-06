package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Article struct {
	ID         string     `json:"id" db:"post_id"`
	Author     NullString `json:"author" db:"author"`
	CreateTime NullTime   `json:"created_at" db:"create_time"`
	Like       int        `json:"liked" db:"like_amount"`
}

func (db *DB) Get(id string) (interface{}, error) {

	article := Article{}
	err := db.QueryRowx("SELECT * FROM article_infos WHERE post_id = ?", id).StructScan(&article)

	switch {
	case err == sql.ErrNoRows:
		log.Println("Article Not Found")
		err = errors.New("Article Not Found")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Successful get article: %s\n", id)
	}
	return article, err
}
