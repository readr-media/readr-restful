package routes

import (
	"bytes"
	"errors"
	"log"
	"testing"
	"time"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/readr-media/readr-restful/models"
)

type mockFollowingAPI struct{}

type followDS struct {
	ID     int64
	Object int64
}

var mockFollowingDS = map[string][]followDS{
	"post":    []followDS{},
	"member":  []followDS{},
	"project": []followDS{},
}

func (a *mockFollowingAPI) AddFollowing(params models.FollowArgs) error {
	store, ok := mockFollowingDS[params.Resource]
	if !ok {
		return errors.New("Resource Not Supported")
	}

	store = append(store, followDS{ID: params.Subject, Object: params.Object})
	return nil
}
func (a *mockFollowingAPI) DeleteFollowing(params models.FollowArgs) error {
	store, ok := mockFollowingDS[params.Resource]
	if !ok {
		return errors.New("Resource Not Supported")
	}

	for index, follow := range store {
		if follow.ID == params.Subject && follow.Object == params.Object {
			store = append(store[:index], store[index+1:]...)
		}
	}
	return nil
}
func (a *mockFollowingAPI) GetFollowing(params models.GetFollowingArgs) (followings []interface{}, err error) {

	switch {
	case params.MemberId == 0:
		return nil, errors.New("Not Found")
	case params.Resource == "member":
		return []interface{}{
			models.Member{ID: 72, MemberID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}},
		}, nil
	case params.Resource == "post":
		switch params.Type {
		case "":
			return []interface{}{
				models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{71, true}, PublishStatus: models.NullInt{2, true}},
				models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{72, true}, PublishStatus: models.NullInt{2, true}},
			}, nil
		case "review":
			return []interface{}{
				models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{71, true}, PublishStatus: models.NullInt{2, true}},
			}, nil
		case "news":
			return []interface{}{
				models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{72, true}, PublishStatus: models.NullInt{2, true}},
			}, nil
		}
	case params.Resource == "project":
		return []interface{}{
			models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
		}, nil
	default:
		return nil, nil
	}
	return nil, nil
}
func (a *mockFollowingAPI) GetFollowed(args models.GetFollowedArgs) (interface{}, error) {
	type followedCount struct {
		Resourceid int
		Count      int
		Follower   []int
	}
	switch {
	case args.Ids[0] == "1001":
		return []followedCount{}, nil
	case args.Resource == "member":
		return []followedCount{
			followedCount{71, 1, []int{72}},
			followedCount{72, 1, []int{71}},
		}, nil
	case args.Resource == "post":
		switch args.Type {
		case "":
			return []followedCount{
				followedCount{42, 2, []int{71, 72}},
				followedCount{84, 1, []int{71}},
			}, nil
		case "review":
			return []followedCount{
				followedCount{42, 2, []int{71, 72}},
			}, nil
		case "news":
			return []followedCount{
				followedCount{84, 1, []int{71}},
			}, nil
		}
		return nil, nil
	case args.Resource == "project":
		switch len(args.Ids) {
		case 1:
			return []followedCount{
				followedCount{840, 1, []int{72}},
			}, nil
		case 2:
			return []followedCount{
				followedCount{420, 2, []int{71, 72}},
				followedCount{840, 1, []int{72}},
			}, nil
		}
		return nil, nil
	default:
		return nil, nil
	}
}
func (*mockFollowingAPI) GetFollowMap(args models.GetFollowMapArgs) (list []models.FollowingMapItem, err error) {
	switch {
	case args.Resource == "member":
		return []models.FollowingMapItem{
			models.FollowingMapItem{[]string{"72"}, []string{"42"}},
			models.FollowingMapItem{[]string{"71"}, []string{"84"}},
		}, nil
	case args.Resource == "post":
		switch args.Type {
		case "":
			return []models.FollowingMapItem{
				models.FollowingMapItem{[]string{"72"}, []string{"42"}},
				models.FollowingMapItem{[]string{"71"}, []string{"42", "84"}},
			}, nil
		case "review":
			return []models.FollowingMapItem{
				models.FollowingMapItem{[]string{"71", "72"}, []string{"42"}},
			}, nil
		case "news":
			return []models.FollowingMapItem{
				models.FollowingMapItem{[]string{"71"}, []string{"84"}},
			}, nil
		}
		return []models.FollowingMapItem{}, nil
	case args.Resource == "project":
		return []models.FollowingMapItem{
			models.FollowingMapItem{[]string{"71"}, []string{"420"}},
			models.FollowingMapItem{[]string{"72"}, []string{"420", "840"}},
		}, nil
	default:
		return []models.FollowingMapItem{}, errors.New("Resource Not Supported")
	}
}

func initFollowTest() {
	mockPostDSBack = mockPostDS

	for _, params := range []models.Member{
		models.Member{ID: 70, MemberID: "followtest0@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest0@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b58e-f06e59fd8467"},
		models.Member{ID: 71, MemberID: "followtest1@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest1@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b59e-f06e59fd8467"},
		models.Member{ID: 72, MemberID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b60e-f06e59fd8467"},
	} {
		_, err := models.MemberAPI.InsertMember(params)
		if err != nil {
			log.Printf("Insert member fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Post{
		models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{71, true}, PublishStatus: models.NullInt{2, true}},
		models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{72, true}, PublishStatus: models.NullInt{2, true}},
	} {
		_, err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Printf("Insert post fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Project{
		models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
		models.Project{ID: 840, PostID: 84, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2016, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
	} {
		err := models.ProjectAPI.InsertProject(params)
		if err != nil {
			log.Printf("Insert Project fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.FollowArgs{
		models.FollowArgs{Resource: "member", Subject: 71, Object: 72},
		models.FollowArgs{Resource: "post", Subject: 71, Object: 42},
		models.FollowArgs{Resource: "post", Subject: 71, Object: 84},
		models.FollowArgs{Resource: "project", Subject: 71, Object: 420},
		models.FollowArgs{Resource: "member", Subject: 72, Object: 71},
		models.FollowArgs{Resource: "post", Subject: 72, Object: 42},
		models.FollowArgs{Resource: "project", Subject: 72, Object: 420},
		models.FollowArgs{Resource: "project", Subject: 72, Object: 840},
	} {
		err := models.FollowingAPI.AddFollowing(params)
		if err != nil {
			log.Printf("Init test case fail. Error: %v", err)
		}
	}
}

func clearFollowTest() {
	//restore the backuped data
	mockPostDS = mockPostDSBack
}

func TestFollowingGet(t *testing.T) {

	initFollowTest()

	type CaseIn struct {
		Resource     string `json:"resource,omitempty"`
		ResourceType string `json:"resource_type, omitempty"`
		Subject      int64  `json:"subject,omitempty"`
	}

	type CaseOut struct {
		httpcode int
		resp     string
	}

	var TestRouteName = "/following/byuser"
	var TestRouteMethod = "GET"

	var TestFollowingGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{"GetFollowingPostOK", CaseIn{"post", "", 71}, CaseOut{http.StatusOK, `[{"id":42,"author":71,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":0,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2},{"id":84,"author":72,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2}]`}},
		{"GetFollowingPostReviewOK", CaseIn{"post", "review", 71}, CaseOut{http.StatusOK, `[{"id":42,"author":71,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":0,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2}]`}},
		{"GetFollowingPostNewsOK", CaseIn{"post", "news", 71}, CaseOut{http.StatusOK, `[{"id":84,"author":72,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2}]`}},
		{"GetFollowingProjectOK", CaseIn{"project", "", 71}, CaseOut{http.StatusOK, `[{"id":420,"created_at":null,"updated_at":"2015-11-10T23:00:00Z","updated_by":null,"published_at":null,"post_id":42,"like_amount":null,"comment_amount":null,"active":1,"hero_image":null,"title":null,"description":null,"author":null,"og_title":null,"og_description":null,"og_image":null,"project_order":null,"status":null,"slug":null,"views":null,"publish_status":2,"progress":null,"memo_points":null}]`}},
		{"GetFollowingFollowerNotExist", CaseIn{"project", "", 0}, CaseOut{http.StatusOK, `[]`}},
	}

	for _, testcase := range TestFollowingGetCases {

		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Body.String() != testcase.out.resp {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.resp, w.Body.String(), testcase.name)
			t.Fail()
		}

	}

	clearFollowTest()
}

func TestFollowedGet(t *testing.T) {

	type CaseIn struct {
		ResourceName string   `json:"resource",omitempty`
		ResourceType string   `json:"resource_type",omitempty`
		ResourceId   []string `json:"ids",omitempty`
	}

	type CaseOut struct {
		httpcode int
		resp     string
	}

	var TestRouteName = "/following/byresource"
	var TestRouteMethod = "GET"

	var TestFollowingGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{"GetFollowedPostOK", CaseIn{"post", "", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":42,"Count":2,"Follower":[71,72]},{"Resourceid":84,"Count":1,"Follower":[71]}]`}},
		{"GetFollowedPostReviewOK", CaseIn{"post", "review", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":42,"Count":2,"Follower":[71,72]}]`}},
		{"GetFollowedPostNewsOK", CaseIn{"post", "news", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":84,"Count":1,"Follower":[71]}]`}},
		{"GetFollowedMemberOK", CaseIn{"member", "", []string{"71", "72"}}, CaseOut{http.StatusOK, `[{"Resourceid":71,"Count":1,"Follower":[72]},{"Resourceid":72,"Count":1,"Follower":[71]}]`}},
		{"GetFollowedProjectSingleOK", CaseIn{"project", "", []string{"840"}}, CaseOut{http.StatusOK, `[{"Resourceid":840,"Count":1,"Follower":[72]}]`}},
		{"GetFollowedProjectOK", CaseIn{"project", "", []string{"420", "840"}}, CaseOut{http.StatusOK, `[{"Resourceid":420,"Count":2,"Follower":[71,72]},{"Resourceid":840,"Count":1,"Follower":[72]}]`}},
		{"GetFollowedMissingResource", CaseIn{"", "", []string{"420", "840"}}, CaseOut{http.StatusBadRequest, `{"Error":"Unsupported Resource"}`}},
		{"GetFollowedMissingID", CaseIn{"post", "", []string{}}, CaseOut{http.StatusBadRequest, `{"Error":"Bad Resource ID"}`}},
		{"GetFollowedPostNotExist", CaseIn{"post", "", []string{"1001", "1000"}}, CaseOut{http.StatusOK, `[]`}},
		{"GetFollowedPostStringID", CaseIn{"post", "", []string{"unintegerable"}}, CaseOut{http.StatusBadRequest, `{"Error":"Bad Resource ID"}`}},
		{"GetFollowedProjectStringID", CaseIn{"project", "", []string{"unintegerable"}}, CaseOut{http.StatusBadRequest, `{"Error":"Bad Resource ID"}`}},
	}

	for _, testcase := range TestFollowingGetCases {

		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Body.String() != testcase.out.resp {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.resp, w.Body.String(), testcase.name)
			t.Fail()
		}
	}

	clearFollowTest()
}

func TestFollowMap(t *testing.T) {

	type CaseIn struct {
		ResourceName string    `json:"resource",omitempty`
		ResourceType string    `json:"resource_type",omitempty`
		UpdatedAfter time.Time `json:"updated_after",omitempty`
	}

	type CaseOut struct {
		httpcode int
		resp     string
	}

	var TestRouteName = "/following/map"
	var TestRouteMethod = "GET"

	var TestFollowingGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{"GetFollowMapPostOK", CaseIn{"post", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["72"],"resource_ids":["42"]},{"member_ids":["71"],"resource_ids":["42","84"]}],"resource":"post"}`}},
		{"GetFollowMapPostReviewOK", CaseIn{"post", "review", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["71","72"],"resource_ids":["42"]}],"resource":"post"}`}},
		{"GetFollowMapPostNewsOK", CaseIn{"post", "news", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["71"],"resource_ids":["84"]}],"resource":"post"}`}},
		{"GetFollowMapResourceUnknown", CaseIn{"unknown source", "news", time.Time{}}, CaseOut{http.StatusBadRequest, `{"Error":"Resource Not Supported"}`}},
		{"GetFollowMapMemberOK", CaseIn{"member", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["72"],"resource_ids":["42"]},{"member_ids":["71"],"resource_ids":["84"]}],"resource":"member"}`}},
		{"GetFollowMapProjectOK", CaseIn{"project", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["71"],"resource_ids":["420"]},{"member_ids":["72"],"resource_ids":["420","840"]}],"resource":"project"}`}},
	}

	for _, testcase := range TestFollowingGetCases {

		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Body.String() != testcase.out.resp {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.resp, w.Body.String(), testcase.name)
			t.Fail()
		}
	}

	clearFollowTest()
}

func TestFollowingAddDelete(t *testing.T) {

	type PubsubWrapperMessage struct {
		ID   string   `json:"messageId"`
		Attr []string `json:"attributes"`
		Body []byte   `json:"data"`
	}

	type PubsubWrapper struct {
		Subscription string               `json:"subscription"`
		Message      PubsubWrapperMessage `json:"message"`
	}

	type CaseIn struct {
		Action   string `json:action,omitempty`
		Resource string `json:resource,omitempty`
		Subject  string `json:subject,omitempty`
		Object   string `json:object,omitempty`
	}

	type CaseOut struct {
		httpcode int
		Error    string
	}

	var TestRouteName = "/restful/pubsub"
	var TestRouteMethod = "POST"

	var TestFollowingGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{"AddFollowingPostOK", CaseIn{"follow", "post", "70", "84"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingMemberOK", CaseIn{"follow", "member", "70", "72"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingProjectOK", CaseIn{"follow", "project", "70", "840"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingMissingResource", CaseIn{"follow", "", "70", "72"}, CaseOut{http.StatusOK, `{"Error":"Resource Not Supported"}`}},
		{"AddFollowingMissingAction", CaseIn{"", "member", "70", "72"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"AddFollowingWrongIDForPost", CaseIn{"follow", "post", "70", "zexal"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"DeleteFollowingPostOK", CaseIn{"unfollow", "post", "70", "84"}, CaseOut{http.StatusOK, ""}},
		{"DeleteFollowingMemberOK", CaseIn{"unfollow", "member", "70", "72"}, CaseOut{http.StatusOK, ""}},
		{"DeleteFollowingProjectOK", CaseIn{"unfollow", "project", "70", "840"}, CaseOut{http.StatusOK, ""}},
	}

	for _, testcase := range TestFollowingGetCases {
		bodyJsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		jsonStr, err := json.Marshal(&PubsubWrapper{"subs", PubsubWrapperMessage{"1", []string{"1"}, bodyJsonStr}})
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}
	}

	clearFollowTest()
}
