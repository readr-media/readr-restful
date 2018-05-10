package models

import (
	"fmt"
	"log"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type Notification struct {
	ID           string `redis:"id" json:"id"`
	SubjectID    string `redis:"subject_id" json:"subject_id"`
	Nickname     string `redis:"nickname" json:"nickname"`
	ProfileImage string `redis:"profile_image" json:"profile_image"`
	ObjectName   string `redis:"object_name" json:"object_name"`
	ObjectType   string `redis:"object_type" json:"object_type"`
	ObjectID     string `redis:"object_id" json:"object_id"`
	PostType     string `redis:"post_type" json:"post_type"`
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

func (n *Notification) SetCommentSubjects(s CommentorInfo) {
	n.SubjectID = s.ID.String
	n.Nickname = s.Nickname.String
	n.ProfileImage = s.Image.String
}

func (n *Notification) SetCommentObjects(s CommentInfo) {
	n.ObjectName = s.AuthorNickname
	n.ObjectType = s.ResourceType
	n.ObjectID = s.ResourceID
	n.PostType = s.ResourcePostType
}

type Notifications map[string]Notification

func (n Notifications) Send() {
	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	for k, v := range n {

		ns := [][]byte{}
		key := fmt.Sprint("notify_", k)

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
