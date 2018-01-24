package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
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
	LinkName        NullString `json:"link_name" db:"link_name"`
}

type PostArgs struct {
	BasicArgs
	Active string `form:"active"`
	Author string `form:"author"`
}

type PostUpdateArgs struct {
	IDs       []int    `json:"ids"`
	UpdatedBy string   `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt NullTime `json:"-" db:"updated_at"`
	Active    NullInt  `json:"-" db:"active"`
}

func (args *PostArgs) parse(prefix string) (restrict string, whereValues []interface{}) {
	input := reflect.ValueOf(args).Elem()
	whereString := make([]string, 0)
	for i := 0; i < input.NumField(); i++ {
		tag := input.Type().Field(i).Tag.Get("form")
		switch tag {
		case "active":
			if args.Active != "" {
				tmp := map[string][]uint32{}
				err := json.Unmarshal([]byte(args.Active), &tmp)
				if err != nil {
					fmt.Println("active ", err.Error())
				} else {
					for operator, values := range tmp {
						whereString = append(whereString, fmt.Sprintf("%s.active %s (?)", prefix, operatorHelper(operator)))
						whereValues = append(whereValues, values)
					}
				}
			}
		case "author":
			if args.Author != "" {
				tmp := map[string][]string{}
				err := json.Unmarshal([]byte(args.Author), &tmp)
				if err != nil {
					fmt.Println("author: ", err.Error())
				} else {
					for operator, values := range tmp {
						whereString = append(whereString, fmt.Sprintf("%s.author %s (?)", prefix, operatorHelper(operator)))
						whereValues = append(whereValues, values)
					}
				}
			}
		}
	}
	if len(whereString) > 1 {
		restrict = strings.Join(whereString, " AND ")
	} else if len(whereString) == 1 {
		restrict = whereString[0]
	}
	return restrict, whereValues
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	DeletePost(id uint32) error
	GetPosts(args PostArgs) (result []PostMember, err error)
	GetPost(id uint32) (PostMember, error)
	//GetPosts(maxResult uint8, page uint16, sortMethod string) ([]PostMember, error)
	InsertPost(p Post) error
	UpdateAll(req PostUpdateArgs) error
	UpdatePost(p Post) error
}

// UpdatedBy wraps Member for embedded field updated_by
// in the usage of anonymous struct in PostMember
type UpdatedBy Member
type PostMember struct {
	Post
	Member    `json:"author" db:"author"`
	UpdatedBy `json:"updated_by" db:"updated_by"`
}

// func (a *postAPI) GetPosts(maxResult uint8, page uint16, sortMethod string, where string) ([]PostMember, error) {
func (a *postAPI) GetPosts(req PostArgs) (result []PostMember, err error) {

	var singlePost PostMember
	fmt.Println(req)
	whereClauses, whereValues := req.parse("posts")
	tags := getStructDBTags("full", Member{})

	authorField := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedByField := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)
	query := fmt.Sprintf(`SELECT posts.*, %s, %s FROM posts 
		LEFT JOIN members AS author ON posts.author = author.member_id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.member_id 
		WHERE %s `,
		strings.Join(authorField, ","), strings.Join(updatedByField, ","), whereClauses)

	// To give adaptability to where clauses, have to use ... operator here
	// Therefore split query into two parts, assembling them after sqlx.Rebind
	query, args, err := sqlx.In(query, whereValues...)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)

	// Attach the order part to query with expanded amounts of placeholder.
	// Append limit and offset to args slice
	query = query + fmt.Sprintf(`ORDER BY %s LIMIT ? OFFSET ?`, orderByHelper(req.Sorting))
	args = append(args, req.MaxResult, (req.Page-1)*uint16(req.MaxResult))
	rows, err := DB.Queryx(query, args...)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	for rows.Next() {
		err = rows.StructScan(&singlePost)
		if err != nil {
			result = []PostMember{}
			return result, err
		}
		result = append(result, singlePost)

	}
	// err = DB.Select(&result, query, args.MaxResult, (args.Page-1)*uint16(args.MaxResult))
	if len(result) == 0 {
		result = []PostMember{}
		err = errors.New("Posts Not Found")
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
		err = nil
	}
	return post, err
}

func (a *postAPI) InsertPost(p Post) error {
	// query, _ := generateSQLStmt("insert", "posts", p)

	tags := getStructDBTags("full", Post{})
	query := fmt.Sprintf(`INSERT INTO posts (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

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

	fmt.Println(p)
	// query, err := generateSQLStmt("partial_update", "posts", p)
	tags := getStructDBTags("partial", p)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE posts SET %s WHERE post_id = :post_id`,
		strings.Join(fields, ", "))

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

func (a *postAPI) UpdateAll(req PostUpdateArgs) error {

	query, args, err := sqlx.In(`UPDATE posts SET updated_by = ?, updated_at = ?, active = ? WHERE post_id IN (?)`, req.UpdatedBy, req.UpdatedAt, req.Active, req.IDs)
	if err != nil {
		return err
	}
	query = DB.Rebind(query)
	result, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > int64(len(req.IDs)) {
		return errors.New("More Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Posts Not Found")
	}
	return nil
}
