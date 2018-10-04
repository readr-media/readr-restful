package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

/* ================================================ Follower API ================================================ */

type FollowArgs struct {
	Resource string
	Subject  int64
	Object   int64
	Type     int
	Emotion  int
}

type GetFollowInterface interface {
	get() (*sqlx.Rows, error)
	scan(*sqlx.Rows) (interface{}, error)
}

type GetFollowingArgs struct {
	MemberID  int64  `form:"id" json:"id"`
	Mode      string `form:"mode"`
	TargetIDs []int
	Active    map[string][]int
	Resource
}

func (g *GetFollowingArgs) get() (*sqlx.Rows, error) {

	var osql = struct {
		base      string
		condition []string
		args      []interface{}
		printargs []interface{}
	}{
		base: `SELECT %s FROM %s AS t 
		INNER JOIN following AS f ON t.%s = f.target_id %s
		WHERE %s ORDER BY f.created_at DESC;`,
		printargs: []interface{}{g.Table, g.PrimaryKey},
		condition: []string{"f.type = ?", "f.member_id = ?", "f.emotion = ?"},
		args:      []interface{}{g.FollowType, g.MemberID, 0},
	}

	if g.Mode == "id" {
		osql.printargs = append([]interface{}{fmt.Sprint("t.", g.PrimaryKey)}, osql.printargs...)
	} else {
		osql.printargs = append([]interface{}{"t.*, f.created_at AS followed_at"}, osql.printargs...)
	}
	if g.Active != nil {
		for k, v := range g.Active {
			osql.condition = append(osql.condition, fmt.Sprintf("t.active %s (?)", operatorHelper(k)))
			osql.args = append(osql.args, v)
		}
	}
	if g.ResourceName == "post" {
		if val, ok := config.Config.Models.PostType[g.ResourceType]; ok {
			osql.condition = append(osql.condition, "t.type = ?")
			osql.args = append(osql.args, val)
		} else if g.ResourceType != "" {
			return nil, errors.New("Invalid Post Type")
		}
	}
	// Append project's tags for each following projects
	if g.ResourceName == "project" {
		osql.printargs[0] = fmt.Sprintf("%s, tag.tags AS tags", osql.printargs[0].(string))
		osql.printargs = append(osql.printargs, fmt.Sprintf(
			`
		 LEFT JOIN (
			SELECT target_id, GROUP_CONCAT(t.tag_content separator '||') as tags 
			FROM tagging 
			LEFT JOIN tags AS t 
				ON tagging.tag_id = t.tag_id 
			WHERE type = %d 
			GROUP BY target_id
		) AS tag 
			ON tag.target_id = t.%s
		`, config.Config.Models.TaggingType[g.ResourceName], g.PrimaryKey))
	} else {
		osql.printargs = append(osql.printargs, "")
	}
	if len(g.TargetIDs) > 0 {
		osql.condition = append(osql.condition, "f.target_id IN (?)")
		osql.args = append(osql.args, g.TargetIDs)
	}
	osql.printargs = append(osql.printargs, strings.Join(osql.condition, " AND "))

	query, args, err := sqlx.In(fmt.Sprintf(osql.base, osql.printargs...), osql.args...)
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (g *GetFollowingArgs) scan(rows *sqlx.Rows) (interface{}, error) {

	var (
		followings []interface{}
		err        error
		IDs        []int
	)

	for rows.Next() {
		if g.Mode == "id" {
			var i int
			err = rows.Scan(&i)
			followings = append(followings, i)
		} else {
			switch g.ResourceName {
			case "post":
				var post struct {
					Post
					FollowedAt NullTime `json:"followed_at" db:"followed_at"`
				}
				err = rows.StructScan(&post)
				followings = append(followings, post)
			case "project":
				var project struct {
					Project
					Tags       []string   `json:"tags"`
					TagString  NullString `json:"-" db:"tags"`
					FollowedAt NullTime   `json:"followed_at" db:"followed_at"`
				}
				err = rows.StructScan(&project)
				if project.TagString.Valid {
					project.Tags = strings.Split(project.TagString.String, "||")
				}
				followings = append(followings, project)
			case "member":
				var member struct {
					Member
					FollowedAt NullTime `json:"followed_at" db:"followed_at"`
				}
				err = rows.StructScan(&member)
				followings = append(followings, member)
			case "memo":
				var memo struct {
					Memo
					FollowedAt NullTime `json:"followed_at" db:"followed_at"`
				}
				err = rows.StructScan(&memo)
				followings = append(followings, memo)
			case "report":
				var report struct {
					Report
					FollowedAt NullTime `json:"followed_at" db:"followed_at"`
				}
				err = rows.StructScan(&report)
				followings = append(followings, report)
			case "tag":
				var tag struct {
					Tag
					FollowedAt NullTime `json:"followed_at" db:"followed_at"`
				}
				err = rows.StructScan(&tag)
				IDs = append(IDs, tag.ID)
				followings = append(followings, tag)
			default:
				return nil, errors.New("Unsupported Resource")
			}
		}
		if err != nil {
			log.Println(err.Error())
			return followings, err
		}
	}

	if g.ResourceName == "tag" && len(IDs) != 0 && g.Mode != "id" {
		followings = make([]interface{}, 0)
		tagDetails, err := TagAPI.GetTags(GetTagsArgs{
			ShowStats:     true,
			ShowResources: true,
			IDs:           IDs,
			PostFields:    sqlfields{"post_id", "publish_status", "published_at", "title", "type"},
			ProjectFields: sqlfields{"project_id", "publish_status", "published_at", "title", "slug", "status", "hero_image"},
			ReportFields:  sqlfields{"id", "publish_status", "published_at", "title", "hero_image", "project_id", "slug"},
		})
		if err != nil {
			log.Println("Error getting tag info when updating hottags:", err)
			return nil, err
		}

		for _, tag := range tagDetails {
			for _, following := range followings {
				f := following.(struct {
					Tag
					FollowedAt NullTime `json:"followed_at" db:"followed_at"`
				})
				if f.Tag.ID == tag.ID {
					followings = append(followings, struct {
						TagRelatedResources
						FollowedAt NullTime `json:"followed_at"`
					}{tag, f.FollowedAt})
					continue
				}
			}
		}
	}

	return followings, err
}

type GetFollowedArgs struct {
	IDs []int64 `json:"ids"`
	Resource
}

func (g *GetFollowedArgs) get() (*sqlx.Rows, error) {

	var osql = struct {
		base      string
		condition []string
		join      []string
		args      []interface{}
	}{
		base: `SELECT f.target_id, COUNT(m.id) as count, 
		GROUP_CONCAT(m.id SEPARATOR ',') as follower FROM following as f 
		LEFT JOIN %s WHERE %s GROUP BY f.target_id;`,
		condition: []string{"f.target_id IN (?)", "f.type = ?", "f.emotion = ?"},
		join:      []string{"members AS m ON f.member_id = m.id"},
		args:      []interface{}{g.IDs, g.FollowType, g.Emotion},
	}

	query, args, err := sqlx.In(fmt.Sprintf(osql.base, strings.Join(osql.join, " LEFT JOIN "), strings.Join(osql.condition, " AND ")), osql.args...)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)

	return DB.Queryx(query, args...)
}

func (g *GetFollowedArgs) scan(rows *sqlx.Rows) (interface{}, error) {

	var (
		followed []FollowedCount
		err      error
	)
	for rows.Next() {
		var (
			resourceID int64
			count      int
			follower   string
		)
		err = rows.Scan(&resourceID, &count, &follower)
		if err != nil {
			log.Fatalln(err.Error())
			return nil, err
		}
		var followers []int64
		for _, v := range strings.Split(follower, ",") {
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			followers = append(followers, int64(i))
		}
		followed = append(followed, FollowedCount{resourceID, count, followers})
	}
	return followed, err
}

type GetFollowMapArgs struct {
	UpdateAfter time.Time `form:"updated_after" json:"updated_after"`
	Resource
}

func (g *GetFollowMapArgs) get() (*sqlx.Rows, error) {
	var osql = struct {
		base      string
		condition []string
		join      []string
		args      []interface{}
	}{
		base: `SELECT GROUP_CONCAT(member_resource.member_id) AS member_ids, member_resource.resource_ids
			FROM (
				SELECT GROUP_CONCAT(f.target_id) AS resource_ids, m.id AS member_id 
				FROM following AS f
				LEFT JOIN %s
				WHERE %s
				GROUP BY m.id
				) AS member_resource
			GROUP BY member_resource.resource_ids;`,
		join:      []string{"members AS m ON f.member_id = m.id", fmt.Sprintf("%s AS t ON f.target_id = t.%s", g.Table, g.PrimaryKey)},
		condition: []string{"m.active = ?", "m.post_push = ?", "f.type = ?"},
		args:      []interface{}{config.Config.Models.Members["active"], 1, g.FollowType},
	}

	switch g.ResourceName {
	case "member":
		osql.join = append(osql.join, "posts AS p ON f.target_id = p.author")
		osql.condition = append(osql.condition, "t.active = ?", "p.active = ?", "p.publish_status = ?", "p.updated_at > ?")
		osql.args = append(osql.args, config.Config.Models.Members["active"], config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"], g.UpdateAfter)
	case "post":
		osql.condition = append(osql.condition, "t.active = ?", "t.publish_status = ?", "t.updated_at > ?")
		osql.args = append(osql.args, config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"], g.UpdateAfter)
	case "project":
		osql.condition = append(osql.condition, "t.active = ?", "t.publish_status = ?", "t.updated_at > ?")
		osql.args = append(osql.args, config.Config.Models.ProjectsActive["active"], config.Config.Models.ProjectsPublishStatus["publish"], g.UpdateAfter)
	}

	rows, err := DB.Queryx(fmt.Sprintf(osql.base, strings.Join(osql.join, " LEFT JOIN "), strings.Join(osql.condition, " AND ")), osql.args...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return rows, err
}

func (g *GetFollowMapArgs) scan(rows *sqlx.Rows) (interface{}, error) {

	var (
		list []FollowingMapItem
		err  error
	)
	for rows.Next() {
		var memberIDs, resourceIDs string
		if err = rows.Scan(&memberIDs, &resourceIDs); err != nil {
			log.Println(err)
			return []FollowingMapItem{}, err
		}
		list = append(list, FollowingMapItem{
			strings.Split(memberIDs, ","),
			strings.Split(resourceIDs, ","),
		})
	}
	return list, err
}

type GetFollowerMemberIDsArgs struct {
	ID         int64
	FollowType int
	Emotions   []int
}

func (g *GetFollowerMemberIDsArgs) get() (*sqlx.Rows, error) {

	query, args, err := sqlx.In(`SELECT member_id FROM following WHERE target_id = ? AND type = ? AND emotion IN (?);`, g.ID, g.FollowType, g.Emotions)
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, args...)
	if err != nil {
		log.Printf("Error: %v get Follower for id:%d, type:%d\n", err.Error(), g.ID, g.FollowType)
	}
	return rows, err

}

func (g *GetFollowerMemberIDsArgs) scan(rows *sqlx.Rows) (interface{}, error) {
	var (
		result []int
		err    error
	)
	for rows.Next() {
		var follower int
		err = rows.Scan(&follower)
		if err != nil {
			log.Printf("Error: %v scan for id:%d, type:%d\n", err.Error(), g.ID, g.FollowType)
			return nil, err
		}
		result = append(result, follower)
	}
	return result, err
}

type Resource struct {
	ResourceName string `form:"resource" json:"resource"`
	ResourceType string `form:"resource_type" json:"resource_type, omitempty"`
	Table        string
	PrimaryKey   string
	FollowType   int
	Emotion      int
}

type FollowedCount struct {
	ResourceID int64   `json:"ResourceID"`
	Count      int     `json:"Count"`
	Followers  []int64 `json:"Followers"`
}

type FollowingMapItem struct {
	Followers   []string `json:"member_ids" db:"member_ids"`
	ResourceIDs []string `json:"resource_ids" db:"resource_ids"`
}

type followingAPI struct{}

type FollowingAPIInterface interface {
	Get(params GetFollowInterface) (interface{}, error)
	Insert(params FollowArgs) error
	Update(params FollowArgs) error
	Delete(params FollowArgs) error
}

func (f *followingAPI) Get(params GetFollowInterface) (result interface{}, err error) {

	var rows *sqlx.Rows

	rows, err = params.get()
	if err != nil {
		log.Println("Error Get Follow with params.get()")
		return nil, err
	}
	return params.scan(rows)
}

func (f *followingAPI) Insert(params FollowArgs) (err error) {

	query := `INSERT INTO following (member_id, target_id, type, emotion) VALUES ( ?, ?, ?, ?);`

	result, err := DB.Exec(query, params.Subject, params.Object, params.Type, params.Emotion)
	if err != nil {
		sqlerr, ok := err.(*mysql.MySQLError)
		if ok && sqlerr.Number == 1062 {
			return DuplicateError
		}
		log.Println(err.Error())
		return InternalServerError
	}
	changed, err := result.RowsAffected()
	if err != nil {
		log.Println(err.Error())
		return InternalServerError
	}
	if changed == 0 {
		return SQLInsertionFail
	}

	return nil
}

func (f *followingAPI) Update(params FollowArgs) (err error) {

	result, err := DB.Exec(`UPDATE following SET emotion = ? WHERE member_id = ? AND target_id = ? AND type = ? AND emotion != 0;`, params.Emotion, params.Subject, params.Object, params.Type)
	if err != nil {
		log.Println(err.Error())
		return InternalServerError
	}
	changed, err := result.RowsAffected()
	if err != nil {
		log.Println(err.Error())
		return InternalServerError
	}
	if changed == 0 {
		return SQLUpdateFail
	}
	return nil
}

func (f *followingAPI) Delete(params FollowArgs) (err error) {

	query := `DELETE FROM following WHERE member_id = ? AND target_id = ? AND type = ? AND emotion = ?;`
	_, err = DB.Exec(query, params.Subject, params.Object, params.Type, params.Emotion)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

var FollowingAPI FollowingAPIInterface = new(followingAPI)
