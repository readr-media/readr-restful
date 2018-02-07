package models

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

type postFollower struct {
	Follower NullString `db:"follower"`
}

type latestPostCache struct {
	Key string
}

func (c *latestPostCache) Insert(post Post) {
	if fmt.Sprint(post.Active.Int) != fmt.Sprint(PostStatus["active"]) {
		return
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	keys, err := RedisHelper.GetRedisKeys(c.Key + "[^follower]*")
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

	if _, err := conn.Do("HMSET", redis.Args{}.Add(c.Key+fmt.Sprint(post.ID)).AddFlat(&post)...); err != nil {
		log.Printf("Error delete cache from redis: %v", err)
		return
	}

}

func (c *latestPostCache) Update(post Post) {

	if post.Active.Valid {
		c.SyncFromDataStorage()
		return
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	keys, err := RedisHelper.GetRedisKeys(c.Key + "*")
	if err != nil {
		log.Println(err)
		return
	}

	conn.Send("MULTI")
	for _, key := range keys {
		if key == c.Key+fmt.Sprint(post.ID) {
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

	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache change to redis: %v", err)
		return
	}
}

func (c *latestPostCache) Delete(post_id uint32) {

	keys, err := RedisHelper.GetRedisKeys(c.Key + fmt.Sprint(post_id))
	if err != nil {
		log.Println(err)
		return
	}

	if len(keys) > 0 {
		c.SyncFromDataStorage()
	}

}

func (c *latestPostCache) UpdateMulti(params PostUpdateArgs) {

	if params.Active.Valid {
		c.SyncFromDataStorage()
		return
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	keys, err := RedisHelper.GetRedisKeys(c.Key + "*")
	if err != nil {
		log.Println(err)
		return
	}

	conn.Send("MULTI")
	for _, key := range keys {
		for _, id := range params.IDs {
			if key == c.Key+fmt.Sprint(id) {
				if params.UpdatedBy != "" {
					conn.Send("HSET", key, "updated_by", params.UpdatedBy)
				}
				if params.UpdatedAt.Valid == true {
					conn.Send("HSET", key, "updated_at", params.UpdatedAt)
				}
				if params.Active.Valid == true {
					conn.Send("HSET", key, "active", params.Active)
				}
			}
		}
	}
	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache change to redis: %v", err)
		return
	}
}

func (c *latestPostCache) UpdateFollowing(action string, user_id string, post_id string) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	if action == "follow" {
		if _, err := conn.Do("SADD", c.Key+"follower_"+fmt.Sprint(post_id), user_id); err != nil {
			log.Printf("Error update follower of cached post: %v", err)
			return
		}
	} else if action == "unfollow" {
		if _, err := conn.Do("SREM", c.Key+"follower_"+fmt.Sprint(post_id), user_id); err != nil {
			log.Printf("Error update follower of cached post: %v", err)
			return
		}
	}
}

func (c *latestPostCache) SyncFromDataStorage() {
	conn := RedisHelper.Conn()
	defer conn.Close()

	keys, err := RedisHelper.GetRedisKeys(c.Key + "*")
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

	rows, err := DB.Queryx("SELECT p.*, f.follower FROM posts as p LEFT JOIN (SELECT post_id, GROUP_CONCAT(member_id) as follower FROM following_posts GROUP BY post_id) as f ON p.post_id = f.post_id WHERE active=" + fmt.Sprint(PostStatus["active"]) + " ORDER BY updated_at DESC LIMIT 20;")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	conn.Send("MULTI")
	for rows.Next() {
		var postF struct {
			Post
			postFollower
		}
		err = rows.StructScan(&postF)
		if err != nil {
			log.Printf("Error scan posts from db: %v", err)
			return
		}

		conn.Send("HMSET", redis.Args{}.Add(c.Key+fmt.Sprint(postF.ID)).AddFlat(&postF)...)
		conn.Send("SADD", redis.Args{}.Add(c.Key+"follower_"+fmt.Sprint(postF.ID)).AddFlat(postF.Follower)...)

	}
	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

var PostCache latestPostCache = latestPostCache{"postcache_fp_"}
