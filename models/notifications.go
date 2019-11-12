package models

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/utils"
)

type UpdateNotificationArgs struct {
	IDs      []string       `json:"ids"`
	MemberID string         `redis:"member_id" json:"member_id"`
	Read     rrsql.NullBool `redis:"read" json:"read"`
}

type Notification struct {
	Member       int    `redis:"-" json:"-"`
	ID           string `redis:"id" json:"id"`
	SubjectID    string `redis:"subject_id" json:"subject_id"`
	Nickname     string `redis:"nickname" json:"nickname"`
	ProfileImage string `redis:"profile_image" json:"profile_image"`
	ObjectName   string `redis:"object_name" json:"object_name,omitempty"`
	ObjectType   string `redis:"object_type" json:"object_type,omitempty"`
	ObjectID     string `redis:"object_id" json:"object_id,omitempty"`
	ObjectSlug   string `redis:"object_slug" json:"object_slug,omitempty"`
	PostType     string `redis:"post_type" json:"post_type,omitempty"`
	EventType    string `redis:"event_type" json:"event_type"`
	Timestamp    string `redis:"timestamp" json:"timestamp"`
	Read         bool   `redis:"read" json:"read"`
}

func NewNotification(event string, member_id int) Notification {
	tz, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Println("Load timezone location error")
	}
	return Notification{
		Timestamp: time.Now().In(tz).Format("20060102150405"),
		EventType: event,
		Member:    member_id,
		Read:      false,
	}
}

type Notifications []Notification

func (n Notifications) Send() {
	ids := make([]int, 0, len(n))
	for _, v := range n {
		ids = append(ids, v.Member)
	}

	if len(ids) == 0 {
		return //no need further actions
	}

	mailMapping, err := n.getMemberMail(ids)
	if err != nil {
		log.Printf("Error getting mail list: %v", err)
		return
	}

	conn := RedisHelper.WriteConn()
	defer conn.Close()

	conn.Send("MULTI")
	nsm := groupNotifications(n)
	for memberID, ns := range nsm {
		key := fmt.Sprint("notify_", mailMapping[memberID])

		redisNs, err := getNotificationsFromRedis(key)
		if err != nil {
			return
		}
		for _, v := range ns {
			for _, redisV := range redisNs {
				if redisV.ObjectType == v.ObjectType && redisV.ObjectID == v.ObjectID && redisV.EventType == v.EventType {
					msg, _ := json.Marshal(redisV)
					conn.Send("LREM", redis.Args{}.Add(key).Add("1").Add(msg)...)
					break
				}
			}

			UUID, _ := utils.NewUUIDv4()
			v.ID = UUID.String()
			msg, err := json.Marshal(v)
			if err != nil {
				log.Printf("Error marshaling notification comment event: %v", err)
			}
			conn.Send("LPUSH", redis.Args{}.Add(key).Add(msg)...)
		}
		conn.Send("LTRIM", redis.Args{}.Add(key).Add(0).Add(config.Config.Redis.Cache.NotificationCount-1)...)
	}
	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}
}

func (n Notifications) getMemberMail(ids []int) (result map[int]string, err error) {
	type memberMail struct {
		ID   int    `db:"id"`
		Mail string `db:"mail"`
	}
	result = make(map[int]string)
	query := "SELECT id, mail FROM members WHERE id IN (?);"
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	rows, err := rrsql.DB.Queryx(query, args...)

	for rows.Next() {
		var mm memberMail
		if err = rows.StructScan(&mm); err != nil {
			return result, err
		}
		result[mm.ID] = mm.Mail
	}

	return result, nil
}

type commentHandler struct{}

func (c *commentHandler) ReadNotifications(arg UpdateNotificationArgs) error {
	readConn := RedisHelper.ReadConn()
	writeConn := RedisHelper.WriteConn()
	defer readConn.Close()
	defer writeConn.Close()

	CommentNotifications := [][]byte{}

	key := fmt.Sprint("notify_", arg.MemberID)

	res, err := redis.Values(readConn.Do("LRANGE", key, 0, config.Config.Redis.Cache.NotificationCount-1))
	if err != nil {
		log.Printf("Error getting redis key: %s , %v", key, err)
		return err
	}
	if err = redis.ScanSlice(res, &CommentNotifications); err != nil {
		log.Printf("Error scan redis key: %s , %v", key, err)
		return err
	}

	if len(arg.IDs) > 0 {
		for _, v := range arg.IDs {
			k, err := strconv.Atoi(v)
			if err != nil {
				log.Printf("Error convert ids into integer index: %v", err)
				continue
			}
			var cn Notification
			if err := json.Unmarshal(CommentNotifications[k], &cn); err != nil {
				log.Printf("Error scan redis comment notification: %s , %v", CommentNotifications[k], err)
				continue
			}
			cn.Read = true
			msg, err := json.Marshal(cn)
			if err != nil {
				log.Printf("Error dump redis comment notification: %v", err)
				continue
			}
			CommentNotifications[k] = msg
		}
	} else {
		for k, v := range CommentNotifications {
			var cn Notification
			if err := json.Unmarshal(v, &cn); err != nil {

				log.Printf("Error scan redis comment notification: %s , %v", v, err)
				continue
			}
			cn.Read = true
			msg, err := json.Marshal(cn)
			if err != nil {
				log.Printf("Error dump redis comment notification: %v", err)
				continue
			}
			CommentNotifications[k] = msg
		}
	}

	writeConn.Do("DEL", fmt.Sprint("notify_", arg.MemberID))
	writeConn.Send("MULTI")
	for _, v := range CommentNotifications {
		writeConn.Send("RPUSH", redis.Args{}.Add(fmt.Sprint("notify_", arg.MemberID)).Add(v)...)
	}
	writeConn.Send("LTRIM", redis.Args{}.Add(fmt.Sprint("notify_", arg.MemberID)).Add(0).Add(config.Config.Redis.Cache.NotificationCount-1)...)
	if _, err := redis.Values(writeConn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return err
	}

	return nil
}

var CommentHandler = commentHandler{}

func groupNotifications(ns Notifications) map[int][]Notification {
	nsm := make(map[int][]Notification)
OuterLoop:
	for _, v := range ns {
		for k, u := range nsm[v.Member] {
			if v.ObjectType == u.ObjectType && v.ObjectID == u.ObjectID && v.EventType == u.EventType {
				nsm[v.Member][k] = v
				break OuterLoop
			}
		}
		nsm[v.Member] = append(nsm[v.Member], v)
	}

	return nsm
}

func getNotificationsFromRedis(key string) (redisNs []Notification, err error) {

	conn := RedisHelper.ReadConn()
	defer conn.Close()

	llen, err := RedisHelper.GetRedisListLength(key)
	if err != nil {
		return redisNs, err
	}
	if llen > config.Config.Redis.Cache.NotificationCount {
		llen = config.Config.Redis.Cache.NotificationCount - 1
	} else {
		llen -= 1
	}
	res, err := redis.Values(conn.Do("LRANGE", key, 0, llen))
	if err != nil {
		log.Printf("Error getting redis key: %s , %v", key, err)
		return redisNs, err
	}

	redisNsByte := [][]byte{}
	if err = redis.ScanSlice(res, &redisNsByte); err != nil {
		log.Printf("Error scan redis key: %s , %v", key, err)
		return redisNs, err
	}

	for _, redisNByte := range redisNsByte {
		var redisN Notification
		if err := json.Unmarshal(redisNByte, &redisN); err != nil {
			log.Printf("Error scan redis comment notification: %s , %v", string(redisNByte), err)
			break
		}
		redisNs = append(redisNs, redisN)
	}
	return redisNs, nil
}
