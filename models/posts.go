package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

var PostStatus map[string]interface{}

// Post could use json:"omitempty" tag to ignore null field
// However, struct type field like NullTime, NullString must be declared as pointer,
// like *NullTime, *NullString to be used with omitempty
type Post struct {
	ID              uint32     `json:"id" db:"post_id"`
	Author          NullString `json:"author" db:"author"`
	CreatedAt       NullTime   `json:"created_at" db:"created_at"`
	LikeAmount      NullInt    `json:"like_amount" db:"like_amount"`
	CommentAmount   NullInt    `json:"comment_amount" db:"comment_amount"`
	Title           NullString `json:"title" db:"title"`
	Content         NullString `json:"content" db:"content"`
	Link            NullString `json:"link" db:"link"`
	OgTitle         NullString `json:"og_title" db:"og_title"`
	OgDescription   NullString `json:"og_description" db:"og_description"`
	OgImage         NullString `json:"og_image" db:"og_image"`
	Active          NullInt    `json:"active" db:"active"`
	UpdatedAt       NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy       NullString `json:"updated_by" db:"updated_by"`
	PublishedAt     NullTime   `json:"published_at" db:"published_at"`
	LinkTitle       NullString `json:"link_title" db:"link_title"`
	LinkDescription NullString `json:"link_description" db:"link_description"`
	LinkImage       NullString `json:"link_image" db:"link_image"`
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	GetPosts(maxResult uint8, page uint16, sortMethod string) ([]PostMember, error)
	GetPost(id uint32) (PostMember, error)
	InsertPost(p Post) error
	UpdatePost(p Post) error
	DeletePost(id uint32) error
	SetMultipleActive(ids []uint32, active int) error
}

// UpdatedBy wraps Member for embedded field updated_by
// in the usage of anonymous struct in PostMember
type UpdatedBy Member
type PostMember struct {
	Post
	Member    `json:"author" db:"author"`
	UpdatedBy `json:"updated_by" db:"updated_by"`
}

func (a *postAPI) GetPosts(maxResult uint8, page uint16, sortMethod string) ([]PostMember, error) {

	var (
		result []PostMember
		err    error
	)

	tags := getStructDBTags("full", Member{})
	author := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedBy := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)
	query := fmt.Sprintf(`SELECT posts.*, %s, %s FROM posts 
		LEFT JOIN members AS author ON posts.author = author.member_id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.member_id 
		where posts.active != %d ORDER BY %s LIMIT ? OFFSET ?`,
		strings.Join(author, ","), strings.Join(updatedBy, ","), int(PostStatus["deactive"].(float64)), orderByHelper(sortMethod))

	err = DB.Select(&result, query, maxResult, (page-1)*uint16(maxResult))
	if err != nil || len(result) == 0 {
		result = []PostMember{}
		err = errors.New("Posts Not Found")
		fmt.Println(err)
		fmt.Println(result)
	}
	return result, err
}

func (a *postAPI) GetPost(id uint32) (PostMember, error) {

	post := PostMember{}
	tags := getStructDBTags("full", Member{})
	author := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedBy := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)
	query := fmt.Sprintf(`SELECT posts.*, %s, %s FROM posts 
		LEFT JOIN members AS author ON posts.author = author.member_id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.member_id 
		WHERE post_id = ?`,
		strings.Join(author, ","), strings.Join(updatedBy, ","))

	// fmt.Println(query)
	// query, _ := generateSQLStmt("left_join", "posts", "members")
	err := DB.Get(&post, query, id)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Post Not Found")
		post = PostMember{}
	case err != nil:
		log.Fatal(err)
		post = PostMember{}
	default:
		// fmt.Printf("Successfully get post: %v\n", id)
		err = nil
	}
	return post, err
}

func (a *postAPI) InsertPost(p Post) error {
	// query, _ := generateSQLStmt("insert", "posts", p)

	tags := getStructDBTags("full", Post{})
	query := fmt.Sprintf(`INSERT INTO posts (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))
	// fmt.Println("insert: ", query)
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
	return err
}

func (a *postAPI) UpdatePost(p Post) error {

	// query, err := generateSQLStmt("partial_update", "posts", p)
	tags := getStructDBTags("partial", p)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE posts SET %s WHERE post_id = :post_id`,
		strings.Join(fields, ", "))
	// if err != nil {
	// 	return errors.New("Generate SQL statement failed")
	// }
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
	return err
}

func (a *postAPI) DeletePost(id uint32) error {

	// result := Post{}
	result, err := DB.Exec(fmt.Sprintf("UPDATE posts SET active = %d WHERE post_id = ?", int(PostStatus["deactive"].(float64))), id)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Post Not Found")
	}
	return err
}

func (a *postAPI) SetMultipleActive(ids []uint32, active int) error {
	prep := fmt.Sprintf("Update posts SET active = %d WHERE post_id IN (?);", active)
	query, args, err := sqlx.In(prep, ids)
	if err != nil {
		return err
	}
	query = DB.Rebind(query)
	result, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > int64(len(ids)) {
		return errors.New("More Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Posts Not Found")
	}
	return nil
}
