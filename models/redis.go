package models

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var redisPool *redis.Pool

func RedisConn(config map[string]string) {
	redisPool = &redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", config["url"], redis.DialPassword(config["password"]))
		},
	}
}

func GetRedisConn() redis.Conn {
	return redisPool.Get()
}
