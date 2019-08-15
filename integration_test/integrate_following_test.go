//+build integration

package main

import (
	"fmt"
	"log"
	"testing"
	"time"

	"encoding/json"
	"net/http"

	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
)

func TestFollowing(t *testing.T) {

	gd := Golden{}
	gd.SetUpdate(*update)

	var mockedFollowings = []routes.PubsubFollowMsgBody{
		routes.PubsubFollowMsgBody{
			Resource: "member",
			Emotion:  "follow",
			Subject:  1, // member ID
			Object:   2, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "member",
			Emotion:  "follow",
			Subject:  2, // member ID
			Object:   1, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "post",
			Emotion:  "follow",
			Subject:  1, // member ID
			Object:   1, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "post",
			Emotion:  "follow",
			Subject:  1, // member ID
			Object:   2, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "post",
			Emotion:  "follow",
			Subject:  2, // member ID
			Object:   2, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "project",
			Emotion:  "follow",
			Subject:  1, // member ID
			Object:   2, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "project",
			Emotion:  "follow",
			Subject:  2, // member ID
			Object:   1, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "project",
			Emotion:  "follow",
			Subject:  2, // member ID
			Object:   2, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "post",
			Emotion:  "like",
			Subject:  1, // member ID
			Object:   2, // resource ID
		},
		routes.PubsubFollowMsgBody{
			Resource: "project",
			Emotion:  "like",
			Subject:  1, // member ID
			Object:   2, // resource ID
		},
	}

	var mockedPosts = []models.Post{
		models.Post{
			ID:            1,
			Author:        models.NullInt{2, true},
			Title:         models.NullString{"Test post 01", true},
			Content:       models.NullString{"<p>Test post content 01</p>", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://dev.readr.tw/post/1", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 2, 3, 4, 5, 6, time.UTC), Valid: true},
		},
		models.Post{
			ID:            2,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test report 01", true},
			Content:       models.NullString{"<p>Test report content 01</p>", true},
			Type:          models.NullInt{4, true},
			Link:          models.NullString{"http://dev.readr.tw/project/report_slug_1", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 3, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{1, true},
			Slug:          models.NullString{"report_slug_1", true},
		},
		models.Post{
			ID:            3,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test memo 01", true},
			Type:          models.NullInt{5, true},
			Link:          models.NullString{"http://dev.readr.tw/series/project_slug_2/3", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 4, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{2, true},
			Slug:          models.NullString{"project_slug_2", true},
		},
		models.Post{
			ID:            4,
			Author:        models.NullInt{2, true},
			Title:         models.NullString{"Test post 02", true},
			Content:       models.NullString{"<p>Test post content 02</p>", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://dev.readr.tw/post/2", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 2, 3, 4, 5, 6, time.UTC), Valid: true},
		},
	}
	var mockedMembers = []models.Member{
		models.Member{
			ID:       1,
			UUID:     "uuid1",
			Mail:     models.NullString{"testmember01@test.cc", true},
			Nickname: models.NullString{"test_member_01", true},
			MemberID: "testmember01@test.cc",
		}, models.Member{
			ID:       2,
			UUID:     "uuid2",
			Mail:     models.NullString{"testmember02@test.cc", true},
			Nickname: models.NullString{"test_member_02", true},
			MemberID: "testmember02@test.cc",
		}, models.Member{
			ID:       3,
			UUID:     "uuid3",
			Mail:     models.NullString{"testmember03@test.cc", true},
			Nickname: models.NullString{"test_member_03", true},
			MemberID: "testmember03@test.cc",
		},
	}
	var mockedProjects = []models.Project{
		models.Project{
			ID:            1,
			Slug:          models.NullString{"project_slug_1", true},
			Title:         models.NullString{"project_title_1", true},
			PublishStatus: models.NullInt{1, true},
		},
		models.Project{
			ID:            2,
			Slug:          models.NullString{"project_slug_2", true},
			Title:         models.NullString{"project_title_2", true},
			PublishStatus: models.NullInt{2, true},
		},
	}
	var mockedTags = []models.Tag{
		models.Tag{
			Text: "tag_1",
		},
	}

	transformCommentPubsubMsg := func(name string, msgType string, method string, body []byte) genericRequestTestcase {
		meta := routes.PubsubMessageMeta{
			Subscription: "sub",
			Message: routes.PubsubMessageMetaBody{
				ID:   "1",
				Body: body,
				Attr: map[string]string{"type": msgType, "action": method},
			},
		}
		return genericRequestTestcase{name: name, method: "POST", url: "/restful/pubsub", body: meta}
	}

	init := func() func() {
		for _, v := range mockedPosts {
			_, err := models.PostAPI.InsertPost(v)
			if err != nil {
				t.Fatalf("init post data fail: %s ", err.Error())
			}
			err = models.PostAPI.UpdateAuthors(v, []models.AuthorInput{models.AuthorInput{MemberID: v.Author, Type: models.NullInt{0, true}}})
			if err != nil {
				t.Fatalf("init post author fail: %s ", err.Error())
			}
		}
		for _, v := range mockedMembers {
			_, err := models.MemberAPI.InsertMember(v)
			if err != nil {
				log.Println(err.Error())
				t.Fatalf("init member data fail: %s ", err.Error())
			}
		}
		for _, v := range mockedProjects {
			err := models.ProjectAPI.InsertProject(v)
			if err != nil {
				t.Fatalf("init project data fail: %s ", err.Error())
			}
		}
		for _, v := range mockedTags {
			_, err := models.TagAPI.InsertTag(v)
			if err != nil {
				t.Fatalf("init tag data fail: %s ", err.Error())
			}
		}
		for _, v := range mockedFollowings {
			followBody, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("InitFollowing Error when marshaling input parameters")
			}
			msgType := "emotion"
			actionType := "insert"
			if v.Emotion == "follow" {
				msgType = v.Emotion
				actionType = "follow"
			}
			tc := transformCommentPubsubMsg("InitFollowing", msgType, actionType, followBody)
			genericDoRequest(tc, t)
		}
		return flushDB
	}
	t.Run("GetFollowings", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"FollowingPostOK", "GET", `/following/user?resource=post&id=1`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[3], mockedFollowings[2]}}},
			genericRequestTestcase{"FollowingPostReviewOK", "GET", `/following/user?resource=post&resource_type=review&id=1`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[2]}}},
			genericRequestTestcase{"FollowingPostNewsOK", "GET", `/following/user?resource=memo&id=1`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[3]}}},
			genericRequestTestcase{"FollowingProjectOK", "GET", `/following/user?resource=project&id=1`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[5]}}},
			genericRequestTestcase{"FollowingWithTargetIDsOK", "GET", `/following/user?resource=project&id=1&target_ids=[1,2,3]`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[5]}}},
			genericRequestTestcase{"FollowingWithModeIDOK", "GET", `/following/user?resource=project&id=1&mode=id`, ``, http.StatusOK, `{"_items":[2]}`, nil},
			genericRequestTestcase{"FollowingMultipleRes", "GET", `/following/user?resource=["post", "project"]&id=1`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[3], mockedFollowings[2], mockedFollowings[5]}}},
			genericRequestTestcase{"FollowingMaxresultPaging", "GET", `/following/user?resource=["post", "project"]&id=1&max_result=1&page=2`, ``, http.StatusOK, ``, []interface{}{[]routes.PubsubFollowMsgBody{mockedFollowings[3]}}},
			genericRequestTestcase{"FollowingBadID", "GET", `/following/user?resource=post&max_result=1`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`, nil},
			genericRequestTestcase{"FollowingBadType", "GET", `/following/user?resource=["post", "aaa"]&id=1`, ``, http.StatusBadRequest, `{"Error":"Bad Following Type"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequestByte(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					gd.AssertOrUpdate(t, resp)
				} else {
					assertByteHelper(t, tc.name, "request result", []byte(tc.resp), resp)
				}
			})
		}
	})
	t.Run("GetFollowed", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"FollowedPostOK", "GET", `/following/resource?resource=post&ids=[1,2]`, ``, http.StatusOK, ``, []interface{}{[]models.FollowedCount{
				models.FollowedCount{ResourceID: 1, Count: 1, Followers: []int64{1}},
				models.FollowedCount{ResourceID: 2, Count: 2, Followers: []int64{1, 2}},
			}}},
			genericRequestTestcase{"FollowedMemberOK", "GET", `/following/resource?resource=member&ids=[1,2]`, ``, http.StatusOK, ``, []interface{}{[]models.FollowedCount{
				models.FollowedCount{ResourceID: 1, Count: 1, Followers: []int64{2}},
				models.FollowedCount{ResourceID: 2, Count: 1, Followers: []int64{1}},
			}}},
			genericRequestTestcase{"FollowedProjectSingleOK", "GET", `/following/resource?resource=project&ids=[2]`, ``, http.StatusOK, ``, []interface{}{[]models.FollowedCount{
				models.FollowedCount{ResourceID: 2, Count: 2, Followers: []int64{1, 2}},
			}}},
			genericRequestTestcase{"FollowedProjectOK", "GET", `/following/resource?resource=project&ids=[1,2]`, ``, http.StatusOK, ``, []interface{}{[]models.FollowedCount{
				models.FollowedCount{ResourceID: 1, Count: 1, Followers: []int64{2}},
				models.FollowedCount{ResourceID: 2, Count: 2, Followers: []int64{1, 2}},
			}}},
			genericRequestTestcase{"FollowedMissingResource", "GET", `/following/resource?ids=[1,2]`, ``, http.StatusBadRequest, `{"Error":"Unsupported Resource"}`, nil},
			genericRequestTestcase{"FollowedMissingID", "GET", `/following/resource?resource=post&ids=[]`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`, nil},
			genericRequestTestcase{"FollowedPostNotExist", "GET", `/following/resource?resource=post&ids=[1000,1001]&resource_type=news`, ``, http.StatusOK, `{"_items":null}`, nil},
			genericRequestTestcase{"FollowedPostStringID", "GET", `/following/resource?resource=post&ids=[unintegerable]`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`, nil},
			genericRequestTestcase{"FollowedProjectStringID", "GET", `/following/resource?resource=project&ids=[unintegerable]`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`, nil},
			genericRequestTestcase{"FollowedProjectInvalidEmotion", "GET", `/following/resource?resource=project&ids=[1,2]&resource_type=review&emotion=angry`, ``, http.StatusBadRequest, `{"Error":"Unsupported Emotion"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequestByte(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					gd.AssertOrUpdate(t, resp)
				} else {
					assertByteHelper(t, tc.name, "request result", []byte(tc.resp), resp)
				}
			})
		}
	})
	t.Run("InsertFollow", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"FollowingPostOK", "follow", `/restful/pubsub`, `{"resource":"post","subject":2,"object":4}`, http.StatusOK, ``, []interface{}{"post", 4}},
			genericRequestTestcase{"FollowingTagOK", "follow", `/restful/pubsub`, `{"resource":"tag","subject":2,"object":1}`, http.StatusOK, ``, []interface{}{"tag", 1}},
			genericRequestTestcase{"FollowingMissingResource", "follow", `/restful/pubsub`, `{"resource":"","subject":2,"object":1}`, http.StatusOK, `{"Error":"Unsupported Resource"}`, nil},
			genericRequestTestcase{"FollowingMissingAction", "", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1}`, http.StatusOK, `{"Error":"Bad Request"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "follow", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "InsertFollowVarification",
						method: "GET",
						url:    fmt.Sprintf(`/following/user?id=2&mode=id&resource=%s`, tc.misc[0].(string))}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("UnFollow", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UnFollowPostOK", "unfollow", `/restful/pubsub`, `{"resource":"post","subject":1,"object":2}`, http.StatusOK, ``, []interface{}{"post", 2}},
			genericRequestTestcase{"UnFollowMemberOK", "unfollow", `/restful/pubsub`, `{"resource":"member","subject":1,"object":2}`, http.StatusOK, ``, []interface{}{"member", 2}},
			genericRequestTestcase{"UnFollowProjectOK", "unfollow", `/restful/pubsub`, `{"resource":"project","subject":1,"object":2}`, http.StatusOK, ``, []interface{}{"project", 2}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "follow", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "UnFollowVarification",
						method: "GET",
						url:    fmt.Sprintf(`/following/user?id=1&mode=id&resource=%s`, tc.misc[0].(string))}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("InsertEmotion", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"LikePostOK", "insert", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"like"}`, http.StatusOK, ``, []interface{}{"post", "like"}}, //resourceType, memberID, objectID, emotion
			genericRequestTestcase{"DislikePostOK", "insert", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"dislike"}`, http.StatusOK, ``, []interface{}{"post", "dislike"}},
			genericRequestTestcase{"LikeMemberFail", "insert", `/restful/pubsub`, `{"resource":"member","subject":2,"object":4}`, http.StatusOK, `{"Error":"Emotion Not Available For Member"}`, nil},
			genericRequestTestcase{"UnknownEmotion", "insert", `/restful/pubsub`, `{"resource":"post","subject":2,"object":4,"emotion":"unknown"}`, http.StatusOK, `{"Error":"Unsupported Emotion"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "emotion", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "InsertEmotionVarification",
						method: "GET",
						url: fmt.Sprintf(`/following/resource?resource=%s&ids=[1]&emotion=%s`,
							tc.misc[0].(string),
							tc.misc[1].(string),
						)}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("UpdateEmotion", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateEmotionOK", "update", `/restful/pubsub`, `{"resource":"post","subject":1,"object":2,"emotion":"dislike"}`, http.StatusOK, ``, []interface{}{"post", "dislike"}}, //resourceType, memberID, objectID, emotion
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "emotion", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "UpdateEmotionVarification",
						method: "GET",
						url: fmt.Sprintf(`/following/resource?resource=%s&ids=[2]&emotion=%s`,
							tc.misc[0].(string),
							tc.misc[1].(string),
						)}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("DeleteEmotion", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteLikePostOK", "delete", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"like"}`, http.StatusOK, ``, []interface{}{"post", "like"}}, //resourceType, memberID, objectID, emotion
			genericRequestTestcase{"DeleteDislikePostOK", "delete", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"dislike"}`, http.StatusOK, ``, []interface{}{"post", "dislike"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "emotion", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "UpdateEmotionVarification",
						method: "GET",
						url: fmt.Sprintf(`/following/resource?resource=%s&ids=[1]&emotion=%s`,
							tc.misc[0].(string),
							tc.misc[1].(string),
						)}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
}
