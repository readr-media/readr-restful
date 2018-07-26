package models

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/readr-media/readr-restful/config"
)

type FollowCacheInterface interface {
	Update(input GetFollowedArgs, followeds []FollowedCount)
	Revoke(actionType string, resource string, emotion int, object int64)
}

type followCache struct {
	redisIndexKey string
	redisFieldKey string
}

func (f followCache) Update(input GetFollowedArgs, followeds []FollowedCount) {
	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	for _, followed := range followeds {

		s, _ := json.Marshal(followed)
		conn.Send("HSET", redis.Args{}.
			Add(fmt.Sprintf(f.redisIndexKey, input.Emotion)).
			Add(fmt.Sprintf(f.redisFieldKey, input.ResourceName, followed.ResourceID)).
			Add(s)...)
	}

	_, err := conn.Do("EXEC")
	if err != nil {
		log.Printf("Error update cache to redis: %v", err)
		return
	}
}

func (f followCache) Revoke(actionType string, resource string, emotion int, object int64) {
	var emotions []int
	if actionType != "update" {
		emotions = []int{emotion}
	} else {
		emotions = []int{
			config.Config.Models.Emotions["like"],
			config.Config.Models.Emotions["dislike"],
		}
	}

	conn := RedisHelper.Conn()
	defer conn.Close()
	conn.Send("MULTI")

	for _, emotion := range emotions {
		conn.Send("HDEL", fmt.Sprintf(f.redisIndexKey, emotion), fmt.Sprintf(f.redisFieldKey, resource, object))
	}

	_, err := conn.Do("EXEC")
	if err != nil {
		log.Printf("Error revoke cache from redis: %v", err)
		return
	}
}

var FollowCache FollowCacheInterface = followCache{redisIndexKey: "followcache_%d", redisFieldKey: "%s_%d"}
