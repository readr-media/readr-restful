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
	return DB.Queryx(query, args...)
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
	GetFollowMap(args GetFollowMapArgs) ([]FollowingMapItem, error)
}

type FollowArgs struct {
	Resource string
	Subject  int64
	Object   int64
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
