package routes

import (
	// "encoding/json"
	"errors"
	// "log"
	// "net/http"
	// "os"
	// "testing"
	// "time"

	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

var mockMemoDS []models.Memo

type mockMemoAPI struct{}

func (m *mockMemoAPI) CountMemos(args *models.MemoGetArgs) (count int, err error) {
	switch {
	case len(args.Author) > 0 && len(args.Project) > 0:
		return 1, nil
	case len(args.Author) > 0 && len(args.Project) == 0:
		return 3, nil
	case len(args.Author) == 0 && len(args.Project) > 0:
		return 2, nil
	default:
		return 6, nil
	}
	return 0, nil
}
func (m *mockMemoAPI) GetMemo(id int) (memo models.Memo, err error) {

	for _, v := range mockMemoDS {
		if int(v.ID) == id {
			return v, nil
		}
	}
	return models.Memo{}, errors.New("Not Found")
}
func (m *mockMemoAPI) GetMemos(args *models.MemoGetArgs) (memos []models.MemoDetail, err error) {
	switch {
	case len(args.IDs) == 1:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}}}, nil
	case len(args.Author) > 0 && len(args.Project) > 0:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case len(args.Author) > 0 && len(args.Project) == 0:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case len(args.Author) == 0 && len(args.Project) > 0:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 3, Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case args.Sorting == "-author,post_id":
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 3, Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case len(args.Slugs) == 1:
		return []models.MemoDetail{

			models.MemoDetail{Memo: models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case args.Page == 2:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case args.MaxResult == 1:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case len(args.ProjectPublishStatus) > 0:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	case args.Keyword == "中文":
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	default:
		return []models.MemoDetail{
			models.MemoDetail{Memo: models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 3, Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
			models.MemoDetail{Memo: models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	}
	return []models.MemoDetail{}, nil
}
func (m *mockMemoAPI) InsertMemo(memo models.Memo) (lastID int, err error) {
	if memo.ID == 0 {
		memo.ID = uint32(len(mockMemoDS) + 1)
	}
	for _, v := range mockMemoDS {
		if v.ID == memo.ID {
			return 0, errors.New("Duplicate entry")
		}
	}
	mockMemoDS = append(mockMemoDS, memo)

	return len(mockMemoDS), nil
}
func (m *mockMemoAPI) UpdateMemo(memo models.Memo) (err error) {
	for _, v := range mockMemoDS {
		if v.ID == memo.ID {
			v.Title = memo.Title
			return nil
		}
	}
	return nil
}
func (m *mockMemoAPI) UpdateMemos(args models.MemoUpdateArgs) (err error) { return nil }
func (m *mockMemoAPI) SchedulePublish() ([]int, error)                    { return []int{}, nil }
func (m *mockMemoAPI) PublishHandler(ids []int) error                     { return nil }
func (m *mockMemoAPI) UpdateHandler(ids []int, params ...int64) error     { return nil }

// func TestRouteMemos(t *testing.T) {

// 	if os.Getenv("db_driver") == "mysql" {
// 		_, _ = models.DB.Exec("truncate table projects;")
// 	} else {
// 		mockProjectDS = []models.Project{}
// 	}

// 	for _, memo := range []models.Memo{
// 		models.Memo{Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 		models.Memo{Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 		models.Memo{Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 		models.Memo{Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 	} {
// 		_, err := models.MemoAPI.InsertMemo(memo)
// 		if err != nil {
// 			log.Printf("Init memo test fail %s", err.Error())
// 		}
// 	}

// 	for _, params := range []models.Member{
// 		models.Member{ID: 131, MemberID: "MemoTestDefault1@mirrormedia.mg", Active: rrsql.NullInt{1, true}, PostPush: rrsql.NullBool{true, true}, UpdatedAt: rrsql.NullTime{time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: rrsql.NullString{"MemoTestDefault1@mirrormedia.mg", true}, Points: rrsql.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b88e-f06e59fd8131", TalkID: rrsql.NullString{"abc1d5b1-da54-4200-b58e-f06e59fd8131", true}},
// 		models.Member{ID: 132, MemberID: "MemoTestDefault2@mirrormedia.mg", Active: rrsql.NullInt{1, true}, PostPush: rrsql.NullBool{true, true}, UpdatedAt: rrsql.NullTime{time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: rrsql.NullString{"MemoTestDefault2@mirrormedia.mg", true}, Points: rrsql.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b59e-f06e59fd8132", TalkID: rrsql.NullString{"abc1d5b1-da54-4200-b59e-f06e59fd8132", true}},
// 		models.Member{ID: 135, MemberID: "MemoTestDefault3@mirrormedia.mg", Active: rrsql.NullInt{1, true}, PostPush: rrsql.NullBool{true, true}, UpdatedAt: rrsql.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: rrsql.NullString{"MemoTestDefault3@mirrormedia.mg", true}, Points: rrsql.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b60e-f06e59fd8135", TalkID: rrsql.NullString{"abc1d5b1-da54-4200-b60e-f06e59fd8135", true}},
// 	} {
// 		_, err := models.MemberAPI.InsertMember(params)
// 		if err != nil {
// 			log.Printf("Insert member fail when init test case. Error: %v", err)
// 		}
// 	}

// 	for _, params := range []models.Project{
// 		models.Project{ID: 420, Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Test project for memo", true}, Slug: rrsql.NullString{"testproject", true}},
// 		models.Project{ID: 421, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{3, true}, Title: rrsql.NullString{"Test project for memo2", true}, Slug: rrsql.NullString{"testproject2", true}},
// 	} {
// 		err := models.ProjectAPI.InsertProject(params)
// 		if err != nil {
// 			log.Printf("Insert project fail when init test case. Error: %v", err)
// 		}
// 	}

// 	asserter := func(resp string, tc genericTestcase, t *testing.T) {
// 		type response struct {
// 			Items []models.Memo `json:"_items"`
// 		}

// 		var Response response
// 		var expected []models.Memo = tc.resp.([]models.Memo)

// 		err := json.Unmarshal([]byte(resp), &Response)
// 		if err != nil {
// 			t.Errorf("%s, Unexpected result body: %v", resp, err.Error())
// 		}

// 		if len(Response.Items) != len(expected) {
// 			t.Errorf("%s expect memo length to be %v but get %v", tc.name, len(expected), len(Response.Items))
// 			return
// 		}

// 		for i, respmemo := range Response.Items {
// 			expmemo := expected[i]
// 			if respmemo.ID == expmemo.ID &&
// 				respmemo.Title == expmemo.Title &&
// 				respmemo.Active == expmemo.Active &&
// 				respmemo.ProjectID == expmemo.ProjectID {
// 				continue
// 			}
// 			t.Errorf("%s, expect to get %v, but %v ", tc.name, expmemo, respmemo)
// 		}
// 	}

// 	t.Run("InsertMemo", func(t *testing.T) {
// 		for _, testcase := range []genericTestcase{
// 			genericTestcase{"InsertMemoOK", "POST", "/memo", `{"id":100,"title":"MemoTest2","content":"MemoTest2","author":132, "project_id":421}`, http.StatusOK, ``},
// 			genericTestcase{"InsertMemoOK", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":131, "project_id":420}`, http.StatusOK, ``},
// 			genericTestcase{"InsertMemoDupe", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":131, "project_id":420}`, http.StatusBadRequest, `{"Error":"Memo ID Already Taken"}`},
// 			genericTestcase{"InsertMemoNoProject", "POST", "/memo", `{"title":"MemoTest1","author":131}`, http.StatusBadRequest, `{"Error":"Invalid Project"}`},
// 		} {
// 			genericDoTest(testcase, t, asserter)
// 		}
// 	})
// 	t.Run("PutMemo", func(t *testing.T) {
// 		for _, testcase := range []genericTestcase{
// 			genericTestcase{"PutMemoOK", "PUT", "/memo", `{"id":101,"title":"MemoTestMod","updated_by":131}`, http.StatusOK, ``},
// 			genericTestcase{"PutMemoUTF8", "PUT", "/memo", `{"id":101,"title":"順便測中文","updated_by":133}`, http.StatusOK, ``},
// 			genericTestcase{"PutMemoScheduleNoTime", "PUT", "/memo", `{"id":100,"updated_by":131,"publish_status":3}`, http.StatusBadRequest, `{"Error":"Invalid Publish Time"}`},
// 			genericTestcase{"PutMemoSchedule", "PUT", "/memo", `{"id":100,"updated_by":131,"publish_status":3,"published_at":"2046-01-05T00:42:42+00:00"}`, http.StatusOK, ``}, //published_at is time string in RFC3339 format
// 			genericTestcase{"PutMemoPublishNoContent", "PUT", "/memo", `{"id":101,"updated_by":131,"publish_status":2}`, http.StatusBadRequest, `{"Error":"Invalid Memo Content"}`},
// 			genericTestcase{"PutMemoNoUpdater", "PUT", "/memo", `{"id":101,"title":"NoUpdater"}`, http.StatusBadRequest, `{"Error":"Neither updated_by or author is valid"}`},
// 		} {
// 			genericDoTest(testcase, t, asserter)
// 		}
// 	})
// 	t.Run("GetMemo", func(t *testing.T) {
// 		for _, testcase := range []genericTestcase{
// 			// genericTestcase{"GetMemoOK", "GET", "/memo/1", ``, http.StatusOK, `{"_items":{"id":1,"created_at":null,"comment_amount":null,"title":"MemoTestDefault1","content":null,"link":null,"author":131,"project_id":420,"active":1,"updated_at":null,"updated_by":null,"published_at":null,"publish_status":null,"memo_order":null}}`},
// 			genericTestcase{"GetMemoOK", "GET", "/memo/1", ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 		} {
// 			genericDoTest(testcase, t, asserter)
// 		}
// 	})
// 	t.Run("GetMemos", func(t *testing.T) {
// 		for _, testcase := range []genericTestcase{
// 			genericTestcase{"GetMemoDefaultOK", "GET", "/memos?sort=project_id", ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 3, Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1", ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1&page=2", ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoSortMultipleOK", "GET", "/memos?sort=-author,memo_id", ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 3, Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoSortInvalidOption", "GET", "/memos?sort=meow", ``, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`},
// 			genericTestcase{"GetMemoFilterAuthor", "GET", `/memos?author=[135,132]`, ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoFilterProject", "GET", `/memos?project_id=[422]`, ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 3, Title: rrsql.NullString{"MemoTestDefault3", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoFilterMultipleCondition", "GET", `/memos?active={"$nin":[0]}&author=[135,132]&project_id=[422]`, ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 4, Title: rrsql.NullString{"MemoTestDefault4", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{422, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoWithSlug", "GET", `/memos?slugs=["testproject"]`, ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 1, Title: rrsql.NullString{"MemoTestDefault1", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 				models.Memo{ID: 2, Title: rrsql.NullString{"MemoTestDefault2", true}, Author: rrsql.NullInt{135, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoWithMemoStatusAndProjectStatus", "GET", `/memos?project_publish_status={"$in":[3]}`, ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 100, Title: rrsql.NullString{"MemoTest2", true}, Author: rrsql.NullInt{132, true}, ProjectID: rrsql.NullInt{421, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 			genericTestcase{"GetMemoWithKeyword", "GET", `/memos?keyword=中文`, ``, http.StatusOK, []models.Memo{
// 				models.Memo{ID: 101, Title: rrsql.NullString{"順便測中文", true}, Author: rrsql.NullInt{131, true}, ProjectID: rrsql.NullInt{420, true}, Active: rrsql.NullInt{1, true}},
// 			}},
// 		} {
// 			genericDoTest(testcase, t, asserter)
// 		}
// 	})
// 	t.Run("GetMemoCount", func(t *testing.T) {
// 		for _, testcase := range []genericTestcase{
// 			genericTestcase{"GetMemoCountOK", "GET", "/memos/count", ``, http.StatusOK, `{"_meta":{"total":6}}`},
// 			genericTestcase{"GetMemoCountFilterAuthor", "GET", `/memos/count?author=[135,132]`, ``, http.StatusOK, `{"_meta":{"total":3}}`},
// 			genericTestcase{"GetMemoCountFilterProject", "GET", `/memos/count?project_id=[422]`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
// 			genericTestcase{"GetMemoCountFilterMultipleCondition", "GET", `/memos/count?active={"$nin":[0]}&author=[135,132]&project_id=[422]`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
// 		} {
// 			genericDoTest(testcase, t, asserter)
// 		}
// 	})
// 	t.Run("DeleteMemo", func(t *testing.T) {
// 		for _, testcase := range []genericTestcase{
// 			genericTestcase{"DeleteMemoOK", "DELETE", "/memo/1", ``, http.StatusOK, ``},
// 			genericTestcase{"DeleteMemoOK", "DELETE", "/memos", `{"ids":[2,3],"updated_by":131}`, http.StatusOK, ``},
// 			genericTestcase{"DeleteMemoNoUpdater", "DELETE", "/memos", `{"ids":[1,2,3]}`, http.StatusBadRequest, `{"Error":"Updater Not Specified"}`},
// 			genericTestcase{"DeleteMemoNoID", "DELETE", "/memos", `{"updated_by":131}`, http.StatusBadRequest, `{"Error":"ID List Empty"}`},
// 		} {
// 			genericDoTest(testcase, t, asserter)
// 		}
// 	})

// 	if os.Getenv("db_driver") == "mysql" {
// 		_, _ = models.DB.Exec("truncate table projects;")
// 	} else {
// 		mockProjectDS = []models.Project{}
// 	}

// }
