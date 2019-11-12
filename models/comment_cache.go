package models

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
)

type CommentCacheInterface interface {
	Obtain() (comments []CommentAuthor, err error)
	Insert(comment CommentAuthor) (err error)
	// Update(comment CommentAuthor) (err error)
	// Revoke(comment_id int) (err error)
	Generate() (err error)
}

type commentCache struct {
	redisIndexKey string
}

func (c commentCache) Obtain() (comments []CommentAuthor, err error) {
	conn := RedisHelper.ReadConn()
	defer conn.Close()

	CommentCacheBytes := [][]byte{}

	res, err := redis.Values(conn.Do("LRANGE", c.redisIndexKey, "0", config.Config.Redis.Cache.LatestCommentCount-1))
	if err != nil {
		log.Printf("Error getting redis key when getting comment cache: %s , %v", c.redisIndexKey, err)
		return comments, err
	}
	if err = redis.ScanSlice(res, &CommentCacheBytes); err != nil {
		log.Printf("Error scan redis key when getting comment cache: %s , %v", c.redisIndexKey, err)
		return comments, err
	}

	for _, v := range CommentCacheBytes {
		var ca CommentAuthor
		if err := json.Unmarshal(v, &ca); err != nil {
			log.Printf("Error scan redis comment notification: %s , %v", v, err)
			continue
		}
		comments = append(comments, ca)
	}

	return comments, nil
}

func (c commentCache) Insert(comment CommentAuthor) (err error) {
	conn := RedisHelper.WriteConn()
	defer conn.Close()

	conn.Send("MULTI")

	msg, err := json.Marshal(comment)
	if err != nil {
		log.Printf("Error marshaling comment to cache string: %v", err)
		return err
	}
	conn.Send("LPUSH", redis.Args{}.Add(c.redisIndexKey).Add(msg)...)
	conn.Send("LTRIM", redis.Args{}.Add(c.redisIndexKey).Add(0).Add(config.Config.Redis.Cache.LatestCommentCount-1)...)

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert comment cache to redis: %v", err)
		return err
	}

	return nil
}

func (c commentCache) Update(comment CommentAuthor) (err error) {
	return nil
}

func (c commentCache) Revoke(comment_id int) (err error) {
	return nil
}

func (c commentCache) Generate() (err error) {
	query := fmt.Sprintf(`
		SELECT c.id FROM comments as c 
		LEFT JOIN comments as pc ON pc.id = c.parent_id
		WHERE c.resource NOT IN (
			SELECT CONCAT('%s/series/',p.slug,'/',m.memo_id) 
			FROM memos AS m 
				LEFT JOIN projects AS p ON p.project_id = m.project_id 
			WHERE p.status != %d 
		)
			AND c.active = %d AND c.status = %d 
			AND (pc.active = %d OR pc.active IS NULL) 
		ORDER BY c.created_at DESC 
		LIMIT 20;`,
		config.Config.DomainName,
		config.Config.Models.ProjectsStatus["done"],
		config.Config.Models.Comment["active"],
		config.Config.Models.CommentStatus["show"],
		config.Config.Models.Comment["active"],
	)

	rows, err := rrsql.DB.Queryx(query)
	if err != nil {
		log.Printf("Fail to query comment indexes when updating latest comments: %v \n", err.Error())
		return err
	}

	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			log.Printf("Fail to scan comment index when updating latest comments: %v \n", err.Error())
			return err
		}
		ids = append([]int{id}, ids...)
	}

	for _, id := range ids {
		commentAuthor, err := CommentAPI.GetComment(id)
		if err != nil {
			log.Printf("Fail to get comment when updating latest comments: %v \n", err.Error())
			return err
		}
		CommentCache.Insert(commentAuthor)
	}
	return nil
}

var CommentCache CommentCacheInterface = commentCache{redisIndexKey: "commentcache_latest"}
