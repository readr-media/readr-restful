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

func TestReport(t *testing.T) {

	var mockedPosts = []models.Post{
		models.Post{
			ID:            1,
			Title:         models.NullString{"Test report 01", true},
			Content:       models.NullString{"Test report content 01", true},
			Type:          models.NullInt{4, true},
			Link:          models.NullString{"http://dev.readr.tw/project/report_slug1", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{1, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 2, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{1, true},
		},
		models.Post{
			ID:            2,
			Title:         models.NullString{"Test report 02", true},
			Content:       models.NullString{"Test report content 02", true},
			Type:          models.NullInt{4, true},
			Link:          models.NullString{"http://dev.readr.tw/project/report_slug2", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 3, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{2, true},
			Slug:          models.NullString{"report_slug2", true},
		},
		models.Post{
			ID:            3,
			Title:         models.NullString{"Test report 03", true},
			Content:       models.NullString{"Test report content 03", true},
			Type:          models.NullInt{4, true},
			Link:          models.NullString{"http://dev.readr.tw/project/report_slug3", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2018, 1, 4, 3, 4, 5, 6, time.UTC), Valid: true},
			ProjectID:     models.NullInt{1, true},
			Slug:          models.NullString{"report_slug3", true},
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

	var mockedReportAuthors = map[int][]int{
		1: []int{1},
		2: []int{1, 2},
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
		for k, v := range mockedReportAuthors {
			err := models.ReportAPI.InsertAuthors(k, v)
			if err != nil {
				t.Fatalf("init project data fail: %s ", err.Error())
			}
		}
		return flushDB
	}

	t.Run("GetReport", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetReportBasicOK", "GET", "/report/list", ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2], mockedPosts[1], mockedPosts[0]}},
			},
			genericRequestTestcase{"GetReportMaxResultOK", "GET", "/report/list?max_result=1", ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2]}},
			},
			genericRequestTestcase{"GetReportOffsetOK", "GET", "/report/list?max_result=1&page=2", ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[1]}},
			},
			genericRequestTestcase{"GetReportWithIDsOK", "GET", `/report/list?ids=[2,1]`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[1], mockedPosts[0]}},
			},
			genericRequestTestcase{"GetReportWithIDsNotFound", "GET", "/report/list?ids=[9527]", ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{}},
			},
			genericRequestTestcase{"GetReportWithSlugs", "GET", `/report/list?report_slugs=["report_slug2"]`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[1]}},
			},
			genericRequestTestcase{"GetReportWithMultipleSlugs", "GET", `/report/list?report_slugs=["report_slug2","report_slug3"]`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2], mockedPosts[1]}},
			},
			genericRequestTestcase{"GetReportWithProjectSlugs", "GET", `/report/list?project_slugs=["project_slug_2"]`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[1]}},
			},
			genericRequestTestcase{"GetReportWithReportPublishStatus", "GET", `/report/list?report_publish_status={"$in":[1]}`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[0]}},
			},
			genericRequestTestcase{"GetReportWithProjectPublishStatus", "GET", `/report/list?project_publish_status={"$in":[1]}`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2], mockedPosts[0]}},
			},
			genericRequestTestcase{"GetReportWithMultipleSorting", "GET", `/report/list?sort=-slug,id`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[0], mockedPosts[1], mockedPosts[2]}},
			},
			genericRequestTestcase{"GetReportKeywordMatchTitle", "GET", `/report/list?keyword=03&active={"$in":[0,1]}`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2]}},
			},
			genericRequestTestcase{"GetReportKeywordMatchID", "GET", `/report/list?keyword=3`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2]}},
			},
			genericRequestTestcase{"GetReportWithAuthorsFieldsSet", "GET", `/report/list?ids=[1,2]&fields=["id","nickname"]`, ``, http.StatusOK, ``,
				[]interface{}{
					[]models.Post{mockedPosts[1], mockedPosts[0]},
					[][]string{[]string{"test_member_01", "test_member_02"}, []string{"test_member_01"}},
				}},
			genericRequestTestcase{"GetReportWithAuthorsFull", "GET", `/report/list?ids=[1,32767]&mode=full`, ``, http.StatusOK, ``,
				[]interface{}{
					[]models.Post{mockedPosts[0]},
					[][]string{[]string{"test_member_01"}},
					[][]string{[]string{"testmember01@test.cc"}},
				}},
			genericRequestTestcase{"GetReportWithAuthorsInvalidFields", "GET", `/report/list?fields=["cat"]`, ``, http.StatusBadRequest, `{"Error":"Invalid Fields"}`,
				[]interface{}{}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					var Response struct {
						Items []models.ReportAuthors `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Errorf("%s, Unexpected result body: %v", resp)
					}

					var expected []models.Post = tc.misc[0].([]models.Post)
					assertIntHelper(t, tc.name, "result length", len(expected), len(Response.Items))
					for i, r := range Response.Items {
						assertIntHelper(t, tc.name, "report ID", int(expected[i].ID), int(r.Report.ID))
						assertIntHelper(t, tc.name, "report project ID", int(expected[i].ProjectID.Int), int(r.Project.ID))
						assertStringHelper(t, tc.name, "report content", expected[i].Content.String, r.Report.Content.String)

						if tc.name == "GetReportWithAuthorsFieldsSet" {
							expectedNames := tc.misc[1].([][]string)
							assertIntHelper(t, tc.name, "author info amount", len(expectedNames[i]), len(r.Authors))
							if len(expectedNames[i]) == len(r.Authors) {
								for authorIndex, author := range r.Authors {
									assertStringHelper(t, tc.name, "author nickname", expectedNames[i][authorIndex], author.Nickname.String)
								}
							}
						}

						if tc.name == "GetReportWithAuthorsFull" {
							expectedNames := tc.misc[1].([][]string)
							expectedMails := tc.misc[2].([][]string)
							assertIntHelper(t, tc.name, "author amount", len(expectedNames[i]), len(r.Authors))
							assertIntHelper(t, tc.name, "author amount", len(expectedMails[i]), len(r.Authors))
							if len(expectedNames) == len(r.Authors) && len(expectedMails) == len(r.Authors) {
								for authorIndex, author := range r.Authors {
									assertStringHelper(t, tc.name, "author nickname", expectedNames[i][authorIndex], author.Nickname.String)
									assertStringHelper(t, tc.name, "author mail", expectedMails[i][authorIndex], author.Mail.String)
								}
							}
						}
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("CountReports", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetReportCount", "GET", `/report/count`, ``, http.StatusOK, `{"_meta":{"total":3}}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)
			})
		}
	})
	t.Run("InsertReport", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"PostReportOK", "POST", "/report", `{"title":"PostReportOK","post_id":4,"like_amount":0,"comment_amount":0,"active":1,"slug":"", "project_id":1}`, http.StatusOK, ``, []interface{}{"PostReportOK"}},
			genericRequestTestcase{"PostReportNonActive", "POST", "/report", `{"title":"PostReportNonActive","post_id":4,"slug":"testreportslug02"}`, http.StatusOK, ``, []interface{}{"PostReportNonActive"}},
			genericRequestTestcase{"PostReportEmptyBody", "POST", "/report", ``, http.StatusBadRequest, `{"Error":"Invalid Report"}`, nil},
			genericRequestTestcase{"PostReportNoTitle", "POST", "/report", ``, http.StatusBadRequest, `{"Error":"Invalid Report"}`, nil},
			genericRequestTestcase{"PostReportSlugDupe", "POST", "/report", `{"slug":"testreportslug02", "title":"Dupe"}`, http.StatusBadRequest, `{"Error":"Duplicate Slug"}`, nil},
			genericRequestTestcase{"PostReportPublishedWithoutSlug", "POST", "/report", `{"publish_status": 2, "title":"PostReportPublishedWithoutSlug"}`, http.StatusBadRequest, `{"Error":"Must Have Slug Before Publish"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)

				if statusCodeOKHelper(code) {
					args := models.GetReportArgs{
						MaxResult: 1,
						Sorting:   "-post_id",
					}
					args.Fields = args.FullAuthorTags()
					result, err := models.ReportAPI.GetReports(args)
					if err != nil {
						t.Fatalf("Error getting the latest report")
					}
					assertStringHelper(t, tc.name, "title", tc.misc[0].(string), result[0].Title.String)

				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})
	t.Run("UpdateReport", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateReportOK", "PUT", "/report", `{"id":1,"title":"Modified","active":1}`, http.StatusOK, ``, []interface{}{"Modified"}},
			genericRequestTestcase{"UpdateReportID0", "PUT", "/report", `{"id":0}`, http.StatusBadRequest, `{"Error":"Invalid Report Data"}`, nil},
			genericRequestTestcase{"UpdateReportInvalidPublishStatus", "PUT", "/report", `{"id":1, "publish_status":987}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`, nil},
			genericRequestTestcase{"UpdateReportNotExist", "PUT", "/report", `{"id":32767,"title":"NotExist"}`, http.StatusBadRequest, `{"Error":"Report Not Found"}`, nil},
			genericRequestTestcase{"UpdateReportInvalidActive", "PUT", "/report", `{"id":1,"active":3}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`, nil},
			genericRequestTestcase{"UpdatePublishReportWithNoSlug", "PUT", "/report", `{"id":1,"publish_status":2}`, http.StatusBadRequest, `{"Error":"Must Have Slug Before Publish"}`, nil},
			genericRequestTestcase{"UpdateReportStatusOK", "PUT", "/report", `{"id":2,"publish_status":2}`, http.StatusOK, ``, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					if tc.name == "UpdateReportOK" {
						args := models.GetReportArgs{
							MaxResult: 1,
							IDs:       []int{1},
						}
						args.Fields = args.FullAuthorTags()
						result, err := models.ReportAPI.GetReports(args)
						if err != nil {
							t.Fatalf("Error getting the latest report")
						}
						assertStringHelper(t, tc.name, "report title", tc.misc[0].(string), result[0].Title.String)
					}
				}
			})
		}
	})
	t.Run("DeleteReport", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteReportOK", "DELETE", "/report/1", ``, http.StatusOK, ``, []interface{}{0}},
			genericRequestTestcase{"DeleteReportNotExist", "DELETE", "/report/32767", ``, http.StatusNotFound, `{"Error":"Report Not Found"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					args := models.GetReportArgs{
						MaxResult: 1,
						IDs:       []int{1},
					}
					args.Fields = args.FullAuthorTags()
					result, err := models.ReportAPI.GetReports(args)
					if err != nil {
						t.Fatalf("Error getting the latest report")
					}
					assertIntHelper(t, tc.name, "report active", int(result[0].Active.Int), tc.misc[0].(int))
				}
			})
		}
	})
	t.Run("Insert||UpdateReportAuthors", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"PostReportAuthorsOK", "POST", `/report/author`, `{"report_id":3, "author_ids":[1]}`, http.StatusOK, ``, []interface{}{[]string{"test_member_01"}, []string{"testmember01@test.cc"}}},
			genericRequestTestcase{"PostReportAuthorsInvalidParameters", "POST", "/report/author", `{"report_id":1000010}`, http.StatusBadRequest, `{"Error":"Insufficient Parameters"}`, nil},
			genericRequestTestcase{"PutReportAuthorsOK", "PUT", `/report/author`, `{"report_id":3, "author_ids":[2]}`, http.StatusOK, ``, []interface{}{[]string{"test_member_02"}, []string{"testmember02@test.cc"}}},
			genericRequestTestcase{"PutReportAuthorsInvalidParameters", "PUT", "/report/author", `{"author_ids":[1000010]}`, http.StatusBadRequest, `{"Error":"Insufficient Parameters"}`, nil},
		} {
			code, resp := genericDoRequest(tc, t)
			assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
			assertStringHelper(t, tc.name, "request result", tc.resp, resp)

			if statusCodeOKHelper(code) {
				args := models.GetReportArgs{
					MaxResult: 1,
					IDs:       []int{3},
				}
				args.Fields = args.FullAuthorTags()
				result, err := models.ReportAPI.GetReports(args)
				if err != nil {
					t.Fatalf("Error getting the latest report")
				}
				expectedNames := tc.misc[0].([]string)
				expectedMails := tc.misc[1].([]string)
				assertIntHelper(t, tc.name, "author amount", len(expectedNames), len(result[0].Authors))
				assertIntHelper(t, tc.name, "author amount", len(expectedMails), len(result[0].Authors))
				if len(expectedNames) == len(result[0].Authors) && len(expectedMails) == len(result[0].Authors) {
					for authorIndex, author := range result[0].Authors {
						assertStringHelper(t, tc.name, "author nickname", expectedNames[authorIndex], author.Nickname.String)
						assertStringHelper(t, tc.name, "author mail", expectedMails[authorIndex], author.Mail.String)
					}
				}
			}
		}
	})
}
