package models

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisHelper struct {
	pool *redis.Pool
}

var RedisHelper = redisHelper{}

func (r *redisHelper) Conn() redis.Conn {
	return r.pool.Get()
}

func (r *redisHelper) GetRedisKeys(key string) ([]string, error) {
	conn := r.Conn()
	defer conn.Close()

	var keys []string

	res, err := redis.Values(conn.Do("KEYS", key))
	if err != nil {
		log.Printf("Error get redis keys: %v", err)
		return keys, err
	}

	if err := redis.ScanSlice(res, &keys); err != nil {
		log.Printf("Error scan redis keys: %v", err)
		return keys, err
	}

	return keys, nil
}

func RedisConn(config map[string]string) {
	RedisHelper = redisHelper{&redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", config["url"], redis.DialPassword(config["password"]))
		},
	}}
}
