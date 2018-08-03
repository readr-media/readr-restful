package models

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

type Tag struct {
	ID              int      `json:"id" db:"tag_id"`
	Text            string   `json:"text" db:"tag_content"`
	CreatedAt       NullTime `json:"created_at" db:"created_at"`
	UpdatedAt       NullTime `json:"updated_at" db:"updated_at"`
	UpdatedBy       NullInt  `json:"updated_by" db:"updated_by"`
	Active          NullInt  `json:"active" db:"active"`
	RelatedReviews  NullInt  `json:"related_reviews" db:"related_reviews"`
	RelatedNews     NullInt  `json:"related_news" db:"related_news"`
	RelatedProjects NullInt  `json:"related_projects" db:"related_projects"`
}

type TagInterface interface {
	ToggleTags(args UpdateMultipleTagsArgs) error
	GetTags(args GetTagsArgs) ([]Tag, error)
	InsertTag(tag Tag) (int, error)
	UpdateTag(tag Tag) error
	UpdateTagging(resourceType int, targetID int, tagIDs []int) error
	CountTags(args GetTagsArgs) (int, error)
}

type GetTagsArgs struct {
	MaxResult   uint8  `form:"max_result" json:"max_result"`
	Page        uint16 `form:"page" json:"page"`
	Sorting     string `form:"sort" json:"sort"`
	Keyword     string `form:"keyword" json:"keyword"`
	ShowStats   bool   `form:"stats" json:"stats"`
	TaggingType int    `form:"tagging_type" json:"tagging_type" db:"tagging_type"`
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

	if !utils.ValidateStringArgs(a.Sorting, "-?(text|updated_at|created_at|related_reviews|related_news)") {
		return errors.New("Bad Sorting Option")
	}
	if a.TaggingType != 0 && !utils.ValidateTaggingType(a.TaggingType) {
		return errors.New("Invalid Tagging Type")
	}
	return nil
}

type UpdateMultipleTagsArgs struct {
	IDs       []int    `json:"ids"`
	UpdatedBy string   `form:"updated_by" json:"updated_by" db:"updated_by"`
	UpdatedAt NullTime `json:"-" db:"updated_at"`
	Active    string   `json:"-" db:"active"`
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

func (t *tagApi) GetTags(args GetTagsArgs) (tags []Tag, err error) {

	var query bytes.Buffer

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

		if args.TaggingType == 0 {
			query.WriteString(fmt.Sprintf(` WHERE ta.active=%d `, config.Config.Models.Tags["active"]))
		} else {
			query.WriteString(fmt.Sprintf(` LEFT JOIN tagging AS tg ON ta.tag_id = tg.tag_id WHERE ta.active=%d AND tg.type = :tagging_type`, config.Config.Models.Tags["active"]))
		}
	} else {
		// query.WriteString(fmt.Sprintf(`SELECT ta.* FROM tags as ta WHERE ta.active=%d `, int(TagStatus["active"].(float64))))
		if args.TaggingType == 0 {
			//Avoid unnecessary join
			query.WriteString(fmt.Sprintf(`SELECT ta.* FROM tags as ta WHERE ta.active=%d `, config.Config.Models.Tags["active"]))
		} else {
			query.WriteString(fmt.Sprintf(`SELECT ta.* FROM tags as ta LEFT JOIN tagging AS tg ON ta.tag_id = tg.tag_id WHERE ta.active=%d AND tg.type = :tagging_type`, config.Config.Models.Tags["active"]))
		}
	}

	if args.Keyword != "" {
		query.WriteString(` AND ta.tag_content LIKE :keyword`)
		args.Keyword = "%" + args.Keyword + "%"
	}

	args.Page = (args.Page - 1) * uint16(args.MaxResult)
	query.WriteString(fmt.Sprintf(` ORDER BY %s LIMIT :maxresult OFFSET :page;`, orderByHelper(args.Sorting)))

	rows, err := DB.NamedQuery(query.String(), args)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	tags = []Tag{}
	for rows.Next() {
		var singleTag Tag
		err = rows.StructScan(&singleTag)
		if err != nil {
			tags = []Tag{}
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
		tags = append(tags, singleTag)
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

// var TagStatus map[string]interface{}
var TagAPI TagInterface = new(tagApi)
