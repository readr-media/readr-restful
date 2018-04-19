package models

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gopkg.in/mgo.v2/bson"
)

type CommentEvent struct {
	UpdatedAt  NullTime   `json:"updated_at" redis:"updated_at"`
	CreateAt   NullTime   `json:"created_at" redis:"created_at"`
	Body       NullString `json:"body" redis:"body"`
	AssetID    NullString `json:"asset_id" redis:"asset_id"`
	ParentID   NullString `json:"parent_id" redis:"parent_id"`
	AuthorID   NullString `json:"author_id" redis:"author_id"`
	ReplyCount NullInt    `json:"reply_count" redis:"reply_count"`
	Status     NullString `json:"status" redis:"status"`
	ID         NullString `json:"id" redis:"id"`
	Visible    NullBool   `json:"visible" redis:"visible"`
}

type CommentInfo struct {
	ID           string `id`
	ParentAuthor string `parent_author`
	ResourceType string `resource`
	ResourceID   string
	Commentors   []string `commentors`
}

func (c *CommentInfo) parse() {
	assetUrl := c.ResourceType

	r := regexp.MustCompile(`\/\/[a-zA-Z0-9_.]*\/(.*)\/[0-9]*$`)
	c.ResourceType = r.FindStringSubmatch(assetUrl)[1]

	s := regexp.MustCompile(`\/\/[a-zA-Z0-9_.]*\/.*\/([0-9]*)$`)
	c.ResourceID = s.FindStringSubmatch(assetUrl)[1]
}

type CommentNotification struct {
	ID           string `redis:"id"`
	Nickname     string `redis:"nickname"`
	ProfileImage string `redis:"profile_image"`
	ObjectName   string `redis:"object_name"`
	ObjectType   string `redis:"object_type"`
	ObjectID     string `redis:"object_id"`
	PostType     string `redis:"post_type"`
	EventType    string `redis:"event_type"`
	Timestamp    string `redis:"timestamp"`
	Read         bool   `redis:"read"`
}

func NewCommentNotification() CommentNotification {
	return CommentNotification{
		ID:        time.Now().Format("20060102150405"),
		Timestamp: time.Now().Format("20060102150405"),
		Read:      false,
	}
}

type commentHandler struct{}

func (c *commentHandler) SubscribeToChannel(channel string) {
	ctx, cancel := context.WithCancel(context.Background())

	ticker := time.NewTicker(1 * time.Minute)
	ticker.Stop()

	go func() {
		for _ = range ticker.C {
			select {
			case <-ctx.Done():
				break
			default:
			}
		}
		log.Println("SubscribeToChannel execution finished.")
		return
	}()

	err := RedisHelper.Subscribe(ctx, cancel,
		func(channel string, message []byte) error {
			var c CommentEvent
			if err := json.Unmarshal(message, &c); err != nil {
				log.Printf("Error scan redis comment event: %v", err)
			}
			log.Printf("channel: %s, message: %s\n", channel, c)
			return nil
		}, channel)

	if err != nil {
		log.Println(err.Error())
		cancel()
	}
}

func (c *commentHandler) CreateNotifications(comment CommentEvent) {
	var CommentNotifications map[string]CommentNotification
	//var Users []string
	//func (c *commentHandler) CreateNotifications(comment CommentEvent) {
	// Information to collect:
	// 1. parent_id -> author -> is author? -> [comment_reply_author, comment_reply]
	// 2. asset_id -> find all comment user id -> [comment_comment]
	// 3. asset_id 判斷 resource -> if post -> find user who follows the post -> [follow_post_reply]
	// 3. asset_id 判斷 resource -> if post -> find author -> find user who follows the author -> [follow_author_reply]
	// 3. asset_id 判斷 resource -> if project -> find user who follows the project -> [follow_project_reply]
	// 3. asset_id 判斷 resource -> if memo -> find project -> find user who follows the project -> [follow_memo_reply]

	// Setps:
	// Query mongodb, get parent comment author id, all comment author id, resource type and id
	// Find followers by calling follwoing API
	// Find user nickname, profile_image
	// Write to redis

	session := MongoSession.Get()
	mongoConn := session.DB("talk").C("comments")
	pipe := mongoConn.Pipe([]bson.M{
		bson.M{"$match": bson.M{"id": comment.ID}},
		bson.M{"$lookup": bson.M{"from": "assets", "localField": "asset_id", "foreignField": "id", "as": "asset"}},
		bson.M{"$lookup": bson.M{"from": "comments", "localField": "parent_id", "foreignField": "id", "as": "parent"}},
		bson.M{"$lookup": bson.M{"from": "comments", "localField": "asset_id", "foreignField": "asset_id", "as": "comments"}},
		bson.M{"$unwind": "$asset"},
		bson.M{"$unwind": "$parent"},
		bson.M{"$project": bson.M{"_id": false, "id": "$id", "resource": "$asset.url", "parent_author": "$parent.author_id", "commentors": "$comments.author_id"}},
	})
	var commentInfo CommentInfo
	pipe.One(&commentInfo)

	commentInfo.parse()
	switch commentInfo.ResourceType {
	case "post":
		var postFollowers []string
		err := DB.Get(&postFollowers, fmt.Sprintf("SELECT member_id FROM following_posts WHERE post_id = %s;", commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get postFollowers", commentInfo.ResourceID)
			return
		}

		var author NullString
		var postType NullString
		rows, err := DB.Query(fmt.Sprintf("SELECT author, post_type FROM posts WHERE post_id = %s LIMIT 1;", commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get post info", commentInfo.ResourceID)
			return
		}
		for rows.Next() {
			err = rows.Scan(&author, &postType)
		}

		var authorNickname NullString
		err = DB.Get(&authorNickname, fmt.Sprintf("SELECT nickname FROM members WHERE member_id = %s LIMIT 1;", author.String))
		if err != nil {
			log.Println("Error get authorNickname", author.String)
			return
		}

		var authorFollowers []string
		err = DB.Get(&authorFollowers, fmt.Sprintf("SELECT member_id FROM following_members WHERE custom_editor = %s;", author.String))
		if err != nil {
			log.Println("Error get authorFollowers", author.String)
			return
		}

		for _, v := range postFollowers {
			c := NewCommentNotification()
			c.EventType = "follow_post_reply"
			CommentNotifications[v] = c
		}

		for _, v := range authorFollowers {
			c := NewCommentNotification()
			c.EventType = "follow_member_reply"
			CommentNotifications[v] = c
		}

		query, args, err := sqlx.In("SELECT member_id FROM members WHERE talk_id IN (?);", commentInfo.Commentors)
		if err != nil {
			log.Println("Error in mixing sql `in` query", commentInfo.Commentors)
			return
		}

		query = DB.Rebind(query)
		rows, err = DB.Query(query, args...)
		if err != nil {
			return
		}

		for rows.Next() {
			var memberID string
			err = rows.Scan(&memberID)
			if err != nil {
				log.Println("Error scan member", rows)
				return
			}
			c := NewCommentNotification()
			c.EventType = "comment_comment"
			CommentNotifications[memberID] = c
		}

		if comment.ParentID.Valid {
			var parentCommentor string
			err = DB.Get(&parentCommentor, fmt.Sprintf("SELECT member_id FROM members WHERE talk_id = %s;", comment.ParentID.String))
			if err != nil {
				log.Println("Error get memberid by talkid", comment.ParentID.String)
				return
			}
			c := NewCommentNotification()
			if author.String == parentCommentor {
				c.EventType = "comment_reply_author"
			} else {
				c.EventType = "comment_reply"
			}
			CommentNotifications[parentCommentor] = c
		}

		for _, c := range CommentNotifications {
			c.ObjectName = authorNickname.String
			c.ObjectType = commentInfo.ResourceType
			c.ObjectID = commentInfo.ResourceID
			c.PostType = postType.String
		}

		break
	case "project":

		var projectFollowers []string
		err := DB.Get(&projectFollowers, fmt.Sprintf("SELECT member_id FROM following_projects WHERE project_id = %s;", commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get projectFollowers", commentInfo.ResourceID)
			return
		}

		var projectTitle NullString
		err = DB.Get(&projectTitle, fmt.Sprintf("SELECT title FROM projects WHERE project_id = %s LIMIT 1;", commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get project title", commentInfo.ResourceID)
			return
		}

		for _, v := range projectFollowers {
			c := NewCommentNotification()
			c.EventType = "follow_project_reply"
			CommentNotifications[v] = c
		}

		for _, c := range CommentNotifications {
			c.ObjectName = projectTitle.String
			c.ObjectType = commentInfo.ResourceType
			c.ObjectID = commentInfo.ResourceID
			c.PostType = ""
		}

		query, args, err := sqlx.In("SELECT member_id FROM members WHERE talk_id IN (?);", commentInfo.Commentors)
		if err != nil {
			log.Println("Error in mixing sql `in` query", commentInfo.Commentors)
			return
		}

		query = DB.Rebind(query)
		rows, err := DB.Query(query, args...)
		if err != nil {
			return
		}

		for rows.Next() {
			var memberID string
			err = rows.Scan(&memberID)
			if err != nil {
				log.Println("Error get memberid by talkid", rows)
				return
			}
			c := NewCommentNotification()
			c.EventType = "comment_comment"
			CommentNotifications[memberID] = c
		}

		if comment.ParentID.Valid {
			var parentCommentor string
			err = DB.Get(&parentCommentor, fmt.Sprintf("SELECT member_id FROM members WHERE talk_id = %s;", comment.ParentID.String))
			if err != nil {
				log.Println("Error get memberid by talkid", comment.ParentID.String)
				return
			}
			c := NewCommentNotification()
			c.EventType = "comment_reply"
			CommentNotifications[parentCommentor] = c
		}

		break
	default:
		break
	}

	keys := make([]string, len(CommentNotifications))
	i := 0
	for k := range CommentNotifications {
		keys[i] = k
		i++
	}

	query, args, err := sqlx.In("SELECT member_id, nickname, profile_image FROM members WHERE member_id IN (?);", keys)
	if err != nil {
		log.Println("Error get member profiles building `in` query", keys)
		return
	}

	query = DB.Rebind(query)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return
	}

	for rows.Next() {
		var memberID string
		var nickName string
		var profileImage string
		err = rows.Scan(&memberID, &nickName, &profileImage)
		if err != nil {
			log.Println("Error get member profiles", keys)
			return
		}
		n := CommentNotifications[memberID]
		n.Nickname = nickName
		n.ProfileImage = profileImage
		CommentNotifications[memberID] = n
	}

	conn := RedisHelper.Conn()
	defer conn.Close()

	conn.Send("MULTI")
	for k, v := range CommentNotifications {
		conn.Send("HMSET", redis.Args{}.Add(fmt.Sprint("notify_", k)).AddFlat(&v)...)
	}
	if _, err := redis.Values(conn.Do("EXEC")); err != nil {
		log.Printf("Error insert cache to redis: %v", err)
		return
	}

}

//func (c *commentHandler) UpdateCommentCount(event_type string, comment CommentEvent) {}

var CommentHandler = commentHandler{}
