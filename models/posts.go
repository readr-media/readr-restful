package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

// var PostStatus map[string]interface{}
// var PostType map[string]interface{}
// var PostPublishStatus map[string]interface{}

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
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	DeletePost(id uint32) error
	GetPosts(args *PostArgs) (result []TaggedPostMember, err error)
	GetPost(id uint32) (TaggedPostMember, error)
	InsertPost(p Post) (int, error)
	UpdateAll(req PostUpdateArgs) error
	UpdatePost(p Post) error
	Count(req *PostArgs) (result int, err error)
	Hot() (result []HotPost, err error)
	SchedulePublish() error
}

type TaggedPostMember struct {
	PostMember
	Tags NullString `json:"-" db:"tags"`
}

type HotPost struct {
	Post
	AuthorNickname     NullString `json:"author_nickname" redis:"author_nickname"`
	AuthorProfileImage NullString `json:"author_profileImage" redis:"author_profileImage"`
}

func (t *TaggedPostMember) MarshalJSON() ([]byte, error) {
	type TPM TaggedPostMember
	type tag struct {
		ID      int    `json:"id"`
		Content string `json:"text"`
	}
	var Tags []tag

	if t.Tags.Valid != false {
		tas := strings.Split(t.Tags.String, ",")
		for _, value := range tas {
			t := strings.Split(value, ":")
			id, _ := strconv.Atoi(t[0])
			Tags = append(Tags, tag{ID: id, Content: t[1]})
		}
	}
	return json.Marshal(&struct {
		LastSeen []tag `json:"tags"`
		*TPM
	}{
		LastSeen: Tags,
		TPM:      (*TPM)(t),
	})
}

// Currently not used. Need to be modified
// func (t *TaggedPostMember) UnmarshalJSON(text []byte) error {
// 	if err := json.Unmarshal(text, *t); err != nil {
// 		return err
// 	}
// 	if t.PostMember.Member.ID != 0 {
// 		t.PostMember.Post.Author = NullInt{Int: t.PostMember.Member.ID, Valid: true}
// 	}
// 	if t.PostMember.UpdatedBy.ID != 0 {
// 		t.PostMember.Post.UpdatedBy = NullInt{Int: t.PostMember.UpdatedBy.ID, Valid: true}
// 	}
// 	return nil
// }

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

// type UpdatedBy Member
type PostMember struct {
	Post
	Member    MemberBasic `json:"author" db:"author"`
	UpdatedBy MemberBasic `json:"updated_by" db:"updated_by"`
}

type PostUpdateArgs struct {
	IDs           []int    `json:"ids"`
	UpdatedBy     string   `form:"updated_by" json:"updated_by" db:"updated_by"`
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
	if p.UpdatedBy != "" {
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
	Active        map[string][]int   `form:"active"`
	PublishStatus map[string][]int   `form:"publish_status"`
	Author        map[string][]int64 `form:"author"`
	Type          map[string][]int   `form:"type"`
}

func (p *PostArgs) Default() (result *PostArgs) {
	return &PostArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (p *PostArgs) DefaultActive() {
	// p.Active = map[string][]int{"$nin": []int{int(PostStatus["deactive"].(float64))}}
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
	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (a *postAPI) GetPosts(req *PostArgs) (result []TaggedPostMember, err error) {

	restricts, values := req.parse()
	tags := getStructDBTags("full", MemberBasic{})
	authorField := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedByField := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)

	authorIDQuery := strings.Split(authorField[0], " ")
	authorField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, authorIDQuery[0], authorIDQuery[1])
	updatedByIDQuery := strings.Split(updatedByField[0], " ")
	updatedByField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, updatedByIDQuery[0], updatedByIDQuery[1])
	query := fmt.Sprintf(`SELECT posts.*, %s, %s, t.tags as tags  FROM posts
		LEFT JOIN members AS author ON posts.author = author.id
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.id
		LEFT JOIN (
			SELECT pt.target_id as post_id, GROUP_CONCAT(CONCAT(t.tag_id, ":", t.tag_content) SEPARATOR ',') as tags
			FROM tagging as pt LEFT JOIN tags as t ON t.tag_id = pt.tag_id WHERE pt.type=%d 
			GROUP BY pt.target_id
		) AS t ON t.post_id = posts.post_id
		WHERE %s `,
		strings.Join(authorField, ","), strings.Join(updatedByField, ","), config.Config.Models.TaggingType["post"], restricts)

	// To give adaptability to where clauses, have to use ... operator here
	// Therefore split query into two parts, assembling them after sqlx.Rebind
	query, args, err := sqlx.In(query, values...)
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
	return result, err
}

func (a *postAPI) GetPost(id uint32) (TaggedPostMember, error) {

	post := TaggedPostMember{}
	tags := getStructDBTags("full", MemberBasic{})
	author := makeFieldString("get", `author.%s "author.%s"`, tags)
	updatedBy := makeFieldString("get", `updated_by.%s "updated_by.%s"`, tags)

	authorIDQuery := strings.Split(author[0], " ")
	author[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, authorIDQuery[0], authorIDQuery[1])
	updatedByIDQuery := strings.Split(updatedBy[0], " ")
	updatedBy[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, updatedByIDQuery[0], updatedByIDQuery[1])

	query := fmt.Sprintf(`SELECT posts.*, %s, %s, t.tags as tags FROM posts
		LEFT JOIN members AS author ON posts.author = author.id 
		LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.id 
		LEFT JOIN (
			SELECT pt.target_id as post_id, GROUP_CONCAT(CONCAT(t.tag_id, ":", t.tag_content) SEPARATOR ',') as tags 
			FROM tagging as pt LEFT JOIN tags as t ON t.tag_id = pt.tag_id 
			GROUP BY pt.target_id
		) AS t ON t.post_id = posts.post_id WHERE posts.post_id = ?`,
		strings.Join(author, ","), strings.Join(updatedBy, ","))

	err := DB.Get(&post, query, id)
	if err != nil {
		log.Println(err.Error())
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
	return post, err
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

	// Only insert a post when it's published
	if !p.Active.Valid || p.Active.Int != int64(config.Config.Models.Posts["deactive"]) {
		if p.PublishStatus.Valid && p.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"]) {
			if p.ID == 0 {
				p.ID = uint32(lastID)
			}
			go PostCache.Insert(p)
			// Write to new post data to search feed
			post, err := PostAPI.GetPost(p.ID)
			if err != nil {
				return 0, err
			}
			go Algolia.InsertPost([]TaggedPostMember{post})
			go NotificationGen.GeneratePostNotifications(post)
		}
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

	go PostCache.Update(p)

	// if (p.PublishStatus.Valid == true && p.PublishStatus.Int != int64(PostPublishStatus["publish"].(float64))) || (p.Active.Valid == true && p.Active.Int != int64(PostStatus["active"].(float64))) {
	if (p.PublishStatus.Valid && p.PublishStatus.Int != int64(config.Config.Models.PostPublishStatus["publish"])) ||
		(p.Active.Valid && p.Active.Int != int64(config.Config.Models.Posts["active"])) {
		// Case: Set a post to unpublished state, Delete the post from cache/searcher
		go Algolia.DeletePost([]int{int(p.ID)})
	} else if p.PublishStatus.Valid || p.Active.Valid {
		// Case: Publish a post. Read whole post from database, then store to cache/searcher
		// Case: Update a post.
		tpm, err := a.GetPost(p.ID)
		if err != nil {
			return err
		}
		activeStatus := tpm.PostMember.Post.Active
		publishStatus := tpm.PostMember.Post.PublishStatus
		if publishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"]) &&
			activeStatus.Int == int64(config.Config.Models.Posts["active"]) {
			go Algolia.InsertPost([]TaggedPostMember{tpm})
			go NotificationGen.GeneratePostNotifications(tpm)
		}
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

	go PostCache.UpdateAll(req)

	// if (req.PublishStatus.Valid == true && req.PublishStatus.Int != int64(PostPublishStatus["publish"].(float64))) || (req.Active.Valid == true && req.Active.Int != int64(PostStatus["active"].(float64))) {
	if (req.PublishStatus.Valid && req.PublishStatus.Int != int64(config.Config.Models.PostPublishStatus["publish"])) ||
		(req.Active.Valid && req.Active.Int != int64(config.Config.Models.Posts["active"])) {
		// Case: Set a post to unpublished state, Delete the post from cache/searcher
		go Algolia.DeletePost(req.IDs)
	} else if req.Active.Valid || req.PublishStatus.Valid {
		// Case: Publish posts. Read those post from database, then store to cache/searcher
		tpms := []TaggedPostMember{}
		for _, id := range req.IDs {
			tpm, err := a.GetPost(uint32(id))
			if err != nil {
				return err
			}
			tpms = append(tpms, tpm)

			if tpm.PublishStatus.Int == int64(config.Config.Models.PostPublishStatus["publish"]) &&
				tpm.Active.Int == int64(config.Config.Models.Posts["active"]) {
				go NotificationGen.GeneratePostNotifications(tpm)
			}
		}
		go Algolia.InsertPost(tpms)
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

func (a *postAPI) SchedulePublish() error {
	_, err := DB.Exec("UPDATE posts SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime);")
	if err != nil {
		return err
	}
	return nil
}
