package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/readr-media/readr-restful/models"
)

type mockCommentAPI struct{}

func (c *mockCommentAPI) GetComments(args *models.GetCommentArgs) (result []models.CommentAuthor, err error) {

	var mockCommentResult = []models.CommentAuthor{
		models.CommentAuthor{models.Comment{ID: 1, Body: models.NullString{"Comment No.1", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{91, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}}, models.NullString{"commenttest1", true}, models.NullString{"", false}, models.NullInt{0, false}, models.NullInt{0, false}},
		models.CommentAuthor{models.Comment{ID: 2, Body: models.NullString{"Comment No.2", true}, Resource: models.NullString{"http://dev.readr.tw/post/91", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}}, models.NullString{"commenttest2", true}, models.NullString{"", true}, models.NullInt{0, false}, models.NullInt{0, false}},
		models.CommentAuthor{models.Comment{ID: 3, Body: models.NullString{"Comment No.3", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}, Status: models.NullInt{int64(models.CommentStatus["hide"].(float64)), true}}, models.NullString{"commenttest2", true}, models.NullString{"", true}, models.NullInt{0, false}, models.NullInt{0, false}},
	}

	switch len(args.Author) {
	case 3:
		return []models.CommentAuthor{mockCommentResult[0], mockCommentResult[2]}, nil
	case 2:
		return []models.CommentAuthor{mockCommentResult[0]}, nil
	case 1:
		return []models.CommentAuthor{mockCommentResult[2]}, nil
	}
	return result, err
}

func (c *mockCommentAPI) GetComment(id int) (comment models.CommentAuthor, err error) {
	if id == 1 {
		return models.CommentAuthor{models.Comment{ID: 1, Body: models.NullString{"Comment No.1", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{91, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}}, models.NullString{"commenttest1", true}, models.NullString{"pi1", true}, models.NullInt{2, true}, models.NullInt{0, true}}, nil
	} else {
		return comment, errors.New("Comment Not Found")
	}
}
func (c *mockCommentAPI) InsertComment(comment models.Comment) (id int64, err error) { return id, err }
func (c *mockCommentAPI) UpdateComment(comment models.Comment) (err error)           { return err }
func (c *mockCommentAPI) UpdateComments(req models.CommentUpdateArgs) (err error)    { return err }

func (c *mockCommentAPI) GetReportedComments(args *models.GetReportedCommentArgs) (result []models.ReportedCommentAuthor, err error) {

	var mockCommentResult = []models.CommentAuthor{
		models.CommentAuthor{models.Comment{ID: 1, Body: models.NullString{"Comment No.1", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{91, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}}, models.NullString{"commenttest1", true}, models.NullString{"", false}, models.NullInt{0, false}, models.NullInt{0, false}},
		models.CommentAuthor{models.Comment{ID: 2, Body: models.NullString{"Comment No.2", true}, Resource: models.NullString{"http://dev.readr.tw/post/91", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}, IP: models.NullString{"5.6.7.8", true}}, models.NullString{"commenttest2", true}, models.NullString{"pi2", true}, models.NullInt{3, true}, models.NullInt{0, true}},
		models.CommentAuthor{models.Comment{ID: 3, Body: models.NullString{"Comment No.3", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}, Status: models.NullInt{int64(models.CommentStatus["hide"].(float64)), true}}, models.NullString{"commenttest2", true}, models.NullString{"", true}, models.NullInt{0, false}, models.NullInt{0, false}},
	}

	var mockReports = []models.ReportedComment{
		models.ReportedComment{ID: 1, CommentID: 2, Reporter: models.NullInt{92, true}, IP: models.NullString{"1.2.3.4", true}},
		models.ReportedComment{ID: 2, CommentID: 2, Reporter: models.NullInt{90, true}},
	}

	switch len(args.Reporter) {
	case 1:

		result = append(result, models.ReportedCommentAuthor{Comment: mockCommentResult[1], Report: mockReports[1]})
		return result, err
	case 0:
		result = append(result, models.ReportedCommentAuthor{Comment: mockCommentResult[1], Report: mockReports[0]})
		result = append(result, models.ReportedCommentAuthor{Comment: mockCommentResult[1], Report: mockReports[1]})
		//result = append(result, models.ReportedCommentAuthor{models.CommentAuthor{models.Comment{ID: 2, Author: 92, Body: models.NullString{"Comment No.2", true}, Resource: models.NullString{"http://dev.readr.tw/post/91", true}, Active: models.NullInt{1, true}}, "commenttest2", "pi2", 3, 0}, models.NullString{"", false}, 0, 92, models.NullInt{0, false}})
		//result = append(result, models.ReportedCommentAuthor{models.CommentAuthor{models.Comment{ID: 2, Author: 92, Body: models.NullString{"Comment No.2", true}, Resource: models.NullString{"http://dev.readr.tw/post/91", true}, Active: models.NullInt{1, true}}, "commenttest2", "pi2", 3, 0}, models.NullString{"", false}, 0, 90, models.NullInt{0, false}})
		return result, err
	}
	return result, err
}
func (c *mockCommentAPI) InsertReportedComments(report models.ReportedComment) (id int64, err error) {
	return id, err
}
func (c *mockCommentAPI) UpdateReportedComments(report models.ReportedComment) (err error) {
	return err
}
func (c *mockCommentAPI) UpdateCommentAmountByResource(resource string, action string) (err error) {
	return err
}
func (c *mockCommentAPI) UpdateCommentAmountByIDs(ids []int) (err error) {
	return err
}

func TestRouteComments(t *testing.T) {
	log.Println("test start")

	var mockComments = []models.Comment{
		models.Comment{ID: 1, Body: models.NullString{"Comment No.1", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{91, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}},
		models.Comment{ID: 2, Body: models.NullString{"Comment No.2", true}, Resource: models.NullString{"http://dev.readr.tw/post/91", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}, IP: models.NullString{"5.6.7.8", true}},
		models.Comment{ID: 3, Body: models.NullString{"Comment No.3", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}, Status: models.NullInt{int64(models.CommentStatus["hide"].(float64)), true}},
	}

	var mockCommentResult = []models.CommentAuthor{
		models.CommentAuthor{models.Comment{ID: 1, Body: models.NullString{"Comment No.1", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{91, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}}, models.NullString{"commenttest1", true}, models.NullString{"", false}, models.NullInt{0, false}, models.NullInt{0, false}},
		models.CommentAuthor{models.Comment{ID: 2, Body: models.NullString{"Comment No.2", true}, Resource: models.NullString{"http://dev.readr.tw/post/91", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}}, models.NullString{"commenttest2", true}, models.NullString{"", false}, models.NullInt{0, false}, models.NullInt{0, false}},
		models.CommentAuthor{models.Comment{ID: 3, Body: models.NullString{"Comment No.3", true}, Resource: models.NullString{"http://dev.readr.tw/post/90", true}, Author: models.NullInt{92, true}, Active: models.NullInt{int64(models.CommentActive["active"].(float64)), true}, Status: models.NullInt{int64(models.CommentStatus["hide"].(float64)), true}}, models.NullString{"commenttest2", true}, models.NullString{"", false}, models.NullInt{0, false}, models.NullInt{0, false}},
	}

	var mockReports = []models.ReportedComment{
		models.ReportedComment{ID: 1, CommentID: 2, Reporter: models.NullInt{92, true}, IP: models.NullString{"1.2.3.4", true}},
		models.ReportedComment{ID: 2, CommentID: 2, Reporter: models.NullInt{90, true}},
	}

	for _, params := range []models.Member{
		models.Member{ID: 90, MemberID: "commenttest0@mirrormedia.mg", Active: models.NullInt{1, true}, Role: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"commenttest0@mirrormedia.mg", true}, Points: models.NullInt{0, true}, TalkID: models.NullString{"abc1d5b1-da54-4200-b90e-f06e59fd9487", true}, ProfileImage: models.NullString{"pi0", true}, Nickname: models.NullString{"commenttest0", true}, UUID: "abc1d5b1-da54-4200-b90e-f06e59fd9487"},
		models.Member{ID: 91, MemberID: "commenttest1@mirrormedia.mg", Active: models.NullInt{1, true}, Role: models.NullInt{2, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"commenttest1@mirrormedia.mg", true}, Points: models.NullInt{0, true}, TalkID: models.NullString{"abc1d5b1-da54-4200-b91e-f06e59fd9487", true}, ProfileImage: models.NullString{"pi1", true}, Nickname: models.NullString{"commenttest1", true}, UUID: "abc1d5b1-da54-4200-b91e-f06e59fd9487"},
		models.Member{ID: 92, MemberID: "commenttest2@mirrormedia.mg", Active: models.NullInt{1, true}, Role: models.NullInt{3, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"commenttest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}, TalkID: models.NullString{"abc1d5b1-da54-4200-b92e-f06e59fd9487", true}, ProfileImage: models.NullString{"pi2", true}, Nickname: models.NullString{"commenttest2", true}, UUID: "abc1d5b1-da54-4200-b92e-f06e59fd9487"},
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

	for _, params := range mockComments {
		_, err := models.CommentAPI.InsertComment(params)
		if err != nil {
			log.Printf("Init test case fail. Error: %v", err)
		}
	}

	for _, params := range mockReports {
		_, err := models.CommentAPI.InsertReportedComments(params)
		if err != nil {
			log.Printf("Init test case fail. Error: %v", err)
		}
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.CommentAuthor `json:"_items"`
		}

		var Response response
		var expected []models.CommentAuthor = tc.resp.([]models.CommentAuthor)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", resp)
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect member length to be %v but get %v", tc.name, len(expected), len(Response.Items))
			return
		}

		for i, resp := range Response.Items {
			exp := expected[i]
			if resp.ID == exp.ID &&
				resp.AuthorNickname == exp.AuthorNickname &&
				resp.Body == exp.Body &&
				resp.Resource == exp.Resource &&
				resp.Status == exp.Status &&
				resp.Active == exp.Active {
				continue
			}
			t.Errorf("%s, expect to get %v, but %v ", tc.name, exp, resp)
		}
	}

	transformPubsub := func(tc genericTestcase) genericTestcase {
		meta := PubsubMessageMeta{
			Subscription: "sub",
			Message: PubsubMessageMetaBody{
				ID:   "1",
				Body: []byte(tc.body.(string)),
				Attr: map[string]string{"type": "comment", "action": tc.method},
			},
		}

		return genericTestcase{tc.name, "POST", "/restful/pubsub", meta, tc.httpcode, tc.resp}
	}

	t.Run("GetComment", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetCommentOK", "GET", `/comment?author=[90,91,92]&resource=["http://dev.readr.tw/post/90"]&sort=-updated_at`, ``, http.StatusOK, []models.CommentAuthor{mockCommentResult[0], mockCommentResult[2]}},
			genericTestcase{"GetCommentMultipleResourceOK", "GET", `/comment?author=[90,91]&resource=["http://dev.readr.tw/post/90", "http://dev.readr.tw/post/91"]&sort=-updated_at`, ``, http.StatusOK, []models.CommentAuthor{mockCommentResult[0]}},
			genericTestcase{"GetCommentFilterStatusOK", "GET", `/comment?author=[92]&status={"$in":[0]}`, ``, http.StatusOK, []models.CommentAuthor{mockCommentResult[2]}},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetSingleComment", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetCommentOK", "GET", "/comment/1", ``, http.StatusOK, `{"_items":{"id":1,"author":91,"body":"Comment No.1","og_title":null,"og_description":null,"og_image":null,"like_amount":null,"parent_id":null,"resource":"http://dev.readr.tw/post/90","status":null,"active":1,"updated_at":null,"created_at":null,"ip":null,"author_nickname":"commenttest1","author_image":"pi1","author_role":2,"comment_amount":0}}`},
			genericTestcase{"GetCommentNotfound", "GET", "/comment/101", ``, http.StatusNotFound, `{"Error":"Comment Not Found"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetReport", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetReportOK", "GET", "/reported_comment", ``, http.StatusOK, `{"_items":[{"comments":{"id":2,"author":92,"body":"Comment No.2","og_title":null,"og_description":null,"og_image":null,"like_amount":null,"parent_id":null,"resource":"http://dev.readr.tw/post/91","status":null,"active":1,"updated_at":null,"created_at":null,"ip":"5.6.7.8","author_nickname":"commenttest2","author_image":"pi2","author_role":3,"comment_amount":0},"reported":{"id":1,"comment_id":2,"reporter":92,"reason":null,"solved":null,"updated_at":null,"created_at":null,"ip":"1.2.3.4"}},{"comments":{"id":2,"author":92,"body":"Comment No.2","og_title":null,"og_description":null,"og_image":null,"like_amount":null,"parent_id":null,"resource":"http://dev.readr.tw/post/91","status":null,"active":1,"updated_at":null,"created_at":null,"ip":"5.6.7.8","author_nickname":"commenttest2","author_image":"pi2","author_role":3,"comment_amount":0},"reported":{"id":2,"comment_id":2,"reporter":90,"reason":null,"solved":null,"updated_at":null,"created_at":null,"ip":null}}]}`},
			genericTestcase{"GetReportOK", "GET", "/reported_comment?reporter=[90]", ``, http.StatusOK, `{"_items":[{"comments":{"id":2,"author":92,"body":"Comment No.2","og_title":null,"og_description":null,"og_image":null,"like_amount":null,"parent_id":null,"resource":"http://dev.readr.tw/post/91","status":null,"active":1,"updated_at":null,"created_at":null,"ip":"5.6.7.8","author_nickname":"commenttest2","author_image":"pi2","author_role":3,"comment_amount":0},"reported":{"id":2,"comment_id":2,"reporter":90,"reason":null,"solved":null,"updated_at":null,"created_at":null,"ip":null}}]}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("InsertComment", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"InsertCommentOK", "post", "/comment", `{"body":"成功","resource":"http://dev.readr.tw/post/90","author":91}`, http.StatusOK, ``},
			genericTestcase{"InsertCommentWithIPOK", "post", "/comment", `{"body":"成功2","resource":"http://dev.readr.tw/post/92","author":92,"ip":"1.2.3.4"}`, http.StatusOK, ``},
			genericTestcase{"InsertCommentMissingRequired", "post", "/comment", `{"body":"成功","author":91}`, http.StatusOK, `{"Error":"Missing Required Parameters"}`},
			genericTestcase{"InsertCommentWithCreatedAt", "post", "/comment", `{"body":"成功，created_at 被無視","resource":"http://dev.readr.tw/post/90","author":91,"created_at":"2046-01-05T00:42:42+00:00"}`, http.StatusOK, ``},
		} {
			genericDoTest(transformPubsub(testcase), t, asserter)
		}
	})
	t.Run("InsertCommentWithUrl", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"InsertCommentWithUrlOK", "post", "/comment", `{"body":"https://developers.facebook.com/","resource":"http://dev.readr.tw/post/90","author":91}`, http.StatusOK, ``},
			//genericTestcase{"InsertCommentWithUnicodeOK", "post", "/comment", `{"body":"https://medium.com/@evonneyifangtsai/短評xdite參選台北市長-84b391b3bfae","resource":"http://dev.readr.tw/post/90","author":91}`, http.StatusOK, ``},
			genericTestcase{"InsertCommentWithMultipleUrlOK", "post", "/comment", `{"body":"https://www.readr.tw/post/274 http://news.ltn.com.tw/news/focus/paper/1191781","resource":"http://dev.readr.tw/post/90","author":91}`, http.StatusOK, ``},
			//genericTestcase{"PutCommentWithUrlOK", "put", "/comment", `{"id": 1, "body":"https://medium.com/@evonneyifangtsai/"}`, http.StatusOK, ``},
		} {
			genericDoTest(transformPubsub(testcase), t, asserter)
		}
	})
	t.Run("UpdateComment", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"UpdateCommentOK", "put", "/comment", `{"id":1, "body":"modified"}`, http.StatusOK, ``},
			genericTestcase{"UpdateCommentMissingID", "put", "/comment", `{"solved":1}`, http.StatusOK, `{"Error":"Invalid Parameters"}`},
			genericTestcase{"UpdateAuthorFail", "put", "/comment", `{"id":1, "author":90}`, http.StatusOK, `{"Error":"Invalid Parameters"}`},
		} {
			genericDoTest(transformPubsub(testcase), t, asserter)
		}
	})
	t.Run("UpdateComments", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"UpdateCommentOK", "putstatus", "/comment/status", `{"ids":[1,2,3], "status":0}`, http.StatusOK, ``},
			genericTestcase{"UpdateCommentNoIDs", "putstatus", "/comment/status", `{"status":0}`, http.StatusOK, `{"Error":"ID List Empty"}`},
		} {
			genericDoTest(transformPubsub(testcase), t, asserter)
		}
	})
	t.Run("DeleteComments", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"DeleteCommentOK", "delete", "/comment", `{"ids":[1,2]}`, http.StatusOK, ``},
		} {
			genericDoTest(transformPubsub(testcase), t, asserter)
		}
	})
	t.Run("InsertReport", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"InsertReportOK", "POST", "/reported_comment", `{"comment_id":1, "reporter":91}`, http.StatusOK, ``},
			genericTestcase{"InsertReportMissingCommentID", "POST", "/reported_comment", `{"reporter":91}`, http.StatusBadRequest, `{"Error":"Missing Required Parameters."}`},
			genericTestcase{"InsertReportMissingReporter", "POST", "/reported_comment", `{"comment_id":1}`, http.StatusBadRequest, `{"Error":"Missing Required Parameters."}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("UpdateReport", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"UpdateReportOK", "PUT", "/reported_comment", `{"id":1, "solved":1}`, http.StatusOK, ``},
			genericTestcase{"UpdateReportMissingID", "PUT", "/reported_comment", `{"solved":1}`, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`},
			genericTestcase{"UpdateReporterFail", "PUT", "/reported_comment", `{"id":1, "reporter":90}`, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})

	log.Println("init finished")

}

func TestPubsubComments(t *testing.T) {

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

	for _, memo := range []models.Memo{
		models.Memo{ID: 92, Title: models.NullString{"CommentTestDefault1", true}, Author: models.NullInt{92, true}, Project: models.NullInt{920, true}, Active: models.NullInt{1, true}},
	} {
		err := models.MemoAPI.InsertMemo(memo)
		if err != nil {
			log.Printf("Init memo test fail %s", err.Error())
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

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		//log.Println("ok")
		return
	}

	transformPubsub := func(tc genericTestcase) genericTestcase {
		meta := PubsubMessageMeta{
			Subscription: "sub",
			Message: PubsubMessageMetaBody{
				ID:   "1",
				Body: []byte(tc.body.(string)),
				Attr: map[string]string{"type": "comment", "action": tc.method},
			},
		}

		return genericTestcase{tc.name, "POST", "/restful/pubsub", meta, tc.httpcode, tc.resp}
	}

	if os.Getenv("db_driver") == "mysql" {
		t.Run("Comments", func(t *testing.T) {
			for _, testcase := range []genericTestcase{

				genericTestcase{"post_reply", "post", "/comment", `{"body":"base","resource":"http://test.readr.tw/post/90","author":91,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"comment_reply_author", "post", "/comment", `{"body":"comment_reply_author","resource":"http://test.readr.tw/post/90","parent_id":1,"author":90,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"comment_reply", "post", "/comment", `{"body":"comment_reply","resource":"http://test.readr.tw/post/90","parent_id":1,"author":92,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"comment_comment", "post", "/comment", `{"body":"comment_reply","resource":"http://test.readr.tw/post/90","author":92,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_member_reply", "post", "/comment", `{"body":"follow_member_reply","resource":"http://test.readr.tw/post/90","author":92,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_post_reply", "post", "/comment", `{"body":"follow_post_reply","resource":"http://test.readr.tw/post/92","author":90,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_project_reply", "post", "/comment", `{"body":"follow_project_reply","resource":"http://test.readr.tw/project/920","author":90,"status":"NONE","vidible":true}`, http.StatusOK, ``},
				genericTestcase{"follow_memo_reply", "post", "/comment", `{"body":"follow_memo_reply","resource":"http://test.readr.tw/memo/92","author":90,"status":"NONE","vidible":true}`, http.StatusOK, ``},
			} {
				genericDoTest(transformPubsub(testcase), t, asserter)
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

	if os.Getenv("db_driver") == "mysql" {
		_, _ = models.DB.Exec("truncate table memos;")
	} else {
		mockMemoDS = []models.Memo{}
	}

}
