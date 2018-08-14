package models

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"database/sql"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

type Tag struct {
	ID              int      `json:"id" db:"tag_id" redis:"id"`
	Text            string   `json:"text" db:"tag_content" redis:"tag_content"`
	CreatedAt       NullTime `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt       NullTime `json:"updated_at" db:"updated_at" redis:"updated_at"`
	UpdatedBy       NullInt  `json:"updated_by" db:"updated_by" redis:"updated_by"`
	Active          NullInt  `json:"active" db:"active"`
	RelatedReviews  NullInt  `json:"related_reviews" db:"related_reviews"`
	RelatedNews     NullInt  `json:"related_news" db:"related_news"`
	RelatedProjects NullInt  `json:"related_projects" db:"related_projects"`
}

type TagInterface interface {
	ToggleTags(args UpdateMultipleTagsArgs) error
	GetTags(args GetTagsArgs) ([]TagRelatedResources, error)
	InsertTag(tag Tag) (int, error)
	UpdateTag(tag Tag) error
	UpdateTagging(resourceType int, targetID int, tagIDs []int) error
	CountTags(args GetTagsArgs) (int, error)
	GetHotTags() ([]TagRelatedResources, error)
	UpdateHotTags() error
}

type tagApi struct{}

func (t *tagApi) inCondition(isIn bool) string {
	if isIn {
		return "IN"
	} else {
		return "NOT IN"
	}
}

func (t *tagApi) ToggleTags(args UpdateMultipleTagsArgs) error {

	query := fmt.Sprintf("UPDATE tags SET active=%s WHERE tag_id IN (?);", args.Active)
	query, sqlArgs, err := sqlx.In(query, args.IDs)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	query = DB.Rebind(query)

	_, err = DB.Exec(query, sqlArgs...)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

type GetTagsArgs struct {
	MaxResult     uint8     `form:"max_result" json:"max_result"`
	Page          uint16    `form:"page" json:"page"`
	Sorting       string    `form:"sort" json:"sort"`
	Keyword       string    `form:"keyword" json:"keyword"`
	ShowStats     bool      `form:"stats" json:"stats"`
	ShowResources bool      `form:"tagged_resources" json:"tagged_resources"`
	TaggingType   int       `form:"tagging_type" json:"tagging_type" db:"tagging_type"`
	IDs           []int     `form:"ids" json:"-"`
	PostFields    sqlfields `form:"post_fields"`
	ProjectFields sqlfields `form:"project_fields"`
}

func DefaultGetTagsArgs() GetTagsArgs {
	return GetTagsArgs{
		MaxResult: 50,
		Page:      1,
		Sorting:   "-updated_at",
		ShowStats: false,
	}
}

func (a *GetTagsArgs) ValidateGet() error {

	if !utils.ValidateStringArgs(a.Sorting, "-?(text|updated_at|created_at|related_reviews|related_news|related_projects)") {
		return errors.New("Bad Sorting Option")
	}
	if a.TaggingType != 0 && !utils.ValidateTaggingType(a.TaggingType) {
		return errors.New("Invalid Tagging Type")
	}
	return nil
}

func (g *GetTagsArgs) FullPostTags() (result []string) {
	return getStructDBTags("full", Post{})
}

func (g *GetTagsArgs) FullProjectTags() (result []string) {
	return getStructDBTags("full", Project{})
}

type TagRelatedResources struct {
	Tag
	TagPosts    []tagPost    `json:"tagged_posts"`
	TagProjects []tagProject `json:"tagged_projects"`
}

func (t *TagRelatedResources) MarshalJSON() ([]byte, error) {

	values := make(map[string]interface{})

	utils.MarshalIgnoreNullNullable(t.Tag, values)
	utils.MarshalIgnoreNullNullable(*t, values)

	return json.Marshal(values)

}

type tagPost struct {
	Post  `db:"post"`
	TagID int `db:"tag_id"`
}

func (t *tagPost) MarshalJSON() ([]byte, error) {
	values := make(map[string]interface{})
	utils.MarshalIgnoreNullNullable(t.Post, values)
	return json.Marshal(values)
}

type tagProject struct {
	Project `db:"project"`
	TagID   int `db:"tag_id"`
}

func (t *tagProject) MarshalJSON() ([]byte, error) {
	values := make(map[string]interface{})
	utils.MarshalIgnoreNullNullable(t.Project, values)
	return json.Marshal(values)
}

func (t *tagApi) GetTags(args GetTagsArgs) (tags []TagRelatedResources, err error) {

	var query bytes.Buffer
	var queryArgs []interface{}

	if args.ShowStats {
		base := `
		SELECT ta.*, pt.related_reviews, pt.related_news, jt.related_projects FROM tags as ta 
		LEFT JOIN (SELECT t.tag_id as tag_id,
			COUNT(CASE WHEN p.type=%d THEN p.post_id END) as related_reviews,
			COUNT(CASE WHEN p.type=%d THEN p.post_id END) as related_news 
			FROM tagging as t LEFT JOIN posts as p ON t.target_id=p.post_id 
			WHERE t.type=%d GROUP BY t.tag_id ) as pt ON ta.tag_id = pt.tag_id 
		LEFT JOIN (SELECT t.tag_id as tag_id,
			COUNT(p.project_id) as related_projects 
			FROM tagging as t LEFT JOIN projects as p ON t.target_id=p.project_id 
			WHERE t.type=%d GROUP BY t.tag_id ) as jt ON ta.tag_id = jt.tag_id 
		`

		query.WriteString(fmt.Sprintf(base,
			config.Config.Models.PostType["review"],
			config.Config.Models.PostType["news"],
			config.Config.Models.TaggingType["post"],
			config.Config.Models.TaggingType["project"],
		))

	} else {
		query.WriteString(`SELECT ta.* FROM tags as ta `)
	}

	if args.TaggingType == 0 {
		query.WriteString(fmt.Sprintf(` WHERE ta.active=%d `, config.Config.Models.Tags["active"]))
	} else {
		query.WriteString(fmt.Sprintf(` LEFT JOIN tagging AS tg ON ta.tag_id = tg.tag_id WHERE ta.active=%d AND tg.type = ?`, config.Config.Models.Tags["active"]))
		queryArgs = append(queryArgs, args.TaggingType)
	}

	if args.Keyword != "" {
		query.WriteString(` AND ta.tag_content LIKE ?`)
		args.Keyword = "%" + args.Keyword + "%"
		queryArgs = append(queryArgs, args.Keyword)
	}

	if len(args.IDs) > 0 {
		query.WriteString(` AND ta.tag_id IN (?)`)
		queryArgs = append(queryArgs, args.IDs)
	}

	if args.Sorting != "" {
		query.WriteString(fmt.Sprintf(` ORDER BY %s`, orderByHelper(args.Sorting)))
	}

	if args.MaxResult != 0 {
		query.WriteString(` LIMIT ?`)
		queryArgs = append(queryArgs, args.MaxResult)

		if args.Page != 0 {
			args.Page = (args.Page - 1) * uint16(args.MaxResult)
			query.WriteString(` OFFSET ?`)
			queryArgs = append(queryArgs, args.Page)
		}
	}

	queryString, queryArgs, err := sqlx.In(query.String(), queryArgs...)
	if err != nil {
		log.Println("Error parsing IN query when get tag info when updating hottags:", err)
		return nil, err
	}
	queryString = DB.Rebind(queryString)

	rows, err := DB.Queryx(queryString, queryArgs...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	tag_ids := make([]int, 0)
	tags = []TagRelatedResources{}
	for rows.Next() {
		var singleTag Tag
		err = rows.StructScan(&singleTag)
		if err != nil {
			tags = []TagRelatedResources{}
			log.Println(err.Error())
			return tags, err
		}
		if args.ShowStats {
			if !singleTag.RelatedNews.Valid {
				singleTag.RelatedNews = NullInt{0, true}
			}
			if !singleTag.RelatedReviews.Valid {
				singleTag.RelatedReviews = NullInt{0, true}
			}
		}
		tag_ids = append(tag_ids, singleTag.ID)
		tags = append(tags, TagRelatedResources{Tag: singleTag})
	}

	if args.ShowResources {
		// Get Related Post
		relatedPostQuery := fmt.Sprintf(`
			SELECT tg.tag_id, %s 
			FROM tagging AS tg 
			LEFT JOIN posts AS p ON p.post_id = tg.target_id 
			WHERE tg.type = %d AND tg.tag_id IN (?) AND p.active = %d`,
			args.PostFields.GetFields(`p.%s "post.%s"`),
			config.Config.Models.TaggingType["post"],
			config.Config.Models.Posts["active"],
		)
		relatedPostQuery, relatedPostArgs, err := sqlx.In(relatedPostQuery, tag_ids)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		relatedPostQuery = DB.Rebind(relatedPostQuery)
		rows, err := DB.Queryx(relatedPostQuery, relatedPostArgs...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var tp tagPost
			err = rows.StructScan(&tp)
			if err != nil {
				log.Fatalln("Error scan TagPost when getting Tags", err)
				return nil, err
			}
			for k, v := range tags {
				if tp.TagID == v.ID {
					tags[k].TagPosts = append(v.TagPosts, tp)
					break
				}
			}
		}

		// Get Related Projects
		relatedProjectQuery := fmt.Sprintf(`
			SELECT tg.tag_id, %s 
			FROM tagging AS tg 
			LEFT JOIN projects AS p ON p.project_id = tg.target_id 
			WHERE tg.type = %d AND tg.tag_id IN (?) AND p.active = %d`,
			args.ProjectFields.GetFields(`p.%s "project.%s"`),
			config.Config.Models.TaggingType["project"],
			config.Config.Models.ProjectsActive["active"],
		)
		relatedProjectQuery, relatedProjectArgs, err := sqlx.In(relatedProjectQuery, tag_ids)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		relatedProjectQuery = DB.Rebind(relatedProjectQuery)
		rows, err = DB.Queryx(relatedProjectQuery, relatedProjectArgs...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var tp tagProject
			err = rows.StructScan(&tp)
			if err != nil {
				log.Fatalln("Error scan TagProject when getting Tags", err)
				return nil, err
			}
			for k, v := range tags {
				if tp.TagID == v.ID {
					tags[k].TagProjects = append(v.TagProjects, tp)
					break
				}
			}
		}
	}
	return tags, nil
}

func (t *tagApi) InsertTag(tag Tag) (int, error) {
	var existTag Tag
	// query := fmt.Sprint("SELECT * FROM tags WHERE active=", TagStatus["active"].(float64), " AND BINARY tag_content=?;")
	query := fmt.Sprint("SELECT * FROM tags WHERE active=", config.Config.Models.Tags["active"], " AND BINARY tag_content=?;")
	err := DB.Get(&existTag, query, tag.Text)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if existTag.ID > 0 {
		return 0, DuplicateError
	}

	query = fmt.Sprintf(`INSERT INTO tags (tag_content, updated_by) VALUES (?, ?);`)

	result, err := DB.Exec(query, tag.Text, tag.UpdatedBy)
	if err != nil {
		sqlerr, ok := err.(*mysql.MySQLError)
		if ok && sqlerr.Number == 1062 {
			return 0, DuplicateError
		} else {
			return 0, err
		}
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a tag: %v", err)
	}

	return int(lastID), nil
}

type UpdateMultipleTagsArgs struct {
	IDs       []int    `json:"ids"`
	UpdatedBy string   `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt NullTime `json:"-" db:"updated_at"`
	Active    string   `json:"-" db:"active"`
}

func (t *tagApi) UpdateTag(tag Tag) error {

	var existTag Tag
	// query := fmt.Sprint("SELECT * FROM tags WHERE active=", TagStatus["active"].(float64), " AND BINARY tag_content=?;")
	query := fmt.Sprint("SELECT * FROM tags WHERE active=", config.Config.Models.Tags["active"], " AND BINARY tag_content=?;")
	err := DB.Get(&existTag, query, tag.Text)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existTag.ID > 0 {
		return DuplicateError
	}

	dbTags := getStructDBTags("partial", tag)
	fields := makeFieldString("update", `%s = :%s`, dbTags)
	query = fmt.Sprintf(`UPDATE tags SET %s WHERE tag_id = :tag_id`,
		strings.Join(fields, ", "))

	result, err := DB.NamedExec(query, tag)
	if err != nil {
		sqlerr, ok := err.(*mysql.MySQLError)
		if ok && sqlerr.Number == 1062 {
			return DuplicateError
		} else {
			return err
		}
	}

	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return MultipleRowAffectedError
	} else if rowCnt == 0 {
		return ItemNotFoundError
	}

	return nil
}

func (t *tagApi) UpdateTagging(resourceType int, targetID int, tagIDs []int) error {
	//To add new tags and eliminate unwanted tags, we need to perfom two sql queries
	//The update is success only if all query succeed, to make sure this, we use transaction.

	if !utils.ValidateTaggingType(resourceType) {
		return errors.New("Invalid Resource Type")
	}

	tx, err := DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v", err)
		return err
	}

	_ = tx.MustExec(fmt.Sprintf("DELETE FROM tagging WHERE target_id=%d;", targetID))

	if len(tagIDs) > 0 {
		var insqueryBuffer bytes.Buffer
		var insargs []interface{}
		insqueryBuffer.WriteString("INSERT IGNORE INTO tagging (type, tag_id, target_id) VALUES ")
		for index, tagID := range tagIDs {
			insqueryBuffer.WriteString("( ?, ?, ? )")
			insargs = append(insargs, resourceType, tagID, targetID)
			if index < len(tagIDs)-1 {
				insqueryBuffer.WriteString(",")
			} else {
				insqueryBuffer.WriteString(";")
			}
		}
		_ = tx.MustExec(insqueryBuffer.String(), insargs...)
	}
	tx.Commit()

	//If post tag updated, Write to new post data to search feed

	if resourceType == config.Config.Models.TaggingType["post"] {
		post, err := PostAPI.GetPost(uint32(targetID))
		if err != nil {
			return err
		}
		go Algolia.InsertPost([]TaggedPostMember{post})
	}

	return nil
}

func (a *tagApi) CountTags(args GetTagsArgs) (result int, err error) {
	var query bytes.Buffer
	query.WriteString(fmt.Sprintf(`SELECT COUNT(*) FROM tags WHERE active=%d `, config.Config.Models.Tags["active"]))

	if args.Keyword != "" {
		query.WriteString(` AND tag_content LIKE ?`)
		args.Keyword = "%" + args.Keyword + "%"
		err = DB.Get(&result, query.String(), args.Keyword)
	} else {
		err = DB.Get(&result, query.String())
	}

	if err != nil {
		return 0, err
	}

	return result, err
}

func (a *tagApi) GetHotTags() (tags []TagRelatedResources, error error) {
	result, err := RedisHelper.GetHotTags("tag_hot_%d", 20)
	if err != nil {
		log.Printf("Error getting popular list: %v", err)
		return result, err
	}
	return result, err
}

type tagRes struct {
	TagFollows int
	TagScore   int
	PostIDs    []int
	ProjectIDs []int
}

type sortableItem struct {
	ID  int
	Key int
}

type sortableList []sortableItem

func (s sortableList) Len() int           { return len(s) }
func (s sortableList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortableList) Less(i, j int) bool { return s[i].Key > s[j].Key }

type resStats struct {
	Clicks   int
	Emotions int
	Comments int
	Score    int
}

func (a *tagApi) UpdateHotTags() error {
	var tagResources = make(map[int]tagRes, 0)
	var tagResourcesStats = map[string]map[int]resStats{
		"post":    map[int]resStats{},
		"project": map[int]resStats{},
	}

	// Get Tag and Reources
	query := "SELECT tag_id, type, GROUP_CONCAT(target_id) FROM tagging GROUP BY tag_id, type;"
	rows, err := DB.Queryx(query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var resIDs string
		var tagID, resType int
		err = rows.Scan(&tagID, &resType, &resIDs)
		if err != nil {
			log.Fatalln("Scan tag info error when updating hot tags", err)
			return err
		}

		if _, ok := tagResources[tagID]; ok {
			if resType == config.Config.Models.TaggingType["post"] {
				res := tagResources[tagID]
				res.PostIDs = parseIntSlice(resIDs)
				tagResources[tagID] = res
				for _, u := range tagResources[tagID].PostIDs {
					tagResourcesStats["post"][u] = resStats{}
				}
			} else if resType == config.Config.Models.TaggingType["project"] {
				res := tagResources[tagID]
				res.ProjectIDs = parseIntSlice(resIDs)
				tagResources[tagID] = res
				for _, u := range tagResources[tagID].ProjectIDs {
					tagResourcesStats["project"][u] = resStats{}
				}
			}
		} else {
			if resType == config.Config.Models.TaggingType["post"] {
				t := tagRes{PostIDs: parseIntSlice(resIDs)}
				tagResources[tagID] = t
				for _, u := range t.PostIDs {
					tagResourcesStats["post"][u] = resStats{}
				}
			} else if resType == config.Config.Models.TaggingType["project"] {
				t := tagRes{ProjectIDs: parseIntSlice(resIDs)}
				tagResources[tagID] = t
				for _, u := range t.ProjectIDs {
					tagResourcesStats["project"][u] = resStats{}
				}
			}
		}
	}

	// Generate Tag-Related Resource List
	postResourceIDs := getMapKeySlice(tagResourcesStats["post"])
	projectResourceIDs := getMapKeySlice(tagResourcesStats["project"])

	postResourceIDs64 := mapResourceIDint64(postResourceIDs)
	projectResourceIDs64 := mapResourceIDint64(projectResourceIDs)

	_, postResourceStrings := mapResourceString("post", postResourceIDs)
	orderedProjectIDs, projectResourceStrings := mapResourceString("project", projectResourceIDs)

	projectSlugIDMap := map[string]int{}
	for i := 0; i < len(projectResourceStrings); i++ {
		projectSlugIDMap[projectResourceStrings[i]] = orderedProjectIDs[i]
	}

	// ES
	esClient, err := ESConn(map[string]string{
		"url": config.Config.ES.Url,
	})
	if err != nil {
		return err
	}

	esFilterQuery := elastic.NewBoolQuery()
	queryParams := []interface{}{}
	for _, v := range postResourceStrings {
		queryParams = append(queryParams, v)
	}
	for _, v := range projectResourceStrings {
		queryParams = append(queryParams, v)
	}
	esFilterQuery = esFilterQuery.Must([]elastic.Query{
		elastic.NewTermsQuery("jsonPayload.curr-url.keyword", queryParams...),
		elastic.NewTermQuery("jsonPayload.event-type.keyword", "click")}...)
	esCnstScoreQuery := elastic.NewConstantScoreQuery(esFilterQuery)

	esTermAggsQuery := elastic.NewTermsAggregation()
	esTermAggsQuery = esTermAggsQuery.CollectionMode("breadth_first")
	esTermAggsQuery = esTermAggsQuery.Field("jsonPayload.curr-url.keyword")
	esTermAggsQuery = esTermAggsQuery.Size(10000)

	searchResult, err := esClient.Search().Query(esCnstScoreQuery).Aggregation("counts", esTermAggsQuery).Do(context.Background())
	if err != nil {
		panic(err)
	}

	agg, found := searchResult.Aggregations.Terms("counts")
	if !found {
		log.Fatalf("we should have a terms aggregation called %q", "counts")
	}
	for _, bucket := range agg.Buckets {
		t, ID := utils.ParseResourceInfo(bucket.Key.(string))
		if t == "post" {
			iID, _ := strconv.Atoi(ID)
			trs := tagResourcesStats[t][iID]
			trs.Clicks = int(bucket.DocCount)
			tagResourcesStats[t][iID] = trs
		} else if t == "project" {
			iID := projectSlugIDMap[bucket.Key.(string)]
			trs := tagResourcesStats[t][iID]
			trs.Clicks = int(bucket.DocCount)
			tagResourcesStats[t][iID] = trs
		}
	}

	// Resource Following
	postFollowResult, err := FollowingAPI.Get(
		&GetFollowedArgs{
			IDs: postResourceIDs64,
			Resource: Resource{
				FollowType: config.Config.Models.FollowingType["post"],
				Emotion:    0,
			},
		})
	if err != nil {
		log.Println("Fail getting followed members when updating hot tags:", err)
		return err
	}

	projectFollowResult, err := FollowingAPI.Get(
		&GetFollowedArgs{
			IDs: projectResourceIDs64,
			Resource: Resource{
				FollowType: config.Config.Models.FollowingType["project"],
				Emotion:    0,
			},
		})
	if err != nil {
		log.Println("Fail getting followed members when updating hot tags:", err)
		return err
	}

	for _, v := range postFollowResult.([]FollowedCount) {
		ResourceID := int(v.ResourceID)
		res := tagResourcesStats["post"][ResourceID]
		res.Emotions = v.Count
		tagResourcesStats["post"][ResourceID] = res
	}

	for _, v := range projectFollowResult.([]FollowedCount) {
		ResourceID := int(v.ResourceID)
		res := tagResourcesStats["project"][ResourceID]
		res.Emotions = v.Count
		tagResourcesStats["project"][ResourceID] = res
	}

	tagIDs64 := make([]int64, 0)
	for k, _ := range tagResources {
		tagIDs64 = append(tagIDs64, int64(k))
	}

	tagFollowResult, err := FollowingAPI.Get(
		&GetFollowedArgs{
			IDs: tagIDs64,
			Resource: Resource{
				FollowType: config.Config.Models.FollowingType["tag"],
				Emotion:    0,
			},
		})
	if err != nil {
		log.Println("Fail getting followed members when updating hot tags:", err)
		return err
	}

	for _, v := range tagFollowResult.([]FollowedCount) {
		ResourceID := int(v.ResourceID)
		res := tagResources[ResourceID]
		res.TagFollows = v.Count
		tagResources[ResourceID] = res
	}

	// Comment Count
	postCCQuery := "SELECT post_id, IFNULL(comment_amount, 0) AS comment_amount FROM posts WHERE post_id IN (?);"
	postCCQuery, args, err := sqlx.In(postCCQuery, postResourceIDs)
	if err != nil {
		log.Println("Error parsing IN query when get post Commentcount when updating hottags:", err)
		return err
	}
	postCCQuery = DB.Rebind(postCCQuery)
	rows, err = DB.Queryx(postCCQuery, args...)
	if err != nil {
		log.Println("Error get post Commentcount when updating hottags:", err)
		return err
	}

	for rows.Next() {
		var postID, commentCount int
		if err = rows.Scan(&postID, &commentCount); err != nil {
			log.Println("Error scaning query result when updating hottags:", err)
			return err
		}
		res := tagResourcesStats["post"][postID]
		res.Comments = commentCount
		tagResourcesStats["post"][postID] = res
	}

	projectCCQuery := "SELECT project_id, IFNULL(comment_amount, 0) AS comment_amount FROM projects WHERE project_id IN (?);"
	projectCCQuery, args, err = sqlx.In(projectCCQuery, projectResourceIDs)
	if err != nil {
		log.Println("Error parsing IN query when get post Commentcount when updating hottags:", err)
		return err
	}
	projectCCQuery = DB.Rebind(projectCCQuery)
	rows, err = DB.Queryx(projectCCQuery, args...)
	if err != nil {
		log.Println("Error get post Commentcount when updating hottags:", err)
		return err
	}

	for rows.Next() {
		var projectID, commentCount int
		if err = rows.Scan(&projectID, &commentCount); err != nil {
			log.Println("Error scaning query result when updating hottags:", err)
			return err
		}
		res := tagResourcesStats["project"][projectID]
		res.Comments = commentCount
		tagResourcesStats["project"][projectID] = res
	}

	// Calculate tag score
	for k, v := range tagResourcesStats["post"] {
		v.Score = v.Clicks + v.Emotions + v.Comments
		tagResourcesStats["post"][k] = v
	}
	for k, v := range tagResourcesStats["project"] {
		v.Score = v.Clicks + v.Emotions + v.Comments
		tagResourcesStats["project"][k] = v
	}

	for k, v := range tagResources {
		v.TagScore += v.TagFollows
		for _, postID := range v.PostIDs {
			v.TagScore += tagResourcesStats["post"][postID].Score
		}
		for _, projectID := range v.ProjectIDs {
			v.TagScore += tagResourcesStats["project"][projectID].Score
		}
		tagResources[k] = v
	}

	// Sort Score
	var sl sortableList
	for k, v := range tagResources {
		sl = append(sl, sortableItem{k, v.TagScore})
	}
	sort.Sort(sl)
	limit := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}(len(sl), 20)

	var tagIDs []int
	for _, v := range sl[:limit] {
		tagIDs = append(tagIDs, v.ID)
	}

	// Get Tag Info
	getTagArgs := GetTagsArgs{
		ShowStats:     true,
		ShowResources: true,
		IDs:           tagIDs,
		PostFields:    sqlfields{"post_id", "publish_status", "published_at", "title", "type"},
		ProjectFields: sqlfields{"project_id", "publish_status", "published_at", "title", "slug", "status"},
	}

	tagDetails, err := TagAPI.GetTags(getTagArgs)
	if err != nil {
		log.Println("Error getting tag info when updating hottags:", err)
		return err
	}

	// Store to Redis
	// Write post data, post followers, post score to Redis
	conn := RedisHelper.Conn()
	defer conn.Close()

	if _, err := conn.Do("DEL", redis.Args{}.Add("tagcache_hot")); err != nil {
		log.Printf("Error delete cache from redis: %v", err)
		return err
	}

	conn.Send("MULTI")
	for _, tagInfo := range tagDetails {
		for rank, tagID := range tagIDs {
			if tagInfo.ID == tagID {
				tagInfos, _ := json.Marshal(tagInfo)
				conn.Send("HSET", redis.Args{}.Add("tagcache_hot").Add(rank+1).Add(tagInfos)...)
				break
			}
		}
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return err
	}

	return nil
}

func parseIntSlice(ids string) []int {
	ss := strings.Split(ids, ",")
	is := make([]int, len(ss))
	for i, v := range ss {
		id, _ := strconv.Atoi(v)
		is[i] = id
	}
	return is
}

func getMapKeySlice(m map[int]resStats) []int {
	ks := make([]int, len(m))
	i := 0
	for k := range m {
		ks[i] = k
		i++
	}
	return ks
}

func mapResourceString(resType string, resIDs []int) (orderedResIDs []int, resStrings []string) {
	if resType == "post" {
		for _, resID := range resIDs {
			resStrings = append(resStrings, utils.GenerateResourceInfo(resType, resID, ""))
		}
	} else if resType == "project" {
		query := fmt.Sprintf("SELECT project_id, slug FROM projects WHERE project_id IN (?) AND active = %d", config.Config.Models.Tags["active"])
		query, args, err := sqlx.In(query, resIDs)
		if err != nil {
			log.Println("Error parsing IN query when mapResourceString:", err)
			return orderedResIDs, resStrings
		}
		query = DB.Rebind(query)
		rows, err := DB.Queryx(query, args...)
		if err != nil {
			log.Println("Error querying project slugs when mapResourceString:", err)
			return orderedResIDs, resStrings
		}

		for rows.Next() {
			var id int
			var slug string
			if err = rows.Scan(&id, &slug); err != nil {
				log.Println("Error scaning query result when mapResourceString:", err)
				return []int{}, []string{}
			}
			resStrings = append(resStrings, utils.GenerateResourceInfo(resType, id, slug))
			orderedResIDs = append(orderedResIDs, id)
		}
	}
	return orderedResIDs, resStrings
}

func mapResourceIDint64(res []int) (res64 []int64) {
	res64 = make([]int64, len(res))
	for i := 0; i < len(res); i++ {
		res64[i] = int64(res[i])
	}
	return res64
}

// var TagStatus map[string]interface{}
var TagAPI TagInterface = new(tagApi)
