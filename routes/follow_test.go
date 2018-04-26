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
	ID     string
	Object string
}

var mockFollowingDS = map[string][]followDS{
	"post":    []followDS{},
	"member":  []followDS{},
	"project": []followDS{},
}

func (a *mockFollowingAPI) AddFollowing(params map[string]string) error {
	store, ok := mockFollowingDS[params["resource"]]
	if !ok {
		log.Fatalln("unexpected error")
	}

	store = append(store, followDS{ID: params["subject"], Object: params["object"]})
	return nil
}
func (a *mockFollowingAPI) DeleteFollowing(params map[string]string) error {
	store, ok := mockFollowingDS[params["resource"]]
	if !ok {
		log.Fatalln("unexpected error")
	}

	for index, follow := range store {
		if follow.ID == params["subject"] && follow.Object == params["object"] {
			store = append(store[:index], store[index+1:]...)
		}
	}
	return nil
}
func (a *mockFollowingAPI) GetFollowing(params map[string]string) (followings []interface{}, err error) {

	switch {
	case params["subject"] == "unknown@user.who":
		return nil, errors.New("Not Found")
	case params["resource"] == "member":
		return []interface{}{
			models.Member{MemberID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}},
		}, nil
	case params["resource"] == "post":
		switch params["resource_type"] {
		case "":
			return []interface{}{
				models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullString{"followtest1@mirrormedia.mg", true}},
				models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullString{"followtest2@mirrormedia.mg", true}},
			}, nil
		case "review":
			return []interface{}{
				models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullString{"followtest1@mirrormedia.mg", true}},
			}, nil
		case "news":
			return []interface{}{
				models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullString{"followtest2@mirrormedia.mg", true}},
			}, nil
		}
	case params["resource"] == "project":
		return []interface{}{
			models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}},
		}, nil
	default:
		return nil, nil
	}
	return nil, nil
}
func (a *mockFollowingAPI) GetFollowed(args models.GetFollowedArgs) (interface{}, error) {
	type followedCount struct {
		Resourceid string
		Count      int
		Follower   []string
	}
	switch {
	case args.Ids[0] == "1001":
		return []followedCount{}, nil
	case args.Resource == "member":
		return []followedCount{
			followedCount{"followtest1@mirrormedia.mg", 1, []string{"followtest2@mirrormedia.mg"}},
			followedCount{"followtest2@mirrormedia.mg", 1, []string{"followtest1@mirrormedia.mg"}},
		}, nil
	case args.Resource == "post":
		switch args.Type {
		case "":
			return []followedCount{
				followedCount{"42", 2, []string{"followtest1@mirrormedia.mg", "followtest2@mirrormedia.mg"}},
				followedCount{"84", 1, []string{"followtest1@mirrormedia.mg"}},
			}, nil
		case "review":
			return []followedCount{
				followedCount{"42", 2, []string{"followtest1@mirrormedia.mg", "followtest2@mirrormedia.mg"}},
			}, nil
		case "news":
			return []followedCount{
				followedCount{"84", 1, []string{"followtest1@mirrormedia.mg"}},
			}, nil
		}
		return nil, nil
	case args.Resource == "project":
		switch len(args.Ids) {
		case 1:
			return []followedCount{
				followedCount{"840", 1, []string{"followtest2@mirrormedia.mg"}},
			}, nil
		case 2:
			return []followedCount{
				followedCount{"420", 2, []string{"followtest1@mirrormedia.mg", "followtest2@mirrormedia.mg"}},
				followedCount{"840", 1, []string{"followtest2@mirrormedia.mg"}},
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
			models.FollowingMapItem{[]string{"followtest2@mirrormedia.mg"}, []string{"42"}},
			models.FollowingMapItem{[]string{"followtest1@mirrormedia.mg"}, []string{"84"}},
		}, nil
	case args.Resource == "post":
		switch args.Type {
		case "":
			return []models.FollowingMapItem{
				models.FollowingMapItem{[]string{"followtest2@mirrormedia.mg"}, []string{"42"}},
				models.FollowingMapItem{[]string{"followtest1@mirrormedia.mg"}, []string{"42", "84"}},
			}, nil
		case "review":
			return []models.FollowingMapItem{
				models.FollowingMapItem{[]string{"followtest1@mirrormedia.mg", "followtest2@mirrormedia.mg"}, []string{"42"}},
			}, nil
		case "news":
			return []models.FollowingMapItem{
				models.FollowingMapItem{[]string{"followtest1@mirrormedia.mg"}, []string{"84"}},
			}, nil
		}
		return []models.FollowingMapItem{}, nil
	case args.Resource == "project":
		return []models.FollowingMapItem{
			models.FollowingMapItem{[]string{"followtest1@mirrormedia.mg"}, []string{"420"}},
			models.FollowingMapItem{[]string{"followtest2@mirrormedia.mg"}, []string{"840", "420"}},
		}, nil
	default:
		return []models.FollowingMapItem{}, errors.New("Resource Not Supported")
	}
}

func initFollowTest() {
	mockMemberDSBack = mockMemberDS
	mockPostDSBack = mockPostDS

	for _, params := range []models.Member{
		models.Member{MemberID: "followtest0@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest0@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b58e-f06e59fd8467"},
		models.Member{MemberID: "followtest1@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest1@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b59e-f06e59fd8467"},
		models.Member{MemberID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b60e-f06e59fd8467"},
	} {
		err := models.MemberAPI.InsertMember(params)
		if err != nil {
			log.Printf("Insert member fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Post{
		models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullString{"followtest1@mirrormedia.mg", true}},
		models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullString{"followtest2@mirrormedia.mg", true}},
	} {
		_, err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Printf("Insert post fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []models.Project{
		models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}},
		models.Project{ID: 840, PostID: 84, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2016, time.November, 10, 23, 0, 0, 0, time.UTC), true}},
	} {
		err := models.ProjectAPI.InsertProject(params)
		if err != nil {
			log.Printf("Insert Project fail when init test case. Error: %v", err)
		}
	}

	for _, params := range []map[string]string{
		map[string]string{"resource": "member", "subject": "followtest1@mirrormedia.mg", "object": "followtest2@mirrormedia.mg"},
		map[string]string{"resource": "post", "subject": "followtest1@mirrormedia.mg", "object": "42"},
		map[string]string{"resource": "post", "subject": "followtest1@mirrormedia.mg", "object": "84"},
		map[string]string{"resource": "project", "subject": "followtest1@mirrormedia.mg", "object": "420"},
		map[string]string{"resource": "member", "subject": "followtest2@mirrormedia.mg", "object": "followtest1@mirrormedia.mg"},
		map[string]string{"resource": "post", "subject": "followtest2@mirrormedia.mg", "object": "42"},
		map[string]string{"resource": "project", "subject": "followtest2@mirrormedia.mg", "object": "420"},
		map[string]string{"resource": "project", "subject": "followtest2@mirrormedia.mg", "object": "840"},
	} {
		err := models.FollowingAPI.AddFollowing(params)
		if err != nil {
			log.Printf("Init test case fail. Error: %v", err)
		}
	}
}

func clearFollowTest() {
	//restore the backuped data
	mockMemberDS = mockMemberDSBack
	mockPostDS = mockPostDSBack
}

func TestFollowingGet(t *testing.T) {

	initFollowTest()

	type CaseIn struct {
		Resource     string `json:"resource,omitempty"`
		ResourceType string `json:"resource_type, omitempty"`
		Subject      string `json:"subject,omitempty"`
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
		{"GetFollowingPostOK", CaseIn{"post", "", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, `[{"id":42,"author":"followtest1@mirrormedia.mg","created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":0,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":null},{"id":84,"author":"followtest2@mirrormedia.mg","created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":null}]`}},
		{"GetFollowingPostReviewOK", CaseIn{"post", "review", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, `[{"id":42,"author":"followtest1@mirrormedia.mg","created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":0,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":null}]`}},
		{"GetFollowingPostNewsOK", CaseIn{"post", "news", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, `[{"id":84,"author":"followtest2@mirrormedia.mg","created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":null}]`}},
		{"GetFollowingProjectOK", CaseIn{"project", "", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, `[{"id":420,"created_at":null,"updated_at":"2015-11-10T23:00:00Z","updated_by":null,"published_at":null,"post_id":42,"like_amount":null,"comment_amount":null,"active":1,"hero_image":null,"title":null,"description":null,"author":null,"og_title":null,"og_description":null,"og_image":null,"project_order":null,"status":null,"slug":null,"views":null,"publish_status":null,"progress":null,"memo_points":null}]`}},
		{"GetFollowingFollowerNotExist", CaseIn{"project", "", "unknown@user.who"}, CaseOut{http.StatusOK, `[]`}},
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

	initFollowTest()

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
		{"GetFollowedPostOK", CaseIn{"post", "", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":"42","Count":2,"Follower":["followtest1@mirrormedia.mg","followtest2@mirrormedia.mg"]},{"Resourceid":"84","Count":1,"Follower":["followtest1@mirrormedia.mg"]}]`}},
		{"GetFollowedPostReviewOK", CaseIn{"post", "review", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":"42","Count":2,"Follower":["followtest1@mirrormedia.mg","followtest2@mirrormedia.mg"]}]`}},
		{"GetFollowedPostNewsOK", CaseIn{"post", "news", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":"84","Count":1,"Follower":["followtest1@mirrormedia.mg"]}]`}},
		{"GetFollowedMemberOK", CaseIn{"member", "", []string{"followtest1@mirrormedia.mg", "followtest2@mirrormedia.mg"}}, CaseOut{http.StatusOK, `[{"Resourceid":"followtest1@mirrormedia.mg","Count":1,"Follower":["followtest2@mirrormedia.mg"]},{"Resourceid":"followtest2@mirrormedia.mg","Count":1,"Follower":["followtest1@mirrormedia.mg"]}]`}},
		{"GetFollowedProjectSingleOK", CaseIn{"project", "", []string{"840"}}, CaseOut{http.StatusOK, `[{"Resourceid":"840","Count":1,"Follower":["followtest2@mirrormedia.mg"]}]`}},
		{"GetFollowedProjectOK", CaseIn{"project", "", []string{"420", "840"}}, CaseOut{http.StatusOK, `[{"Resourceid":"420","Count":2,"Follower":["followtest1@mirrormedia.mg","followtest2@mirrormedia.mg"]},{"Resourceid":"840","Count":1,"Follower":["followtest2@mirrormedia.mg"]}]`}},
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

	initFollowTest()

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
		{"GetFollowMapPostOK", CaseIn{"post", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["followtest2@mirrormedia.mg"],"resource_ids":["42"]},{"member_ids":["followtest1@mirrormedia.mg"],"resource_ids":["42","84"]}],"resource":"post"}`}},
		{"GetFollowMapPostReviewOK", CaseIn{"post", "review", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["followtest1@mirrormedia.mg","followtest2@mirrormedia.mg"],"resource_ids":["42"]}],"resource":"post"}`}},
		{"GetFollowMapPostNewsOK", CaseIn{"post", "news", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["followtest1@mirrormedia.mg"],"resource_ids":["84"]}],"resource":"post"}`}},
		{"GetFollowMapResourceUnknown", CaseIn{"unknown source", "news", time.Time{}}, CaseOut{http.StatusBadRequest, `{"Error":"Resource Not Supported"}`}},
		{"GetFollowMapMemberOK", CaseIn{"member", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["followtest2@mirrormedia.mg"],"resource_ids":["42"]},{"member_ids":["followtest1@mirrormedia.mg"],"resource_ids":["84"]}],"resource":"member"}`}},
		{"GetFollowMapProjectOK", CaseIn{"project", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["followtest1@mirrormedia.mg"],"resource_ids":["420"]},{"member_ids":["followtest2@mirrormedia.mg"],"resource_ids":["840","420"]}],"resource":"project"}`}},
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

	initFollowTest()

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
		{"AddFollowingPostOK", CaseIn{"follow", "post", "followtest0@mirrormedia.mg", "84"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingMemberOK", CaseIn{"follow", "member", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingProjectOK", CaseIn{"follow", "project", "followtest0@mirrormedia.mg", "840"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingMissingResource", CaseIn{"follow", "", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"AddFollowingMissingAction", CaseIn{"", "member", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"AddFollowingWrongIDForPost", CaseIn{"follow", "post", "followtest0@mirrormedia.mg", "zexal"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"DeleteFollowingPostOK", CaseIn{"unfollow", "post", "followtest0@mirrormedia.mg", "84"}, CaseOut{http.StatusOK, ""}},
		{"DeleteFollowingMemberOK", CaseIn{"unfollow", "member", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, ""}},
		{"DeleteFollowingProjectOK", CaseIn{"unfollow", "project", "followtest0@mirrormedia.mg", "840"}, CaseOut{http.StatusOK, ""}},
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
