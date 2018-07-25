package models

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

type postCacheType interface {
	Key() string
	Insert(post Post)
	SyncFromDataStorage()
}

type postCache struct {
	caches []postCacheType
}

func (p *postCache) Register(cahceType postCacheType) {

	p.caches = append(p.caches, cahceType)
}

func (p *postCache) Insert(post Post) {

	// if fmt.Sprint(post.Active.Int) != fmt.Sprint(PostStatus["active"]) {
	if fmt.Sprint(post.Active.Int) != fmt.Sprint(config.Config.Models.Posts["active"]) {
		return
	}
	// if fmt.Sprint(post.PublishStatus.Int) != fmt.Sprint(PostPublishStatus["publish"]) {
	if fmt.Sprint(post.PublishStatus.Int) != fmt.Sprint(config.Config.Models.PostPublishStatus["publish"]) {
		return
	}
	for _, cache := range p.caches {
		cache.Insert(post)
	}
}

func (p *postCache) Update(post Post) {
	if post.Active.Valid || post.PublishStatus.Valid {
		for _, cache := range p.caches {
			cache.SyncFromDataStorage()
		}
		return
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	for _, cache := range p.caches {
		keys, err := RedisHelper.GetRedisKeys(fmt.Sprint(cache.Key(), "[^follower]*"))
		if err != nil {
			log.Println(err)
			return
		}
		for _, key := range keys {
			if key == fmt.Sprint(cache.Key(), fmt.Sprint(post.ID)) {
				src := reflect.ValueOf(post)
				for i := 0; i < src.NumField(); i++ {
					field := src.Type().Field(i)
					value := src.Field(i).Interface()
					switch value := value.(type) {
					case string:
						if value != "" {
							conn.Send("HSET", key, field.Tag.Get("redis"), value)
						}
					case NullString:
						if value.Valid {
							conn.Send("HSET", key, field.Tag.Get("redis"), value)
						}
					case NullTime:
						if value.Valid {
							conn.Send("HSET", key, field.Tag.Get("redis"), value)
						}
					case NullInt:
						if value.Valid {
							conn.Send("HSET", key, field.Tag.Get("redis"), value)
						}
					case NullBool:
						if value.Valid {
							conn.Send("HSET", key, field.Tag.Get("redis"), value)
						}
					case bool, int, uint32:
						conn.Send("HSET", key, field.Tag.Get("redis"), value)
					default:
						fmt.Println("unrecognised format: ", src.Field(i).Type())
					}
				}
				break
			}
		}
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache change to redis: %v", err)
		return
	}
}

func (p *postCache) Delete(post_id uint32) {
	for _, cache := range p.caches {
		keys, err := RedisHelper.GetRedisKeys(fmt.Sprint(cache.Key(), fmt.Sprint(post_id)))
		if err != nil {
			log.Println(err)
			return
		}
		if len(keys) > 0 {
			cache.SyncFromDataStorage()
		}
	}
}

func (p *postCache) UpdateAll(params PostUpdateArgs) {
	if params.Active.Valid || params.PublishStatus.Valid {
		p.SyncFromDataStorage()
		return
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	for _, cache := range p.caches {
		keys, err := RedisHelper.GetRedisKeys(fmt.Sprint(cache.Key(), "[^follower]*"))
		if err != nil {
			log.Println(err)
			return
		}
		for _, key := range keys {
			for _, id := range params.IDs {
				if key == fmt.Sprint(cache.Key, fmt.Sprint(id)) {
					if params.UpdatedBy != "" {
						conn.Send("HSET", key, "updated_by", params.UpdatedBy)
					}
					if params.UpdatedAt.Valid == true {
						conn.Send("HSET", key, "updated_at", params.UpdatedAt)
					}
					if params.PublishedAt.Valid == true {
						conn.Send("HSET", key, "published_at", params.PublishedAt)
					}
					break
				}
			}
		}
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache change to redis: %v", err)
		return
	}
}

func (p *postCache) UpdateFollowing(action string, user_id int64, post_id int64) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	for _, cache := range p.caches {
		if action == "follow" {
			if _, err := conn.Do("SADD", fmt.Sprint(cache.Key(), "follower_", fmt.Sprint(post_id)), user_id); err != nil {
				log.Printf("Error update follower of cached post: %v", err)
				return
			}
		} else if action == "unfollow" {
			if _, err := conn.Do("SREM", fmt.Sprint(cache.Key(), "follower_", fmt.Sprint(post_id)), user_id); err != nil {
				log.Printf("Error update follower of cached post: %v", err)
				return
			}
		}
	}
}

func (p *postCache) SyncFromDataStorage() {

	for _, cache := range p.caches {
		/* Do not delete old caches
		conn := RedisHelper.Conn()
		defer conn.Close()

		keys, err := RedisHelper.GetRedisKeys(fmt.Sprint(cache.Key(), "*"))
		if err != nil {
			log.Println(err)
			return
		}

		if len(keys) > 0 {
			if _, err := conn.Do("DEL", redis.Args{}.AddFlat(keys)...); err != nil {
				log.Printf("Error delete cache from redis: %v", err)
				return
			}
		}
		*/
		cache.SyncFromDataStorage()
	}
}

type latestPostCache struct {
	key string
}

func (c latestPostCache) Key() string {
	return c.key
}
func (c latestPostCache) Insert(post Post) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	keys, err := RedisHelper.GetRedisKeys(fmt.Sprint(c.key, "[^follower]*"))
	if err != nil {
		log.Println(err)
		return
	}

	if len(keys) >= 20 {
		conn.Send("MULTI")
		for _, key := range keys {
			conn.Send("MGET", key, "updated_at")
		}
		res, err := redis.Values(conn.Do("EXEC"))
		if err != nil {
			log.Printf("Error insert cache to redis: %v", err)
			return
		}
		var updateAts []time.Time
		if err := redis.ScanSlice(res, &updateAts); err != nil {
			log.Printf("Error scan keys cache: %v", err)
			return
		}
		var (
			oldestTime  = updateAts[0]
			oldestIndex = 0
		)
		for index, time := range updateAts {
			if time.Before(oldestTime) {
				oldestTime = time
				oldestIndex = index
			}
		}
		if _, err := conn.Do("DEL", keys[oldestIndex]); err != nil {
			log.Printf("Error delete cache from redis: %v", err)
			return
		}
	}

	if _, err := conn.Do("HMSET", redis.Args{}.Add(fmt.Sprint(c.key, post.ID)).AddFlat(&post)...); err != nil {
		log.Printf("Error delete cache from redis: %v", err)
		return
	}

}

func (c latestPostCache) SyncFromDataStorage() {

	conn := RedisHelper.Conn()
	defer conn.Close()

	//rows, err := DB.Queryx("SELECT p.*, f.follower FROM posts as p LEFT JOIN (SELECT post_id, GROUP_CONCAT(id) as follower FROM following_posts GROUP BY post_id) as f ON p.post_id = f.post_id WHERE active=" + fmt.Sprint(PostStatus["active"]) + " ORDER BY updated_at DESC LIMIT 20;")
	// rows, err := DB.Queryx(fmt.Sprint("SELECT p.*, f.follower, m.nickname AS m_name, m.profile_image AS m_image FROM posts as p LEFT JOIN (SELECT post_id, GROUP_CONCAT(id SEPARATOR ',') as follower FROM following_posts GROUP BY post_id) as f ON p.post_id = f.post_id LEFT JOIN members AS m ON m.id = p.author WHERE p.active=", fmt.Sprint(PostStatus["active"]), " AND p.publish_status=", fmt.Sprint(PostPublishStatus["publish"]), " ORDER BY updated_at DESC LIMIT 20;"))
	rows, err := DB.Queryx(fmt.Sprintf(`SELECT p.*, f.follower, m.nickname AS m_name, m.profile_image AS m_image FROM posts AS p 
		LEFT JOIN (SELECT target_id, GROUP_CONCAT(id SEPARATOR ',') AS follower FROM following WHERE type = %d GROUP BY target_id) AS f ON p.post_id = f.target_id 
		LEFT JOIN members AS m ON m.id = p.author WHERE p.active=%d AND p.publish_status=%d ORDER BY updated_at DESC LIMIT 20;`, config.Config.Models.FollowingType["post"], config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"]))
	if err != nil {
		log.Println(err.Error())
		return
	}

	conn.Send("MULTI")
	rows_index := 0
	for rows.Next() {
		var postF struct {
			Post
			PostFollower       NullString `db:"follower"`
			AuthorNickname     NullString `db:"m_name"`
			AuthorProfileImage NullString `db:"m_image"`
		}
		err = rows.StructScan(&postF)
		if err != nil {
			log.Printf("Error scan posts from db: %v", err)
			return
		}
		rows_index += 1
		conn.Send("HMSET", redis.Args{}.Add(fmt.Sprint(c.key, rows_index)).Add("author_nickname").Add(postF.AuthorNickname).Add("author_profileImage").Add(postF.AuthorProfileImage).AddFlat(&postF.Post)...)
		conn.Send("SADD", redis.Args{}.Add(fmt.Sprint(c.key, "follower_", rows_index)).AddFlat(postF.PostFollower.String)...)

	}
	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

type hottestPostCache struct {
	key string
}

type commentCount struct {
	Url   string `bson:"id,omitempty"`
	Count int    `bson:"count,omitempty"`
}
type followCount struct {
	ID           int `db:"post_id"`
	FollowCount  int `db:"follow_count"`
	CommentCount int `db:"comment_count"`
}

type postScore struct {
	ID    int
	Index int
	Score float64
}

type postScores []postScore
type postScoreIndex map[int]postScore

func (p postScores) Len() int           { return len(p) }
func (p postScores) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p postScores) Less(i, j int) bool { return p[i].Score > p[j].Score }

func (c hottestPostCache) Key() string {
	return c.key
}

func (c hottestPostCache) Insert(post Post) {
	return
}

func (c hottestPostCache) SyncFromDataStorage() {

	PostScores := postScores{}
	PostScoreIndex := postScoreIndex{}

	// Read follow count from Mysql
	// rows, err := DB.Queryx(fmt.Sprint("SELECT p.post_id, p.comment_amount AS comment_count, IFNULL(f.count,0) as follow_count FROM posts AS p LEFT JOIN (SELECT post_id, count(*) as count FROM post_tags GROUP BY post_id) AS f ON f.post_id = p.post_id WHERE p.active =", fmt.Sprint(PostStatus["active"], " AND p.publish_status=", fmt.Sprint(PostPublishStatus["publish"]))))
	rows, err := DB.Queryx(fmt.Sprint("SELECT p.post_id, IFNULL(p.comment_amount,0) AS comment_count, IFNULL(f.count,0) as follow_count FROM posts AS p LEFT JOIN (SELECT post_id, count(*) as count FROM post_tags GROUP BY post_id) AS f ON f.post_id = p.post_id WHERE p.active =", fmt.Sprint(config.Config.Models.Posts["active"], " AND p.publish_status=", fmt.Sprint(config.Config.Models.PostPublishStatus["publish"]))))
	if err != nil {
		log.Println(err.Error())
		return
	}

	for rows.Next() {
		var count followCount
		err = rows.StructScan(&count)
		if err != nil {
			log.Printf("Error scan follow count from db: %v", err)
			return
		}
		PostScores = append(PostScores, postScore{ID: count.ID, Score: 0.6*float64(count.FollowCount) + 0.4*float64(count.CommentCount)})
	}

	// Sort post Score
	sort.Sort(PostScores)
	limit := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}(len(PostScores), 20)

	var hotPosts []int
	for i, post := range PostScores[:limit] {
		hotPosts = append(hotPosts, post.ID)
		post.Index = i + 1
		PostScoreIndex[post.ID] = post
	}

	// Fetching post and follower data for hottest posts
	if len(hotPosts) <= 0 {
		return
	}
	// query, args, err := sqlx.In(fmt.Sprint("SELECT p.*, f.follower, m.nickname AS m_name, m.profile_image AS m_image FROM posts as p LEFT JOIN (SELECT post_id, GROUP_CONCAT(id SEPARATOR ',') as follower FROM following_posts GROUP BY post_id) as f ON p.post_id = f.post_id LEFT JOIN members AS m ON m.id = p.author WHERE p.active=", fmt.Sprint(PostStatus["active"]), " AND p.publish_status=", fmt.Sprint(PostPublishStatus["publish"]), " AND p.post_id IN (?);"), hotPosts)
	query, args, err := sqlx.In(fmt.Sprintf(`SELECT p.*, f.follower, m.nickname AS m_name, m.profile_image AS m_image FROM posts AS p 
		LEFT JOIN (SELECT target_id, GROUP_CONCAT(id SEPARATOR ',') as follower FROM following WHERE type = %d GROUP BY target_id) as f ON p.post_id = f.target_id 
		LEFT JOIN members AS m ON m.id = p.author WHERE p.active= %d AND p.publish_status=%d AND p.post_id IN (?);`, config.Config.Models.FollowingType["post"], config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"]), hotPosts)
	if err != nil {
		log.Println("error to build `in` query when fetching post cache data ", err)
		return
	}
	query = DB.Rebind(query)
	rows, err = DB.Queryx(query, args...)
	if err != nil {
		log.Println("error to query post when fetching post cache data ", err)
		return
	}

	// Write post data, post followers, post score to Redis
	conn := RedisHelper.Conn()
	defer conn.Close()
	conn.Send("MULTI")
	for rows.Next() {
		var postF struct {
			Post
			PostFollower       NullString `db:"follower"`
			AuthorNickname     NullString `db:"m_name"`
			AuthorProfileImage NullString `db:"m_image"`
		}
		err = rows.StructScan(&postF)
		if err != nil {
			log.Printf("Error scan posts from db: %v", err)
			return
		}
		postI := PostScoreIndex[int(postF.ID)].Index
		conn.Send("HMSET", redis.Args{}.Add(fmt.Sprint(c.key, postI)).Add("author_nickname").Add(postF.AuthorNickname).Add("author_profileImage").Add(postF.AuthorProfileImage).AddFlat(&postF.Post)...)
		conn.Send("SADD", redis.Args{}.Add(fmt.Sprint(c.key, "follower_", postI)).AddFlat(postF.PostFollower.String)...)
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

var PostCache postCache = postCache{}

func InitPostCache() {
	PostCache.Register(latestPostCache{"postcache_fp_"})
	PostCache.Register(hottestPostCache{"postcache_hot_"})
	PostCache.SyncFromDataStorage()
}
