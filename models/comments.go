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
	rresult := r.FindStringSubmatch(assetUrl)
	if len(rresult) < 2 {
		c.ResourceType = ""
	} else {
		c.ResourceType = rresult[1]
	}

	s := regexp.MustCompile(`\/\/[a-zA-Z0-9_.]*\/.*\/([0-9]*)$`)
	sresult := s.FindStringSubmatch(assetUrl)
	if len(sresult) < 2 {
		c.ResourceID = ""
	} else {
		c.ResourceID = sresult[1]
	}
}

type CommentNotification struct {
	ID           string `redis:"id" json:"id"`
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
			var commentEvent CommentEvent
			if err := json.Unmarshal(message, &commentEvent); err != nil {
				log.Printf("Error scan redis comment event: %v", err)
			}
			if channel == "commentAdded" {
				c.CreateNotifications(commentEvent)
			}
			return nil
		}, channel)

	if err != nil {
		log.Println(err.Error())
		cancel()
	}
}

func (c *commentHandler) CreateNotifications(comment CommentEvent) {
	CommentNotifications := make(map[string]CommentNotification)

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
		rows, err := DB.Query(fmt.Sprintf(`SELECT member_id FROM following_posts WHERE post_id=%s;`, commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get postFollowers", commentInfo.ResourceID, err.Error())
		}
		for rows.Next() {
			var follower string
			err = rows.Scan(&follower)
			if err != nil {
				log.Println("Error scan postFollowers", commentInfo.ResourceID, err.Error())
			}
			postFollowers = append(postFollowers, follower)
		}

		var author NullString
		var postType NullString
		rows, err = DB.Query(fmt.Sprintf(`SELECT author, type FROM posts WHERE post_id=%s LIMIT 1;`, commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get post info", commentInfo.ResourceID, err.Error())
			return
		}
		for rows.Next() {
			err = rows.Scan(&author, &postType)
		}

		var authorNickname NullString
		err = DB.Get(&authorNickname, fmt.Sprintf(`SELECT nickname FROM members WHERE member_id="%s" LIMIT 1;`, author.String))
		if err != nil {
			log.Println("Error get authorNickname", author.String, err.Error())
			return
		}

		var authorFollowers []string
		rows, err = DB.Query(fmt.Sprintf(`SELECT member_id FROM following_members WHERE custom_editor="%s";`, author.String))
		if err != nil {
			log.Println("Error get authorFollowers", author.String, err.Error())
		}
		for rows.Next() {
			var follower string
			err = rows.Scan(&follower)
			if err != nil {
				log.Println("Error scan authorFollowers", author.String, err.Error())
			}
			authorFollowers = append(authorFollowers, follower)
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

		if len(commentInfo.Commentors) > 0 {
			query, args, err := sqlx.In(`SELECT member_id FROM members WHERE talk_id IN (?);`, commentInfo.Commentors)
			if err != nil {
				log.Println("Error in mixing sql `in` query", commentInfo.Commentors, err.Error())
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
					log.Println("Error scan member", rows, err.Error())
					return
				}
				c := NewCommentNotification()
				c.EventType = "comment_comment"
				CommentNotifications[memberID] = c
			}
		}

		if comment.ParentID.Valid {
			var parentCommentor string
			err = DB.Get(&parentCommentor, fmt.Sprintf(`SELECT member_id FROM members WHERE talk_id="%s";`, comment.ParentID.String))
			if err != nil {
				log.Println("Error get memberid by talkid", comment.ParentID.String, err.Error())
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

		for k, v := range CommentNotifications {
			v.ObjectName = authorNickname.String
			v.ObjectType = commentInfo.ResourceType
			v.ObjectID = commentInfo.ResourceID
			v.PostType = postType.String
			CommentNotifications[k] = v
		}

		break
	case "project":
		var projectFollowers []string
		rows, err := DB.Query(fmt.Sprintf(`SELECT member_id FROM following_projects WHERE project_id=%s;`, commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get projectFollowers", commentInfo.ResourceID, err.Error())
		}
		for rows.Next() {
			var follower string
			err = rows.Scan(&follower)
			if err != nil {
				log.Println("Error scan authorFollowers", commentInfo.ResourceID, err.Error())
			}
			projectFollowers = append(projectFollowers, follower)
		}

		var projectTitle NullString
		err = DB.Get(&projectTitle, fmt.Sprintf(`SELECT title FROM projects WHERE project_id=%s LIMIT 1;`, commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get project title", commentInfo.ResourceID, err.Error())
			return
		}

		for _, v := range projectFollowers {
			c := NewCommentNotification()
			c.EventType = "follow_project_reply"
			CommentNotifications[v] = c
		}

		if len(commentInfo.Commentors) > 0 {
			query, args, err := sqlx.In(`SELECT member_id FROM members WHERE talk_id IN (?);`, commentInfo.Commentors)
			if err != nil {
				log.Println("Error in mixing sql `in` query", commentInfo.Commentors, err.Error())
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
					log.Println("Error get memberid by talkid", rows, err.Error())
					return
				}
				c := NewCommentNotification()
				c.EventType = "comment_comment"
				CommentNotifications[memberID] = c
			}
		}

		if comment.ParentID.Valid {
			var parentCommentor string
			err = DB.Get(&parentCommentor, fmt.Sprintf(`SELECT member_id FROM members WHERE talk_id="%s";`, comment.ParentID.String))
			if err != nil {
				log.Println("Error get memberid by talkid", comment.ParentID.String, err.Error())
				return
			}
			c := NewCommentNotification()
			c.EventType = "comment_reply"
			CommentNotifications[parentCommentor] = c
		}

		for k, v := range CommentNotifications {
			v.ObjectName = projectTitle.String
			v.ObjectType = commentInfo.ResourceType
			v.ObjectID = commentInfo.ResourceID
			v.PostType = ""
			CommentNotifications[k] = v
		}

		break
	default:
		break
	}

	if len(CommentNotifications) > 0 {
		keys := make([]string, len(CommentNotifications))
		i := 0
		for k := range CommentNotifications {
			keys[i] = k
			i++
		}

		query, args, err := sqlx.In(`SELECT member_id, nickname, profile_image FROM members WHERE member_id IN (?);`, keys)
		if err != nil {
			log.Println("Error get member profiles building `in` query", keys, err.Error())
			return
		}

		query = DB.Rebind(query)
		rows, err := DB.Query(query, args...)
		if err != nil {
			return
		}

		for rows.Next() {
			var memberID string
			var nickName NullString
			var profileImage NullString
			err = rows.Scan(&memberID, &nickName, &profileImage)
			if err != nil {
				log.Println("Error get member profiles", keys, err.Error())
				return
			}
			n := CommentNotifications[memberID]
			n.Nickname = nickName.String
			n.ProfileImage = profileImage.String
			CommentNotifications[memberID] = n
		}

		conn := RedisHelper.Conn()
		defer conn.Close()

		conn.Send("MULTI")
		for k, v := range CommentNotifications {
			msg, err := json.Marshal(v)
			if err != nil {
				log.Printf("Error marshaling notification comment event: %v", err)
			}
			conn.Send("LPUSH", redis.Args{}.Add(fmt.Sprint("notify_", k)).Add(msg)...)
			conn.Send("LTRIM", redis.Args{}.Add(fmt.Sprint("notify_", k)).Add(0).Add(49)...)
		}
		if _, err := redis.Values(conn.Do("EXEC")); err != nil {
			log.Printf("Error insert cache to redis: %v", err)
			return
		}
	}
}

//func (c *commentHandler) UpdateCommentCount(event_type string, comment CommentEvent) {}

var CommentHandler = commentHandler{}
