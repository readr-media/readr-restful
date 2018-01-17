package models

import (
	"errors"
	"log"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

/* Types of Followings */

type follow interface {
	Insert() (sql.Result, error)
	Delete() (sql.Result, error)
	GetFollowings(map[string]string) (interface{}, error)
	GetFollowed([]string) (*sqlx.Rows, error)
}

type followedCount struct {
	Resourceid string
	Count      int
}

type followPost struct {
	ID     string
	Object string
}

func (f followPost) Insert() (sql.Result, error) {
	query := "INSERT INTO following_posts (member_id, post_id) VALUES ( ? , ? );"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followPost) Delete() (sql.Result, error) {
	query := "DELETE FROM following_posts WHERE member_id = ? AND post_id = ?;"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followPost) GetFollowings(params map[string]string) (interface{}, error) {
	rows, err := DB.Queryx("SELECT m.* from posts as m INNER JOIN following_posts as f ON m.post_id = f.post_id WHERE m.active = 1 AND f.member_id = ? ORDER BY f.created_at DESC;", f.ID)
	if err != nil {
		return nil, err
	}

	var followings []Post
	for rows.Next() {
		var post Post
		err = rows.StructScan(&post)
		if err != nil {
			log.Fatalln(err.Error())
		}
		followings = append(followings, post)
	}

	if len(followings) == 0 {
		err = errors.New("Not Found")
	}

	return followings, err
}

func (f followPost) GetFollowed(ids []string) (*sqlx.Rows, error) {
	//rows, err := DB.Queryx("SELECT m.*, IFNULL(f.count, 0) FROM posts as m LEFT JOIN (SELECT post_id, COUNT(post_id) as count from following_posts GROUP BY post_id) as f ON f.post_id = m.post_id WHERE m.post_id IN ? ;", ids)
	query, args, err := sqlx.In("SELECT post_id, COUNT(post_id) as count FROM following_posts WHERE post_id in ( ? ) GROUP BY post_id;", ids)
	if err != nil {
		return nil, err
	}

	rows, err := DB.Queryx(query, args...)

	return rows, err
}

type followMember struct {
	ID     string
	Object string
}

func (f followMember) Insert() (sql.Result, error) {
	query := "INSERT INTO following_members (member_id, custom_editor) VALUES ( ? , ? );"
	return DB.Exec(query, f.ID, f.Object)
}

func (f followMember) Delete() (sql.Result, error) {
	query := "DELETE FROM following_members WHERE member_id = ? AND custom_editor = ?;"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followMember) GetFollowings(params map[string]string) (interface{}, error) {
	rows, err := DB.Queryx("SELECT m.* from members as m INNER JOIN following_members as f ON m.member_id = f.custom_editor WHERE m.active > 0 AND f.member_id = ? ORDER BY f.created_at DESC;", f.ID)
	if err != nil {
		return nil, err
	}

	var followings []Member
	for rows.Next() {
		var member Member
		err = rows.StructScan(&member)
		if err != nil {
			log.Fatalln(err.Error())
		}
		followings = append(followings, member)
	}

	if len(followings) == 0 {
		err = errors.New("Not Found")
	}

	return followings, err
}
func (f followMember) GetFollowed(ids []string) (*sqlx.Rows, error) {
	query, args, err := sqlx.In("SELECT member_id, COUNT(member_id) as count FROM following_members WHERE member_id in ( ? ) GROUP BY member_id;", ids)
	if err != nil {
		return nil, err
	}

	rows, err := DB.Queryx(query, args...)

	return rows, err
}

type followProject struct {
	ID     string
	Object string
}

func (f followProject) Insert() (sql.Result, error) {
	query := "INSERT INTO following_projects (member_id, project_id) VALUES ( ? , ? );"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followProject) Delete() (sql.Result, error) {
	query := "DELETE FROM following_projects WHERE member_id = ? AND project_id = ?;"
	return DB.Exec(query, f.ID, f.Object)
}
func (f followProject) GetFollowings(params map[string]string) (interface{}, error) {
	rows, err := DB.Queryx("SELECT m.* from projects as m INNER JOIN following_projects as f ON m.project_id = f.project_id WHERE m.active = 1  AND f.member_id = ? ORDER BY f.created_at DESC;", f.ID)
	if err != nil {
		return nil, err
	}

	var followings []Project
	for rows.Next() {
		var project Project
		err = rows.StructScan(&project)
		if err != nil {
			log.Fatalln(err.Error())
		}
		followings = append(followings, project)
	}

	if len(followings) == 0 {
		err = errors.New("Not Found")
	}

	return followings, err
}
func (f followProject) GetFollowed(ids []string) (*sqlx.Rows, error) {
	query, args, err := sqlx.In("SELECT project_id, COUNT(project_id) as count FROM following_projects WHERE project_id in ( ? ) GROUP BY project_id;", ids)
	if err != nil {
		return nil, err
	}

	rows, err := DB.Queryx(query, args...)

	return rows, err
}

/*
Follower Factory
*/

func followPostFactory(config map[string]string) follow {
	return followPost{
		ID:     config["subject"],
		Object: config["object"],
	}
}
func followMemberFactory(config map[string]string) follow {
	return followMember{
		ID:     config["subject"],
		Object: config["object"],
	}
}

func followProjectFactory(config map[string]string) follow {
	return followProject{
		ID:     config["subject"],
		Object: config["object"],
	}
}

var followFactories = map[string]func(conf map[string]string) follow{
	"post":    followPostFactory,
	"member":  followMemberFactory,
	"project": followProjectFactory,
}

func CreateFollow(config map[string]string) (follow, error) {
	followFactory, ok := followFactories[config["resource"]]
	if !ok {
		return nil, errors.New("Resource Not Found")
	}
	return followFactory(config), nil
}

/*
Follower API
*/

var (
	DuplicateError      = errors.New("Duplicate Entry")
	SQLInsertionFail    = errors.New("SQL Insertion Fail")
	InternalServerError = errors.New("Internal Server Error")
)

type FollowingAPIInterface interface {
	GetFollowing(params map[string]string) (interface{}, error)
	GetFollowed(resource string, ids []string) (interface{}, error)
	AddFollowing(params map[string]string) error
	DeleteFollowing(params map[string]string) error
}

type followingAPI struct{}

func (*followingAPI) GetFollowing(params map[string]string) (interface{}, error) {
	follow, err := CreateFollow(params)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	result, err := follow.GetFollowings(params)

	return result, err
}
func (*followingAPI) GetFollowed(resource string, ids []string) (interface{}, error) {
	follow, err := CreateFollow(map[string]string{"resource": resource})
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	rows, err := follow.GetFollowed(ids)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	followed := []followedCount{}
	for rows.Next() {
		var (
			resourceId string
			count      int
		)
		err = rows.Scan(&resourceId, &count)
		if err != nil {
			log.Fatalln(err.Error())
			return nil, err
		}
		followed = append(followed, followedCount{resourceId, count})
	}

	return followed, nil
}
func (*followingAPI) AddFollowing(params map[string]string) error {
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

	return nil
}
func (*followingAPI) DeleteFollowing(params map[string]string) error {
	follow, err := CreateFollow(params)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	_, err = follow.Delete()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

var FollowingAPI FollowingAPIInterface = new(followingAPI)
