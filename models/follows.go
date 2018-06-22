package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

/* ================================================ Types of Followings ================================================ */

type follow interface {
	Delete() (sql.Result, error)
	GetFollowed(args GetFollowedArgs) (*sqlx.Rows, error)
	GetFollowings(args GetFollowingArgs) (*sqlx.Rows, error)
	GetFollowerMemberIDs(id string) ([]int, error)
	GetMap(args GetFollowMapArgs) (*sqlx.Rows, error)
	Insert() (sql.Result, error)
}

type followPost struct {
	ID     int64
	Object int64
}

func (f followPost) Insert() (sql.Result, error) {
	query := "INSERT INTO following_posts (member_id, post_id) VALUES ( ? , ? );"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followPost) Delete() (sql.Result, error) {
	query := "DELETE FROM following_posts WHERE member_id = ? AND post_id = ?;"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followPost) GetFollowings(params GetFollowingArgs) (*sqlx.Rows, error) {
	post_type := f.postTypeHelper(params.Type)
	return DB.Queryx(fmt.Sprintf("SELECT p.* from posts as p INNER JOIN following_posts as f ON p.post_id = f.post_id WHERE p.active = 1 AND f.member_id = ? %s ORDER BY f.created_at DESC;", post_type), f.ID)
}
func (f followPost) GetFollowed(params GetFollowedArgs) (*sqlx.Rows, error) {
	post_type := f.postTypeHelper(params.Type)
	query, args, err := sqlx.In(fmt.Sprintf("SELECT f.post_id, COUNT(m.id) as count, GROUP_CONCAT(m.id SEPARATOR ',') as follower FROM posts AS p LEFT JOIN following_posts as f ON f.post_id = p.post_id LEFT JOIN members as m ON f.member_id = m.id WHERE f.post_id IN (?) %s GROUP BY f.post_id;", post_type), params.Ids)
	if err != nil {
		return nil, err
	}
	return DB.Queryx(query, args...)
}
func (f followPost) GetFollowerMemberIDs(id string) (result []int, err error) {
	rows, err := DB.Query(fmt.Sprintf(`SELECT member_id AS member_id FROM following_posts WHERE post_id=%s;`, id))
	if err != nil {
		log.Println("Error get postFollowers", id, err.Error())
	}
	for rows.Next() {
		var follower int
		err = rows.Scan(&follower)
		if err != nil {
			log.Println("Error scan postFollowers", id, err.Error())
		}
		result = append(result, follower)
	}
	return result, err
}
func (f followPost) GetMap(params GetFollowMapArgs) (*sqlx.Rows, error) {
	post_type := f.postTypeHelper(params.Type)
	query := fmt.Sprintf(`
		SELECT GROUP_CONCAT(member_resource.member_id) AS member_ids, member_resource.resource_ids
		FROM (
			SELECT GROUP_CONCAT(fp.post_id) AS resource_ids, m.id AS member_id
			FROM following_posts AS fp
			LEFT JOIN members AS m ON fp.member_id = m.id
			LEFT JOIN posts AS p ON p.post_id = fp.post_id
			WHERE m.active = ?
				AND m.post_push = ?
				AND p.active = ?
				AND p.publish_status = ?
				AND p.updated_at > ?
				%s
			GROUP BY m.id
			) AS member_resource
		GROUP BY member_resource.resource_ids;`, post_type)
	return DB.Queryx(query, int(MemberStatus["active"].(float64)), 1, int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)), params.UpdateAfter)
}
func (f followPost) postTypeHelper(post_type string) string {
	if t := PostType[post_type]; t == nil {
		return ""
	} else {
		return fmt.Sprintf(" AND p.type = %d", int(t.(float64)))
	}
}

type followMember struct {
	ID     int64
	Object int64
}

func (f followMember) Insert() (sql.Result, error) {
	query := "INSERT INTO following_members (member_id, custom_editor) VALUES ( ? , ? );"
	return DB.Exec(query, f.ID, f.Object)
}

func (f followMember) Delete() (sql.Result, error) {
	query := "DELETE FROM following_members WHERE member_id = ? AND custom_editor = ?;"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followMember) GetFollowings(params GetFollowingArgs) (*sqlx.Rows, error) {
	return DB.Queryx("SELECT m.* from members as m INNER JOIN following_members as f ON m.id = f.custom_editor WHERE m.active > 0 AND f.member_id = ? ORDER BY f.created_at DESC;", f.ID)
}
func (f followMember) GetFollowed(params GetFollowedArgs) (*sqlx.Rows, error) {
	query, args, err := sqlx.In("SELECT f.custom_editor, COUNT(m.id) as count, GROUP_CONCAT(m.id SEPARATOR ',') as follower FROM following_members as f LEFT JOIN members as m ON f.member_id = m.id WHERE f.custom_editor IN (?) GROUP BY f.custom_editor;", params.Ids)
	if err != nil {
		return nil, err
	}
	fmt.Printf("followMember GetFollwed sql:%s,\nargs:%v\n", query, args)
	return DB.Queryx(query, args...)
}
func (f followMember) GetFollowerMemberIDs(id string) (result []int, err error) {
	rows, err := DB.Query(fmt.Sprintf(`SELECT member_id FROM following_members WHERE custom_editor="%s";`, id))
	if err != nil {
		log.Println("Error get authorFollowers", id, err.Error())
	}
	for rows.Next() {
		var follower int
		err = rows.Scan(&follower)
		if err != nil {
			log.Println("Error scan authorFollowers", id, err.Error())
		}
		result = append(result, follower)
	}
	return result, err
}
func (f followMember) GetMap(params GetFollowMapArgs) (*sqlx.Rows, error) {
	query := `
		SELECT GROUP_CONCAT(member_resource.member_id) AS member_ids, member_resource.resource_ids
		FROM (
			SELECT GROUP_CONCAT(p.post_id) AS resource_ids, m.id AS member_id
			FROM following_members AS f
			LEFT JOIN members AS m ON f.member_id = m.id
			LEFT JOIN members AS e ON f.custom_editor = e.id
			LEFT JOIN posts AS p ON f.custom_editor = p.author
			WHERE m.active = ?
				AND m.post_push = ?
				AND e.active = ?
				AND p.active = ?
				AND p.publish_status = ?
				AND p.updated_at > ?
			GROUP BY m.id ) AS member_resource
		GROUP BY member_resource.resource_ids;`
	return DB.Queryx(query, int(MemberStatus["active"].(float64)), 1, int(MemberStatus["active"].(float64)), int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)), params.UpdateAfter)
}

type followProject struct {
	ID     int64
	Object int64
}

func (f followProject) Insert() (sql.Result, error) {
	query := "INSERT INTO following_projects (member_id, project_id) VALUES ( ? , ? );"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followProject) Delete() (sql.Result, error) {
	query := "DELETE FROM following_projects WHERE member_id = ? AND project_id = ?;"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followProject) GetFollowings(params GetFollowingArgs) (*sqlx.Rows, error) {
	return DB.Queryx("SELECT m.* from projects as m INNER JOIN following_projects as f ON m.project_id = f.project_id WHERE m.active = 1  AND f.member_id = ? ORDER BY f.created_at DESC;", f.ID)
}
func (f followProject) GetFollowed(params GetFollowedArgs) (*sqlx.Rows, error) {
	query, args, err := sqlx.In("SELECT f.project_id, COUNT(m.id) as count, GROUP_CONCAT(m.id SEPARATOR ',') as follower FROM following_projects as f LEFT JOIN members as m ON f.member_id = m.id WHERE f.project_id IN (?) GROUP BY f.project_id;", params.Ids)
	if err != nil {
		return nil, err
	}
	return DB.Queryx(query, args...)
}
func (f followProject) GetFollowerMemberIDs(id string) (result []int, err error) {
	rows, err := DB.Query(fmt.Sprintf(`SELECT member_id FROM following_projects WHERE project_id=%s;`, id))
	if err != nil {
		log.Println("Error get projectFollowers", id, err.Error())
	}
	for rows.Next() {
		var follower int
		err = rows.Scan(&follower)
		if err != nil {
			log.Println("Error scan authorFollowers", id, err.Error())
		}
		result = append(result, follower)
	}
	return result, err
}
func (f followProject) GetMap(params GetFollowMapArgs) (*sqlx.Rows, error) {
	query := `
		SELECT GROUP_CONCAT(member_resource.member_id) AS member_ids, member_resource.resource_ids
		FROM (
			SELECT GROUP_CONCAT(f.project_id) AS resource_ids, m.id AS member_id 
			FROM following_projects AS f
			LEFT JOIN members AS m ON f.member_id = m.id
			LEFT JOIN projects AS p ON f.project_id = p.project_id
			WHERE m.active = ?
				AND m.post_push = ?
				AND p.active = ?
				AND p.publish_status = ?
				AND p.updated_at > ?
			GROUP BY m.id
			) AS member_resource
		GROUP BY member_resource.resource_ids;`
	return DB.Queryx(query, int(MemberStatus["active"].(float64)), 1, int(ProjectActive["active"].(float64)), int(ProjectPublishStatus["publish"].(float64)), params.UpdateAfter)
}

/* ================================================ Follower Factory ================================================ */

func followPostFactory(args FollowArgs) follow {
	return followPost{
		ID:     args.Subject,
		Object: args.Object,
	}
}
func followMemberFactory(args FollowArgs) follow {
	return followMember{
		ID:     args.Subject,
		Object: args.Object,
	}
}
func followProjectFactory(args FollowArgs) follow {
	return followProject{
		ID:     args.Subject,
		Object: args.Object,
	}
}

var followFactories = map[string]func(conf FollowArgs) follow{
	"post":    followPostFactory,
	"member":  followMemberFactory,
	"project": followProjectFactory,
}

func CreateFollow(args FollowArgs) (follow, error) {
	followFactory, ok := followFactories[args.Resource]
	if !ok {
		return nil, errors.New("Resource Not Supported")
	}
	return followFactory(args), nil
}

/* ================================================ Follower API ================================================ */

type FollowingAPIInterface interface {
	AddFollowing(params FollowArgs) error
	DeleteFollowing(params FollowArgs) error
	GetFollowing(params GetFollowingArgs) ([]interface{}, error)
	GetFollowed(args GetFollowedArgs) (interface{}, error)
	GetFollowerMemberIDs(resourceType string, id string) ([]int, error)
	GetFollowMap(args GetFollowMapArgs) ([]FollowingMapItem, error)

	PseudoAddFollowing(params FollowArgs) error
	PseudoDeleteFollowing(params FollowArgs) error
	PseudoGetFollowed(params GetFollowArgs) (interface{}, error)
	PseudoGetFollowing(params GetFollowArgs) (following []interface{}, err error)
	PseudoGetFollowerMemberIDs(id int64, followType int) (result []int, err error)
	PseudoGetFollowMap(args GetFollowArgs) (list []FollowingMapItem, err error)
}

type FollowArgs struct {
	Resource string
	Subject  int64
	Object   int64
	Type     int
	Emotion  int
}

type GetFollowingArgs struct {
	MemberId int64  `json:"subject"`
	Resource string `json:"resource"`
	Type     string `json:"resource_type,omitempty"`
}

type GetFollowedArgs struct {
	Ids      []string `json:"ids"`
	Resource string   `json:"resource"`
	Type     string   `json:"resource_type,omitempty"`
}

type GetFollowMapArgs struct {
	Resource    string    `json:"resource"`
	Type        string    `json:"resource_type,omitempty"`
	UpdateAfter time.Time `json:"updated_after"`
}

type followedCount struct {
	Resourceid int64
	Count      int
	Follower   []int64
}

type FollowerInfo struct {
	MemberId string `json:"member_ids" db:"member_ids"`
	Name     string `json:"resource_ids" db:"resource_ids"`
	Mail     string
}

type FollowingMapItem struct {
	Followers   []string `json:"member_ids" db:"member_ids"`
	ResourceIDs []string `json:"resource_ids" db:"resource_ids"`
}

type followingAPI struct{}

// ============================= Refactor ======================
type GetFollowArgs struct {
	IDs         []int64   `json:"ids"`
	UpdateAfter time.Time `form:"updated_after" json:"updated_after"`
	Type        int
	Method      string
	Active      map[string][]int
	Resource
}

type Resource struct {
	ResourceName string `form:"resource" json:"resource"`
	ResourceType string `form:"resource_type" json:"resource_type, omitempty"` //review, report ...
	TableName    string
	KeyName      string
}

func (g GetFollowArgs) following() (result string, values []interface{}) {

	var osql = struct {
		base      string
		condition []string
		args      []interface{}
		printargs []interface{}
	}{
		base: `SELECT t.* FROM %s AS t 
		INNER JOIN following AS f ON t.%s = f.target_id
		WHERE %s ORDER BY f.created_at DESC;`,
		printargs: []interface{}{g.TableName, g.KeyName},
		condition: []string{"f.type = ?", "f.member_id = ?"},
		args:      []interface{}{g.Type, g.IDs[0]},
	}
	if g.Active != nil {
		for k, v := range g.Active {
			osql.condition = append(osql.condition, fmt.Sprintf("t.active %s (?)", operatorHelper(k)))
			osql.args = append(osql.args, v)
		}
	}
	osql.printargs = append(osql.printargs, strings.Join(osql.condition, " AND "))

	return fmt.Sprintf(osql.base, osql.printargs...), osql.args
}

func (g GetFollowArgs) followed() (result string, values []interface{}) {

	var osql = struct {
		base      string
		condition []string
		join      []string
		args      []interface{}
	}{
		base: `SELECT f.target_id, COUNT(m.id) as count, 
		GROUP_CONCAT(m.id SEPARATOR ',') as follower FROM following as f 
		LEFT JOIN %s WHERE %s GROUP BY f.target_id;`,
		condition: []string{"f.target_id IN (?)", "f.type = ?"},
		join:      []string{"members AS m ON f.member_id = m.id"},
		args:      []interface{}{g.IDs, g.Type},
	}
	if g.ResourceName == "post" {
		osql.join = append(osql.join, "posts AS p ON f.target_id = p.post_id")

		if t := PostType[g.ResourceType]; t != nil {
			osql.condition = append(osql.condition, "p.type = ?")
			osql.args = append(osql.args, int(t.(float64)))
		}
	}
	result = fmt.Sprintf(osql.base, strings.Join(osql.join, " LEFT JOIN "), strings.Join(osql.condition, " AND "))
	return result, osql.args
}

func (g GetFollowArgs) getmap() (result string, values []interface{}) {

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
		join:      []string{"members AS m ON f.member_id = m.id", fmt.Sprintf("%s AS t ON f.target_id = t.%s", g.TableName, g.KeyName)},
		condition: []string{"m.active = ?", "m.post_push = ?"},
		args:      []interface{}{int(MemberStatus["active"].(float64)), 1},
	}

	switch g.ResourceName {
	case "member":
		osql.join = append(osql.join, "posts AS p ON f.target_id = p.author")
		osql.condition = append(osql.condition, "t.active = ?", "p.active = ?", "p.publish_status = ?", "p.updated_at > ?")
		osql.args = append(osql.args, int(MemberStatus["active"].(float64)), int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)), g.UpdateAfter)
	case "post":
		osql.condition = append(osql.condition, "t.active = ?", "t.publish_status = ?", "t.updated_at > ?")
		osql.args = append(osql.args, int(PostStatus["active"].(float64)), int(PostPublishStatus["publish"].(float64)), g.UpdateAfter)
	case "project":
		osql.condition = append(osql.condition, "t.active = ?", "t.publish_status = ?", "t.updated_at > ?")
		osql.args = append(osql.args, int(ProjectActive["active"].(float64)), int(ProjectPublishStatus["publish"].(float64)), g.UpdateAfter)
	}
	result = fmt.Sprintf(osql.base, strings.Join(osql.join, " LEFT JOIN "), strings.Join(osql.condition, " AND "))
	return result, osql.args
}

func (f *followingAPI) PseudoAddFollowing(params FollowArgs) (err error) {
	query := `INSERT INTO following_posts (member_id, target_id, type, emotion) VALUES ( ? , ?, ?, ?);`

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

func (f *followingAPI) PseudoDeleteFollowing(params FollowArgs) (err error) {

	query := `DELETE FROM following_posts WHERE member_id = ? AND target_id = ? AND type = ? AND emotion = ?;`
	_, err = DB.Exec(query, params.Subject, params.Object, params.Type, params.Emotion)
	if err != nil {
		log.Fatal(err)
	}

	if params.Resource == "post" {
		go PostCache.UpdateFollowing("unfollow", params.Subject, params.Object)
	}

	return err
}

// Following is the subject a user following
func (f *followingAPI) PseudoGetFollowing(params GetFollowArgs) (followings []interface{}, err error) {

	query, args := params.following()

	query, args, err = sqlx.In(query, args...)
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		switch params.ResourceName {
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
		}

		if err != nil {
			log.Println(err.Error())
			return followings, err
		}
	}

	if len(followings) == 0 {
		err = errors.New("Not Found")
	}

	return followings, err
}

func (f *followingAPI) PseudoGetFollowed(params GetFollowArgs) (result interface{}, err error) {

	query, args := params.followed()

	query, args, err = sqlx.In(query, args...)
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}

	followed := []interface{}{}
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
		followed = append(followed, struct {
			ResourceID int64
			Count      int
			Followers  []int64
		}{resourceID, count, followers})
	}
	return followed, nil
}

func (f *followingAPI) GetFollowing(params GetFollowingArgs) (followings []interface{}, err error) {
	follow, err := CreateFollow(FollowArgs{Resource: params.Resource, Subject: params.MemberId})
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	rows, err := follow.GetFollowings(params)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	for rows.Next() {
		if params.Resource == "post" {
			var post Post
			err = rows.StructScan(&post)
			followings = append(followings, post)
		} else if params.Resource == "project" {
			var project Project
			err = rows.StructScan(&project)
			followings = append(followings, project)
		} else if params.Resource == "member" {
			var member Member
			err = rows.StructScan(&member)
			followings = append(followings, member)
		}

		if err != nil {
			log.Println(err.Error())
			return followings, err
		}
	}

	if len(followings) == 0 {
		err = errors.New("Not Found")
	}

	return followings, err
}
func (f *followingAPI) PseudoGetFollowerMemberIDs(id int64, followType int) (result []int, err error) {

	rows, err := DB.Query(`SELECT member_id FROM following WHERE target_id = ? AND type = ?;`, id, followType)
	if err != nil {
		log.Printf("Error: %v get Follower for id:%d, type:%d\n", err.Error(), id, followType)
	}
	for rows.Next() {
		var follower int
		err = rows.Scan(&follower)
		if err != nil {
			log.Printf("Error: %v scan for id:%d, type:%d\n", err.Error(), id, followType)
		}
		result = append(result, follower)
	}
	return result, err
}

func (f *followingAPI) PseudoGetFollowMap(params GetFollowArgs) (list []FollowingMapItem, err error) {

	query, args := params.getmap()

	rows, err := DB.Queryx(query, args...)
	if err != nil {
		log.Println(err.Error())
		return []FollowingMapItem{}, err
	}

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
	return list, nil
}

func (f *followingAPI) GetFollowed(args GetFollowedArgs) (interface{}, error) {
	follow, err := CreateFollow(FollowArgs{Resource: args.Resource})
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	rows, err := follow.GetFollowed(args)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	followed := []followedCount{}
	for rows.Next() {
		var (
			resourceId int64
			count      int
			follower   string
		)
		err = rows.Scan(&resourceId, &count, &follower)
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
		followed = append(followed, followedCount{resourceId, count, followers})
	}
	return followed, nil
}
func (f *followingAPI) GetFollowerMemberIDs(resourceType string, id string) ([]int, error) {
	follow, err := CreateFollow(FollowArgs{Resource: resourceType})
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	ids, err := follow.GetFollowerMemberIDs(id)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return ids, nil
}
func (f *followingAPI) GetFollowMap(args GetFollowMapArgs) (list []FollowingMapItem, err error) {
	follow, err := CreateFollow(FollowArgs{Resource: args.Resource})
	if err != nil {
		log.Println(err.Error())
		return []FollowingMapItem{}, err
	}

	rows, err := follow.GetMap(args)
	if err != nil {
		log.Println(err.Error())
		return []FollowingMapItem{}, err
	}

	for rows.Next() {
		var member_ids, resource_ids string
		if err = rows.Scan(&member_ids, &resource_ids); err != nil {
			log.Println(err)
			return []FollowingMapItem{}, err
		}
		list = append(list, FollowingMapItem{
			strings.Split(member_ids, ","),
			strings.Split(resource_ids, ","),
		})
	}
	return list, nil
}

func (f *followingAPI) AddFollowing(params FollowArgs) error {
	follow, err := CreateFollow(params)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	result, err := follow.Insert()
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

func (*followingAPI) DeleteFollowing(params FollowArgs) error {
	follow, err := CreateFollow(params)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	_, err = follow.Delete()
	if err != nil {
		log.Fatal(err)
	}

	if params.Resource == "post" {
		go PostCache.UpdateFollowing("unfollow", params.Subject, params.Object)
	}

	return err
}

var FollowingAPI FollowingAPIInterface = new(followingAPI)
