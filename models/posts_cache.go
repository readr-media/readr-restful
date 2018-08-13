package models

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

type postCacheType interface {
	Key() string
	Insert(post fullCachePost)
	SyncFromDataStorage()
}

type postCache struct {
	caches []postCacheType
}

// fullCachePost holds the full fields a post cached should have.
// There should be Author, Comment, Post, and Tag now.
type fullCachePost struct {
	TaggedPostMember
	HeadComments []CommentAuthor `json:"comments" db:"comments"`
}

func assembleCachePost(postIDs []uint32) (posts []fullCachePost, err error) {

	targetPost, err := PostAPI.GetPost(postID)
	if err != nil {
		log.Printf("postCache failed to get post:%d in Insert phase\n", postID)
		return fullCachePost{}, err
	}

	setIntraMax := func(max int) func(*GetCommentArgs) {
		return func(arg *GetCommentArgs) {
			arg.IntraMax = max
		}
	}
	setResource := func(resource string) func(*GetCommentArgs) {
		return func(arg *GetCommentArgs) {
			arg.Resource = append(arg.Resource, resource)
		}
	}
	commentResource := utils.GenerateResourceInfo("post", int(postID), "")
	args, err := NewGetCommentArgs(setIntraMax(2), setResource(commentResource))
	if err != nil {
		log.Printf("AssembleCachePost Error:%s\n", err.Error())
	}
	commentsFromTargetPost, err := CommentAPI.GetComments(args)
	if err != nil {
		log.Printf("AssembleCachePost get comments error:%s\n", err.Error())
	}

	post.TaggedPostMember = targetPost
	post.HeadComments = commentsFromTargetPost

	return post, nil
}

func (p *postCache) Register(cacheType postCacheType) {

	p.caches = append(p.caches, cacheType)
}

func (p *postCache) Insert(postID uint32) {

	post, err := assembleCachePost(postID)
	if err != nil {
		log.Printf("Error Insert postCache:%v\n", err)
		return
	}

	if post.Active.Valid &&
		fmt.Sprint(post.Active.Int) != fmt.Sprint(config.Config.Models.Posts["active"]) {
		return
	}
	if post.PublishStatus.Valid &&
		fmt.Sprint(post.PublishStatus.Int) != fmt.Sprint(config.Config.Models.PostPublishStatus["publish"]) {
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

	for _, cache := range p.caches {
		var postIDs []int
		res, err := redis.Values(conn.Do("smembers", fmt.Sprint(cache.Key(), "ids")))
		if err != nil {
			log.Println("Fail to scan redis set: ", err)
			return
		}

		if err := redis.ScanSlice(res, &postIDs); err != nil {
			log.Printf("Error scan redis keys: %v", err)
			return
		}

		for _, id := range postIDs {
			if id == int(post.ID) {
				cache.SyncFromDataStorage()
				break
			}
		}
	}
}

func (p *postCache) Delete(post_id uint32) {

	conn := RedisHelper.Conn()
	defer conn.Close()

	for _, cache := range p.caches {
		res, err := redis.Values(conn.Do("smembers", fmt.Sprint(cache.Key(), "ids")))
		if err != nil {
			log.Println("Fail to scan redis set: ", err)
			return
		}
		var postIDs []int
		if err := redis.ScanSlice(res, &postIDs); err != nil {
			log.Printf("Error scan redis keys: %v", err)
			return
		}
		for _, id := range postIDs {
			if id == int(post_id) {
				cache.SyncFromDataStorage()
				break
			}
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

	for _, cache := range p.caches {
		res, err := redis.Values(conn.Do("smembers", fmt.Sprint(cache.Key(), "ids")))
		if err != nil {
			log.Println("Fail to scan redis set: ", err)
			return
		}
		var postIDs []int
		if err := redis.ScanSlice(res, &postIDs); err != nil {
			log.Printf("Error scan redis keys: %v", err)
			return
		}

		for _, id := range postIDs {
			for _, pid := range params.IDs {
				if id == pid {
					cache.SyncFromDataStorage()
					return
				}
			}
		}
	}
}

/*
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
*/

func (p *postCache) SyncFromDataStorage() {

	for _, cache := range p.caches {
		cache.SyncFromDataStorage()
	}
}

type latestPostCache struct {
	key string
}

func (c latestPostCache) Key() string {
	return c.key
}

func (c latestPostCache) Insert(post fullCachePost) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	postString, _ := json.Marshal(post)
	postCacheMap, err := c.getHashMap(c.Key())
	if err != nil {
		fmt.Println("Error get post cache map:", c.Key(), err.Error())
		return
	}

	conn.Send("HSET", redis.Args{}.Add(c.Key()).Add("1").Add(postString)...)
	for k, v := range postCacheMap {
		ki, err := strconv.Atoi(k)
		if err != nil {
			fmt.Println("Error parse string to integer:", k)
			continue
		}
		if ki+1 <= 20 {
			conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(ki+1).Add(v)...)
		}
	}

	postCacheIndex, err := c.getHashMap(fmt.Sprintf("%s_index", c.Key()))
	if err != nil {
		fmt.Println("Error get post cache index:", fmt.Sprintf("%s_index", c.Key()), err.Error())
		return
	}

	conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(post.ID).Add(1)...)
	for k, v := range postCacheIndex {
		vi, err := strconv.Atoi(v)
		if err != nil {
			fmt.Println("Error parse string to integer:", k)
			continue
		}
		if vi+1 <= 20 {
			conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(k).Add(vi+1)...)
		}
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

func (c *latestPostCache) getHashMap(key string) (map[string]string, error) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	resMap, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		fmt.Println("Error hgetall redis key:", c.Key(), err.Error())
	}
	return resMap, err
}

func (c latestPostCache) SyncFromDataStorage() {

	conn := RedisHelper.Conn()
	defer conn.Close()

	rows, err := DB.Queryx(fmt.Sprintf(`
		SELECT p.*, m.nickname AS m_name, m.profile_image AS m_image 
		FROM posts AS p 
		LEFT JOIN members AS m ON m.id = p.author 
		WHERE p.active=%d AND p.publish_status=%d 
		ORDER BY updated_at DESC LIMIT 20;`,
		config.Config.Models.Posts["active"],
		config.Config.Models.PostPublishStatus["publish"],
	))
	if err != nil {
		log.Println(err.Error())
		return
	}

	conn.Send("MULTI")

	postIndex := 0
	for rows.Next() {
		var postF struct {
			Post
			AuthorNickname     NullString `db:"m_name"`
			AuthorProfileImage NullString `db:"m_image"`
		}
		err = rows.StructScan(&postF)
		if err != nil {
			log.Printf("Error scan posts from db: %v", err)
			return
		}

		postFString, err := json.Marshal(postF)
		if err != nil {
			fmt.Sprintf("Error marshal postF struct when updating post cache", err.Error())
		}
		conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(postIndex).Add(postFString)...)
		conn.Send("HSET", redis.Args{}.Add(fmt.Sprint(c.key, "_index")).Add(postF.ID).Add(postIndex)...)
		postIndex += 1

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

func (c hottestPostCache) Insert(post fullCachePost) {
	return
}

func (c hottestPostCache) SyncFromDataStorage() {

	PostScores := postScores{}
	PostScoreIndex := postScoreIndex{}

	// Read follow count from Mysql
	query := fmt.Sprintf(`
		SELECT p.post_id, IFNULL(p.comment_amount,0) AS comment_count, IFNULL(f.count,0) as follow_count 
		FROM posts AS p 
		LEFT JOIN (
			SELECT target_id, count(*) as count 
			FROM following 
			WHERE type = %d GROUP BY target_id
		) AS f ON f.target_id = p.post_id 
		WHERE p.active =%d AND p.publish_status=%d
		`, config.Config.Models.FollowingType["post"], config.Config.Models.Posts["active"], config.Config.Models.PostPublishStatus["publish"])
	rows, err := DB.Queryx(query)
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

	query, args, err := sqlx.In(fmt.Sprintf(`
		SELECT p.*, m.nickname AS m_name, m.profile_image AS m_image 
		FROM posts AS p 
		LEFT JOIN members AS m ON m.id = p.author 
			WHERE p.active= %d AND p.publish_status=%d AND p.post_id IN (?);`,
		config.Config.Models.Posts["active"],
		config.Config.Models.PostPublishStatus["publish"]),
		hotPosts,
	)
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

	var postIDs []uint32
	for rows.Next() {
		var postF struct {
			Post
			AuthorNickname     NullString `db:"m_name"`
			AuthorProfileImage NullString `db:"m_image"`
		}
		err = rows.StructScan(&postF)
		if err != nil {
			log.Printf("Error scan posts from db: %v", err)
			return
		}
		postIDs = append(postIDs, postF.Post.ID)
		postI := PostScoreIndex[int(postF.ID)].Index
		conn.Send("HMSET", redis.Args{}.Add(fmt.Sprint(c.key, postI)).Add("author_nickname").Add(postF.AuthorNickname).Add("author_profileImage").Add(postF.AuthorProfileImage).AddFlat(&postF.Post)...)

		postFString, err := json.Marshal(postF)
		if err != nil {
			fmt.Sprintf("Error marshal postF struct when updating post cache", err.Error())
		}

		for postIndex, v := range postIDs {
			if v == postF.ID {
				conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(postIndex).Add(postFString)...)
				conn.Send("HSET", redis.Args{}.Add(fmt.Sprint(c.key, "_index")).Add(postF.ID).Add(postIndex)...)
			}
		}
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

var PostCache postCache = postCache{}

func InitPostCache() {
	PostCache.Register(latestPostCache{"postcache_latest"})
	PostCache.Register(hottestPostCache{"postcache_hot"})
	PostCache.SyncFromDataStorage()
}
