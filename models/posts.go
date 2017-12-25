package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Post struct {
	ID            uint32     `json:"id" db:"post_id"`
	Author        NullString `json:"author" db:"author"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
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
	PublishedAt   NullString `json:"published_at" db:"published_at"`
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	GetPosts(maxResult uint8, page uint16, sortMethod string) ([]Post, error)
	GetPost(id uint32) (Post, error)
	InsertPost(p Post) error
	UpdatePost(p Post) error
	DeletePost(id uint32) (Post, error)
}

func (a *postAPI) GetPosts(maxResult uint8, page uint16, sortMethod string) ([]Post, error) {

	var (
		result     []Post
		err        error
		sortString string
	)
	switch sortMethod {
	case "updated_at":
		sortString = "updated_at"
	case "-updated_at":
		sortString = "updated_at DESC"
	default:
		sortString = "updated_at DESC"
	}
	limitBase := (page - 1) * uint16(maxResult)
	limitIncrement := page * uint16(maxResult)

	err = DB.Select(&result, "SELECT * FROM posts ORDER BY ? LIMIT ?, ?", sortString, limitBase, limitIncrement)
	if err != nil || len(result) == 0 {
		result = []Post{}
		err = errors.New("Posts Not Found")
	}
	return result, err
}

func (a *postAPI) GetPost(id uint32) (Post, error) {
	post := Post{}
	err := DB.QueryRowx("SELECT * FROM posts WHERE post_id = ?", id).StructScan(&post)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Post Not Found")
		post = Post{}
	case err != nil:
		log.Fatal(err)
		post = Post{}
	default:
		fmt.Printf("Successfully get post: %v\n", id)
		err = nil
	}
	return post, err
}

func (a *postAPI) InsertPost(p Post) error {
	query, _ := generateSQLStmt(p, "insert", "posts")
	result, err := DB.NamedExec(query, p)
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
		return errors.New("Post Not Found")
	}
	return nil
}

func (a *postAPI) UpdatePost(p Post) error {

	query, err := generateSQLStmt(p, "partial_update", "posts")
	fmt.Println(query)
	if err != nil {
		return errors.New("Generate SQL statement failed")
	}
	result, err := DB.NamedExec(query, p)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Post Not Found")
	}
	return nil
}

func (a *postAPI) DeletePost(id uint32) (Post, error) {

	result := Post{}
	_, err := DB.Exec("UPDATE posts SET active = 0 WHERE post_id = ?", id)
	if err != nil {
		log.Println(err)
	} else {
		err = nil
	}
	return result, err
}
