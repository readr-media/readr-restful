package routes

import (
	"errors"
	"testing"
	"time"

	"net/http"

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

func (a *mockFollowingAPI) Get(params models.GetFollowInterface) (result interface{}, err error) {

	switch params := params.(type) {
	case *models.GetFollowingArgs:
		result, err = getFollowing(params)
	case *models.GetFollowedArgs:
		result, err = getFollowed(params)
	case *models.GetFollowMapArgs:
		result, err = getFollowMap(params)
	case *models.GetFollowerMemberIDsArgs:
		result, err = getFollowerMemberIDs(params)
	default:
		return nil, errors.New("Unsupported Query Args")
	}
	return result, err
}

func (a *mockFollowingAPI) Insert(params models.FollowArgs) error {

	store, ok := mockFollowingDS[params.Resource]
	if !ok {
		return errors.New("Resource Not Supported")
	}

	store = append(store, followDS{ID: params.Subject, Object: params.Object})
	return nil
}

func (a *mockFollowingAPI) Update(params models.FollowArgs) error {
	return nil
}

func (a *mockFollowingAPI) Delete(params models.FollowArgs) error {

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

func getFollowing(params *models.GetFollowingArgs) (followings []interface{}, err error) {
	switch {
	case params.MemberID == 0:
		return nil, errors.New("Not Found")
	case params.ResourceName == "member":
		return []interface{}{
			models.Member{ID: 72, MemberID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}},
		}, nil
	case params.ResourceName == "post":
		switch params.ResourceType {
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
	case params.ResourceName == "project":
		return []interface{}{
			models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
		}, nil
	default:
		return nil, nil
	}
	return nil, nil
}

func getFollowed(args *models.GetFollowedArgs) (interface{}, error) {

	switch {
	case args.IDs[0] == 1001:
		return []models.FollowedCount{}, nil
	case args.ResourceName == "member":
		return []models.FollowedCount{
			models.FollowedCount{71, 1, []int64{72}},
			models.FollowedCount{72, 1, []int64{71}},
		}, nil
	case args.ResourceName == "post":
		switch args.ResourceType {
		case "":
			return []models.FollowedCount{
				models.FollowedCount{42, 2, []int64{71, 72}},
				models.FollowedCount{84, 1, []int64{71}},
			}, nil
		case "review":
			return []models.FollowedCount{
				models.FollowedCount{42, 2, []int64{71, 72}},
			}, nil
		case "news":
			return []models.FollowedCount{
				models.FollowedCount{84, 1, []int64{71}},
			}, nil
		}
		return nil, nil
	case args.ResourceName == "project":
		switch len(args.IDs) {
		case 1:
			return []models.FollowedCount{
				models.FollowedCount{840, 1, []int64{72}},
			}, nil
		case 2:
			return []models.FollowedCount{
				models.FollowedCount{420, 2, []int64{71, 72}},
				models.FollowedCount{840, 1, []int64{72}},
			}, nil
		}
		return nil, nil
	default:
		return nil, nil
	}
}

func getFollowerMemberIDs(args *models.GetFollowerMemberIDsArgs) ([]int, error) {
	return []int{}, nil
}

func getFollowMap(args *models.GetFollowMapArgs) (list []models.FollowingMapItem, err error) {

	switch {
	case args.ResourceName == "member":
		return []models.FollowingMapItem{
			models.FollowingMapItem{[]string{"72"}, []string{"42"}},
			models.FollowingMapItem{[]string{"71"}, []string{"84"}},
		}, nil
	case args.ResourceName == "post":
		switch args.ResourceType {
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
	case args.ResourceName == "project":
		return []models.FollowingMapItem{
			models.FollowingMapItem{[]string{"71"}, []string{"420"}},
			models.FollowingMapItem{[]string{"72"}, []string{"420", "840"}},
		}, nil
	default:
		return []models.FollowingMapItem{}, errors.New("Resource Not Supported")
	}
}

type mockFollowCache struct{}

func (m mockFollowCache) Update(i models.GetFollowedArgs, f []models.FollowedCount)            {}
func (m mockFollowCache) Revoke(actionType string, resource string, emotion int, object int64) {}

// func initFollowTest() {
// 	mockPostDSBack = mockPostDS

// 	for _, params := range []models.Member{
// 		models.Member{ID: 70, MemberID: "followtest0@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest0@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b88e-f06e59fd8467", TalkID: models.NullString{"abc1d5b1-da54-4200-b58e-f06e59fd8467", true}},
// 		models.Member{ID: 71, MemberID: "followtest1@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest1@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b59e-f06e59fd8467", TalkID: models.NullString{"abc1d5b1-da54-4200-b59e-f06e59fd8467", true}},
// 		models.Member{ID: 72, MemberID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}, PostPush: models.NullBool{true, true}, UpdatedAt: models.NullTime{time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Mail: models.NullString{"followtest2@mirrormedia.mg", true}, Points: models.NullInt{0, true}, UUID: "abc1d5b1-da54-4200-b60e-f06e59fd8467", TalkID: models.NullString{"abc1d5b1-da54-4200-b60e-f06e59fd8467", true}},
// 	} {
// 		_, err := models.MemberAPI.InsertMember(params)
// 		if err != nil {
// 			log.Printf("Insert member fail when init test case. Error: %v", err)
// 		}
// 	}

// 	for _, params := range []models.Post{
// 		models.Post{ID: 42, Active: models.NullInt{1, true}, Type: models.NullInt{0, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{71, true}, PublishStatus: models.NullInt{2, true}},
// 		models.Post{ID: 84, Active: models.NullInt{1, true}, Type: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), true}, Author: models.NullInt{72, true}, PublishStatus: models.NullInt{2, true}},
// 	} {
// 		_, err := models.PostAPI.InsertPost(params)
// 		if err != nil {
// 			log.Printf("Insert post fail when init test case. Error: %v", err)
// 		}
// 	}

// 	for _, params := range []models.Project{
// 		models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
// 		models.Project{ID: 840, PostID: 84, Active: models.NullInt{1, true}, UpdatedAt: models.NullTime{time.Date(2016, time.November, 10, 23, 0, 0, 0, time.UTC), true}, PublishStatus: models.NullInt{2, true}},
// 	} {
// 		err := models.ProjectAPI.InsertProject(params)
// 		if err != nil {
// 			log.Printf("Insert Project fail when init test case. Error: %v", err)
// 		}
// 	}

// 	for _, params := range []models.FollowArgs{
// 		models.FollowArgs{Resource: "member", Subject: 71, Object: 72},
// 		models.FollowArgs{Resource: "post", Subject: 71, Object: 42},
// 		models.FollowArgs{Resource: "post", Subject: 71, Object: 84},
// 		models.FollowArgs{Resource: "project", Subject: 71, Object: 420},
// 		models.FollowArgs{Resource: "member", Subject: 72, Object: 71},
// 		models.FollowArgs{Resource: "post", Subject: 72, Object: 42},
// 		models.FollowArgs{Resource: "project", Subject: 72, Object: 420},
// 		models.FollowArgs{Resource: "project", Subject: 72, Object: 840},
// 	} {
// 		err := models.FollowingAPI.Insert(params)
// 		if err != nil {
// 			log.Printf("Init test case fail. Error: %v", err)
// 		}
// 	}
// }

// {"GetFollowingPostOK", CaseIn{"post", "", 71}, CaseOut{http.StatusOK, `[{"id":42,"author":71,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":0,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2},{"id":84,"author":72,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2}]`}},
// {"GetFollowingPostReviewOK", CaseIn{"post", "review", 71}, CaseOut{http.StatusOK, `[{"id":42,"author":71,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":0,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2}]`}},
// {"GetFollowingPostNewsOK", CaseIn{"post", "news", 71}, CaseOut{http.StatusOK, `[{"id":84,"author":72,"created_at":null,"like_amount":null,"comment_amount":null,"title":null,"content":null,"type":1,"link":null,"og_title":null,"og_description":null,"og_image":null,"active":1,"updated_at":"2009-11-10T23:00:00Z","updated_by":null,"published_at":null,"link_title":null,"link_description":null,"link_image":null,"link_name":null,"video_id":null,"video_views":null,"publish_status":2}]`}},
// {"GetFollowingProjectOK", CaseIn{"project", "", 71}, CaseOut{http.StatusOK, `[{"id":420,"created_at":null,"updated_at":"2015-11-10T23:00:00Z","updated_by":null,"published_at":null,"post_id":42,"like_amount":null,"comment_amount":null,"active":1,"hero_image":null,"title":null,"description":null,"author":null,"og_title":null,"og_description":null,"og_image":null,"project_order":null,"status":null,"slug":null,"views":null,"publish_status":2,"progress":null,"memo_points":null}]`}},
// {"GetFollowingFollowerNotExist", CaseIn{"project", "", 0}, CaseOut{http.StatusOK, `[]`}},

// {"GetFollowedPostOK", CaseIn{"post", "", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":42,"Count":2,"Follower":[71,72]},{"Resourceid":84,"Count":1,"Follower":[71]}]`}},
// {"GetFollowedPostReviewOK", CaseIn{"post", "review", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":42,"Count":2,"Follower":[71,72]}]`}},
// {"GetFollowedPostNewsOK", CaseIn{"post", "news", []string{"42", "84"}}, CaseOut{http.StatusOK, `[{"Resourceid":84,"Count":1,"Follower":[71]}]`}},
// {"GetFollowedMemberOK", CaseIn{"member", "", []string{"71", "72"}}, CaseOut{http.StatusOK, `[{"Resourceid":71,"Count":1,"Follower":[72]},{"Resourceid":72,"Count":1,"Follower":[71]}]`}},
// {"GetFollowedProjectSingleOK", CaseIn{"project", "", []string{"840"}}, CaseOut{http.StatusOK, `[{"Resourceid":840,"Count":1,"Follower":[72]}]`}},
// {"GetFollowedProjectOK", CaseIn{"project", "", []string{"420", "840"}}, CaseOut{http.StatusOK, `[{"Resourceid":420,"Count":2,"Follower":[71,72]},{"Resourceid":840,"Count":1,"Follower":[72]}]`}},
// {"GetFollowedMissingResource", CaseIn{"", "", []string{"420", "840"}}, CaseOut{http.StatusBadRequest, `{"Error":"Unsupported Resource"}`}},
// {"GetFollowedMissingID", CaseIn{"post", "", []string{}}, CaseOut{http.StatusBadRequest, `{"Error":"Bad Resource ID"}`}},
// {"GetFollowedPostNotExist", CaseIn{"post", "", []string{"1001", "1000"}}, CaseOut{http.StatusOK, `[]`}},
// {"GetFollowedPostStringID", CaseIn{"post", "", []string{"unintegerable"}}, CaseOut{http.StatusBadRequest, `{"Error":"Bad Resource ID"}`}},
// {"GetFollowedProjectStringID", CaseIn{"project", "", []string{"unintegerable"}}, CaseOut{http.StatusBadRequest, `{"Error":"Bad Resource ID"}`}},

// {"GetFollowMapPostOK", CaseIn{"post", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["72"],"resource_ids":["42"]},{"member_ids":["71"],"resource_ids":["42","84"]}],"resource":"post"}`}},
// {"GetFollowMapPostReviewOK", CaseIn{"post", "review", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["71","72"],"resource_ids":["42"]}],"resource":"post"}`}},
// {"GetFollowMapPostNewsOK", CaseIn{"post", "news", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["71"],"resource_ids":["84"]}],"resource":"post"}`}},
// {"GetFollowMapResourceUnknown", CaseIn{"unknown source", "news", time.Time{}}, CaseOut{http.StatusBadRequest, `{"Error":"Resource Not Supported"}`}},
// {"GetFollowMapMemberOK", CaseIn{"member", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["72"],"resource_ids":["42"]},{"member_ids":["71"],"resource_ids":["84"]}],"resource":"member"}`}},
// {"GetFollowMapProjectOK", CaseIn{"project", "", time.Time{}}, CaseOut{http.StatusOK, `{"list":[{"member_ids":["71"],"resource_ids":["420"]},{"member_ids":["72"],"resource_ids":["420","840"]}],"resource":"project"}`}},

func TestFollowing(t *testing.T) {

	transformPubsub := func(tc genericTestcase) genericTestcase {
		meta := PubsubMessageMeta{
			Subscription: "sub",
			Message: PubsubMessageMetaBody{
				ID:   "1",
				Body: []byte(tc.body.(string)),
				Attr: map[string]string{"type": "follow", "action": tc.method},
			},
		}

		return genericTestcase{tc.name, "POST", "/restful/pubsub", meta, tc.httpcode, tc.resp}
	}

	t.Run("Get", func(t *testing.T) {

		for _, testcase := range []genericTestcase{

			// Only check error message when http status != 200
			// use nil as resp parameter for genericDoTest()
			genericTestcase{"FollowingPostOK", "GET", `/following/user?resource=post&id=71`, ``, http.StatusOK, nil},
			genericTestcase{"FollowingPostReviewOK", "GET", `/following/user?resource=post&resource_type=review&id=71`, ``, http.StatusOK, nil},
			genericTestcase{"FollowingPostNewsOK", "GET", `/following/user?resource=post&resource_type=news&id=71`, ``, http.StatusOK, nil},
			genericTestcase{"FollowingProjectOK", "GET", `/following/user?resource=project&id=71`, ``, http.StatusOK, nil},

			genericTestcase{"FollowedPostOK", "GET", `/following/resource?resource=post&ids=[42,84]&resource_type=news`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedPostReviewOK", "GET", `/following/resource?resource=post&ids=[42,84]&resource_type=review`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedPostNewsOK", "GET", `/following/resource?resource=post&resource_type=news&ids=[42,84]`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedMemberOK", "GET", `/following/resource?resource=member&ids=[42,84]`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedProjectSingleOK", "GET", `/following/resource?resource=project&ids=[840]`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedProjectOK", "GET", `/following/resource?resource=project&ids=[420,840]`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedMissingResource", "GET", `/following/resource?ids=[420,840]`, ``, http.StatusBadRequest, `{"Error":"Unsupported Resource"}`},
			genericTestcase{"FollowedMissingID", "GET", `/following/resource?resource=post&ids=[]`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`},
			genericTestcase{"FollowedPostNotExist", "GET", `/following/resource?resource=post&ids=[1000,1001]&resource_type=news`, ``, http.StatusOK, nil},
			genericTestcase{"FollowedPostStringID", "GET", `/following/resource?resourcea=post&ids=[unintegerable]`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`},
			genericTestcase{"FollowedProjectStringID", "GET", `/following/resource?resource=project&ids=[unintegerable]`, ``, http.StatusBadRequest, `{"Error":"Bad Resource ID"}`},
			genericTestcase{"FollowedProjectInvalidEmotion", "GET", `/following/resource?resource=project&ids=[42,84]&resource_type=review&emotion=angry`, ``, http.StatusBadRequest, `{"Error":"Unsupported Emotion"}`},

			genericTestcase{"FollowMapPostOK", "GET", `/following/map?resource=post`, ``, http.StatusOK, nil},
			genericTestcase{"FollowMapPostReviewOK", "GET", `/following/map?resource=post&resource_type=review`, ``, http.StatusOK, nil},
			genericTestcase{"GetFollowMapPostNewsOK", "GET", `/following/map?resource=post&resource_type=news`, ``, http.StatusOK, nil},
			genericTestcase{"GetFollowMapResourceUnknown", "GET", `/following/map?resource=unknown source&resource_type=news`, ``, http.StatusBadRequest, `{"Error":"Unsupported Resource"}`},
			genericTestcase{"GetFollowMapMemberOK", "GET", `/following/map?resource=member`, ``, http.StatusOK, nil},
			genericTestcase{"GetFollowMapProjectOK", "GET", `/following/map?resource=project`, ``, http.StatusOK, nil},
		} {
			genericDoTest(testcase, t, nil)
		}
	})
	// It seems insert and delete shouldn't be tested here.
	t.Run("Insert", func(t *testing.T) {

		for _, testcase := range []genericTestcase{
			genericTestcase{"FollowingPostOK", "follow", `/restful/pubsub`, `{"resource":"post","subject":70,"object":84}`, http.StatusOK, nil},
			genericTestcase{"FollowingMemberOK", "follow", `/restful/pubsub`, `{"resource":"member","subject":70,"object":72}`, http.StatusOK, nil},
			genericTestcase{"FollowingProjectOK", "follow", `/restful/pubsub`, `{"resource":"project","subject":70,"object":840}`, http.StatusOK, nil},
			genericTestcase{"FollowingTagOK", "follow", `/restful/pubsub`, `{"resource":"tag","subject":70,"object":1}`, http.StatusOK, nil},
			genericTestcase{"FollowingMissingResource", "follow", `/restful/pubsub`, `{"resource":"","subject":70,"object":72}`, http.StatusOK, `{"Error":"Unsupported Resource"}`},
			genericTestcase{"FollowingMissingAction", "", `/restful/pubsub`, `{"resource":"post","subject":70,"object":72}`, http.StatusOK, `{"Error":"Bad Request"}`},
		} {
			genericDoTest(transformPubsub(testcase), t, nil)
		}
	})
	t.Run("Delete", func(t *testing.T) {

		for _, testcase := range []genericTestcase{
			genericTestcase{"FollowingPostOK", "unfollow", `/restful/pubsub`, `{"resource":"post","subject":70,"object":84}`, http.StatusOK, nil},
			genericTestcase{"FollowingMemberOK", "unfollow", `/restful/pubsub`, `{"resource":"member","subject":70,"object":72}`, http.StatusOK, nil},
			genericTestcase{"FollowingProjectOK", "unfollow", `/restful/pubsub`, `{"resource":"project","subject":70,"object":840}`, http.StatusOK, nil},
		} {
			genericDoTest(transformPubsub(testcase), t, nil)
		}
	})
}
