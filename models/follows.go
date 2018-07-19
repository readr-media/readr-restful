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
	MemberID int64 `form:"id" json:"id"`
	Active   map[string][]int
	Resource
}

func (g *GetFollowingArgs) get() (*sqlx.Rows, error) {

	var osql = struct {
		base      string
		condition []string
		args      []interface{}
		printargs []interface{}
	}{
		base: `SELECT t.* FROM %s AS t 
		INNER JOIN following AS f ON t.%s = f.target_id
		WHERE %s ORDER BY f.created_at DESC;`,
		printargs: []interface{}{g.Table, g.PrimaryKey},
		condition: []string{"f.type = ?", "f.member_id = ?", "f.emotion = ?"},
		args:      []interface{}{g.FollowType, g.MemberID, 0},
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
		} else {
			return nil, errors.New("Invalid Post Type")
		}
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
	)

	for rows.Next() {

		switch g.ResourceName {
		case "post":
			var post Post
			err = rows.StructScan(&post)
			followings = append(followings, post)
		case "project":
			var project Project
			err = rows.StructScan(&project)
			followings = append(followings, project)
		case "member":
			var member Member
			err = rows.StructScan(&member)
			followings = append(followings, member)
		case "memo":
			var memo Memo
			err = rows.StructScan(&memo)
			followings = append(followings, memo)
		case "report":
			var report Report
			err = rows.StructScan(&report)
			followings = append(followings, report)
		default:
			return nil, errors.New("Unsupported Resource")
		}

		if err != nil {
			log.Println(err.Error())
			return followings, err
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
		// args:      []interface{}{int(MemberStatus["active"].(float64)), 1, g.FollowType},
	}

	switch g.ResourceName {
	case "member":
		osql.join = append(osql.join, "posts AS p ON f.target_id = p.author")
		osql.condition = append(osql.condition, "t.active = ?", "p.active = ?", "p.publish_status = ?", "p.updated_at > ?")
		osql.args = append(osql.args, config.Config.Models.Members["active"], config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"], g.UpdateAfter)
		// osql.args = append(osql.args, int(MemberStatus["active"].(float64)), int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)), g.UpdateAfter)
	case "post":
		osql.condition = append(osql.condition, "t.active = ?", "t.publish_status = ?", "t.updated_at > ?")
		osql.args = append(osql.args, config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"], g.UpdateAfter)
		// osql.args = append(osql.args, int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)), g.UpdateAfter)
	case "project":
		osql.condition = append(osql.condition, "t.active = ?", "t.publish_status = ?", "t.updated_at > ?")
		osql.args = append(osql.args, config.Config.Models.ProjectsActive["active"], config.Config.Models.ProjectsPublishStatus["publish"], g.UpdateAfter)
		// osql.args = append(osql.args, int(ProjectActive["active"].(float64)), int(ProjectPublishStatus["publish"].(float64)), g.UpdateAfter)
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
}

func (g *GetFollowerMemberIDsArgs) get() (*sqlx.Rows, error) {

	rows, err := DB.Queryx(`SELECT member_id FROM following WHERE target_id = ? AND type = ?;`, g.ID, g.FollowType)
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
	ResourceID int64
	Count      int
	Follower   []int64
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

	if params.Resource == "post" {
		go PostCache.UpdateFollowing("follow", params.Subject, params.Object)
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

	if params.Resource == "post" {
		go PostCache.UpdateFollowing("unfollow", params.Subject, params.Object)
	}
	return err
}

var FollowingAPI FollowingAPIInterface = new(followingAPI)
