package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Post could use json:"omitempty" tag to ignore null field
// However, struct type field like NullTime, NullString must be declared as pointer,
// like *NullTime, *NullString to be used with omitempty
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
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	GetPosts(maxResult uint8, page uint16, sortMethod string) ([]PostMember, error)
	GetPost(id uint32) (PostMember, error)
	InsertPost(p Post) error
	UpdatePost(p Post) error
	DeletePost(id uint32) error
}

// UpdatedBy wraps Member for embedded field updated_by
// in the usage of anonymous struct in PostMember
type UpdatedBy Member
type PostMember struct {
	Post
	Member    `json:"author" db:"author"`
	UpdatedBy `json:"updated_by" db:"updated_by"`
}

func makeFieldString(mode string, pattern string, tags []string, params string) (result []string) {
	switch mode {
	case "get":
		for _, field := range tags {
			result = append(result, fmt.Sprintf(pattern, params, field, params, field))
		}
	case "update":
		for _, value := range tags {
			result = append(result, fmt.Sprintf(pattern, value, value))
		}
	}
	return result
}

func (a *postAPI) GetPosts(maxResult uint8, page uint16, sortMethod string) ([]PostMember, error) {

	var (
		result     []PostMember
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
	// query, _ := generateSQLStmt("get_all", "posts", sortString)

	tags := getStructDBTags("full", Member{})
	author := makeFieldString("get", `%s.%s "%s.%s"`, tags, "author")
	updatedBy := makeFieldString("get", `%s.%s "%s.%s"`, tags, "updated_by")
	query := fmt.Sprintf(`SELECT posts.*, %s, %s FROM posts 
		LEFT JOIN members AS author ON posts.author = author.user_id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.user_id 
		where posts.active != 0 ORDER BY %s LIMIT ? OFFSET ?`,
		strings.Join(author, ","), strings.Join(updatedBy, ","), sortString)

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
	author := makeFieldString("get", `%s.%s "%s.%s"`, tags, "author")
	updatedBy := makeFieldString("get", `%s.%s "%s.%s"`, tags, "updated_by")
	query := fmt.Sprintf(`SELECT posts.*, %s, %s FROM posts 
		LEFT JOIN members AS author ON posts.author = author.user_id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.user_id 
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
	fields := makeFieldString("update", `%s = :%s`, tags, "")
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
	result, err := DB.Exec("UPDATE posts SET active = 0 WHERE post_id = ?", id)
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
