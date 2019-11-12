package models

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type redisHelper struct {
	readPool  *redis.Pool
	writePool *redis.Pool
}

var RedisHelper = redisHelper{}

func (r *redisHelper) ReadConn() redis.Conn {
	return r.readPool.Get()
}

func (r *redisHelper) WriteConn() redis.Conn {
	return r.writePool.Get()
}

func (r *redisHelper) GetRedisKeys(key string) ([]string, error) {
	conn := r.ReadConn()
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

func (r *redisHelper) GetRedisListLength(key string) (int, error) {
	conn := r.ReadConn()
	defer conn.Close()

	l, err := redis.Int(conn.Do("LLEN", key))
	if err != nil {
		log.Printf("Error getting length of redis list: %s, %v", key, err)
		return 0, err
	}
	return l, err
}

func (r *redisHelper) getOrderedHashes(keysTemplate string, quantity int) (result [][]interface{}, err error) {
	conn := r.ReadConn()
	defer conn.Close()
	conn.Send("MULTI")

	for i := 1; i <= quantity; i++ {
		key := fmt.Sprintf(keysTemplate, i)
		conn.Send("HGETALL", key)
	}

	res, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		log.Printf("Error getting redis key: %v", err)
		return result, err
	}

	for _, v := range res {
		r, err := redis.Values(v, nil)
		if err != nil {
			log.Printf("Error parsing redis result: %v", err)
			continue
		}
		result = append(result, r)
	}

	return result, err
}

/*
func (r *redisHelper) GetHotPosts(keysTemplate string, quantity int) (result []HotPost, err error) {
	res, err := r.getOrderedHashes(keysTemplate, quantity)
	if err != nil {
		log.Printf("Error getting redis key: %v", err)
		return result, err
	}

	for _, v := range res {
		var p HotPost
		if err = redis.ScanStruct(v, &p); err != nil {
			log.Printf("Error scan redis key: %v", err)
			return result, err
		}
		if p.ID == 0 {
			break
		}
		result = append(result, p)
	}

	return result, err
}
*/

func (r *redisHelper) GetHotTags(keysTemplate string, quantity int) (result []TagRelatedResources, err error) {
	conn := r.ReadConn()
	defer conn.Close()

	resMap, err := redis.StringMap(conn.Do("HGETALL", "tagcache_hot"))
	if err != nil {
		fmt.Println("Error hgetall redis key:", "tagcache_hot", err.Error())
	} else {
		result = make([]TagRelatedResources, 20)
		for ranks, tagDetails := range resMap {
			var t TagRelatedResources
			if err = json.Unmarshal([]byte(tagDetails), &t); err != nil {
				log.Printf("Error scan string from redis: %v", err)
				return result, err
			}
			ranki, err := strconv.Atoi(ranks)
			if err != nil || ranki-1 > 20 {
				continue
			}

			result[ranki-1] = t
		}
	}
	return result, err
}

func (r *redisHelper) Subscribe(ctx context.Context, cancel func(), onMessage func(channel string, data []byte) error, channel string) error {

	conn := r.ReadConn()
	psc := redis.PubSubConn{Conn: conn}
	if err := psc.PSubscribe(redis.Args{}.AddFlat(channel)...); err != nil {
		return err
	}

	go func() {
		for {
			msg := psc.Receive()
			switch n := msg.(type) {
			case error:
				log.Println("subscribe error", n.Error())
				cancel()
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					log.Println("on subscribe message handler error", err.Error())
					cancel()
					return
				}
			case redis.PMessage:
				if err := onMessage(n.Channel, n.Data); err != nil {
					log.Println("on subscribe pmessage handler error", err.Error())
					cancel()
					return
				}
			case redis.Subscription:
				log.Println("subscribed: ", channel, n.Count)
				switch n.Count {
				case 0:
					log.Println("no subscribed channel")
					cancel()
					return
				}
			default:
				log.Println(n)
			}
			if ctx.Err() == context.Canceled {
				log.Println("terminated")
				psc.PUnsubscribe(channel)
				conn.Close()
				return
			}
		}
	}()
	return nil
}

func RedisConn(config map[string]string) {
	RedisHelper = redisHelper{
		readPool: &redis.Pool{
			MaxIdle:     30,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", config["read_url"], redis.DialPassword(config["password"]))
			},
		},
		writePool: &redis.Pool{
			MaxIdle:     30,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", config["write_url"], redis.DialPassword(config["password"]))
			},
		},
	}
}
