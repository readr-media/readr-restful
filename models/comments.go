package models

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

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
	ID               string `id`
	ParentAuthor     string `parent_author`
	ResourceType     string `resource`
	ResourceID       string
	AuthorID         int      `db:"id"`
	AuthorMail       string   `db:"member_id"`
	AuthorNickname   string   `db:"nickname"`
	ResourcePostType string   `db:"type"`
	Commentors       []string `commentors`
}

type CommentorInfo struct {
	ID       NullString `db:"member_id"`
	Image    NullString `db:"profile_image"`
	Nickname NullString `db:"nickname"`
}

func (c *CommentInfo) Parse() {
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

type UpdateNotificationArgs struct {
	IDs      []string `json:"ids"`
	MemberID string   `redis:"member_id" json:"member_id"`
	Read     NullBool `redis:"read" json:"read"`
}

type CommentInfoGetter interface {
	GetCommentInfo(comment CommentEvent) (commentInfo CommentInfo)
}

type CommentInfoGet struct{}

func (c *CommentInfoGet) GetCommentInfo(comment CommentEvent) (commentInfo CommentInfo) {

	session := MongoSession.Get()
	mongoConn := session.DB("talk").C("comments")

	if comment.ParentID.Valid {
		pipe := mongoConn.Pipe([]bson.M{
			bson.M{"$match": bson.M{"id": comment.ID.String}},
			bson.M{"$lookup": bson.M{"from": "assets", "localField": "asset_id", "foreignField": "id", "as": "asset"}},
			bson.M{"$lookup": bson.M{"from": "comments", "localField": "parent_id", "foreignField": "id", "as": "parents"}},
			bson.M{"$lookup": bson.M{"from": "comments", "localField": "asset_id", "foreignField": "asset_id", "as": "comments"}},
			bson.M{"$unwind": "$asset"},
			bson.M{"$unwind": "$parents"},
			bson.M{"$project": bson.M{"_id": false, "id": "$id", "resource": "$asset.url", "parent_author": "$parents.author_id", "commentors": "$comments.author_id"}},
		})
		pipe.One(&commentInfo)
	} else {
		pipe := mongoConn.Pipe([]bson.M{
			bson.M{"$match": bson.M{"id": comment.ID.String}},
			bson.M{"$lookup": bson.M{"from": "assets", "localField": "asset_id", "foreignField": "id", "as": "asset"}},
			bson.M{"$lookup": bson.M{"from": "comments", "localField": "asset_id", "foreignField": "asset_id", "as": "comments"}},
			bson.M{"$unwind": "$asset"},
			bson.M{"$project": bson.M{"_id": false, "id": "$id", "resource": "$asset.url", "commentors": "$comments.author_id"}},
		})
		pipe.One(&commentInfo)
	}

	commentInfo.Parse()
	return commentInfo
}

type CommentHandlerStruct struct {
	CommentInfoGetter
}

func (c *CommentHandlerStruct) GetCommentorInfo(talk_id string) (commentorInfo CommentorInfo) {
	err := DB.QueryRowx(fmt.Sprintf(`SELECT member_id, profile_image, nickname FROM members WHERE talk_id="%s" LIMIT 1;`, talk_id)).StructScan(&commentorInfo)
	if err != nil {
		log.Println("Error commentor info", talk_id, err.Error())
	}
	return commentorInfo
}

func (c *CommentHandlerStruct) GetCommentorMemberIDs(ids []string) (result []string) {
	query, args, err := sqlx.In(`SELECT member_id FROM members WHERE talk_id IN (?);`, ids)
	if err != nil {
		log.Println("Error in mixing sql `in` query", ids, err.Error())
		return result
	}
	query = DB.Rebind(query)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return result
	}
	for rows.Next() {
		var memberID string
		err = rows.Scan(&memberID)
		if err != nil {
			log.Println("Error get memberid by talkid", rows, err.Error())
			return result
		}
		result = append(result, memberID)
	}
	return result
}

func (c *CommentHandlerStruct) CreateNotifications(comment CommentEvent) {
	CommentNotifications := Notifications{} //:= make(map[string]Notification)

	//func (c *commentHandler) CreateNotifications(comment CommentEvent) {
	// Information to collect:
	// 1. parent_id -> author -> is author? -> [comment_reply_author, comment_reply]
	// 2. asset_id -> find all comment user id -> [comment_comment]
	// 3. asset_id 判斷 resource -> if post -> find user who follows the post -> [follow_post_reply]
	// 3. asset_id 判斷 resource -> if post -> find author -> find user who follows the author -> [follow_author_reply]
	// 3. asset_id 判斷 resource -> if project -> find user who follows the project -> [follow_project_reply]
	// 3. asset_id 判斷 resource -> if memo -> find project -> find user who follows the project -> [follow_memo_reply]

	// Setps:
	// Find followers by calling follwoing API
	// Find user nickname, profile_image
	// Write to redis

	// Module: Query mongodb, get parent comment author id, all comment author id, resource type and id
	commentInfo := c.GetCommentInfo(comment)

	// Module: Get commentor's Profile Info
	commentorInfo := c.GetCommentorInfo(comment.AuthorID.String)

	switch commentInfo.ResourceType {
	case "post":

		// Module: find post's followers
		i, err := strconv.Atoi(commentInfo.ResourceID)
		if err != nil {
			log.Println("Error parsing post id", commentInfo.ResourceID, err.Error())
		}
		postFollowers, err := FollowingAPI.GetFollowerMemberIDs("post", strconv.Itoa(i))
		if err != nil {
			log.Println("Error get post followers", commentInfo.ResourceID, err.Error())
		}

		// Module: find post post's info
		/*
			var author NullString
			var postType NullString
			var authorNickname NullString
		*/
		//rows, err = DB.Query(fmt.Sprintf(`SELECT m.member_id AS author, m.nickname, type FROM posts LEFT JOIN members AS m ON m.id = posts.author WHERE posts.post_id=%s LIMIT 1;`, commentInfo.ResourceID))
		err = DB.QueryRowx(fmt.Sprintf(`SELECT m.id, m.member_id, m.nickname, type FROM posts LEFT JOIN members AS m ON m.id = posts.author WHERE posts.post_id=%s LIMIT 1;`, commentInfo.ResourceID)).StructScan(&commentInfo)
		if err != nil {
			log.Println("Error get comment info", commentInfo.ResourceID, err.Error())
			return
		}

		/*for rows.Next() {
			err = rows.Scan(&author, &authorNickname, &postType)
		}
		if !author.Valid || !authorNickname.Valid || !postType.Valid {

		}

		commentInfo.AuthorNickname = authorNickname.String
		*/
		// end module

		// Module: find post author's followers
		authorFollowers, err := FollowingAPI.GetFollowerMemberIDs("member", strconv.Itoa(commentInfo.AuthorID))
		if err != nil {
			log.Println("Error get author followers", commentInfo.AuthorID, err.Error())
		}

		for _, v := range postFollowers {
			if v != commentorInfo.ID.String {
				CommentNotifications[v] = NewNotification("follow_post_reply")
			}
		}

		for _, v := range authorFollowers {
			if v != commentorInfo.ID.String {
				CommentNotifications[v] = NewNotification("follow_member_reply")
			}
		}

		if commentInfo.AuthorMail != "" && commentInfo.AuthorMail != commentorInfo.ID.String {
			CommentNotifications[commentInfo.AuthorMail] = NewNotification("post_reply")
		}

		if len(commentInfo.Commentors) > 0 {
			// Module get member_id of all commentors
			memberIDs := c.GetCommentorMemberIDs(commentInfo.Commentors)

			for _, id := range memberIDs {
				if id != commentorInfo.ID.String {
					CommentNotifications[id] = NewNotification("comment_comment")
				}
			}
		}

		if commentInfo.ParentAuthor != "" && commentInfo.ParentAuthor != comment.AuthorID.String {
			// Module: Get parent author's id
			var parentCommentor string
			err = DB.Get(&parentCommentor, fmt.Sprintf(`SELECT member_id FROM members WHERE talk_id="%s";`, commentInfo.ParentAuthor))
			if err != nil {
				log.Println("Error get memberid by talkid", commentInfo.ParentAuthor, err.Error())
				return
			}
			if commentInfo.AuthorMail == commentorInfo.ID.String {
				CommentNotifications[parentCommentor] = NewNotification("comment_reply_author")
			} else {
				CommentNotifications[parentCommentor] = NewNotification("comment_reply")
			}
			// end module
		}

		for k, v := range CommentNotifications {
			v.SetCommentSubjects(commentorInfo)
			v.SetCommentObjects(commentInfo)
			CommentNotifications[k] = v
		}

		break
	case "project":
	case "memo":

		// Module: find post project's followers
		project_id, err := strconv.Atoi(commentInfo.ResourceID)
		if err != nil {
			log.Println("Error convert project id", commentInfo.ResourceID, err.Error())
		}
		projectFollowers, err := FollowingAPI.GetFollowerMemberIDs("project", strconv.Itoa(project_id))
		if err != nil {
			log.Println("Error get project followers", commentInfo.ResourceID, err.Error())
		}

		var projectTitle NullString
		err = DB.Get(&projectTitle, fmt.Sprintf(`SELECT title FROM projects WHERE project_id=%s LIMIT 1;`, commentInfo.ResourceID))
		if err != nil {
			log.Println("Error get project title", commentInfo.ResourceID, err.Error())
			return
		}

		for _, v := range projectFollowers {
			if v != commentorInfo.ID.String {
				if commentInfo.ResourceType == "project" {
					CommentNotifications[v] = NewNotification("follow_project_reply")
				} else if commentInfo.ResourceType == "memo" {
					CommentNotifications[v] = NewNotification("follow_memo_reply")
				}
			}
		}

		if len(commentInfo.Commentors) > 0 {
			// Module get member_id of all commentors
			memberIDs := c.GetCommentorMemberIDs(commentInfo.Commentors)

			for _, id := range memberIDs {
				if id != commentorInfo.ID.String {
					CommentNotifications[id] = NewNotification("comment_comment")
				}
			}
		}

		if commentInfo.ParentAuthor != "" && commentInfo.ParentAuthor != commentorInfo.ID.String {
			var parentCommentor string
			err = DB.Get(&parentCommentor, fmt.Sprintf(`SELECT member_id FROM members WHERE talk_id="%s";`, comment.ParentID.String))
			if err != nil {
				log.Println("Error get memberid by talkid", comment.ParentID.String, err.Error())
				return
			}
			CommentNotifications[parentCommentor] = NewNotification("comment_reply")
		}

		for k, v := range CommentNotifications {
			v.SetCommentSubjects(commentorInfo)
			v.SetCommentObjects(commentInfo)
			CommentNotifications[k] = v
		}

		break
	default:
		break
	}

	if len(CommentNotifications) > 0 {
		CommentNotifications.Send()
	}
}

//func (c *commentHandler) UpdateCommentCount(event_type string, comment CommentEvent) {}
func (c *CommentHandlerStruct) ReadNotifications(arg UpdateNotificationArgs) error {
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

var CommentHandler = CommentHandlerStruct{new(CommentInfoGet)}
