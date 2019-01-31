//+build integration

package main

import (
	"fmt"
	"testing"
	"time"

	"encoding/json"
	"net/http"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
)

func TestPost(t *testing.T) {

	var mockedPosts = []models.Post{
		models.Post{
			ID:            1,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test post 01", true},
			Content:       models.NullString{"Test post content 01", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://test.link.com", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
		},
		models.Post{
			ID:            2,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test report 01", true},
			Content:       models.NullString{"Test report content 01", true},
			Type:          models.NullInt{4, true},
			Link:          models.NullString{"http://dev.readr.tw/project/report_slug", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
		},
		models.Post{
			ID:            3,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test memo 01", true},
			Content:       models.NullString{"Test memo content 01", true},
			Type:          models.NullInt{5, true},
			Link:          models.NullString{"http://dev.readr.tw/series/project_slug_1/1", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
		},
	}
	var mockedMembers = []models.Member{
		models.Member{
			ID:       1,
			UUID:     "uuid1",
			MemberID: "testmember01@test.cc",
			Nickname: models.NullString{"test_member_01", true},
		},
	}
	var mockTags = []models.Tag{
		models.Tag{
			Text: "tag1",
		},
		models.Tag{
			Text: "tag2",
		},
		models.Tag{
			Text: "tag3",
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
				t.Fatalf("init member data fail: %s ", err.Error())
			}
		}
		for _, v := range mockTags {
			_, err := models.TagAPI.InsertTag(v)
			if err != nil {
				t.Fatalf("init tag data fail: %s ", err.Error())
			}
		}
		return flushDB
	}

	t.Run("GetPost", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetPostOK", "GET", `/post/1`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[0],
			}}},
			genericRequestTestcase{"NotExisted", "GET", `/post/12345`, `{"Error":"Post Not Found"}`, http.StatusNotFound, `{"Error":"Post Not Found"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				if statusCodeOKHelper(code) {
					var Response struct {
						Items []models.TaggedPostMember `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Errorf("%s, Unexpected result body: %v", resp)
					}
					var expected []models.Post = tc.misc[0].([]models.Post)
					assertIntHelper(t, tc.name, "result length", len(expected), len(Response.Items))
					for i, r := range Response.Items {
						assertIntHelper(t, tc.name, "post ID", int(expected[i].ID), int(r.Post.ID))
						assertStringHelper(t, tc.name, "post title", expected[i].Title.String, r.Post.Title.String)
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})

	t.Run("InsertPost", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"New", "POST", `/post`, `{"author":1, "title":"New Post", "type":0}`, http.StatusOK, ``, []interface{}{"New Post"}},
			genericRequestTestcase{"EmptyPayload", "POST", `/post`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Post"}`, []interface{}{}},
			genericRequestTestcase{"WithTags", "POST", `/post`, `{"author":1, "title":"Post with tags", "tags":[1,2], "type":0}`, http.StatusOK, ``, []interface{}{"Post with tags", `1:tag1,2:tag2`}},
			genericRequestTestcase{"WithPorojectID", "POST", `/post`, `{"author":1, "title":"Post with project id", "type":4, "project_id":100001, "type":0}`, http.StatusOK, ``, []interface{}{"Post with project id", 100001}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.PostAPI.GetPosts(&models.PostArgs{
						MaxResult: 1,
						Sorting:   "-post_id",
						ShowTag:   true,
					})
					if err != nil {
						t.Fatalf("Error getting latest posts")
					}
					assertStringHelper(t, tc.name, "title", tc.misc[0].(string), result[0].Title.String)
					if tc.name == "WithTags" {
						assertStringHelper(t, tc.name, "tag string", tc.misc[1].(string), result[0].Tags.String)
					} else if tc.name == "WithPorojectID" {
						assertIntHelper(t, tc.name, "project_id", tc.misc[1].(int), int(result[0].ProjectID.Int))
					}
				}
			})
		}
	})

	t.Run("UpdatePost", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateCurrent", "PUT", `/post`, `{"id":1,"title":"Updated Title", "updated_by":1}`, http.StatusOK, ``, []interface{}{"Updated Title"}},
			genericRequestTestcase{"NotExisted", "PUT", `/post`, `{"id":12345, "author":1}`, http.StatusBadRequest, `{"Error":"Post Not Found"}`, nil},
			genericRequestTestcase{"WithoutUpdater", "PUT", `/post`, `{"id":1,"title":""}`, http.StatusBadRequest, `{"Error":"Neither updated_by or author is valid"}`, nil},
			genericRequestTestcase{"UpdateTags", "PUT", `/post`, `{"id":1, "tags":[3], "title":"UpdateTags", "updated_by":1}`, http.StatusOK, ``, []interface{}{"3:tag3"}},
			genericRequestTestcase{"DeleteTags", "PUT", `/post`, `{"id":1, "tags":[], "title":"DeleteTags", "updated_by":1}`, http.StatusOK, ``, []interface{}{""}},
			genericRequestTestcase{"UpdateProjectID", "PUT", `/post`, `{"id":1, "author":1, "project_id":100002}`, http.StatusOK, ``, []interface{}{100002}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.PostAPI.GetPosts(&models.PostArgs{
						MaxResult: 1,
						IDs:       []uint32{uint32(1)},
						ShowTag:   true,
					})
					if err != nil {
						t.Fatalf("Error getting latest posts")
					}
					if tc.name == "UpdateTags" || tc.name == "DeleteTags" {
						assertStringHelper(t, tc.name, "tag string", tc.misc[0].(string), result[0].Tags.String)
					} else if tc.name == "UpdateProjectID" {
						assertIntHelper(t, tc.name, "project_id", tc.misc[0].(int), int(result[0].ProjectID.Int))
					} else if tc.name == "UpdateCurrent" {
						assertStringHelper(t, tc.name, "title", tc.misc[0].(string), result[0].Title.String)
					}
				}
			})
		}
	})

	t.Run("DeletePost", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"DeleteOK", "DELETE", `/post/1`, ``, http.StatusOK, ``, []interface{}{0}},
			genericRequestTestcase{"NotFound", "DELETE", `/post/12345`, ``, http.StatusNotFound, `{"Error":"Post Not Found"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.PostAPI.GetPosts(&models.PostArgs{
						MaxResult: 1,
						IDs:       []uint32{uint32(1)},
					})
					if err != nil {
						t.Fatalf("Error getting latest posts")
					}
					assertIntHelper(t, tc.name, "post active", int(result[0].Post.Active.Int), tc.misc[0].(int))
				}
			})
		}
	})
}

func TestPosts(t *testing.T) {

	var mockedPosts = []models.Post{
		models.Post{
			ID:            1,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test post 01", true},
			Content:       models.NullString{"Test post content 01", true},
			Type:          models.NullInt{1, true},
			Link:          models.NullString{"http://test.link1.com", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2017, 12, 17, 23, 11, 37, 0, time.UTC), Valid: true},
		},
		models.Post{
			ID:            2,
			Author:        models.NullInt{1, true},
			Title:         models.NullString{"Test post 02", true},
			Content:       models.NullString{"Test post content 02", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://test.link2.com", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{2, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2017, 12, 18, 23, 11, 37, 0, time.UTC), Valid: true},
		},
		models.Post{
			ID:            3,
			Author:        models.NullInt{2, true},
			Title:         models.NullString{"Test post 03", true},
			Content:       models.NullString{"Test post content 03", true},
			Type:          models.NullInt{2, true},
			Link:          models.NullString{"http://test.link3.com", true},
			Active:        models.NullInt{1, true},
			PublishStatus: models.NullInt{3, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2017, 12, 19, 23, 11, 37, 0, time.UTC), Valid: true},
		},
		models.Post{
			ID:            4,
			Author:        models.NullInt{2, true},
			Title:         models.NullString{"Test post 04", true},
			Content:       models.NullString{"Test post content 04", true},
			Type:          models.NullInt{0, true},
			Link:          models.NullString{"http://test.link4.com", true},
			Active:        models.NullInt{0, true},
			PublishStatus: models.NullInt{3, true},
			UpdatedAt:     models.NullTime{Time: time.Date(2017, 12, 20, 23, 11, 37, 0, time.UTC), Valid: true},
		},
	}
	var mockedMembers = []models.Member{
		models.Member{
			ID:       1,
			UUID:     "uuid1",
			MemberID: "testmember01@test.cc",
			Nickname: models.NullString{"test_member_01", true},
		},
		models.Member{
			ID:       2,
			UUID:     "uuid2",
			MemberID: "testmember02@test.cc",
			Nickname: models.NullString{"test_member_02", true},
		},
	}
	var mockTags = []models.Tag{
		models.Tag{
			Text: "tag1",
		},
		models.Tag{
			Text: "tag2",
		},
		models.Tag{
			Text: "tag3",
		},
	}

	init := func() func() {
		for _, v := range mockedPosts {
			_, err := models.PostAPI.InsertPost(v)
			if err != nil {
				fmt.Print(err.Error())
				t.Fatalf("init post data fail: %s ", err.Error())
			}
		}
		for _, v := range mockedMembers {
			_, err := models.MemberAPI.InsertMember(v)
			if err != nil {
				fmt.Print(err.Error())
				t.Fatalf("init member data fail: %s ", err.Error())
			}
		}
		for _, v := range mockTags {
			_, err := models.TagAPI.InsertTag(v)
			if err != nil {
				fmt.Print(err.Error())
				t.Fatalf("init tag data fail: %s ", err.Error())
			}
		}
		return flushDB
	}

	t.Run("GetPosts", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdatedAtDescending", "GET", `/posts`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[2], mockedPosts[1], mockedPosts[0]}}},
			genericRequestTestcase{"UpdatedAtAscending", "GET", `/posts?sort=updated_at`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[0], mockedPosts[1], mockedPosts[2]}}},
			genericRequestTestcase{"MaxResult", "GET", `/posts?max_result=2`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[2], mockedPosts[1]}}},
			genericRequestTestcase{"AuthorFilter", "GET", `/posts?author={"$in":[1]}`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[1], mockedPosts[0]}}},
			genericRequestTestcase{"ActiveFilter", "GET", `/posts?active={"$nin":[1]}`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[3]}}},
			genericRequestTestcase{"NotFound", "GET", `/posts?active={"$nin":[0,1]}`, ``, http.StatusOK, ``, []interface{}{[]models.Post{}}},
			genericRequestTestcase{"Type", "GET", `/posts?type={"$in":[1,2]}`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[2], mockedPosts[0]}}},
			genericRequestTestcase{"ShowDetails", "GET", `/posts?show_author=true&show_updater=true&show_tag=true&show_comment=true`, ``, http.StatusOK, ``,
				[]interface{}{[]models.Post{mockedPosts[2], mockedPosts[1], mockedPosts[0]}, []string{"test_member_02", "test_member_01", "test_member_01"}}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)

				if statusCodeOKHelper(code) {
					var Response struct {
						Items []models.TaggedPostMember `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Errorf("%s, Unexpected result body: %v", resp)
					}
					var expected []models.Post = tc.misc[0].([]models.Post)

					assertIntHelper(t, tc.name, "result length", len(expected), len(Response.Items))

					for i, r := range Response.Items {

						assertIntHelper(t, tc.name, "post ID", int(expected[i].ID), int(r.Post.ID))
						assertStringHelper(t, tc.name, "post title", expected[i].Title.String, r.Post.Title.String)

						if tc.name == "ShowDetails" {

							assertStringHelper(t, tc.name, "author name", tc.misc[1].([]string)[i], r.Member.Nickname.String)
							assertStringHelper(t, tc.name, "tags", "", r.Tags.String)

						}
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})

	t.Run("GetActivePosts", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"GetActivePostsOK", "GET", `/posts/active`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[2], mockedPosts[1], mockedPosts[0]}}},
			genericRequestTestcase{"SetActiveOption", "GET", `/posts/active?active={"$nin":[1]}`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[2], mockedPosts[1], mockedPosts[0]}}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)

				if statusCodeOKHelper(code) {
					var Response struct {
						Items []models.TaggedPostMember `json:"_items"`
					}
					err := json.Unmarshal([]byte(resp), &Response)
					if err != nil {
						t.Errorf("%s, Unexpected result body: %v", resp)
					}
					var expected []models.Post = tc.misc[0].([]models.Post)

					assertIntHelper(t, tc.name, "result length", len(expected), len(Response.Items))

					for i, r := range Response.Items {

						assertIntHelper(t, tc.name, "post ID", int(expected[i].ID), int(r.Post.ID))
						assertStringHelper(t, tc.name, "post title", expected[i].Title.String, r.Post.Title.String)
					}
				} else {
					assertStringHelper(t, tc.name, "request result", tc.resp, resp)
				}
			})
		}
	})

	t.Run("DeletePosts", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"SimpleDelete", "DELETE", `/posts?ids=[1,2]`, ``, http.StatusOK, ``, []interface{}{[]uint32{1, 2}}},
			genericRequestTestcase{"EmptyID", "DELETE", `/posts?ids=[]`, ``, http.StatusBadRequest, `{"Error":"ID List Empty"}`, nil},
			genericRequestTestcase{"NotFound", "DELETE", `/posts?ids=[6,7]`, ``, http.StatusNotFound, `{"Error":"Posts Not Found"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.PostAPI.GetPosts(&models.PostArgs{
						MaxResult: 1,
						IDs:       tc.misc[0].([]uint32),
					})
					if err != nil {
						t.Fatalf("Error getting latest posts")
					}
					for _, post := range result {
						assertIntHelper(t, tc.name, "post active", 0, int(post.Active.Int))
					}
				}
			})
		}
	})

	t.Run("UpdatePosts", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"Posts", "PUT", `/posts`, `{"ids":[0,1,2,3]}`, http.StatusOK, ``, []interface{}{[]uint32{0, 1, 2, 3}}},
			genericRequestTestcase{"NotFound", "PUT", `/posts`, `{"ids":[6,7]}`, http.StatusNotFound, `{"Error":"Posts Not Found"}`, nil},
			genericRequestTestcase{"InvalidPayload", "PUT", `/posts`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Request Body"}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				if statusCodeOKHelper(code) {
					result, err := models.PostAPI.GetPosts(&models.PostArgs{
						MaxResult: 1,
						IDs:       tc.misc[0].([]uint32),
					})
					if err != nil {
						t.Fatalf("Error getting latest posts")
					}
					for _, post := range result {
						assertIntHelper(t, tc.name, "post publish_status", 2, int(post.PublishStatus.Int))
					}
				}
			})
		}
	})

	t.Run("CountPosts", func(t *testing.T) {
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"Posts", "GET", `/posts/count`, ``, http.StatusOK, `{"_meta":{"total":3}}`, nil},
			genericRequestTestcase{"Active", "GET", `/posts/count?active={"$in":[0,1]}`, ``, http.StatusOK, `{"_meta":{"total":4}}`, nil},
			genericRequestTestcase{"Author", "GET", `/posts/count?author={"$nin":[1]}`, ``, http.StatusOK, `{"_meta":{"total":1}}`, nil},
			genericRequestTestcase{"MoreThanOneActive", "GET", `/posts/count?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`, nil},
			genericRequestTestcase{"NotEntirelyValidActive", "GET", `/posts/count?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`, nil},
			genericRequestTestcase{"NoValidActive", "GET", `/posts/count?active={"$nin":[-3,-4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`, nil},
			genericRequestTestcase{"Type", "GET", `/posts/count?type={"$in":[1,2]}`, ``, http.StatusOK, `{"_meta":{"total":2}}`, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				code, resp := genericDoRequest(tc, t)
				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)
			})
		}
	})

	t.Run("PostCache", func(t *testing.T) {
		resetRedisKeyHelper(t, "PostCache", []string{"postcache_hot_index", "postcache_hot", "postcache_latest_index", "postcache_latest"})
		defer init()()
		for _, tc := range []genericRequestTestcase{
			genericRequestTestcase{"UpdateCache", "PUT", `/posts/cache`, ``, http.StatusOK, ``, []interface{}{[]models.Post{
				mockedPosts[1], mockedPosts[0]}}},
		} {
			t.Run(tc.name, func(t *testing.T) {

				code, resp := genericDoRequest(tc, t)
				var expected []models.Post = tc.misc[0].([]models.Post)

				assertIntHelper(t, tc.name, "status code", tc.httpcode, code)
				assertStringHelper(t, tc.name, "request result", tc.resp, resp)

				conn := models.RedisHelper.Conn()
				defer conn.Close()

				res, err := redis.Values(conn.Do("hgetall", "postcache_hot"))
				if err != nil {
					t.Fatalf("Fail to scan redis hash: %v", err.Error())
				}

				redisBytes := [][]byte{}
				if err = redis.ScanSlice(res, &redisBytes); err != nil {
					t.Fatalf("Error scan redis hash to slice: %v", err)
				}

				assertIntHelper(t, tc.name, "result length", len(expected), len(redisBytes)/2)
				for i, redisByte := range redisBytes {

					if i%2 != 0 {
						index := (i - 1) / 2
						var post_cache models.TaggedPostMember

						if err := json.Unmarshal(redisByte, &post_cache); err != nil {
							t.Fatalf("Error scan redis comment notification: %s , %v", string(redisByte), err)
						}

						assertIntHelper(t, tc.name, "post ID", int(expected[index].ID), int(post_cache.Post.ID))
						assertStringHelper(t, tc.name, "post title", expected[index].Title.String, post_cache.Post.Title.String)

					}
				}

			})
		}
	})
}
