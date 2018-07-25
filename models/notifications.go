package models

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
)

type UpdateNotificationArgs struct {
	IDs      []string `json:"ids"`
	MemberID string   `redis:"member_id" json:"member_id"`
	Read     NullBool `redis:"read" json:"read"`
}

type Notification struct {
	ID           string `redis:"id" json:"id"`
	SubjectID    int    `redis:"subject_id" json:"subject_id"`
	Nickname     string `redis:"nickname" json:"nickname"`
	ProfileImage string `redis:"profile_image" json:"profile_image"`
	ObjectName   string `redis:"object_name" json:"object_name,omitempty"`
	ObjectType   string `redis:"object_type" json:"object_type,omitempty"`
	ObjectID     int    `redis:"object_id" json:"object_id,omitempty"`
	ObjectSlug   string `redis:"object_slug" json:"object_slug,omitempty"`
	PostType     int    `redis:"post_type" json:"post_type,omitempty"`
	EventType    string `redis:"event_type" json:"event_type"`
	Timestamp    string `redis:"timestamp" json:"timestamp"`
	Read         bool   `redis:"read" json:"read"`
}

func NewNotification(event string) Notification {
	tz, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Println("Load timezone location error")
	}
	return Notification{
		ID:        time.Now().In(tz).Format("20060102150405"),
		Timestamp: time.Now().In(tz).Format("20060102150405"),
		EventType: event,
		Read:      false,
	}
}

type Notifications map[int]Notification

func (n Notifications) Send() {
	ids := make([]int, 0, len(n))
	for k, _ := range n {
		ids = append(ids, k)
	}

	if len(ids) == 0 {
		return //no need further actions
	}

	mailMapping, err := n.getMemberMail(ids)
	if err != nil {
		log.Printf("Error getting mail list: %v", err)
		return
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	for k, v := range n {

		ns := [][]byte{}
		key := fmt.Sprint("notify_", mailMapping[k])

		res, err := redis.Values(conn.Do("LRANGE", key, "0", "49"))
		if err != nil {
			msg, err := json.Marshal(v)
			if err != nil {
				log.Printf("Error marshaling notification comment event: %v", err)
			}
			conn.Send("LPUSH", redis.Args{}.Add(fmt.Sprint(key)).Add(msg)...)
			conn.Send("LTRIM", redis.Args{}.Add(fmt.Sprint(key)).Add(0).Add(49)...)
		} else {
			if err = redis.ScanSlice(res, &ns); err != nil {
				log.Printf("Error scan redis key: %s , %v", key, err)
				return
			}

			for _, kv := range ns {
				var n Notification
				if err := json.Unmarshal(kv, &n); err != nil {
					log.Printf("Error scan redis comment notification: %s , %v", kv, err)
					break
				}
				if n.SubjectID == v.SubjectID && n.ObjectID == v.ObjectID && n.EventType == v.EventType {
					break
				}

				msg, err := json.Marshal(v)
				if err != nil {
					log.Printf("Error marshaling notification comment event: %v", err)
				}
				conn.Send("LPUSH", redis.Args{}.Add(fmt.Sprint(key)).Add(msg)...)
				conn.Send("LTRIM", redis.Args{}.Add(fmt.Sprint(key)).Add(0).Add(49)...)
			}
		}
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
	rows, err := DB.Queryx(query, args...)

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
	conn := RedisHelper.Conn()
	defer conn.Close()

	CommentNotifications := [][]byte{}

	key := fmt.Sprint("notify_", arg.MemberID)

	res, err := redis.Values(conn.Do("LRANGE", key, "0", "49"))
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

	conn.Do("DEL", fmt.Sprint("notify_", arg.MemberID))
	conn.Send("MULTI")
	for _, v := range CommentNotifications {
		conn.Send("RPUSH", redis.Args{}.Add(fmt.Sprint("notify_", arg.MemberID)).Add(v)...)
	}
	conn.Send("LTRIM", redis.Args{}.Add(fmt.Sprint("notify_", arg.MemberID)).Add(0).Add(49)...)
	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return err
	}

	return nil

}

var CommentHandler = commentHandler{}
