package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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

func TestRouteTags(t *testing.T) {

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
		models.Post{ID: 43, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, Author: models.NullString{"AMI@mirrormedia.mg", true}, UpdatedBy: models.NullString{"AMI@mirrormedia.mg", true}},
		models.Post{ID: 44, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, Author: models.NullString{"AMI@mirrormedia.mg", true}, UpdatedBy: models.NullString{"AMI@mirrormedia.mg", true}},
	} {
		_, err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Printf("Insert post fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Member{
		models.Member{ID: "AMI@mirrormedia.mg", Active: models.NullInt{1, true}},
	} {
		err := models.MemberAPI.InsertMember(params)
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
		if err := models.TagAPI.UpdatePostTags(params.post_id, params.tag_ids); err != nil {
			log.Printf("Insert post tag fail when init test case. Error: %v", err)
		}
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.Tag `json:"_items"`
		}

		var Response response
		var expected []models.Tag = tc.resp.([]models.Tag)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", resp)
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

	t.Run("GetTags", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"GetTagBasicOK", "GET", "/tags?stats=0", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
				models.Tag{ID: 3, Text: "tag3", Active: models.NullInt{1, true}},
				models.Tag{ID: 4, Text: "tag4", Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetTagMaxresultOK", "GET", "/tags?stats=0&max_result=1", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetTagPaginationOK", "GET", "/tags?stats=0&max_result=1&page=2", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetTagKeywordAndStatsOK", "GET", "/tags?stats=1&keyword=tag2", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}, RelatedReviews: models.NullInt{0, true}, RelatedNews: models.NullInt{1, true}},
			}},
			genericTestcase{"GetTagSortingOK", "GET", "/tags?keyword=tag&sort=updated_at", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
				models.Tag{ID: 3, Text: "tag3", Active: models.NullInt{1, true}},
				models.Tag{ID: 4, Text: "tag4", Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetTagKeywordNotFound", "GET", "/tags?keyword=1024", ``, http.StatusOK, `{"_items":[]}`},
			genericTestcase{"GetTagUnknownSortingKey", "GET", "/tags?sort=unknown", ``, http.StatusBadRequest, `{"Error":"Bad Sorting Option"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("InsertTag", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"PostTagOK", "POST", "/tags", `{"name":"insert1"}`, http.StatusOK, `{"tag_id":5}`},
			genericTestcase{"PostTagDupe", "POST", "/tags", `{"name":"insert1"}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("UpdateTag", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"UpdateTagOK", "PUT", "/tags", `{"id":1, "text":"text5566"}`, http.StatusOK, ``},
			genericTestcase{"UpdateTagDupe", "PUT", "/tags", `{"id":2, "text":"tag3"}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("DaleteTags", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"DeleteTagOK", "DELETE", "/tags?ids=[1, 2, 3, 4]", ``, http.StatusOK, ``},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	/*
		t.Run("GetPostWithTags", func(t *testing.T) {
			testcases := []genericTestcase{
				genericTestcase{"GetPostWithTagsOK", "GET", "/post/43", ``, http.StatusOK, `{"_items":[{"tags":[{"id":"1","text":"text5566"},{"id":"2","text":"tag2"}],"id":43,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"author":{"id":"AMI@mirrormedia.mg","name":null,"nickname":null,"birthday":null,"gender":null,"work":null,"mail":null,"register_mode":null,"social_id":null,"talk_id":null,"created_at":null,"updated_at":null,"updated_by":null,"description":null,"profile_image":null,"identity":null,"role":null,"active":1,"custom_editor":null,"hide_profile":null,"profile_push":null,"post_push":null,"comment_push":null},"updated_by":{"id":"AMI@mirrormedia.mg","name":null,"nickname":null,"birthday":null,"gender":null,"work":null,"mail":null,"register_mode":null,"social_id":null,"talk_id":null,"created_at":null,"updated_at":null,"updated_by":null,"description":null,"profile_image":null,"identity":null,"role":null,"active":1,"custom_editor":null,"hide_profile":null,"profile_push":null,"post_push":null,"comment_push":null}}]}`},
			}
			for _, tc := range testcases {
				genericDoTest(tc, t, asserter)
			}
		})
	*/
}
