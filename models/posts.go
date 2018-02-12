package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

var PostStatus map[string]interface{}
var PostType map[string]interface{}

// Post could use json:"omitempty" tag to ignore null field
// However, struct type field like NullTime, NullString must be declared as pointer,
// like *NullTime, *NullString to be used with omitempty
type Post struct {
	ID              uint32     `json:"id" db:"post_id" redis:"post_id"`
	Author          NullString `json:"author" db:"author" redis:"author"`
	CreatedAt       NullTime   `json:"created_at" db:"created_at" redis:"created_at"`
	LikeAmount      NullInt    `json:"like_amount" db:"like_amount" redis:"like_amount"`
	CommentAmount   NullInt    `json:"comment_amount" db:"comment_amount" redis:"comment_amount"`
	Title           NullString `json:"title" db:"title" redis:"title"`
	Content         NullString `json:"content" db:"content" redis:"content"`
	Type            NullInt    `json:"type" db:"type" redis:"type"`
	Link            NullString `json:"link" db:"link" redis:"link"`
	OgTitle         NullString `json:"og_title" db:"og_title" redis:"og_title"`
	OgDescription   NullString `json:"og_description" db:"og_description" redis:"og_description"`
	OgImage         NullString `json:"og_image" db:"og_image" redis:"og_image"`
	Active          NullInt    `json:"active" db:"active" redis:"active"`
	UpdatedAt       NullTime   `json:"updated_at" db:"updated_at" redis:"updated_at"`
	UpdatedBy       NullString `json:"updated_by" db:"updated_by" redis:"updated_by"`
	PublishedAt     NullTime   `json:"published_at" db:"published_at" redis:"published_at"`
	LinkTitle       NullString `json:"link_title" db:"link_title" redis:"link_title"`
	LinkDescription NullString `json:"link_description" db:"link_description" redis:"link_description"`
	LinkImage       NullString `json:"link_image" db:"link_image" redis:"link_image"`
	LinkName        NullString `json:"link_name" db:"link_name" redis:"link_name"`
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	DeletePost(id uint32) error
	GetPosts(args PostArgs) (result []PostMember, err error)
	GetPost(id uint32) (TaggedPostMember, error)
	InsertPost(p Post) (int, error)
	UpdateAll(req PostUpdateArgs) error
	UpdatePost(p Post) error
	Count(req PostArgs) (result int, err error)
}

type TaggedPost struct {
	Post
	Tags []int `json:"tags" db:"tags"`
}

type TaggedPostMember struct {
	PostMember
	Tags NullString `json:"-" db:"tags"`
}

func (t *TaggedPostMember) MarshalJSON() ([]byte, error) {
	type TPM TaggedPostMember
	var Tags []map[string]string

	if t.Tags.Valid != false {
		tags := strings.Split(t.Tags.String, ",")
		for _, value := range tags {
			tag := strings.Split(value, ":")
			Tags = append(Tags, map[string]string{
				"id":   tag[0],
				"text": tag[1],
			})
		}
	}
	return json.Marshal(&struct {
		LastSeen []map[string]string `json:"tags"`
		*TPM
	}{
		LastSeen: Tags,
		TPM:      (*TPM)(t),
	})
}

// UpdatedBy wraps Member for embedded field updated_by
// in the usage of anonymous struct in PostMember
type UpdatedBy Member
type PostMember struct {
	Post
	Member    `json:"author" db:"author"`
	UpdatedBy `json:"updated_by" db:"updated_by"`
}

type PostUpdateArgs struct {
	IDs       []int    `json:"ids"`
	UpdatedBy string   `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt NullTime `json:"-" db:"updated_at"`
	Active    NullInt  `json:"-" db:"active"`
}

type PostArgs map[string]interface{}

func (p *PostArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)
	for arg, value := range *p {
		switch arg {
		//	  Count  , GetAll
		case "active", "posts.active":
			for k, v := range value.(map[string][]int) {
				where = append(where, fmt.Sprintf("%s %s (?)", arg, operatorHelper(k)))
				values = append(values, v)
			}
		//      Count, GetAll
		case "author", "posts.author":
			for k, v := range value.(map[string][]string) {
				where = append(where, fmt.Sprintf("%s %s (?)", arg, operatorHelper(k)))
				values = append(values, v)
			}
		case "type", "posts.type":
			for k, v := range value.(map[string][]int) {
				where = append(where, fmt.Sprintf("%s %s (?)", arg, operatorHelper(k)))
				values = append(values, v)
			}
		}
	}
	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (a *postAPI) GetPosts(req PostArgs) (result []PostMember, err error) {

	var singlePost PostMember

	restricts, values := req.parse()

	tags := getStructDBTags("full", Member{})
	authorField := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedByField := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)
	query := fmt.Sprintf(`SELECT posts.*, %s, %s FROM posts
		LEFT JOIN members AS author ON posts.author = author.member_id
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.member_id
		WHERE %s `,
		strings.Join(authorField, ","), strings.Join(updatedByField, ","), restricts)

	// To give adaptability to where clauses, have to use ... operator here
	// Therefore split query into two parts, assembling them after sqlx.Rebind
	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)

	// Attach the order part to query with expanded amounts of placeholder.
	// Append limit and offset to args slice
	query = query + fmt.Sprintf(`ORDER BY %s LIMIT ? OFFSET ?`, orderByHelper(req["sort"].(string)))
	args = append(args, req["max_result"].(uint8), (req["page"].(uint16)-1)*uint16(req["max_result"].(uint8)))
	rows, err := DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		if err = rows.StructScan(&singlePost); err != nil {
			result = []PostMember{}
			return result, err
		}
		result = append(result, singlePost)
	}
	// err = DB.Select(&result, query, args.MaxResult, (args.Page-1)*uint16(args.MaxResult))
	// if len(result) == 0 {
	// 	result = []PostMember{}
	// 	err = errors.New("Posts Not Found")
	// }
	return result, err
}

func (a *postAPI) GetPost(id uint32) (TaggedPostMember, error) {

	post := TaggedPostMember{}
	tags := getStructDBTags("full", Member{})
	author := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedBy := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)
	query := fmt.Sprintf(`SELECT posts.*, %s, %s, tags.tags as tags FROM posts
		LEFT JOIN members AS author ON posts.author = author.member_id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.member_id 
		LEFT JOIN (
			SELECT pt.post_id as post_id, GROUP_CONCAT(CONCAT(t.tag_id, ":", t.tag_content) SEPARATOR ',') as tags 
			FROM post_tags as pt LEFT JOIN tags as t ON t.tag_id = pt.tag_id 
			GROUP BY pt.post_id
		) AS tags ON tags.post_id = posts.post_id WHERE posts.post_id = ?`,
		strings.Join(author, ","), strings.Join(updatedBy, ","))

	err := DB.Get(&post, query, id)
	if err != nil {
		log.Println(err.Error())
		switch {
		case err == sql.ErrNoRows:
			err = errors.New("Post Not Found")
			return TaggedPostMember{}, err
		case err != nil:
			log.Fatal(err)
			return TaggedPostMember{}, err
		default:
			err = nil
		}
	}

	return post, err
}

func (a *postAPI) InsertPost(p Post) (int, error) {
	// query, _ := generateSQLStmt("insert", "posts", p)

	tags := getStructDBTags("full", Post{})
	query := fmt.Sprintf(`INSERT INTO posts (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	result, err := DB.NamedExec(query, p)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, errors.New("Duplicate entry")
		}
		return 0, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return 0, errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return 0, errors.New("Post Not Found")
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a tag: %v", err)
		return 0, err
	}

	PostCache.Insert(p)

	return int(lastID), err
}

func (a *postAPI) UpdatePost(p Post) error {

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

	PostCache.Update(p)

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

	PostCache.Delete(id)

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

	PostCache.UpdateMulti(req)

	return nil
}

func (a *postAPI) Count(req PostArgs) (result int, err error) {

	if len(req) == 0 {
		rows, err := DB.Queryx(`SELECT COUNT(*) FROM posts`)
		if err != nil {
			return 0, err
		}
		for rows.Next() {
			if err = rows.Scan(&result); err != nil {
				return 0, err
			}
		}
	} else if len(req) > 0 {

		restricts, values := req.parse()
		query := fmt.Sprintf(`SELECT COUNT(*) FROM (SELECT post_id FROM posts WHERE %s) AS subquery`, restricts)

		query, args, err := sqlx.In(query, values...)
		if err != nil {
			return 0, err
		}
		query = DB.Rebind(query)
		count, err := DB.Queryx(query, args...)
		if err != nil {
			return 0, err
		}
		for count.Next() {
			if err = count.Scan(&result); err != nil {
				return 0, err
			}
		}
	}
	return result, err
}
