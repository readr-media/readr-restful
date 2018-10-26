package models

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
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
	*TaggedPostMember
	HeadComments []CommentAuthor `json:"comments" db:"comments"`
}

func assembleCachePost(postIDs []uint32) (posts []fullCachePost, err error) {

	postArgs := &PostArgs{}
	postArgs = postArgs.Default()
	postArgs.IDs = postIDs

	targetPosts, err := PostAPI.GetPosts(postArgs)
	if err != nil {
		log.Printf("postCache failed to get posts:%v in Insert phase, %v\n", postIDs, err)
		return []fullCachePost{}, err
	}

	setIntraMax := func(max int) func(*GetCommentArgs) {
		return func(arg *GetCommentArgs) {
			arg.IntraMax = max
		}
	}
	setResource := func(postIDs []uint32) func(*GetCommentArgs) {
		return func(arg *GetCommentArgs) {
			for _, postID := range postIDs {
				commentResource := utils.GenerateResourceInfo("post", int(postID), "")
				arg.Resource = append(arg.Resource, commentResource)
			}
		}
	}
	args, err := NewGetCommentArgs(setIntraMax(2), setResource(postArgs.IDs))
	if err != nil {
		log.Printf("AssembleCachePost Error:%s\n", err.Error())
	}
	commentsFromTargetPosts, err := CommentAPI.GetComments(args)
	if err != nil {
		log.Printf("AssembleCachePost get comments error:%s\n", err.Error())
	}

	for _, targetPost := range targetPosts {
		tpm := targetPost
		post := fullCachePost{TaggedPostMember: &tpm}
		for _, comment := range commentsFromTargetPosts {
			commentResourceID, _ := strconv.ParseUint(strings.TrimPrefix(comment.Resource.String, fmt.Sprintf("%s/post/", config.Config.DomainName)), 10, 32)
			if post.ID == uint32(commentResourceID) {
				post.HeadComments = append(post.HeadComments, comment)
			}
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (p *postCache) Register(cacheType postCacheType) {

	p.caches = append(p.caches, cacheType)
}

func (p *postCache) Insert(postID uint32) {

	posts, err := assembleCachePost([]uint32{postID})
	if err != nil {
		log.Printf("Error Insert postCache:%v\n", err)
		return
	}

	if len(posts) == 0 {
		log.Printf("Error Assemble postCache:%v\n", err)
		return
	}
	for _, cache := range p.caches {
		cache.Insert(posts[0])
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

	postString, _ := json.Marshal(post)

	postCacheMap, err := redis.StringMap(conn.Do("HGETALL", c.Key()))
	if err != nil {
		fmt.Println("Error get post cache map:", c.Key(), err.Error())
	}

	postCacheIndex, err := redis.StringMap(conn.Do("HGETALL", fmt.Sprintf("%s_index", c.Key())))
	if err != nil {
		fmt.Println("Error get post cache index:", fmt.Sprintf("%s_index", c.Key()), err.Error())
	}

	conn.Send("MULTI")
	conn.Send("DEL", redis.Args{}.Add(fmt.Sprint(c.key, "_index")))

	conn.Send("HSET", redis.Args{}.Add(c.Key()).Add("1").Add(postString)...)
	for k, v := range postCacheMap {
		ki, err := strconv.Atoi(k)
		if err != nil {
			fmt.Println("Error parse string to integer:", k, err)
			continue
		}
		if ki+1 <= 20 {
			conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(ki+1).Add(v)...)
		}
	}

	conn.Send("HSET", redis.Args{}.Add(fmt.Sprintf("%s_index", c.Key())).Add(post.ID).Add(1)...)
	for k, v := range postCacheIndex {
		vi, err := strconv.Atoi(v)
		if err != nil {
			fmt.Println("Error parse string to integer:", v, err)
			continue
		}
		if vi+1 <= 20 {
			conn.Send("HSET", redis.Args{}.Add(fmt.Sprintf("%s_index", c.Key())).Add(k).Add(vi+1)...)
		}
	}

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

func (c latestPostCache) SyncFromDataStorage() {

	conn := RedisHelper.Conn()
	defer conn.Close()

	var postIDs []uint32
	err := DB.Select(&postIDs, fmt.Sprintf(`
		SELECT post_id FROM posts WHERE active=%d AND publish_status=%d ORDER BY published_at DESC LIMIT 20;`,
		config.Config.Models.Posts["active"],
		config.Config.Models.PostPublishStatus["publish"],
	))
	if err != nil {
		log.Println(err.Error())
		return
	}

	fullCachePosts, err := assembleCachePost(postIDs)
	if err != nil {
		fmt.Println("Error getting cache post when updating hot posts. PostIDs:", postIDs)
		return
	}

	conn.Send("MULTI")
	conn.Send("DEL", redis.Args{}.Add(fmt.Sprint(c.key, "_index")))
	for _, cachePost := range fullCachePosts {
		for postIndex, postID := range postIDs {
			if postID == cachePost.ID {
				postString, err := json.Marshal(&cachePost)
				if err != nil {
					fmt.Sprintf("Error marshal fullCachePost struct when updating latest post cache", err.Error())
					continue
				}

				conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(postIndex+1).Add(postString)...)
				conn.Send("HSET", redis.Args{}.Add(fmt.Sprint(c.key, "_index")).Add(cachePost.ID).Add(postIndex+1)...)
			}
		}
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

	var hotPosts []uint32
	for i, post := range PostScores[:limit] {
		hotPosts = append(hotPosts, uint32(post.ID))
		post.Index = i + 1
		PostScoreIndex[post.ID] = post
	}

	// Fetching post and follower data for hottest posts
	if len(hotPosts) <= 0 {
		return
	}

	fullCachePosts, err := assembleCachePost(hotPosts)
	if err != nil {
		fmt.Println("Error getting cache post when updating hot posts. PostIDs:", hotPosts)
		return
	}

	// Write post data, post followers, post score to Redis
	conn := RedisHelper.Conn()
	defer conn.Close()
	conn.Send("MULTI")

	for _, cachePost := range fullCachePosts {

		postIndex := PostScoreIndex[int(cachePost.ID)].Index
		postString, err := json.Marshal(&cachePost)
		if err != nil {
			fmt.Sprintf("Error marshal fullCachePost struct when updating hot post cache", err.Error())
			continue
		}

		conn.Send("HSET", redis.Args{}.Add(c.Key()).Add(postIndex).Add(postString)...)
		conn.Send("HSET", redis.Args{}.Add(fmt.Sprint(c.key, "_index")).Add(cachePost.ID).Add(postIndex)...)
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
