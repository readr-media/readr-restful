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
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					var Response struct {
						Items []struct {
							Item struct {
								ID int `json:"id"`
							} `json:"item"`
						} `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Fatalf("%s, Unexpected result body: %v", resp)
						return
					}
					var expected []routes.PubsubFollowMsgBody = tc.misc[0].([]routes.PubsubFollowMsgBody)
					for i, r := range Response.Items {
						assertIntHelper(t, tc.name, "following resource id", int(expected[i].Object), int(r.Item.ID))
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
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
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					var Response struct {
						Items []models.FollowedCount `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Fatalf("%s, Unexpected result body: %v", resp)
						return
					}
					var expected []models.FollowedCount = tc.misc[0].([]models.FollowedCount)
					for i, r := range Response.Items {
						assertIntHelper(t, tc.name, "followed resource id", int(expected[i].ResourceID), int(r.ResourceID))
						assertIntHelper(t, tc.name, "followed count", int(expected[i].Count), int(r.Count))
						for j, follower := range r.Followers {
							assertIntHelper(t, tc.name, "follower id", int(expected[i].Followers[j]), int(follower))
						}
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
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
					rawResult, err := models.FollowingAPI.Get(&models.GetFollowingArgs{
						MemberID:  2,
						Mode:      "id",
						TargetIDs: []int{tc.misc[1].(int)},
						Resources: []string{tc.misc[0].(string)},
					})
					if err != nil {
						t.Fatalf(fmt.Sprintf("Get Following error when testing %s", tc.name))
					}
					result := rawResult.([]int)
					assertIntHelper(t, tc.name, "result length", 1, len(result))
					assertIntHelper(t, tc.name, "following id", tc.misc[1].(int), result[0])
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
					rawResult, err := models.FollowingAPI.Get(&models.GetFollowingArgs{
						MemberID:  1,
						Mode:      "id",
						TargetIDs: []int{tc.misc[1].(int)},
						Resources: []string{tc.misc[0].(string)},
					})
					if err != nil {
						t.Fatalf(fmt.Sprintf("Get Following error when testing %s", tc.name))
					}
					result := rawResult.([]int)
					assertIntHelper(t, tc.name, "result length", 0, len(result))
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("InsertEmotion", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"LikePostOK", "insert", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"like"}`, http.StatusOK, ``, []interface{}{2, 2, []int64{1}, 1}}, //resourceType, memberID, objectID, emotion
			genericRequestTestcase{"DislikePostOK", "insert", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"dislike"}`, http.StatusOK, ``, []interface{}{2, 2, []int64{1}, 2}},
			genericRequestTestcase{"LikeMemberFail", "insert", `/restful/pubsub`, `{"resource":"member","subject":2,"object":4}`, http.StatusOK, `{"Error":"Emotion Not Available For Member"}`, nil},
			genericRequestTestcase{"UnknownEmotion", "insert", `/restful/pubsub`, `{"resource":"post","subject":2,"object":4,"emotion":"unknown"}`, http.StatusOK, `{"Error":"Unsupported Emotion"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "emotion", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					rawResult, err := models.FollowingAPI.Get(&models.GetFollowedArgs{
						tc.misc[2].([]int64),
						models.Resource{
							Emotion:    tc.misc[3].(int),
							FollowType: tc.misc[0].(int),
						},
					})
					if err != nil {
						t.Fatalf(fmt.Sprintf("Get Following error when testing %s", tc.name))
					}
					result := rawResult.([]models.FollowedCount)
					assertIntHelper(t, tc.name, "result length", 1, len(result))
					assertIntHelper(t, tc.name, "emotion maker", tc.misc[1].(int), int(result[0].Followers[0]))
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("UpdateEmotion", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateEmotionOK", "update", `/restful/pubsub`, `{"resource":"post","subject":1,"object":2,"emotion":"dislike"}`, http.StatusOK, ``, []interface{}{2, 1, []int64{2}, 2}}, //resourceType, memberID, objectID, emotion
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "emotion", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					rawResult, err := models.FollowingAPI.Get(&models.GetFollowedArgs{
						tc.misc[2].([]int64),
						models.Resource{
							Emotion:    tc.misc[3].(int),
							FollowType: tc.misc[0].(int),
						},
					})
					if err != nil {
						t.Fatalf(fmt.Sprintf("Get Following error when testing %s", tc.name))
					}
					result := rawResult.([]models.FollowedCount)
					assertIntHelper(t, tc.name, "result length", 1, len(result))
					assertIntHelper(t, tc.name, "emotion maker", tc.misc[1].(int), int(result[0].Followers[0]))
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("DeleteEmotion", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteLikePostOK", "delete", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"like"}`, http.StatusOK, ``, []interface{}{2, []int64{1}, 1}}, //resourceType, memberID, objectID, emotion
			genericRequestTestcase{"DeleteDislikePostOK", "delete", `/restful/pubsub`, `{"resource":"post","subject":2,"object":1,"emotion":"dislike"}`, http.StatusOK, ``, []interface{}{2, []int64{1}, 2}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tctransed := transformCommentPubsubMsg(tc.name, "emotion", tc.method, []byte(tc.body.(string)))
				code, resp := genericDoRequest(tctransed, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) && tc.resp == "" {
					rawResult, err := models.FollowingAPI.Get(&models.GetFollowedArgs{
						tc.misc[1].([]int64),
						models.Resource{
							Emotion:    tc.misc[2].(int),
							FollowType: tc.misc[0].(int),
						},
					})
					if err != nil {
						t.Fatalf(fmt.Sprintf("Get Following error when testing %s", tc.name))
					}
					result := rawResult.([]models.FollowedCount)
					assertIntHelper(t, tc.name, "result length", 0, len(result))
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
}
