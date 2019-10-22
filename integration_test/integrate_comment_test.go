//+build integration

package main

import (
	"log"
	"testing"
	"time"

	"encoding/json"
	"net/http"

	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
)

func TestComment(t *testing.T) {

	gd := Golden{}
	gd.SetUpdate(*update)

	var mockedComments = []models.InsertCommentArgs{
		models.InsertCommentArgs{
			Author:       models.NullInt{1, true},
			Body:         models.NullString{"comment body 01", true},
			ParentID:     models.NullInt{0, false},
			Resource:     models.NullString{"http://dev.readr.tw/post/1", true},
			Status:       models.NullInt{1, true},
			Active:       models.NullInt{1, true},
			ResourceName: models.NullString{"post", true},
			ResourceID:   models.NullInt{1, true},
			UpdatedAt:    models.NullTime{Time: time.Date(2018, 1, 2, 3, 4, 5, 6, time.UTC), Valid: true},
		},
		models.InsertCommentArgs{
			Author:       models.NullInt{1, true},
			Body:         models.NullString{"comment body 02", true},
			ParentID:     models.NullInt{0, false},
			Resource:     models.NullString{"http://dev.readr.tw/project/report_slug_1", true},
			Status:       models.NullInt{1, true},
			Active:       models.NullInt{1, true},
			ResourceName: models.NullString{"report", true},
			ResourceID:   models.NullInt{2, true},
			UpdatedAt:    models.NullTime{Time: time.Date(2018, 1, 3, 3, 4, 5, 6, time.UTC), Valid: true},
		},
		models.InsertCommentArgs{
			Author:       models.NullInt{2, true},
			Body:         models.NullString{"comment body 03", true},
			ParentID:     models.NullInt{0, false},
			Resource:     models.NullString{"http://dev.readr.tw/series/project_slug_2/3", true},
			Status:       models.NullInt{0, true},
			Active:       models.NullInt{0, true},
			ResourceName: models.NullString{"memo", true},
			ResourceID:   models.NullInt{3, true},
			UpdatedAt:    models.NullTime{Time: time.Date(2018, 1, 4, 3, 4, 5, 6, time.UTC), Valid: true},
		},
		models.InsertCommentArgs{
			Author:       models.NullInt{2, true},
			Body:         models.NullString{"comment body 04 child of 01", true},
			ParentID:     models.NullInt{1, true},
			Resource:     models.NullString{"http://dev.readr.tw/post/1", true},
			Status:       models.NullInt{1, true},
			Active:       models.NullInt{1, true},
			ResourceName: models.NullString{"post", true},
			ResourceID:   models.NullInt{1, true},
			UpdatedAt:    models.NullTime{Time: time.Date(2018, 1, 5, 3, 4, 5, 6, time.UTC), Valid: true},
		},
		models.InsertCommentArgs{
			Author:       models.NullInt{2, true},
			Body:         models.NullString{"comment body 05 child of 01", true},
			ParentID:     models.NullInt{1, true},
			Resource:     models.NullString{"http://dev.readr.tw/post/1", true},
			Status:       models.NullInt{0, true},
			Active:       models.NullInt{1, true},
			ResourceName: models.NullString{"post", true},
			ResourceID:   models.NullInt{1, true},
			UpdatedAt:    models.NullTime{Time: time.Date(2018, 1, 6, 3, 4, 5, 6, time.UTC), Valid: true},
		},
		models.InsertCommentArgs{
			Author:       models.NullInt{2, true},
			Body:         models.NullString{"comment body 06 child of 01", true},
			ParentID:     models.NullInt{1, true},
			Resource:     models.NullString{"http://dev.readr.tw/post/1", true},
			Status:       models.NullInt{1, true},
			Active:       models.NullInt{1, true},
			ResourceName: models.NullString{"post", true},
			ResourceID:   models.NullInt{1, true},
			UpdatedAt:    models.NullTime{Time: time.Date(2018, 1, 7, 3, 4, 5, 6, time.UTC), Valid: true},
			IP:           models.NullString{"1.2.3.4", true},
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
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test post 02", true},
			Content:       models.NullString{"Test post content 02", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://dev.readr.tw/posts/2", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{1, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 5, 3, 4, 5, 6, time.UTC), Valid: true},
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

	var mockedReportedComments = []models.ReportedComment{
		models.ReportedComment{
			CommentID: models.NullInt{2, true},
			Reporter:  models.NullInt{1, true},
			Reason:    models.NullString{"Reason1", true},
			Solved:    models.NullInt{0, true},
		},
		models.ReportedComment{
			CommentID: models.NullInt{3, true},
			Reporter:  models.NullInt{1, true},
			Reason:    models.NullString{"Reason2", true},
			Solved:    models.NullInt{0, true},
		},
	}

	transformCommentPubsubMsg := func(name string, method string, body []byte) genericRequestTestcase {
		meta := routes.PubsubMessageMeta{
			Subscription: "sub",
			Message: routes.PubsubMessageMetaBody{
				ID:   "1",
				Body: body,
				Attr: map[string]string{"type": "comment", "action": method},
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
		for _, v := range mockedComments {
			commentBody, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("InitComment Error when marshaling input parameters")
			}
			tc := transformCommentPubsubMsg("InitComment", "post", commentBody)
			genericDoRequest(tc, t)
		}
		for _, v := range mockedReportedComments {
			_, err := models.CommentAPI.InsertReportedComments(v)
			if err != nil {
				t.Fatalf("init reported comments data fail: %s ", err.Error())
			}
		}
		return flushDB
	}
	t.Run("GetSingleComment", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetCommentOK", "GET", "/comment/1", ``, http.StatusOK, ``, []interface{}{mockedComments[0]}},
			genericRequestTestcase{"GetCommentNotfound", "GET", "/comment/101", ``, http.StatusNotFound, `{"Error":"Comment Not Found"}`, nil},
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
	t.Run("GetComments", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetCommentOK", "GET", `/comment?status={"$in":[1]}&author=[1,2]&resource=["http://dev.readr.tw/post/1"]`, ``, http.StatusOK, ``, []interface{}{[]models.InsertCommentArgs{mockedComments[0]}}},
			genericRequestTestcase{"GetCommentMultipleResourceOK", "GET", `/comment?author=[1,2]&resource=["http://dev.readr.tw/post/1", "http://dev.readr.tw/project/report_slug_1"]`, ``, http.StatusOK, ``, []interface{}{[]models.InsertCommentArgs{mockedComments[1], mockedComments[0]}}},
			genericRequestTestcase{"GetChildCommentOK", "GET", `/comment?parent=[1]`, ``, http.StatusOK, ``, []interface{}{[]models.InsertCommentArgs{mockedComments[5], mockedComments[4], mockedComments[3]}}},
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
	t.Run("InsertComments", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"InsertCommentOK", `post`, `/comment`, mockedComments[5], http.StatusOK, ``, []interface{}{[]models.InsertCommentArgs{mockedComments[5]}}},
			genericRequestTestcase{"InsertCommentMissingRequired", "post", "/comment", `{"body":"ok","author":91}`, http.StatusOK, ``, []interface{}{[]models.InsertCommentArgs{mockedComments[5]}}},
			genericRequestTestcase{"InsertCommentWithCreatedAt", "post", "/comment", `{"body":"ok","resource":"http://dev.readr.tw/post/1","author":91,"created_at":"2046-01-05T00:42:42+00:00","resource_name":"post","resource_id":1}`, http.StatusOK, ``,
				[]interface{}{[]models.InsertCommentArgs{
					models.InsertCommentArgs{
						Resource: models.NullString{"http://dev.readr.tw/post/1", true},
						Body:     models.NullString{"ok", true},
						ParentID: models.NullInt{0, false},
					}}},
			},
			/*genericRequestTestcase{"InsertCommentWithUrl", "post", "/comment", `{"body":"https://developers.facebook.com/","resource":"http://dev.readr.tw/post/1","author":1,"resource_name":"post","resource_id":1}`, http.StatusOK, ``,
				[]interface{}{[]models.InsertCommentArgs{
					models.InsertCommentArgs{
						Resource:      models.NullString{"http://dev.readr.tw/post/1", true},
						Body:          models.NullString{`<a href="https://developers.facebook.com/" target="_blank">https://developers.facebook.com/</a>`, true},
						ParentID:      models.NullInt{0, false},
						OgTitle:       models.NullString{"Facebook Developers", true},
						OgDescription: models.NullString{"為 Facebook 用戶開發應用程式。深入探討人工智慧、商業工具、遊戲、開放原始碼、發佈、社交網站硬體、社交網站整合，以及虛擬實境。瞭解 Facebook 的全球開發人員教育訓練和交流計畫。", true},
					}}},
			},*/
		} {
			t.Run(tc.name, func(t *testing.T) {
				var body []byte
				if s, ok := tc.body.(string); ok {
					body = []byte(s)
				} else if commentBody, err := json.Marshal(tc.body); err == nil {
					body = commentBody
				} else {
					if err != nil {
						t.Fatalf("%s Error when marshaling input parameters", tc.name)
					}
				}
				code, resp := genericDoRequest(transformCommentPubsubMsg(tc.name, tc.method, body), t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "InsertCommentsVarification",
						method: "GET",
						url:    `/comment?sort=-id&max_result=1`}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("UpdateComment", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateCommentOK", `put`, `/comment`, `{"id":1, "body":"ok"}`, http.StatusOK, ``, []interface{}{[]string{"ok"}}},
			genericRequestTestcase{"UpdateCommentWithResource", `put`, `/comment`, `{"id":1, "resource":"http://google.com"}`, http.StatusOK, `{"Error":"Invalid Parameters"}`, []interface{}{[]string{"ok"}}},
			genericRequestTestcase{"UpdateCommentWithAuthor", `put`, `/comment`, `{"id":1, "author":2}`, http.StatusOK, `{"Error":"Invalid Parameters"}`, []interface{}{[]string{"ok"}}},
			genericRequestTestcase{"UpdateCommentHideComment", `put`, `/comment`, `{"id":1, "hide":0}`, http.StatusOK, ``, []interface{}{[]string{"ok"}}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(transformCommentPubsubMsg(tc.name, tc.method, []byte(tc.body.(string))), t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				if tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "UpdateCommentVarification",
						method: "GET",
						url:    `/comment?resource=["http://dev.readr.tw/post/1"]`}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				}
			})
		}
	})
	t.Run("UpdateComments", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateCommentOK", "putstatus", "/comment/status", `{"ids":[4,5], "status":0}`, http.StatusOK, ``, nil},
			genericRequestTestcase{"UpdateCommentNoIDs", "putstatus", "/comment/status", `{"status":0}`, http.StatusOK, `{"Error":"ID List Empty"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(transformCommentPubsubMsg(tc.name, tc.method, []byte(tc.body.(string))), t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				if tc.resp == "" {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "UpdateCommentsVarification",
						method: "GET",
						url:    `/comment?parent=[1]&status={"$nin":[0]}`}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("DeleteComments", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteCommentOK", "delete", "/comment", `{"ids":[1,2]}`, http.StatusOK, ``, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(transformCommentPubsubMsg(tc.name, tc.method, []byte(tc.body.(string))), t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				if statusCodeOKHelper(code) {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "DeleteCommentsVarification",
						method: "GET",
						url:    `/comment?resource=["http://dev.readr.tw/post/1","http://dev.readr.tw/project/report_slug_1"]`}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("InsertReport", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"InsertReportOK", "POST", "/reported_comment", `{"comment_id":1, "reporter":1}`, http.StatusOK, ``, []interface{}{[]models.ReportedComment{models.ReportedComment{
				CommentID: models.NullInt{1, true},
				Reporter:  models.NullInt{1, true},
			}}}},
			genericRequestTestcase{"InsertReportMissingCommentID", "POST", "/reported_comment", `{"reporter":91}`, http.StatusBadRequest, `{"Error":"Missing Required Parameters."}`, nil},
			genericRequestTestcase{"InsertReportMissingReporter", "POST", "/reported_comment", `{"comment_id":1}`, http.StatusBadRequest, `{"Error":"Missing Required Parameters."}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "InsertReportVarification",
						method: "GET",
						url:    `/reported_comment?sort=-id&max_result=1`}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("UpdateReport", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateReportOK", "PUT", "/reported_comment", `{"id":1, "solved":1}`, http.StatusOK, ``, []interface{}{[]models.ReportedComment{mockedReportedComments[0]}}},
			genericRequestTestcase{"UpdateReportMissingID", "PUT", "/reported_comment", `{"solved":1}`, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`, nil},
			genericRequestTestcase{"UpdateReporterFail", "PUT", "/reported_comment", `{"id":1, "reporter":90}`, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					vCode, vResp := genericDoRequestByte(genericRequestTestcase{
						name:   "UpdateReportVarification",
						method: "GET",
						url:    `/reported_comment?solved={"$in":[1]}&max_result=1`}, t)
					assertIntHelper(t, tc.name, "verify request status code", http.StatusOK, vCode)
					gd.AssertOrUpdate(t, vResp)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
}
