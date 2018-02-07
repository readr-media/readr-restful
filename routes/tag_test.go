package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

var mockTagDS []models.Tag
var mockPostTagDS []map[string]int

type mockTagAPI struct{}

func (t *mockTagAPI) ToggleTags(ids []int, active string) error { return nil }

func (t *mockTagAPI) GetTags(args models.GetTagsArgs) (tags []models.Tag, err error) {
	var result = mockTagDS
	var offset = int(args.Page-1) * int(args.MaxResult)

	if args.Keyword != "" {
		result = []models.Tag{}
		for _, t := range mockTagDS {
			if strings.HasPrefix(t.Text, args.Keyword) {
				result = append(result, t)
			}
		}
	}

	if offset > len(mockTagDS) {
		result = []models.Tag{}
	} else {
		result = result[offset:]
	}

	if len(mockTagDS) > int(args.MaxResult) {
		result = result[:args.MaxResult]
	}

	for i, _ := range result {
		if args.ShowStats == false {
			result[i].RelatedReviews.Valid = false
			result[i].RelatedNews.Valid = false
		} else {
			result[i].RelatedReviews.Valid = true
			result[i].RelatedNews.Valid = true
		}
	}
	return result, err
}

func (t *mockTagAPI) InsertTag(text string) (int, error) {
	index := len(mockTagDS) + 1
	for _, t := range mockTagDS {
		if t.Text == text {
			return 0, errors.New(`Duplicate Entry`)
		}
	}
	mockTagDS = append(mockTagDS, models.Tag{ID: index, Text: text, Active: models.NullInt{1, true}})
	return index, nil
}

func (t *mockTagAPI) UpdateTag(tag models.Tag) error {
	for _, t := range mockTagDS {
		if t.Text == tag.Text {
			return errors.New(`Duplicate Entry`)
		}
	}
	mockTagDS[tag.ID-1].Text = tag.Text
	return nil
}

func (t *mockTagAPI) UpdatePostTags(postId int, tag_ids []int) error {
	for _, tag_id := range tag_ids {
		for i, t := range mockTagDS {
			if t.ID == tag_id {
				if postId%2 == 0 {
					mockTagDS[i].RelatedReviews = models.NullInt{t.RelatedReviews.Int + 1, true}
				} else {
					mockTagDS[i].RelatedNews = models.NullInt{t.RelatedNews.Int + 1, true}
				}
			}
		}
	}
	return nil
}

type tagtc struct {
	name     string
	method   string
	url      string
	body     string
	httpcode int
	resp     interface{}
}

func TestTags(t *testing.T) {

	tags := []models.Tag{
		models.Tag{Text: "tag1"},
		models.Tag{Text: "tag2"},
		models.Tag{Text: "tag3"},
		models.Tag{Text: "tag4"},
	}

	for _, tag := range tags {
		_, err := models.TagAPI.InsertTag(tag.Text)
		if err != nil {
			log.Printf("Init tag test fail %s", err.Error())
		}
	}

	for _, params := range []models.Post{
		models.Post{ID: 43, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}},
		models.Post{ID: 44, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}},
	} {
		_, err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Printf("Insert post fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []struct {
		post_id int
		tag_ids []int
	}{
		{43, []int{1, 2}},
		{44, []int{1, 3}},
	} {
		err := models.TagAPI.UpdatePostTags(params.post_id, params.tag_ids)
		if err != nil {
			log.Printf("Insert post tag fail when init test case. Error: %v", err)
		}
	}

	t.Run("GetTags", func(t *testing.T) {
		testcases := []tagtc{
			tagtc{"GetTagBasicOK", "GET", "/tags?stats=0", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
				models.Tag{ID: 3, Text: "tag3", Active: models.NullInt{1, true}},
				models.Tag{ID: 4, Text: "tag4", Active: models.NullInt{1, true}},
			}},
			tagtc{"GetTagMaxresultOK", "GET", "/tags?stats=0&max_result=1", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
			}},
			tagtc{"GetTagPaginationOK", "GET", "/tags?stats=0&max_result=1&page=2", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
			}},
			tagtc{"GetTagKeywordAndStatsOK", "GET", "/tags?stats=1&keyword=tag2", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}, RelatedReviews: models.NullInt{0, true}, RelatedNews: models.NullInt{1, true}},
			}},
			tagtc{"GetTagSortingOK", "GET", "/tags?keyword=tag&sort=updated_at", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
				models.Tag{ID: 3, Text: "tag3", Active: models.NullInt{1, true}},
				models.Tag{ID: 4, Text: "tag4", Active: models.NullInt{1, true}},
			}},
			tagtc{"GetTagKeywordNotFound", "GET", "/tags?keyword=1024", ``, http.StatusOK, `{"_items":[]}`},
			tagtc{"GetTagUnknownSortingKey", "GET", "/tags?sort=unknown", ``, http.StatusBadRequest, `{"Error":"Bad Sorting Option"}`},
		}
		for _, tc := range testcases {
			doTagTest(tc, t)
		}
	})
	t.Run("InsertTag", func(t *testing.T) {
		testcases := []tagtc{
			tagtc{"PostTagOK", "POST", "/tags", `{"name":"insert1"}`, http.StatusOK, `{"tag_id":5}`},
			tagtc{"PostTagDupe", "POST", "/tags", `{"name":"insert1"}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
		}
		for _, tc := range testcases {
			doTagTest(tc, t)
		}
	})
	t.Run("UpdateTag", func(t *testing.T) {
		testcases := []tagtc{
			tagtc{"UpdateTagOK", "PUT", "/tags", `{"id":1, "text":"text5566"}`, http.StatusOK, ``},
			tagtc{"UpdateTagDupe", "PUT", "/tags", `{"id":2, "text":"tag3"}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
		}
		for _, tc := range testcases {
			doTagTest(tc, t)
		}
	})
	t.Run("DaleteTags", func(t *testing.T) {
		testcases := []tagtc{
			tagtc{"DeleteTagOK", "DELETE", "/tags?ids=[1, 2, 3, 4]", ``, http.StatusOK, ``},
		}
		for _, tc := range testcases {
			doTagTest(tc, t)
		}
	})
}

func doTagTest(tc tagtc, t *testing.T) {
	t.Run(tc.name, func(t *testing.T) {
		w := httptest.NewRecorder()
		var jsonStr = []byte(tc.body)
		req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != tc.httpcode {
			t.Errorf("%s want %d but get %d", tc.name, tc.httpcode, w.Code)
		}
		switch tc.resp.(type) {
		case string:
			if w.Body.String() != tc.resp {
				t.Errorf("%s expect (error) message %v but get %v", tc.name, tc.resp, w.Body.String())
			}
		default:
			type response struct {
				Items []models.Tag `json:"_items"`
			}

			var Response response
			var expected []models.Tag = tc.resp.([]models.Tag)

			err := json.Unmarshal([]byte(w.Body.String()), &Response)
			if err != nil {
				fmt.Println("active ", err.Error())
			}

			if len(Response.Items) != len(expected) {
				t.Errorf("%s expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
			}

		OuterLoop:
			for _, resptag := range Response.Items {
				for _, exptag := range expected {
					if resptag.ID == exptag.ID &&
						resptag.Text == exptag.Text &&
						resptag.Active == exptag.Active &&
						resptag.RelatedReviews == exptag.RelatedReviews &&
						resptag.RelatedNews == exptag.RelatedNews {
						continue OuterLoop
					}
				}
				t.Errorf("%s, Not expect to get %v ", tc.name, resptag)

			}
		}

	})
}
