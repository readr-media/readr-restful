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
	log.Println(args)
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
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
		}, nil
	case len(args.Author) > 0 && len(args.Project) == 0:
		return []models.Memo{
			models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
		}, nil
	case len(args.Author) == 0 && len(args.Project) > 0:
		return []models.Memo{
			models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
		}, nil
	case args.Sorting == "memo_id":
		return []models.Memo{
			models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
		}, nil
	case args.Page == 2:
		return []models.Memo{
			models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
		}, nil
	case args.MaxResult == 1:
		return []models.Memo{
			models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
		}, nil
	default:
		return []models.Memo{
			models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
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

func TestRouteMemos(t *testing.T) {
	for _, memo := range []models.Memo{
		models.Memo{Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
		models.Memo{Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
		models.Memo{Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
		models.Memo{Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
	} {
		err := models.MemoAPI.InsertMemo(memo)
		if err != nil {
			log.Printf("Init tag test fail %s", err.Error())
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

		log.Println(Response)
		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		}

		for i, resptag := range Response.Items {
			exptag := expected[i]
			if resptag.ID == exptag.ID &&
				resptag.Title == exptag.Title &&
				resptag.Active == exptag.Active &&
				resptag.Author == exptag.Author &&
				resptag.Project == exptag.Project {
				continue
			}

			t.Errorf("%s, expect to get %v, but %v ", tc.name, exptag, resptag)
		}
	}

	t.Run("InsertMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"InsertMemoOK", "POST", "/memo", `{"title":"MemoTest2","author":"EMII", "project_id":421}`, http.StatusOK, ``},
			genericTestcase{"InsertMemoOK", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":"EMI", "project_id":420}`, http.StatusOK, ``},
			genericTestcase{"InsertMemoDupe", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":"EMI", "project_id":420}`, http.StatusBadRequest, `{"Error":"Post ID Already Taken"}`},
			genericTestcase{"InsertMemoNoAuthor", "POST", "/memo", `{"id":102,"title":"MemoTest2"}`, http.StatusBadRequest, `{"Error":"Invalid Author"}`},
			genericTestcase{"InsertMemoNoProject", "POST", "/memo", `{"id":101,"title":"MemoTest1","author":"EMI"}`, http.StatusBadRequest, `{"Error":"Invalid Project"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PutMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"PutMemoOK", "PUT", "/memo", `{"id":101,"title":"MemoTestMod","updated_by":"EMI"}`, http.StatusOK, ``},
			genericTestcase{"PutMemoUTF8", "PUT", "/memo", `{"id":101,"title":"順便測中文","updated_by":"EMIII"}`, http.StatusOK, ``},
			genericTestcase{"PutMemoNoUpdater", "PUT", "/memo", `{"id":101,"title":"NoUpdater"}`, http.StatusBadRequest, `{"Error":"Neither updated_by or author is valid"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetMemoOK", "GET", "/memo/1", ``, http.StatusOK, `{"_items":{"id":1,"created_at":null,"comment_amount":null,"title":"MemoTestDefault1","content":null,"link":null,"author":"EMI","project_id":420,"active":2,"updated_at":null,"updated_by":null,"published_at":null}}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMemos", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetMemoDefaultOK", "GET", "/memos", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			}},
			genericTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
			}},
			genericTestcase{"GetMemoMaxresultOK", "GET", "/memos?max_result=1&page=2", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			}},
			genericTestcase{"GetMemoSortOK", "GET", "/memos?sort=memo_id", ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 1, Title: models.NullString{"MemoTestDefault1", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 101, Title: models.NullString{"順便測中文", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
			}},
			genericTestcase{"GetMemoSortInvalidOption", "GET", "/memos?sort=meow", ``, http.StatusBadRequest, `{"Error":"Invalid Parameters"}`},
			genericTestcase{"GetMemoFilterAuthor", "GET", `/memos?author=["EMV","EMII"]`, ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 5, Title: models.NullString{"MemoTest2", true}, Author: models.NullString{"EMII", true}, Project: models.NullInt{421, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 2, Title: models.NullString{"MemoTestDefault2", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{420, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			}},
			genericTestcase{"GetMemoFilterProject", "GET", `/memos?project_id=[422]`, ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 3, Title: models.NullString{"MemoTestDefault3", true}, Author: models.NullString{"EMI", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			}},
			genericTestcase{"GetMemoFilterMultipleCondition", "GET", `/memos?active={"$nin":[0,1]}&author=["EMV","EMII"]&project_id=[422]`, ``, http.StatusOK, []models.Memo{
				models.Memo{ID: 4, Title: models.NullString{"MemoTestDefault4", true}, Author: models.NullString{"EMV", true}, Project: models.NullInt{422, true}, Active: models.NullInt{2, true}},
			}},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMemoCount", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"GetMemoCountOK", "GET", "/memos/count", ``, http.StatusOK, `{"_meta":{"total":6}}`},
			genericTestcase{"GetMemoCountFilterAuthor", "GET", `/memos/count?author=["EMV","EMII"]`, ``, http.StatusOK, `{"_meta":{"total":3}}`},
			genericTestcase{"GetMemoCountFilterProject", "GET", `/memos/count?project_id=[422]`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
			genericTestcase{"GetMemoCountFilterMultipleCondition", "GET", `/memos/count?active={"$nin":[0,1]}&author=["EMV","EMII"]&project_id=[422]`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PublishMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"PublishMemoOK", "PUT", "/memos", `{"ids":[1,2,3],"updated_by":"EMI"}`, http.StatusOK, ``},
			genericTestcase{"PublishMemoNoUpdater", "PUT", "/memos", `{"ids":[1,2,3]}`, http.StatusBadRequest, `{"Error":"Updater Not Specified"}`},
			genericTestcase{"PublishMemoNoID", "PUT", "/memos", `{"updated_by":"EMI"}`, http.StatusBadRequest, `{"Error":"Empty Memo ID"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("DeleteMemo", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"DeleteMemoOK", "DELETE", "/memos", `{"ids":[1,2,3],"updated_by":"EMI"}`, http.StatusOK, ``},
			genericTestcase{"DeleteMemoNoUpdater", "PUT", "/memos", `{"ids":[1,2,3]}`, http.StatusBadRequest, `{"Error":"Updater Not Specified"}`},
			genericTestcase{"DeleteMemoNoID", "PUT", "/memos", `{"updated_by":"EMI"}`, http.StatusBadRequest, `{"Error":"Empty Memo ID"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
}
