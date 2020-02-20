package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/args"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/pkg/cards"
	"github.com/readr-media/readr-restful/utils"
)

// Post could use json:"omitempty" tag to ignore null field
// However, struct type field like rrsql.NullTime, rrsql.NullString must be declared as pointer,
// like *rrsql.NullTime, *rrsql.NullString to be used with omitempty
type Post struct {
	ID              uint32           `json:"id" db:"post_id" redis:"post_id"`
	Author          rrsql.NullInt    `json:"author" db:"author" redis:"author"`
	CreatedAt       rrsql.NullTime   `json:"created_at" db:"created_at" redis:"created_at"`
	LikeAmount      rrsql.NullInt    `json:"like_amount" db:"like_amount" redis:"like_amount"`
	CommentAmount   rrsql.NullInt    `json:"comment_amount" db:"comment_amount" redis:"comment_amount"`
	Title           rrsql.NullString `json:"title" db:"title" redis:"title"`
	Subtitle        rrsql.NullString `json:"subtitle" db:"subtitle" redis:"subtitle"`
	Content         rrsql.NullString `json:"content" db:"content" redis:"content"`
	Type            rrsql.NullInt    `json:"type" db:"type" redis:"type"`
	Link            rrsql.NullString `json:"link" db:"link" redis:"link"`
	OgTitle         rrsql.NullString `json:"og_title" db:"og_title" redis:"og_title"`
	OgDescription   rrsql.NullString `json:"og_description" db:"og_description" redis:"og_description"`
	OgImage         rrsql.NullString `json:"og_image" db:"og_image" redis:"og_image"`
	Active          rrsql.NullInt    `json:"active" db:"active" redis:"active"`
	UpdatedAt       rrsql.NullTime   `json:"updated_at" db:"updated_at" redis:"updated_at"`
	UpdatedBy       rrsql.NullInt    `json:"updated_by" db:"updated_by" redis:"updated_by"`
	PublishedAt     rrsql.NullTime   `json:"published_at" db:"published_at" redis:"published_at"`
	LinkTitle       rrsql.NullString `json:"link_title" db:"link_title" redis:"link_title"`
	LinkDescription rrsql.NullString `json:"link_description" db:"link_description" redis:"link_description"`
	LinkImage       rrsql.NullString `json:"link_image" db:"link_image" redis:"link_image"`
	LinkName        rrsql.NullString `json:"link_name" db:"link_name" redis:"link_name"`
	VideoID         rrsql.NullString `json:"video_id" db:"video_id" redis:"video_id"`
	VideoViews      rrsql.NullInt    `json:"video_views" db:"video_views" redis:"video_views"`
	PublishStatus   rrsql.NullInt    `json:"publish_status" db:"publish_status" redis:"publish_status"`
	ProjectID       rrsql.NullInt    `json:"project_id" db:"project_id" redis:"project_id"`
	Order           rrsql.NullInt    `json:"post_order" db:"post_order" redis:"post_order"`
	HeroImage       rrsql.NullString `json:"hero_image" db:"hero_image" redis:"hero_image"`
	Slug            rrsql.NullString `json:"slug" db:"slug" redis:"slug"`
	CSS             rrsql.NullString `json:"css" db:"css" redis:"css"`
	JS              rrsql.NullString `json:"javascript" db:"javascript" redis:"javascript"`
}

type FilteredPost struct {
	ID            int              `json:"id" db:"post_id"`
	Authors       []AuthorBasic    `json:"authors,omitempty"`
	Title         rrsql.NullString `json:"title" db:"title"`
	Type          rrsql.NullInt    `json:"type" db:"type"`
	PublishStatus rrsql.NullInt    `json:"publish_status" db:"publish_status"`
	CreatedAt     rrsql.NullTime   `json:"created_at" db:"created_at"`
}

type PostDescription struct {
	Post
	Tags      rrsql.NullIntSlice `json:"tags" db:"tags"`
	Authors   []AuthorInput      `json:"authors" db:"authors"`
	NewsCards []cards.NewsCard   `json:"cards" db:"cards"`
}

type postAPI struct{}

var PostAPI PostInterface = new(postAPI)

type PostInterface interface {
	DeletePost(id uint32) error
	GetPosts(args *PostArgs) (result []TaggedPostMember, err error)
	GetPost(id uint32, args *PostArgs) (TaggedPostMember, error)
	FilterPosts(args *FilterPostArgs) ([]FilteredPost, error)
	InsertPost(p PostDescription) (lastID int, err error)
	UpdateAll(req PostUpdateArgs) error
	UpdatePost(p PostDescription) error
	Count(req args.ArgsParser) (result int, err error)
	//Hot() (result []HotPost, err error)
	UpdateAuthors(p Post, authors []AuthorInput) (err error)
	SchedulePublish() (ids []uint32, err error)
	GetPostAuthor(id uint32) (member Member, err error)
}

type TaggedPostMember struct {
	Post
	UpdatedBy *MemberBasic    `json:"updated_by,omitempty" db:"updated_by"`
	Tags      []TagBasic      `json:"tags,omitempty" db:"tags"`
	Authors   []AuthorBasic   `json:"authors,omitempty"`
	Comment   []CommentAuthor `json:"comments,omitempty"`
	Cards     []postCard      `json:"cards,omitempty"`
	Project   *ProjectBasic   `json:"project,omitempty" db:"project"`
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

type TagBasic struct {
	ID   int    `json:"id" db:"tag_id"`
	Text string `json:"text" db:"tag_content"`
}

type MemberBasic struct {
	ID           int64            `json:"id" db:"id"`
	UUID         rrsql.NullString `json:"uuid" db:"uuid"`
	Nickname     rrsql.NullString `json:"nickname" db:"nickname"`
	ProfileImage rrsql.NullString `json:"profile_image" db:"profile_image"`
	Description  rrsql.NullString `json:"description" db:"description"`
	Role         rrsql.NullInt    `json:"role" db:"role"`
}

type AuthorBasic struct {
	ID           int64            `json:"id" db:"id"`
	UUID         rrsql.NullString `json:"uuid" db:"uuid"`
	Nickname     rrsql.NullString `json:"nickname" db:"nickname"`
	ProfileImage rrsql.NullString `json:"profile_image" db:"profile_image"`
	Description  rrsql.NullString `json:"description" db:"description"`
	Role         rrsql.NullInt    `json:"role" db:"role"`
	Type         rrsql.NullInt    `json:"author_type" db:"author_type"`
	ResourceID   rrsql.NullInt    `json:"resource_id" db:"resource_id"`
}

type ProjectBasic struct {
	ID            rrsql.NullInt    `json:"id" db:"project_id"`
	HeroImage     rrsql.NullString `json:"hero_image" db:"hero_image"`
	Title         rrsql.NullString `json:"title" db:"title"`
	Description   rrsql.NullString `json:"description" db:"description"`
	OgTitle       rrsql.NullString `json:"og_title" db:"og_title"`
	OgDescription rrsql.NullString `json:"og_description" db:"og_description"`
	OgImage       rrsql.NullString `json:"og_image" db:"og_image"`
	Slug          rrsql.NullString `json:"slug" db:"slug" redis:"slug"`
}

type PostUpdateArgs struct {
	IDs           []int          `json:"ids"`
	UpdatedBy     int64          `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt     rrsql.NullTime `json:"-" db:"updated_at"`
	PublishedAt   rrsql.NullTime `json:"-" db:"published_at"`
	Active        rrsql.NullInt  `json:"-" db:"active"`
	PublishStatus rrsql.NullInt  `json:"-" db:"publish_status"`
}

type AuthorInput struct {
	Type     rrsql.NullInt `json:"author_type" db:"author_type"`
	MemberID rrsql.NullInt `json:"member_id" db:"member_id"`
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
	ShowTag       bool               `form:"show_tag"`
	ShowAuthor    bool               `form:"show_author"`
	ShowUpdater   bool               `form:"show_updater"`
	ShowProject   bool               `form:"show_project"`
	ShowCommment  bool               `form:"show_comment"`
	ShowCard      bool               `form:"show_card"`
	ProjectID     int64              `form:"project_id"`
	Slug          string             `form:"slug"`
	IDs           []uint32           `form:"ids"`
	Active        map[string][]int   `form:"active"`
	PublishStatus map[string][]int   `form:"publish_status"`
	Author        map[string][]int64 `form:"author"`
	Type          map[string][]int   `form:"type"`
	Total         bool               `form:"total"`
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

func (a *PostArgs) ParseQuery() (query string, values []interface{}) {
	memberDBTags := rrsql.GetStructDBTags("full", MemberBasic{})
	selectedFields := []string{"posts.*"}
	joinedTables := make([]string, 0)
	var joinedTableString, restricts string

	if a.ShowUpdater {
		updatedByField := rrsql.MakeFieldString("get", `updated_by.%s "updated_by.%s"`, memberDBTags)
		updatedByIDQuery := strings.Split(updatedByField[0], " ")
		updatedByField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, updatedByIDQuery[0], updatedByIDQuery[1])
		selectedFields = append(selectedFields, updatedByField...)
		joinedTables = append(joinedTables, `LEFT JOIN members AS updated_by ON posts.updated_by = updated_by.id`)
	}

	if a.ShowProject {
		projectDBTags := rrsql.GetStructDBTags("full", ProjectBasic{})
		projectField := rrsql.MakeFieldString("get", `project.%s "project.%s"`, projectDBTags)
		projectIDQuery := strings.Split(projectField[0], " ")
		projectField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectIDQuery[0], projectIDQuery[1])
		selectedFields = append(selectedFields, projectField...)
		joinedTables = append(joinedTables, `LEFT JOIN projects AS project ON posts.project_id = project.project_id`)
	}

	if len(joinedTables) > 0 {
		joinedTableString = strings.Join(joinedTables, " ")
	}

	restricts, restrictVals := a.parseRestricts()
	resultLimit, resultLimitVals := a.parseResultLimit()
	values = append(values, restrictVals...)
	values = append(values, resultLimitVals...)

	query = fmt.Sprintf(`
		SELECT %s FROM posts %s %s `,
		strings.Join(selectedFields, ","),
		joinedTableString,
		restricts+resultLimit,
	)

	return query, values
}
func (a *PostArgs) ParseCountQuery() (query string, values []interface{}) {

	if !a.anyFilter() {
		return `SELECT COUNT(*) FROM posts`, values
	} else {
		restricts, values := a.parseRestricts()
		return fmt.Sprintf(`SELECT COUNT(*) FROM (SELECT post_id FROM posts %s) AS subquery`, restricts), values
	}

}

func (p *PostArgs) parseRestricts() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.active", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.PublishStatus != nil {
		for k, v := range p.PublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.publish_status", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Author != nil {
		for k, v := range p.Author {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.author", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Type != nil {
		for k, v := range p.Type {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.type", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Slug != "" {
		where = append(where, fmt.Sprintf("%s = ?", "posts.slug"))
		values = append(values, p.Slug)
	}
	if p.ProjectID >= 0 {
		where = append(where, fmt.Sprintf("%v = ?", "posts.project_id"))
		values = append(values, p.ProjectID)
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
		restricts = fmt.Sprintf("WHERE %s", strings.Join(where, " AND "))
	} else if len(where) == 1 {
		restricts = fmt.Sprintf("WHERE %s", where[0])
	}
	return restricts, values
}

func (p *PostArgs) parseResultLimit() (restricts string, values []interface{}) {

	if p.Sorting != "" {
		tmp := strings.Split(p.Sorting, ",")
		for i, v := range tmp {
			if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
				tmp[i] = "-posts." + v[1:]
			} else {
				tmp[i] = "posts." + v
			}
		}
		p.Sorting = strings.Join(tmp, ",")
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, rrsql.OrderByHelper(p.Sorting))
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

type FilterPostArgs struct {
	FilterArgs
}

func (p *FilterPostArgs) ParseQuery() (query string, values []interface{}) {
	return p.parse(false)
}
func (p *FilterPostArgs) ParseCountQuery() (query string, values []interface{}) {
	return p.parse(true)
}
func (p *FilterPostArgs) parse(doCount bool) (query string, values []interface{}) {
	fields := rrsql.GetStructDBTags("exist", FilteredPost{})
	for k, v := range fields {
		fields[k] = fmt.Sprintf("posts.%s", v)
	}

	restricts, restrictVals := p.parseRestricts()
	limit, limitVals := p.parseLimit()
	values = append(values, restrictVals...)
	values = append(values, limitVals...)

	var joinedTables []string
	if len(p.Tag) > 0 {
		joinedTables = append(joinedTables, fmt.Sprintf(`
		LEFT JOIN tagging AS tagging ON tagging.target_id = posts.post_id AND tagging.type = %d LEFT JOIN tags AS tags ON tags.tag_id = tagging.tag_id
		`, config.Config.Models.TaggingType["post"]))
	}
	if len(p.Author) > 0 {
		joinedTables = append(joinedTables, `LEFT JOIN authors AS authors ON authors.resource_id = posts.post_id LEFT JOIN members AS members ON authors.author_id = members.id `)
	}

	if doCount {
		query = fmt.Sprintf(`
		SELECT %s FROM posts AS posts %s %s`,
			"COUNT(post_id)",
			strings.Join(joinedTables, " "),
			restricts,
		)
		values = restrictVals
	} else {
		query = fmt.Sprintf(`
		SELECT %s FROM posts AS posts %s %s `,
			strings.Join(fields, ","),
			strings.Join(joinedTables, " "),
			restricts+limit,
		)
	}

	return query, values
}

func (p *FilterPostArgs) parseRestricts() (restrictString string, values []interface{}) {
	restricts := make([]string, 0)

	if p.ID != 0 {
		restricts = append(restricts, `CAST(posts.post_id as CHAR) LIKE ?`)
		values = append(values, fmt.Sprintf("%s%d%s", "%", p.ID, "%"))
	}
	if len(p.Title) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.Title {
			subRestricts = append(subRestricts, `posts.title LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("%s%s%s", "(", strings.Join(subRestricts, " OR "), ")"))
	}
	if len(p.Content) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.Content {
			subRestricts = append(subRestricts, `posts.content LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("%s%s%s", "(", strings.Join(subRestricts, " OR "), ")"))
	}
	if len(p.Author) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.Author {
			subRestricts = append(subRestricts, `members.name LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("%s%s%s", "(", strings.Join(subRestricts, " OR "), ")"))
	}
	if len(p.Tag) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.Tag {
			subRestricts = append(subRestricts, `tags.tag_content LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("(%s)", strings.Join(subRestricts, " OR ")))
	}
	if len(p.PublishedAt) != 0 {
		if v, ok := p.PublishedAt["$gt"]; ok {
			restricts = append(restricts, "posts.published_at >= ?")
			values = append(values, v)
		}
		if v, ok := p.PublishedAt["$lt"]; ok {
			restricts = append(restricts, "posts.published_at <= ?")
			values = append(values, v)
		}
	}
	if len(p.UpdatedAt) != 0 {
		if v, ok := p.UpdatedAt["$gt"]; ok {
			restricts = append(restricts, "posts.updated_at >= ?")
			values = append(values, v)
		}
		if v, ok := p.UpdatedAt["$lt"]; ok {
			restricts = append(restricts, "posts.updated_at <= ?")
			values = append(values, v)
		}
	}
	if len(restricts) > 1 {
		restrictString = fmt.Sprintf("WHERE %s", strings.Join(restricts, " AND "))
	} else if len(restricts) == 1 {
		restrictString = fmt.Sprintf("WHERE %s", restricts[0])
	}
	return restrictString, values
}

func (p *FilterPostArgs) parseLimit() (restricts string, values []interface{}) {

	if p.Sorting != "" {
		tmp := strings.Split(p.Sorting, ",")
		for i, v := range tmp {
			if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
				tmp[i] = "-posts." + v[1:]
			} else {
				tmp[i] = "posts." + v
			}
		}
		p.Sorting = strings.Join(tmp, ",")
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, rrsql.OrderByHelper(p.Sorting))
	}

	if p.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, p.MaxResult)
		if p.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (p.Page-1)*p.MaxResult)
		}
	}
	return restricts, values
}

func (a *postAPI) GetPosts(req *PostArgs) (result []TaggedPostMember, err error) {

	query, args := req.ParseQuery()
	// To give adaptability to where clauses, have to use ... operator here
	// Therefore split query into two parts, assembling them after sqlx.Rebind
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, args...)
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

	postIDs := make([]int, 0)
	for _, v := range result {
		postIDs = append(postIDs, int(v.Post.ID))
	}

	if req.ShowCommment {
		comments, err := a.fetchPostComments(postIDs)
		if err == nil {
			for k, v := range result {
				result[k].Comment = comments[int(v.Post.ID)]
			}
		}
	}
	if req.ShowTag {
		tags, err := a.fetchPostTags(postIDs)
		if err == nil {
			for k, v := range result {
				result[k].Tags = tags[int(v.Post.ID)]
			}
		}
	}
	if req.ShowAuthor {
		authors, err := a.fetchPostAuthors(postIDs)
		if err == nil {
			for k, v := range result {
				result[k].Authors = authors[int(v.Post.ID)]
			}
		}
	}
	if req.ShowCard {
		cards, err := a.fetchPostCards(postIDs)
		if err == nil {
			for k, v := range result {
				result[k].Cards = cards[int(v.Post.ID)]
			}
		}
	}

	return result, err
}

func (a *postAPI) GetPost(id uint32, req *PostArgs) (post TaggedPostMember, err error) {

	req.IDs = []uint32{id}
	req.MaxResult = 1
	query, args := req.ParseQuery()

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return post, err
	}
	query = rrsql.DB.Rebind(query)

	err = rrsql.DB.Get(&post, query, args...)
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
	if err == nil {
		post.Comment = comments[int(post.Post.ID)]
	}

	authors, err := a.fetchPostAuthors([]int{int(post.Post.ID)})
	if err == nil {
		post.Authors = authors[int(post.Post.ID)]
	}

	if req.ShowCard {
		cards, err := a.fetchPostCards([]int{int(post.Post.ID)})
		if err == nil {
			post.Cards = cards[int(post.Post.ID)]
		}
	}

	return post, err
}

type postCommentResource struct {
	ID       int              `db:"post_id"`
	Type     rrsql.NullInt    `db:"type"`
	Slug     rrsql.NullString `db:"slug"`
	Resource string           `db:"-"`
}
type postCard struct {
	ID              uint32           `json:"id" db:"id"`
	PostID          uint32           `json:"post_id" db:"post_id"`
	Title           rrsql.NullString `json:"title" db:"title"`
	Description     rrsql.NullString `json:"description" db:"description"`
	BackgroundImage rrsql.NullString `json:"background_image" db:"background_image"`
	BackgroundColor rrsql.NullString `json:"background_color" db:"background_color"`
	Image           rrsql.NullString `json:"image" db:"image"`
	Video           rrsql.NullString `json:"video" db:"video"`
	Order           rrsql.NullInt    `json:"order" db:"order"`
}

func (a *postAPI) fetchPostComments(ids []int) (comments map[int][]CommentAuthor, err error) {
	postResources, err := a.fetchPostCommentResource(ids)
	if err != nil {
		fmt.Printf("Failed fetch slugs according to post, %s", err.Error())
		return comments, err
	}

	comments = make(map[int][]CommentAuthor, 0)
	resources := make([]string, 0)
	for i, postResource := range postResources {
		var typeName string
		switch int(postResource.Type.Int) {
		case config.Config.Models.PostType["memo"]:
			typeName = "memo"
		case config.Config.Models.PostType["report"]:
			typeName = "report"
		default:
			typeName = "post"
		}
		resourceString := utils.GenerateResourceInfo(typeName, postResource.ID, postResource.Slug.String)
		postResources[i].Resource = resourceString
		resources = append(resources, resourceString)
	}
	commentSet, err := CommentAPI.GetComments(&GetCommentArgs{
		IntraMax: 2,
		Resource: resources,
		Sorting:  "-created_at",
	})

	for _, comment := range commentSet {
		for _, postResource := range postResources {
			if postResource.Resource == comment.Resource.String {
				comments[postResource.ID] = append(comments[postResource.ID], comment)
				break
			}
		}
	}
	return comments, err
}

func (a *postAPI) fetchPostCommentResource(ids []int) (result []postCommentResource, err error) {
	query := fmt.Sprintf(`
		SELECT posts.post_id as post_id, posts.type as type, CASE posts.type WHEN %d THEN projects.slug ELSE posts.slug END as slug FROM posts
		LEFT JOIN projects ON posts.project_id = projects.project_id
		WHERE posts.post_id IN (?);`, config.Config.Models.PostType["memo"])

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return result, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var pcr postCommentResource
		if err = rows.StructScan(&pcr); err != nil {
			return result, err
		}
		result = append(result, pcr)
	}
	return result, err
}

func (a *postAPI) fetchPostTags(ids []int) (tags map[int][]TagBasic, err error) {
	//SELECT t.tag_id, t.tag_content, ti.target_id FROM tagging as ti LEFT JOIN tags as t ON ti.tag_id = t.tag_id WHERE ti.type = 1 AND ti.target_id IN (1928, 1892);
	query := fmt.Sprintf(`
		SELECT t.tag_id, t.tag_content, ti.target_id FROM tagging as ti
		LEFT JOIN tags as t ON ti.tag_id = t.tag_id
		WHERE ti.type = %d AND ti.target_id IN (?)`,
		config.Config.Models.TaggingType["post"])

	tags = make(map[int][]TagBasic, 0)

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return tags, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return tags, err
	}

	for rows.Next() {
		var tag struct {
			TagBasic
			TargetID int64 `db:"target_id"`
		}
		e := rows.StructScan(&tag)
		if e != nil {
			fmt.Println("Post has no author or the author data don't have a corresponding member")
			continue
		}
		tags[int(tag.TargetID)] = append(tags[int(tag.TargetID)], tag.TagBasic)
	}

	return tags, err
}

func (a *postAPI) fetchPostAuthors(ids []int) (authors map[int][]AuthorBasic, err error) {
	query := `SELECT members.id "id",members.uuid "uuid",members.nickname "nickname",members.profile_image "profile_image",members.description "description",members.role "role",authors.author_type "author_type",authors.resource_id "resource_id" FROM posts
		LEFT JOIN authors ON posts.post_id = authors.resource_id
		LEFT JOIN members ON authors.author_id = members.id
		WHERE posts.post_id IN (?);`

	authors = make(map[int][]AuthorBasic, 0)

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return authors, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return authors, err
	}

	for rows.Next() {
		var authorb AuthorBasic
		e := rows.StructScan(&authorb)
		if e != nil {
			fmt.Println("Post has no author or the author data don't have a corresponding member")
			continue
		}
		authors[int(authorb.ResourceID.Int)] = append(authors[int(authorb.ResourceID.Int)], authorb)
	}

	return authors, err
}

func (a *postAPI) fetchPostCards(ids []int) (cards map[int][]postCard, err error) {
	query := `SELECT newscards.id "id",newscards.post_id "post_id",newscards.title "title",newscards.description "description",newscards.background_image "background_image",newscards.background_color "background_color",newscards.image "image",newscards.video "video",newscards.order "order" FROM posts 
		INNER JOIN newscards ON posts.post_id = newscards.post_id 
		WHERE posts.post_id IN (?);`

	cards = make(map[int][]postCard, 0)

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return cards, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return cards, err
	}

	for rows.Next() {
		var pc postCard
		e := rows.StructScan(&pc)
		if e != nil {
			return map[int][]postCard{}, err
		}
		cards[int(pc.PostID)] = append(cards[int(pc.PostID)], pc)
	}

	return cards, err
}

func (a *postAPI) FilterPosts(args *FilterPostArgs) (result []FilteredPost, err error) {
	query, values := args.ParseQuery()
	fmt.Println(query, values)
	rows, err := rrsql.DB.Queryx(query, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var singlePost FilteredPost
		if err = rows.StructScan(&singlePost); err != nil {
			return result, err
		}
		result = append(result, singlePost)
	}

	if len(result) > 0 {

		postIDs := make([]int, 0)
		for _, v := range result {
			postIDs = append(postIDs, int(v.ID))
		}

		authors, err := a.fetchPostAuthors(postIDs)
		if err != nil {
			return result, err
		}
		for k, v := range result {
			result[k].Authors = authors[int(v.ID)]
		}
	}

	return result, err
}

func (a *postAPI) insertPostStms() string {

	tags := rrsql.GetStructDBTags("full", Post{})
	query := fmt.Sprintf(`INSERT INTO posts (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	return query
}

func (a *postAPI) InsertPost(p PostDescription) (lastID int, err error) {

	stmts := []*rrsql.PipelineStmt{}

	stmts = append(stmts, &rrsql.PipelineStmt{
		Query:        a.insertPostStms(),
		Args:         []interface{}{},
		NamedArgs:    p.Post,
		NamedExec:    true,
		RowsAffected: true,
		LastInsertId: true,
	})

	if len(p.Authors) > 0 {
		stmts = append(stmts, a.updateAuthorsStms(p.Post, p.Authors)...)
	}

	if p.Tags.Valid {
		stmts = append(stmts, updateTaggingStmts(config.Config.Models.TaggingType["post"], 0, p.Tags.Slice)...)
	}

	cardSyncStmts, err := cards.BuildSyncStmts(p.Post.ID, p.NewsCards)
	if err != nil {
		log.Println(fmt.Sprintf("Update Post Error while building card sync sql query: %s", err.Error()))
		return 0, err
	}
	stmts = append(stmts, cardSyncStmts...)

	err = rrsql.WithTransaction(rrsql.DB.DB, func(tx *sqlx.Tx) error {
		id, _, err := rrsql.RunPipeline(tx, stmts...)
		lastID = int(id)
		return err
	})

	return lastID, err
}

func (a *postAPI) updatePostStms(p PostDescription) string {

	tags := rrsql.GetStructDBTags("partial", p.Post)
	fields := rrsql.MakeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE posts SET %s WHERE post_id = :post_id`,
		strings.Join(fields, ", "))

	return query
}

func (a *postAPI) UpdatePost(p PostDescription) (err error) {

	stmts := []*rrsql.PipelineStmt{}

	stmts = append(stmts, &rrsql.PipelineStmt{
		Query:     a.updatePostStms(p),
		Args:      []interface{}{},
		NamedArgs: p.Post,
		NamedExec: true,
	})

	if len(p.Authors) > 0 {
		stmts = append(stmts, a.updateAuthorsStms(p.Post, p.Authors)...)
	}

	if p.Tags.Valid {
		stmts = append(stmts, updateTaggingStmts(config.Config.Models.TaggingType["post"], int(p.Post.ID), p.Tags.Slice)...)
	}

	cardSyncStmts, err := cards.BuildSyncStmts(p.Post.ID, p.NewsCards)
	if err != nil {
		log.Println(fmt.Sprintf("Update Post Error while building card sync sql query: %s", err.Error()))
		return err
	}
	stmts = append(stmts, cardSyncStmts...)

	err = rrsql.WithTransaction(rrsql.DB.DB, func(tx *sqlx.Tx) error {
		_, _, err := rrsql.RunPipeline(tx, stmts...)
		return err
	})

	return err
}

func (a *postAPI) DeletePost(id uint32) error {

	// result, err := rrsql.DB.Exec(fmt.Sprintf("UPDATE posts SET active = %d WHERE post_id = ?", int(PostStatus["deactive"].(float64))), id)
	result, err := rrsql.DB.Exec(fmt.Sprintf("UPDATE posts SET active = %d WHERE post_id = ?", config.Config.Models.Posts["deactive"]), id)
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
	go SearchFeed.DeletePost([]int{int(id)})

	return err
}

func (a *postAPI) UpdateAll(req PostUpdateArgs) error {
	updateQuery, updateArgs := req.parse()
	updateQuery = fmt.Sprintf("UPDATE posts SET %s ", updateQuery)

	restrictQuery, restrictArgs, err := sqlx.In(`WHERE post_id IN (?)`, req.IDs)
	if err != nil {
		return err
	}

	restrictQuery = rrsql.DB.Rebind(restrictQuery)
	updateArgs = append(updateArgs, restrictArgs...)

	result, err := rrsql.DB.Exec(fmt.Sprintf("%s %s", updateQuery, restrictQuery), updateArgs...)
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

func (a *postAPI) Count(req args.ArgsParser) (result int, err error) {

	query, values := req.ParseCountQuery()

	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = rrsql.DB.Rebind(query)
	count, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return 0, err
	}
	for count.Next() {
		if err = count.Scan(&result); err != nil {
			return 0, err
		}
	}
	return result, err
}

/*
func (a *postAPI) Hot() (result []HotPost, err error) {
	result, err = RedisHelper.GetHotPosts("postcache_hot_%d", 20)
	if err != nil {
		log.Printf("Error getting popular list: %v", err)
		return result, err
	}
	return result, err
}
*/

func (a *postAPI) updateAuthorsStms(post Post, authors []AuthorInput) (stmts []*rrsql.PipelineStmt) {

	// If post has no id, then give a placeholder for sql trasaction
	postIDString := config.Config.SQL.TrasactionIDPlaceholder
	if post.ID != 0 {
		postIDString = strconv.Itoa(int(post.ID))
	}

	authorCodition := make([]string, 0)
	for _, v := range authors {
		authorCodition = append(authorCodition, fmt.Sprintf(`AND NOT (author_id = %d and author_type = %d)`, v.MemberID.Int, v.Type.Int))
	}

	// Add / update auhtors
	authorInsertions := make([]string, 0)
	for _, v := range authors {
		authorInsertions = append(authorInsertions, fmt.Sprintf(`(%s, %d, %d ,%d)`, postIDString, v.MemberID.Int, post.Type.Int, v.Type.Int))
	}

	stmts = append(stmts,
		&rrsql.PipelineStmt{
			Query: fmt.Sprintf(`DELETE FROM authors WHERE resource_id = ? AND resource_type = ? %s ;`, strings.Join(authorCodition, " ")),
			Args:  []interface{}{postIDString, post.Type.Int}},
		&rrsql.PipelineStmt{
			Query: fmt.Sprintf(`INSERT IGNORE authors (resource_id, author_id, resource_type, author_type) VALUES %s;`, strings.Join(authorInsertions, ",")),
		})
	return stmts
}

func (a *postAPI) UpdateAuthors(post Post, authors []AuthorInput) (err error) {
	// Delete non-exist author

	stmts := a.updateAuthorsStms(post, authors)

	err = rrsql.WithTransaction(rrsql.DB.DB, func(tx *sqlx.Tx) error {
		_, _, err := rrsql.RunPipeline(tx, stmts...)
		return err
	})

	return nil
}

func (a *postAPI) SchedulePublish() (ids []uint32, err error) {
	ids = make([]uint32, 0)
	rows, err := rrsql.DB.Queryx(fmt.Sprintf("SELECT post_id FROM posts WHERE publish_status=%d AND type in (%d,%d,%d,%d) AND published_at <= cast(now() as datetime);",
		config.Config.Models.PostPublishStatus["schedule"],
		config.Config.Models.PostType["review"],
		config.Config.Models.PostType["news"],
		config.Config.Models.PostType["video"],
		config.Config.Models.PostType["live"],
	))
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

	_, err = rrsql.DB.Exec(fmt.Sprintf("UPDATE posts SET publish_status=%d WHERE publish_status=%d AND type in (%d,%d,%d,%d) AND published_at <= cast(now() as datetime);",
		config.Config.Models.PostPublishStatus["publish"],
		config.Config.Models.PostPublishStatus["schedule"],
		config.Config.Models.PostType["review"],
		config.Config.Models.PostType["news"],
		config.Config.Models.PostType["video"],
		config.Config.Models.PostType["live"],
	))
	if err != nil {
		log.Println("Schedul publishing posts fail", err)
		return nil, err
	}

	return ids, nil
}

func (a *postAPI) GetPostAuthor(id uint32) (member Member, err error) {
	query := `SELECT m.* FROM members AS m LEFT JOIN posts AS p ON p.author = m.id WHERE p.post_id = ?;`

	err = rrsql.DB.Get(&member, query, id)
	if err != nil {
		return member, err
	}
	return member, nil
}
