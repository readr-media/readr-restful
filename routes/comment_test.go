package routes

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/readr-media/readr-restful/models"
)

type mockCommentAPI struct{}

func (c *mockCommentAPI) GetCommentInfo(comment models.CommentEvent) (commentInfo models.CommentInfo) {
	switch comment.Body.String {
	case "comment_reply_author":
		commentInfo = models.CommentInfo{ParentAuthor: "abc1d5b1-da54-4200-b90e-f06e59fd9487", ResourceType: "https://readr.tw/post/91"}
	case "comment_reply":
		commentInfo = models.CommentInfo{ParentAuthor: "abc1d5b1-da54-4200-b90e-f06e59fd9487", ResourceType: "https://readr.tw/post/91"}
	case "comment_comment":
		commentInfo = models.CommentInfo{ResourceType: "https://readr.tw/post/92", Commentors: []string{"abc1d5b1-da54-4200-b90e-f06e59fd9487"}}
	case "follow_member_reply":
		commentInfo = models.CommentInfo{ResourceType: "https://readr.tw/post/90"}
	case "follow_post_reply":
		commentInfo = models.CommentInfo{ResourceType: "https://readr.tw/post/92"}
	case "follow_project_reply":
		commentInfo = models.CommentInfo{ResourceType: "https://readr.tw/project/920"}
	case "follow_memo_reply":
		commentInfo = models.CommentInfo{ResourceType: "https://readr.tw/memo/920"}
	case "post_reply":
		commentInfo = models.CommentInfo{ParentAuthor: "abc1d5b1-da54-4200-b90e-f06e59fd9487", ResourceType: "https://readr.tw/post/91"}
	}
	commentInfo.Parse()
	return commentInfo
}

func TestRouteComments(t *testing.T) {
	log.Println("test start")

	for _, params := range []models.Member{
		models.Member{ID: 90, MemberID: "commenttest0@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"commenttest0@mirrormedia.mg", true}, Points: models.NullInt{0, true}, TalkID: models.NullString{"abc1d5b1-da54-4200-b90e-f06e59fd9487", true}, ProfileImage: models.NullString{"pi0", true}, Nickname: models.NullString{"commenttest0", true}, UUID: "abc1d5b1-da54-4200-b90e-f06e59fd9487"},
		models.Member{ID: 91, MemberID: "commenttest1@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"commenttest1@mirrormedia.mg", true}, Points: models.NullInt{0, true}, TalkID: models.NullString{"abc1d5b1-da54-4200-b91e-f06e59fd9487", true}, ProfileImage: models.NullString{"pi1", true}, Nickname: models.NullString{"commenttest1", true}, UUID: "abc1d5b1-da54-4200-b91e-f06e59fd9487"},
		models.Member{ID: 92, MemberID: "commenttest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"commenttest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}, TalkID: models.NullString{"abc1d5b1-da54-4200-b92e-f06e59fd9487", true}, ProfileImage: models.NullString{"pi2", true}, Nickname: models.NullString{"commenttest2", true}, UUID: "abc1d5b1-da54-4200-b92e-f06e59fd9487"},
	} {
		_, err := models.MemberAPI.InsertMember(params)
		if err != nil {
			log.Printf("Insert member fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Post{
		models.Post{ID: 90, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{90, true}, PublishStatus: models.NullInt{2, true}},
		models.Post{ID: 91, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{91, true}, PublishStatus: models.NullInt{2, true}},
		models.Post{ID: 92, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{92, true}, PublishStatus: models.NullInt{2, true}},
	} {
		_, err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Printf("Insert post fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Project{
		models.Project{ID: 920, PostID: 91, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
		models.Project{ID: 921, PostID: 92, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2016, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
	} {
		err := models.ProjectAPI.InsertProject(params)
		if err != nil {
			log.Printf("Insert Project fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.FollowArgs{
		models.FollowArgs{Resource: "member", Subject: 91, Object: 90},
		models.FollowArgs{Resource: "post", Subject: 91, Object: 92},
		models.FollowArgs{Resource: "project", Subject: 91, Object: 920},
	} {
		err := models.FollowingAPI.AddFollowing(params)
		if err != nil {
			log.Printf("Init test case fail. Error: %v", err)
		}
	}

	log.Println("init finished")

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		log.Println("ok")
	}
	if os.Getenv("db_driver") == "mysql" {
		t.Run("Comments", func(t *testing.T) {
			for _, testcase := range []genericTestcase{
				genericTestcase{"comment_reply_author", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"comment_reply_author","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b91e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"comment_reply", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"comment_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b92e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"comment_comment", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"comment_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b92e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_member_reply", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"follow_member_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b92e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_post_reply", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"follow_post_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b90e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_project_reply", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"follow_project_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b90e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_memo_reply", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"follow_memo_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b90e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"post_reply", "POST", "/comments", `{"updated_at":"2046-01-05T00:42:42+00:00","created_at":"2046-01-05T00:42:42+00:00","body":"post_reply","asset_id":"post1","author_id":"abc1d5b1-da54-4200-b90e-f06e59fd9487","reply_count":0,"status":"NONE","id":"id","vidible":true}`, http.StatusOK, ``},
			} {
				genericDoTest(testcase, t, asserter)
			}
		})
	}

	for _, params := range []models.FollowArgs{
		models.FollowArgs{Resource: "member", Subject: 91, Object: 90},
		models.FollowArgs{Resource: "post", Subject: 91, Object: 92},
		models.FollowArgs{Resource: "project", Subject: 91, Object: 920},
	} {
		err := models.FollowingAPI.DeleteFollowing(params)
		if err != nil {
			log.Printf("Init test case fail. Error: %v", err)
		}
	}

}
