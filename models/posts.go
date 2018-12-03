package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

// Post could use json:"omitempty" tag to ignore null field
// However, struct type field like NullTime, NullString must be declared as pointer,
// like *NullTime, *NullString to be used with omitempty
type Post struct {
	ID              uint32     `json:"id" db:"post_id" redis:"post_id"`
	Author          NullInt    `json:"author" db:"author" redis:"author"`
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
	UpdatedBy       NullInt    `json:"updated_by" db:"updated_by" redis:"updated_by"`
	PublishedAt     NullTime   `json:"published_at" db:"published_at" redis:"published_at"`
	LinkTitle       NullString `json:"link_title" db:"link_title" redis:"link_title"`
	LinkDescription NullString `json:"link_description" db:"link_description" redis:"link_description"`
	LinkImage       NullString `json:"link_image" db:"link_image" redis:"link_image"`
	LinkName        NullString `json:"link_name" db:"link_name" redis:"link_name"`
	VideoID         NullString `json:"video_id" db:"video_id" redis:"video_id"`
	VideoViews      NullInt    `json:"video_views" db:"video_views" redis:"video_views"`
	PublishStatus   NullInt    `json:"publish_status" db:"publish_status" redis:"publish_status"`
	ProjectID       NullInt    `json:"project_id" db:"project_id" redis:"project_id"`
	Order           NullInt    `json:"post_order" db:"post_order" redis:"post_order"`
	HeroImage       NullString `json:"hero_image" db:"hero_image" redis:"hero_image"`
	Slug            NullString `json:"slug" db:"slug" redis:"slug"`
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	DeletePost(id uint32) error
	GetPosts(args *PostArgs) (result []TaggedPostMember, err error)
	GetPost(id uint32, args *PostArgs) (TaggedPostMember, error)
	InsertPost(p Post) (int, error)
	UpdateAll(req PostUpdateArgs) error
	UpdatePost(p Post) error
	Count(req *PostArgs) (result int, err error)
	Hot() (result []HotPost, err error)
	SchedulePublish() (ids []uint32, err error)
	GetPostAuthor(id uint32) (member Member, err error)
}

// PostTags is the wrap for NullString used especially in TaggedPostMember
// Because it is convenient to implement MarshalJSON to override Tags default JSON output
// The reason I don't use direct alias, ex: type PostTags NullString,
// is because in this way sqlx could not scan values into the type
type PostTags struct{ NullString }

func (pt PostTags) MarshalJSON() ([]byte, error) {
	type tag struct {
		ID      int    `json:"id"`
		Content string `json:"text"`
	}

	var Tags []tag

	if pt.Valid != false {
		tagPairs := strings.Split(pt.String, ",")
		for _, value := range tagPairs {
			t := strings.Split(value, ":")
			id, _ := strconv.Atoi(t[0])
			Tags = append(Tags, tag{ID: id, Content: t[1]})
		}
	}
	return json.Marshal(Tags)
}

type TaggedPostMember struct {
	Post
	Member    *MemberBasic    `json:"author,omitempty" db:"author"`
	UpdatedBy *MemberBasic    `json:"updated_by,omitempty" db:"updated_by"`
	Tags      PostTags        `json:"tags,omitempty" db:"tags"`
	Comment   []CommentAuthor `json:"comments,omitempty"`
}

// ------------ ↓↓↓ Requirement to satisfy LastPNRInterface  ↓↓↓ ------------

// ReturnPublishedAt is created to return published_at and used in pnr API
func (tpm TaggedPostMember) ReturnPublishedAt() time.Time {
	if tpm.PublishedAt.Valid {
		return tpm.PublishedAt.Time
	}
	return time.Time{}
}

// ReturnCreatedAt is created to return created_at and used in pnr API
func (tpm TaggedPostMember) ReturnCreatedAt() time.Time {
	if tpm.CreatedAt.Valid {
		return tpm.CreatedAt.Time
	}
	return time.Time{}
}

// ReturnUpdatedAt is created to return updated_at and used in pnr API
func (tpm TaggedPostMember) ReturnUpdatedAt() time.Time {
	if tpm.UpdatedAt.Valid {
		return tpm.UpdatedAt.Time
	}
	return time.Time{}
}

// ------------ ↑↑↑ End of requirement to satisfy LastPNRInterface  ↑↑↑ ------------

type HotPost struct {
	Post
	AuthorNickname     NullString `json:"author_nickname" redis:"author_nickname"`
	AuthorProfileImage NullString `json:"author_profileImage" redis:"author_profileImage"`
}

// UpdatedBy wraps Member for embedded field updated_by
// in the usage of anonymous struct in PostMember
type MemberBasic struct {
	ID           int64      `json:"id" db:"id"`
	UUID         NullString `json:"uuid" db:"uuid"`
	Nickname     NullString `json:"nickname" db:"nickname"`
	ProfileImage NullString `json:"profile_image" db:"profile_image"`
	Description  NullString `json:"description" db:"description"`
	Role         NullInt    `json:"role" db:"role"`
}

type PostUpdateArgs struct {
	IDs           []int    `json:"ids"`
	UpdatedBy     int64    `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt     NullTime `json:"-" db:"updated_at"`
	PublishedAt   NullTime `json:"-" db:"published_at"`
	Active        NullInt  `json:"-" db:"active"`
	PublishStatus NullInt  `json:"-" db:"publish_status"`
}

func (p *PostUpdateArgs) parse() (updates string, values []interface{}) {
	setQuery := make([]string, 0)

	if p.Active.Valid {
		setQuery = append(setQuery, "active = ?")
		values = append(values, p.Active.Int)
	}
	if p.PublishStatus.Valid {
		setQuery = append(setQuery, "publish_status = ?")
		values = append(values, p.PublishStatus.Int)
	}
	if p.PublishedAt.Valid {
		setQuery = append(setQuery, "published_at = ?")
		values = append(values, p.PublishedAt.Time)
	}
	if p.UpdatedAt.Valid {
		setQuery = append(setQuery, "updated_at = ?")
		values = append(values, p.UpdatedAt.Time)
	}
	if p.UpdatedBy != 0 {
		setQuery = append(setQuery, "updated_by = ?")
		values = append(values, p.UpdatedBy)
	}
	if len(setQuery) > 1 {
		updates = fmt.Sprintf(" %s", strings.Join(setQuery, " , "))
	} else if len(setQuery) == 1 {
		updates = fmt.Sprintf(" %s", setQuery[0])
	}

	return updates, values
}

// type PostArgs map[string]interface{}
type PostArgs struct {
	MaxResult     uint8              `form:"max_result"`
	Page          uint16             `form:"page"`
	Sorting       string             `form:"sort"`
	IDs           []uint32           `form:"ids"`
	ShowTag       bool               `form:"show_tag"`
	ShowAuthor    bool               `form:"show_author"`
	ShowUpdater   bool               `form:"show_updater"`
	ShowCommment  bool               `form:"show_comment"`
	Active        map[string][]int   `form:"active"`
	PublishStatus map[string][]int   `form:"publish_status"`
	Author        map[string][]int64 `form:"author"`
	Type          map[string][]int   `form:"type"`
	Filter        Filter
}

// NewPostArgs return a PostArgs struct with default settings,
// which could be overriden at any time as long as
// there are functions passed in whose input in *PostArgs
func NewPostArgs(options ...func(*PostArgs)) *PostArgs {
	args := PostArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
	for _, option := range options {
		option(&args)
	}
	return &args
}

func (p *PostArgs) Default() (result *PostArgs) {
	return &PostArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (p *PostArgs) DefaultActive() {
	p.Active = map[string][]int{"$nin": []int{config.Config.Models.Posts["deactive"]}}
}

func (p *PostArgs) anyFilter() (result bool) {
	return p.Active != nil || p.PublishStatus != nil || p.Author != nil || p.Type != nil
}

func (p *PostArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.PublishStatus != nil {
		for k, v := range p.PublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Author != nil {
		for k, v := range p.Author {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.author", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Type != nil {
		for k, v := range p.Type {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.type", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.IDs != nil {
		where = append(where, fmt.Sprintf("%s %s (?)", "posts.post_id", "IN"))
		values = append(values, p.IDs)
	}
	if p.Filter != (Filter{}) {
		where = append(where, fmt.Sprintf("posts.%s %s ?", p.Filter.Field, p.Filter.Operator))
		values = append(values, p.Filter.Condition)
	}
	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (p *PostArgs) parseResultLimit() (restricts string, values []interface{}) {

	if p.Sorting != "" {
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, orderByHelper(p.Sorting))
	}

	if p.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, p.MaxResult)
		if p.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (p.Page-1)*uint16(p.MaxResult))
		}
	}
	return restricts, values
}

func (a *postAPI) GetPosts(req *PostArgs) (result []TaggedPostMember, err error) {

	query, args := a.buildGetQuery(req)
	// To give adaptability to where clauses, have to use ... operator here
	// Therefore split query into two parts, assembling them after sqlx.Rebind
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var singlePost TaggedPostMember
		if err = rows.StructScan(&singlePost); err != nil {
			result = []TaggedPostMember{}
			return result, err
		}
		result = append(result, singlePost)
	}

	if req.ShowCommment {
		postIDs := make([]int, 0)
		for _, v := range result {
			postIDs = append(postIDs, int(v.Post.ID))
		}
		comments, err := a.fetchPostComments(postIDs)
		if err != nil {
			return result, err
		}
		for k, v := range result {
			result[k].Comment = comments[int(v.Post.ID)]
		}
	}

	return result, err
}

func (a *postAPI) GetPost(id uint32, req *PostArgs) (post TaggedPostMember, err error) {

	req.IDs = []uint32{id}
	req.MaxResult = 1
	query, args := a.buildGetQuery(req)

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return post, err
	}
	query = DB.Rebind(query)

	err = DB.Get(&post, query, args...)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			err = errors.New("Post Not Found")
			return TaggedPostMember{}, err
		case err != nil:
			log.Println(err.Error())
			return TaggedPostMember{}, err
		default:
			err = nil
		}
	}

	comments, err := a.fetchPostComments([]int{int(post.Post.ID)})
	if err != nil {
		return post, err
	}
	post.Comment = comments[int(post.Post.ID)]

	return post, err
}

func (a *postAPI) fetchPostComments(ids []int) (comments map[int][]CommentAuthor, err error) {
	comments = make(map[int][]CommentAuthor, 0)
	resources := make([]string, 0)
	for _, id := range ids {
		resources = append(resources, utils.GenerateResourceInfo("post", id, ""))
	}
	commentSet, err := CommentAPI.GetComments(&GetCommentArgs{
		IntraMax: 2,
		Resource: resources,
		Sorting:  "-created_at",
	})
	for _, comment := range commentSet {
		_, postIDString := utils.ParseResourceInfo(comment.Resource.String)
		postID, err := strconv.Atoi(postIDString)
		if err != nil {
			return comments, err
		}
		comments[postID] = append(comments[postID], comment)
	}
	return comments, err
}

func (a *postAPI) buildGetQuery(req *PostArgs) (query string, values []interface{}) {
	memberDBTags := getStructDBTags("full", MemberBasic{})
	selectedFields := []string{"posts.*"}
	joinedTables := make([]string, 0)
	var joinedTableString, restricts string

	if req.ShowAuthor {
		authorField := makeFieldString("get", `author.%s "author.%s"`, memberDBTags)
		authorIDQuery := strings.Split(authorField[0], " ")
		authorField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, authorIDQuery[0], authorIDQuery[1])
		selectedFields = append(selectedFields, authorField...)
		joinedTables = append(joinedTables, `LEFT JOIN members AS author ON posts.author = author.id`)
	}

	if req.ShowUpdater {
		updatedByField := makeFieldString("get", `updated_by.%s "updated_by.%s"`, memberDBTags)
		updatedByIDQuery := strings.Split(updatedByField[0], " ")
		updatedByField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, updatedByIDQuery[0], updatedByIDQuery[1])
		selectedFields = append(selectedFields, updatedByField...)
		joinedTables = append(joinedTables, `LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.id`)
	}

	if req.ShowTag {
		selectedFields = append(selectedFields, "t.tags as tags")
		joinedTables = append(joinedTables, fmt.Sprintf(`
		LEFT JOIN (
			SELECT pt.target_id as post_id, 
				GROUP_CONCAT(CONCAT(t.tag_id, ":", t.tag_content) SEPARATOR ',') as tags
			FROM tagging as pt LEFT JOIN tags as t ON t.tag_id = pt.tag_id WHERE pt.type=%d 
			GROUP BY pt.target_id
		) AS t ON t.post_id = posts.post_id
		`, config.Config.Models.TaggingType["post"]))
	}

	if len(joinedTables) > 0 {
		joinedTableString = strings.Join(joinedTables, " ")
	}

	restricts, restrictVals := req.parse()
	resultLimit, resultLimitVals := req.parseResultLimit()
	values = append(values, restrictVals...)
	values = append(values, resultLimitVals...)

	query = fmt.Sprintf(`
		SELECT %s FROM posts %s WHERE %s `,
		strings.Join(selectedFields, ","),
		joinedTableString,
		restricts+resultLimit,
	)

	return query, values
}

func (a *postAPI) InsertPost(p Post) (int, error) {

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
		log.Printf("Fail to get last insert ID when insert a post: %v", err)
		return 0, err
	}

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

	return err
}

func (a *postAPI) DeletePost(id uint32) error {

	// result, err := DB.Exec(fmt.Sprintf("UPDATE posts SET active = %d WHERE post_id = ?", int(PostStatus["deactive"].(float64))), id)
	result, err := DB.Exec(fmt.Sprintf("UPDATE posts SET active = %d WHERE post_id = ?", config.Config.Models.Posts["deactive"]), id)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Post Not Found")
	}

	go PostCache.Delete(id)
	go Algolia.DeletePost([]int{int(id)})

	return err
}

func (a *postAPI) UpdateAll(req PostUpdateArgs) error {
	updateQuery, updateArgs := req.parse()
	updateQuery = fmt.Sprintf("UPDATE posts SET %s ", updateQuery)

	restrictQuery, restrictArgs, err := sqlx.In(`WHERE post_id IN (?)`, req.IDs)
	if err != nil {
		return err
	}

	restrictQuery = DB.Rebind(restrictQuery)
	updateArgs = append(updateArgs, restrictArgs...)

	result, err := DB.Exec(fmt.Sprintf("%s %s", updateQuery, restrictQuery), updateArgs...)
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

func (a *postAPI) Count(req *PostArgs) (result int, err error) {

	if !req.anyFilter() {
		rows, err := DB.Queryx(`SELECT COUNT(*) FROM posts`)
		if err != nil {
			return 0, err
		}
		for rows.Next() {
			if err = rows.Scan(&result); err != nil {
				return 0, err
			}
		}
	} else {

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

func (a *postAPI) Hot() (result []HotPost, err error) {
	result, err = RedisHelper.GetHotPosts("postcache_hot_%d", 20)
	if err != nil {
		log.Printf("Error getting popular list: %v", err)
		return result, err
	}
	return result, err
}

func (a *postAPI) SchedulePublish() (ids []uint32, err error) {
	ids = make([]uint32, 0)
	rows, err := DB.Queryx("SELECT post_id FROM posts WHERE publish_status=3 AND published_at <= cast(now() as datetime);")
	if err != nil {
		log.Println("Getting post error when schedule publishing posts", err)
		return nil, err
	}

	for rows.Next() {
		var i uint32
		if err = rows.Scan(&i); err != nil {
			continue
		}
		ids = append(ids, i)
	}

	if len(ids) == 0 {
		return ids, err
	}

	_, err = DB.Exec("UPDATE posts SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime);")
	if err != nil {
		log.Println("Schedul publishing posts fail", err)
		return nil, err
	}

	return ids, nil
}

func (a *postAPI) GetPostAuthor(id uint32) (member Member, err error) {
	query := `SELECT m.* FROM members AS m LEFT JOIN posts AS p ON p.author = m.id WHERE p.post_id = ?;`

	err = DB.Get(&member, query, id)
	if err != nil {
		return member, err
	}
	return member, nil
}
