package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
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

func convertRedisAssign(dest, src interface{}) error {
	var err error
	b, ok := src.([]byte)
	if !ok {
		return errors.New("RedisScan error assert byte array")
	}
	s := string(b)
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		fmt.Println(string(b), " failed")
	} else {
		s = strings.TrimPrefix(s, "{")
		s = strings.TrimSuffix(s, "}")

		if strings.HasSuffix(s, " true") {
			s = strings.TrimSuffix(s, " true")

			switch d := dest.(type) {
			case *NullString:
				d.String, d.Valid = string(s), true
				return nil
			case *NullTime:
				d.Time, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", s)
				if err != nil {
					fmt.Println(err)
					return err
				}
				d.Valid = true
			case *NullInt:
				d.Int, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					fmt.Println(err)
					return err
				}
				d.Valid = true
			case *NullBool:
				d.Bool, err = strconv.ParseBool(s)
				if err != nil {
					fmt.Println(err)
					return err
				}
				d.Valid = true
			default:
				fmt.Println(s, " non case ", d)
				return errors.New("Cannot parse non-nil nullable type")
			}

		} else if strings.HasSuffix(s, " false") {
			s = strings.TrimSuffix(s, " false")

			switch d := dest.(type) {
			case *NullString:
				d.String, d.Valid = "", false
				return nil
			case *NullTime:
				d.Time, d.Valid = time.Time{}, false
				return nil
			case *NullInt:
				d.Int, d.Valid = 0, false
			case *NullBool:
				d.Valid, d.Valid = false, false
			default:
				fmt.Println(s, " FALSE non case ", d)
				return errors.New("redis conversion error: invalid null* valid field")
			}
		}
	}
	return nil
}

func (r *redisHelper) GetHotPosts(keysTemplate string, quantity int) (result []HotPost, err error) {
	conn := r.Conn()
	defer conn.Close()

	for i := 1; i <= quantity; i++ {
		var p HotPost
		key := fmt.Sprintf(keysTemplate, i)
		res, err := redis.Values(conn.Do("HGETALL", key))
		if err != nil {
			log.Printf("Error getting redis key: %v", err)
			return result, err
		}
		if err = redis.ScanStruct(res, &p); err != nil {
			log.Printf("Error scan redis key: %v", err)
			return result, err
		}
		result = append(result, p)
	}

	return result, err
}

func (r *redisHelper) Subscribe(ctx context.Context, cancel func(), onMessage func(channel string, data []byte) error, channel string) error {

	conn := r.Conn()
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
	RedisHelper = redisHelper{&redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", config["url"], redis.DialPassword(config["password"]))
		},
	}}
}
