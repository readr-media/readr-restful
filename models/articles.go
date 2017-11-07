package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
