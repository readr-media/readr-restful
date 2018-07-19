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

func (t *mockTagAPI) ToggleTags(args models.UpdateMultipleTagsArgs) error {
	for _, id := range args.IDs {
		for index, tags := range mockTagDS {
			if tags.ID == id {
				if args.Active == "0" {
					mockTagDS[index].Active = models.NullInt{1, true}
				} else {
					mockTagDS[index].Active = models.NullInt{0, true}
				}
			}
		}
	}
	return nil
}

func (t *mockTagAPI) GetTags(args models.GetTagsArgs) (tags []models.Tag, err error) {
	var result []models.Tag
	var offset = int(args.Page-1) * int(args.MaxResult)

	for _, t := range mockTagDS {
		if t.Active.Int != 0 {
			result = append(result, t)
		}
	}

	if args.Keyword != "" {
		newResult := []models.Tag{}
		for _, t := range result {
			if strings.HasPrefix(t.Text, args.Keyword) {
				newResult = append(newResult, t)
			}
		}
		result = newResult
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

func (t *mockTagAPI) InsertTag(tag models.Tag) (int, error) {
	index := len(mockTagDS) + 1
	for _, t := range mockTagDS {
		if t.Text == tag.Text && t.Active.Int == 1 {
			return 0, errors.New(`Duplicate Entry`)
		}
	}
	mockTagDS = append(mockTagDS, models.Tag{ID: index, Text: tag.Text, Active: models.NullInt{1, true}})
	return index, nil
}

func (t *mockTagAPI) UpdateTag(tag models.Tag) error {
	for _, t := range mockTagDS {
		if t.Text == tag.Text && t.Active.Int == 1 {
			return errors.New(`Duplicate Entry`)
		}
	}
	if tag.ID > len(mockTagDS) {
		return models.ItemNotFoundError
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

func (t *mockTagAPI) CountTags(args models.GetTagsArgs) (int, error) {
	var result []models.Tag

	for _, t := range mockTagDS {
		if t.Active.Int != 0 {
			result = append(result, t)
		}
	}

	return len(result), nil
}

func TestRouteTags(t *testing.T) {

	tags := []models.Tag{
		models.Tag{Text: "tag1", UpdatedBy: models.NullInt{931, true}},
		models.Tag{Text: "tag2", UpdatedBy: models.NullInt{931, true}},
		models.Tag{Text: "tag3", UpdatedBy: models.NullInt{931, true}},
		models.Tag{Text: "tag4", UpdatedBy: models.NullInt{931, true}},
	}

	for _, tag := range tags {
		_, err := models.TagAPI.InsertTag(tag)
		if err != nil {
			log.Printf("Init tag test fail %s", err.Error())
		}
	}

	for _, params := range []models.Post{
		models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, Author: models.NullInt{931, true}, UpdatedBy: models.NullInt{931, true}},
		models.Post{ID: 43, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, Author: models.NullInt{931, true}, UpdatedBy: models.NullInt{931, true}},
		models.Post{ID: 44, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, Author: models.NullInt{931, true}, UpdatedBy: models.NullInt{931, true}},
	} {
		_, err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Printf("Insert post fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Member{
		models.Member{ID: 931, MemberID: "AMI@mirrormedia.mg", Active: models.NullInt{1, true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b61e-f06e59fd8467"},
	} {
		_, err := models.MemberAPI.InsertMember(params)
		if err != nil {
			log.Printf("Insert member fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []struct {
		post_id int
		tag_ids []int
	}{
		{42, []int{1, 2}},
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
		for _, testcase := range []genericTestcase{
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
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}, RelatedReviews: models.NullInt{1, true}, RelatedNews: models.NullInt{0, true}},
			}},
			genericTestcase{"GetTagSortingOK", "GET", "/tags?keyword=tag&sort=updated_at", ``, http.StatusOK, []models.Tag{
				models.Tag{ID: 1, Text: "tag1", Active: models.NullInt{1, true}},
				models.Tag{ID: 2, Text: "tag2", Active: models.NullInt{1, true}},
				models.Tag{ID: 3, Text: "tag3", Active: models.NullInt{1, true}},
				models.Tag{ID: 4, Text: "tag4", Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetTagKeywordNotFound", "GET", "/tags?keyword=1024", ``, http.StatusOK, `{"_items":[]}`},
			genericTestcase{"GetTagUnknownSortingKey", "GET", "/tags?sort=unknown", ``, http.StatusBadRequest, `{"Error":"Bad Sorting Option"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})

	t.Run("CountTags", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"CountTagsOK", "GET", "/tags/count", ``, http.StatusOK, `{"_meta":{"total":4}}`},
			genericTestcase{"CountTagsOK", "GET", "/tags/count?keyowrd=tag", ``, http.StatusOK, `{"_meta":{"total":4}}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("InsertTag", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"PostTagOK", "POST", "/tags", `{"text":"insert1", "updated_by":931}`, http.StatusOK, `{"tag_id":5}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("UpdateTag", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"UpdateTagOK", "PUT", "/tags", `{"id":5, "text":"text5566", "updated_by":931}`, http.StatusOK, ``},
			genericTestcase{"UpdateTagNoSuchTag", "PUT", "/tags", `{"id":6, "text":"text7788", "updated_by":931}`, http.StatusBadRequest, `{"Error":"Item Not Found"}`},
			genericTestcase{"UpdateTagDupe", "PUT", "/tags", `{"id":2, "text":"tag3", "updated_by":931}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
			genericTestcase{"UpdateTagNoID", "PUT", "/tags", `{"text":"tag3"}`, http.StatusBadRequest, `{"Error":"Updater Not Sepcified"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("DaleteTags", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"DeleteTagOK", "DELETE", "/tags?ids=[1, 2, 3, 4]&updated_by=AMI@mirrormedia.mg", ``, http.StatusOK, ``},
			genericTestcase{"DeleteTagWithoutUpdater", "DELETE", "/tags?ids=[1, 2, 3, 4]", ``, http.StatusBadRequest, `{"Error":"Bad Updater"}`},
			genericTestcase{"DeleteTagNoIds", "DELETE", "/tags?", ``, http.StatusBadRequest, `{"Error":"Bad Tag IDs"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("InsertDupeTag", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"PostTagDupe", "POST", "/tags", `{"text":"text5566","updated_by":931}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
			genericTestcase{"PostSameAsInactiveTagOK", "POST", "/tags", `{"text":"tag1", "updated_by":931}`, http.StatusOK, `{"tag_id":6}`},
			genericTestcase{"PostSameAsActiveTag", "POST", "/tags", `{"text":"text5566", "updated_by":931}`, http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	/*
		t.Run("GetPostWithTags", func(t *testing.T) {
			for _, testcase := range []genericTestcase{
				genericTestcase{"GetPostWithTagsOK", "GET", "/post/43", ``, http.StatusOK, `{"_items":[{"tags":[{"id":"1","text":"text5566"},{"id":"2","text":"tag2"}],"id":43,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"author":{"id":931,"name":null,"nickname":null,"birthday":null,"gender":null,"work":null,"mail":null,"register_mode":null,"social_id":null,"talk_id":null,"created_at":null,"updated_at":null,"updated_by":null,"description":null,"profile_image":null,"identity":null,"role":null,"active":1,"custom_editor":null,"hide_profile":null,"profile_push":null,"post_push":null,"comment_push":null},"updated_by":{"id":931,"name":null,"nickname":null,"birthday":null,"gender":null,"work":null,"mail":null,"register_mode":null,"social_id":null,"talk_id":null,"created_at":null,"updated_at":null,"updated_by":null,"description":null,"profile_image":null,"identity":null,"role":null,"active":1,"custom_editor":null,"hide_profile":null,"profile_push":null,"post_push":null,"comment_push":null}}]}`},
			} {
				genericDoTest(testcase, t, asserter)
			}
		})
	*/
}
