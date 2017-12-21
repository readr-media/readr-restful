package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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

type articleAPI struct{}

var ArticleAPI ArticleInterface = new(articleAPI)

type ArticleInterface interface {
	GetArticle(id string) (Article, error)
	InsertArticle(a Article) error
	UpdateArticle(a Article) error
	DeleteArticle(id string) (Article, error)
}

func (api *articleAPI) GetArticle(id string) (Article, error) {
	article := Article{}
	err := DB.QueryRowx("SELECT * FROM article_infos WHERE post_id = ?", id).StructScan(&article)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Article Not Found")
		article = Article{}
	case err != nil:
		log.Fatal(err)
		article = Article{}
	default:
		fmt.Printf("Successfully get article: %v\n", id)
		err = nil
	}
	return article, err
}

func (api *articleAPI) InsertArticle(a Article) error {
	query, _ := generateSQLStmt(a, "insert", "article_infos")
	result, err := DB.NamedExec(query, a)
	if err != nil {
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

func (api *articleAPI) UpdateArticle(a Article) error {

	query, err := generateSQLStmt(a, "partial_update", "article_infos")
	if err != nil {
		return errors.New("Generate SQL statement failed")
	}
	result, err := DB.NamedExec(query, a)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Article Not Found")
	}
	return nil
}

func (api *articleAPI) DeleteArticle(id string) (Article, error) {

	result := Article{}
	_, err := DB.Exec("UPDATE article_infos SET active = 0 WHERE post_id = ?", id)
	if err != nil {
		log.Println(err)
	} else {
		err = nil
	}
	return result, err
}
