package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"testing"

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
		if v.ID == id {
			return v, nil
		}
	}
	return models.Memo{}, errors.New("Not Found")
}
func (m *mockMemoAPI) GetMemos(args *models.MemoGetArgs) (memos []models.Memo, err error) {
	switch {
	case len(args.Author) > 0 && len(args.Project) > 0:
		return []models.Memo{
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
		}, nil
	case len(args.Author) > 0 && len(args.Project) == 0:
		return []models.Memo{
			models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
		}, nil
	case len(args.Author) == 0 && len(args.Project) > 0:
		return []models.Memo{
			models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
		}, nil
	case args.Sorting == "memo_id":
		return []models.Memo{
			models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
		}, nil
	case args.Page == 2:
		return []models.Memo{
			models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
		}, nil
	case args.MaxResult == 1:
		return []models.Memo{
			models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
		}, nil
	default:
		return []models.Memo{
			models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
		}, nil
	}
	return []models.Memo{}, nil
}
func (m *mockMemoAPI) InsertMemo(memo models.Memo) (err error) {
	if memo.ID == 0 {
		memo.ID = len(mockMemoDS) + 1
	}
	for _, v := range mockMemoDS {
		if v.ID == memo.ID {
			return errors.New("Duplicate entry")
		}
	}
	mockMemoDS = append(mockMemoDS, memo)

	return nil
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

func (a *mockMemoAPI) SchedulePublish() error {
	return nil
}

func TestRouteMemos(t *testing.T) {
	for _, memo := range []models.Memo{
		models.Memo{Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
		models.Memo{Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
		models.Memo{Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
		models.Memo{Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
	} {
		err := models.MemoAPI.InsertMemo(memo)
		if err != nil {
			log.Printf("Init memo test fail %s", err.Error())
		}
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.Memo `json:"_items"`
		}

		var Response response
		var expected []models.Memo = tc.resp.([]models.Memo)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", resp)
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect memo length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		}

		for i, respmemo := range Response.Items {
			expmemo := expected[i]
			if respmemo.ID == expmemo.ID &&
				respmemo.Title == expmemo.Title &&
				respmemo.Active == expmemo.Active &&
				respmemo.Author == expmemo.Author &&
				respmemo.Project == expmemo.Project {
				continue
			}
			log.Println(respmemo.ID, expmemo.ID)
			log.Println(respmemo.Title, expmemo.Title)
			log.Println(respmemo.Active, expmemo.Active)
			log.Println(respmemo.Author, expmemo.Author)
			log.Println(respmemo.Project, expmemo.Project)

			t.Errorf("%s, expect to get %v, but %v ", tc.name, expmemo, respmemo)
		}
	}

	t.Run("InsertMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"InsertMemoOK", "POST", "/memo", `{"id":100,"title":"MemoTest2","content":"MemoTest2","author":132, "project_id":421}`, http.StatusOK, ``},
			genericTestcase{"InsertMemoOK", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":131, "project_id":420}`, http.StatusOK, ``},
			genericTestcase{"InsertMemoDupe", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":131, "project_id":420}`, http.StatusBadRequest, `{"Error":"Memo ID Already Taken"}`},
			genericTestcase{"InsertMemoNoProject", "POST", "/memo", `{"title":"MemoTest1","author":131}`, http.StatusBadRequest, `{"Error":"Invalid Project"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PutMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"PutMemoOK", "PUT", "/memo", `{"id":101,"title":"MemoTestMod","updated_by":131}`, http.StatusOK, ``},
			genericTestcase{"PutMemoUTF8", "PUT", "/memo", `{"id":101,"title":"順便測中文","updated_by":133}`, http.StatusOK, ``},
			genericTestcase{"PutMemoScheduleNoTime", "PUT", "/memo", `{"id":100,"updated_by":131,"publish_status":3}`, http.StatusBadRequest, `{"Error":"Invalid Publish Time"}`},
			genericTestcase{"PutMemoSchedule", "PUT", "/memo", `{"id":100,"updated_by":131,"publish_status":3,"published_at":"2046-01-05T00:42:42+00:00"}`, http.StatusOK, ``}, //published_at is time string in RFC3339 format
			genericTestcase{"PutMemoPublishNoContent", "PUT", "/memo", `{"id":101,"updated_by":131,"publish_status":2}`, http.StatusBadRequest, `{"Error":"Invalid Memo Content"}`},
			genericTestcase{"PutMemoNoUpdater", "PUT", "/memo", `{"id":101,"title":"NoUpdater"}`, http.StatusBadRequest, `{"Error":"Neither updated_by or author is valid"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetMemoOK", "GET", "/memo/1", ``, http.StatusOK, `{"_items":{"id":1,"created_at":null,"comment_amount":null,"title":"MemoTestDefault1","content":null,"link":null,"author":131,"project_id":420,"active":1,"updated_at":null,"updated_by":null,"published_at":null,"publish_status":null,"memo_order":null}}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMemos", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetMemoDefaultOK", "GET", "/memos", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1&page=2", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetMemoSortOK", "GET", "/memos?sort=memo_id", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullInt{131, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetMemoSortInvalidOption", "GET", "/memos?sort=meow", ``, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`},
			genericTestcase{"GetMemoFilterAuthor", "GET", `/memos?author=[135,132]`, ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 100, Title: models.NullString{"MemoTest2", true}, Author: models.NullInt{132, true}, Project: models.NullInt{421, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullInt{135, true}, Project: models.NullInt{420, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetMemoFilterProject", "GET", `/memos?project_id=[422]`, ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullInt{131, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetMemoFilterMultipleCondition", "GET", `/memos?active={"$nin":[0]}&author=[135,132]&project_id=[422]`, ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullInt{135, true}, Project: models.NullInt{422, true}, Active: models.NullInt{1, true}},
			}},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMemoCount", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetMemoCountOK", "GET", "/memos/count", ``, http.StatusOK, `{"_meta":{"total":6}}`},
			genericTestcase{"GetMemoCountFilterAuthor", "GET", `/memos/count?author=[135,132]`, ``, http.StatusOK, `{"_meta":{"total":3}}`},
			genericTestcase{"GetMemoCountFilterProject", "GET", `/memos/count?project_id=[422]`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
			genericTestcase{"GetMemoCountFilterMultipleCondition", "GET", `/memos/count?active={"$nin":[0]}&author=[135,132]&project_id=[422]`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("DeleteMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"DeleteMemoOK", "DELETE", "/memo/1", ``, http.StatusOK, ``},
			genericTestcase{"DeleteMemoOK", "DELETE", "/memos", `{"ids":[2,3],"updated_by":131}`, http.StatusOK, ``},
			genericTestcase{"DeleteMemoNoUpdater", "DELETE", "/memos", `{"ids":[1,2,3]}`, http.StatusBadRequest, `{"Error":"Updater Not Specified"}`},
			genericTestcase{"DeleteMemoNoID", "DELETE", "/memos", `{"updated_by":131}`, http.StatusBadRequest, `{"Error":"ID List Empty"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
}
