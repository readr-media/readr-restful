//+build integration

package main

import (
	"log"
	"testing"
	"time"

	"encoding/json"
	"net/http"

	"github.com/readr-media/readr-restful/models"
)

func TestMemo(t *testing.T) {

	var mockedPosts = []models.Post{
		models.Post{
			ID:            1,
			Author:        models.NullInt{2, true},
			Title:         models.NullString{"Test memo 01", true},
			Content:       models.NullString{"<p>Test memo content 01</p>", true},
			Type:          models.NullInt{5, true},
			Link:          models.NullString{"http://dev.readr.tw/series/project_slug_1/1", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{1, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 2, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{1, true},
			Slug:          models.NullString{"project_slug_1", true},
		},
		models.Post{
			ID:            2,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test memo 02", true},
			Content:       models.NullString{"<p>Test memo content 02</p>", true},
			Type:          models.NullInt{5, true},
			Link:          models.NullString{"http://dev.readr.tw/series/project_slug_2/2", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 3, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{2, true},
			Slug:          models.NullString{"project_slug_2", true},
		},
		models.Post{
			ID:            3,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test memo 03", true},
			Type:          models.NullInt{5, true},
			Link:          models.NullString{"http://dev.readr.tw/series/project_slug_2/2", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 4, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{1, true},
			Slug:          models.NullString{"project_slug_2", true},
		},
		models.Post{
			ID:            4,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test post 01", true},
			Content:       models.NullString{"Test post content 01", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://dev.readr.tw/posts/4", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
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
		return flushDB
	}
	t.Run("GetMemo", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetMemoOK", "GET", "/memo/1", ``, http.StatusOK, ``, []interface{}{mockedPosts[0]}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					var Response struct {
						Items []models.MemoDetail `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Errorf("%s, Unexpected result body: %v", resp)
					}

					var expected models.Post = tc.misc[0].(models.Post)
					assertIntHelper(t, tc.name, "memo ID", int(expected.ID), int(Response.Items[0].ID))
					assertIntHelper(t, tc.name, "memo project ID", int(expected.ProjectID.Int), int(Response.Items[0].Project.ID))
					assertStringHelper(t, tc.name, "memo content", expected.Content.String, Response.Items[0].Content.String)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("GetMemos", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetMemoDefaultOK", "GET", "/memos", ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[2], mockedPosts[1], mockedPosts[0]}}},
			genericRequestTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1", ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[2]}}},
			genericRequestTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1&page=2", ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[1]}}},
			genericRequestTestcase{"GetMemoSortMultipleOK", "GET", "/memos?sort=-author,id", ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[0], mockedPosts[1], mockedPosts[2]}}},
			genericRequestTestcase{"GetMemoSortInvalidOption", "GET", "/memos?sort=meow", ``, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`, nil},
			genericRequestTestcase{"GetMemoFilterAuthor", "GET", `/memos?author=[2]`, ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[0]}}},
			genericRequestTestcase{"GetMemoFilterProject", "GET", `/memos?project_id=[2]`, ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[1]}}},
			genericRequestTestcase{"GetMemoFilterMultipleCondition", "GET", `/memos?active={"$nin":[0]}&author=[1]&project_id=[1]`, ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[2]}}},
			genericRequestTestcase{"GetMemoWithSlug", "GET", `/memos?slugs=["project_slug_1"]`, ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[0]}}},
			genericRequestTestcase{"GetMemoWithMemoStatusAndProjectStatus", "GET", `/memos?project_publish_status={"$in":[1]}`, ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[2], mockedPosts[0]}}},
			genericRequestTestcase{"GetMemoWithKeyword", "GET", `/memos?keyword=memo 02`, ``, http.StatusOK, ``, []interface{}{[]models.Post{mockedPosts[1]}}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					var Response struct {
						Items []models.MemoDetail `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Errorf("%s, Unexpected result body: %v", resp)
					}

					var expected []models.Post = tc.misc[0].([]models.Post)
					assertIntHelper(t, tc.name, "result length", len(expected), len(Response.Items))
					for i, r := range Response.Items {
						assertIntHelper(t, tc.name, "memo ID", int(expected[i].ID), int(r.ID))
						assertIntHelper(t, tc.name, "memo project ID", int(expected[i].ProjectID.Int), int(r.Project.ID))
						assertStringHelper(t, tc.name, "memo content", expected[i].Content.String, r.Content.String)
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("GetMemoCount", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetMemoCountOK", "GET", "/memos/count", ``, http.StatusOK, `{"_meta":{"total":3}}`, nil},
			genericRequestTestcase{"GetMemoCountFilterAuthor", "GET", `/memos/count?author=[1]`, ``, http.StatusOK, `{"_meta":{"total":2}}`, nil},
			genericRequestTestcase{"GetMemoCountFilterProject", "GET", `/memos/count?project_id=[1]`, ``, http.StatusOK, `{"_meta":{"total":2}}`, nil},
			genericRequestTestcase{"GetMemoCountFilterMultipleCondition", "GET", `/memos/count?active={"$nin":[0]}&author=[1]&project_id=[2]`, ``, http.StatusOK, `{"_meta":{"total":1}}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)
			})
		}
	})
	t.Run("InsertMemo", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"InsertMemoOK", "POST", "/memo", `{"title":"MemoInsertTest1","content":"MemoInsertTest1","author":1, "project_id":1}`, http.StatusOK, ``, []interface{}{"MemoInsertTest1"}},
			genericRequestTestcase{"InsertMemoDupe", "POST", "/memo", `{"id":1,"title":"MemoTest1","author":131, "project_id":420}`, http.StatusBadRequest, `{"Error":"Memo ID Already Taken"}`, nil},
			genericRequestTestcase{"InsertMemoNoProject", "POST", "/memo", `{"title":"MemoTest1","author":131}`, http.StatusBadRequest, `{"Error":"Invalid Project"}`, nil},
			genericRequestTestcase{"InsertMemoNoProject", "POST", "/memo", `{"title":"MemoTest1","project_id":420}`, http.StatusBadRequest, `{"Error":"Invalid Updator"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					result, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{
						MaxResult: 1,
						Sorting:   "-post_id",
					})
					if err != nil {
						t.Fatalf("Error getting the latest report")
					}
					//log.Println(result)
					assertStringHelper(t, tc.name, "title", tc.misc[0].(string), result[0].Title.String)
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("PutMemo", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"PutMemoOK", "PUT", "/memo", `{"id":1,"title":"MemoTestMod","updated_by":1}`, http.StatusOK, ``, []interface{}{"MemoTestMod"}},
			genericRequestTestcase{"PutMemoUTF8", "PUT", "/memo", `{"id":1,"title":"中文標題蓋蓋蓋","updated_by":1}`, http.StatusOK, ``, []interface{}{"中文標題蓋蓋蓋"}},
			genericRequestTestcase{"PutMemoScheduleNoTime", "PUT", "/memo", `{"id":1,"updated_by":131,"publish_status":3}`, http.StatusBadRequest, `{"Error":"Invalid Publish Time"}`, nil},
			genericRequestTestcase{"PutMemoSchedule", "PUT", "/memo", `{"id":1,"updated_by":131,"publish_status":3,"published_at":"2046-01-05T00:42:42+00:00"}`, http.StatusOK, ``, nil},
			genericRequestTestcase{"PutMemoPublishNoContent", "PUT", "/memo", `{"id":3,"updated_by":131,"publish_status":2}`, http.StatusBadRequest, `{"Error":"Invalid Memo Content"}`, nil},
			genericRequestTestcase{"PutMemoNoUpdater", "PUT", "/memo", `{"id":1,"title":"NoUpdater"}`, http.StatusBadRequest, `{"Error":"Neither updated_by or author is valid"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					if tc.name == "UpdateReportOK" {
						result, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{
							MaxResult: 1,
							IDs:       []int64{1},
						})
						if err != nil {
							t.Fatalf("Error getting the latest report")
						}
						assertStringHelper(t, tc.name, "memo title", tc.misc[0].(string), result[0].Title.String)
					}
				}
			})
		}
	})
	t.Run("DeleteMemo", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteMemoOK", "DELETE", "/memo/1", ``, http.StatusOK, ``, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{
						MaxResult: 1,
						IDs:       []int64{1},
					})
					if err != nil {
						t.Fatalf("Error getting the latest report")
					}
					assertIntHelper(t, tc.name, "report active", 0, int(result[0].Active.Int))
				}
			})
		}
	})
	t.Run("DeleteMemos", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteMemosOK", "DELETE", "/memos", `{"ids":[2,3],"updated_by":131}`, http.StatusOK, ``, []interface{}{[]int64{2, 3}}},
			genericRequestTestcase{"DeleteMemoNoUpdater", "DELETE", "/memos", `{"ids":[1,2,3]}`, http.StatusBadRequest, `{"Error":"Updater Not Specified"}`, nil},
			genericRequestTestcase{"DeleteMemoNoID", "DELETE", "/memos", `{"updated_by":131}`, http.StatusBadRequest, `{"Error":"ID List Empty"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.MemoAPI.GetMemos(&models.MemoGetArgs{
						IDs: tc.misc[0].([]int64),
					})
					if err != nil {
						t.Fatalf("Error getting the latest report")
					}
					assertIntHelper(t, tc.name, "report active", 0, int(result[0].Active.Int))
				}
			})
		}
	})
}
