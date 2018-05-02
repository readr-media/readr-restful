package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockPostAPI struct {
	mockPostDS []models.TaggedPostMember
}

func (a *mockPostAPI) setup(in interface{}) {
	a.mockPostDS = make([]models.TaggedPostMember, len(in.([]models.TaggedPostMember)))
	copy(a.mockPostDS, in.([]models.TaggedPostMember))
	models.PostAPI = a
}

func (a *mockPostAPI) teardown() {
	a.mockPostDS = nil
}

func (a *mockPostAPI) GetPosts(args *models.PostArgs) (result []models.TaggedPostMember, err error) {
	result = []models.TaggedPostMember{
		a.mockPostDS[3],
		a.mockPostDS[1],
		a.mockPostDS[0],
		a.mockPostDS[2],
	}

	err = nil

	if args.Sorting == "updated_at" {
		result = []models.TaggedPostMember{
			a.mockPostDS[2],
			a.mockPostDS[0],
			a.mockPostDS[1],
			a.mockPostDS[3],
		}
		err = nil
	}
	if args.MaxResult == 2 {
		result = result[:2]
	}

	if reflect.DeepEqual(args.Active, map[string][]int{"$nin": {0, 1}}) {
		return []models.TaggedPostMember{}, nil
	}
	// Active filter
	if reflect.DeepEqual(args.Active, map[string][]int{"$nin": {1}}) {
		result = []models.TaggedPostMember{
			a.mockPostDS[1],
			a.mockPostDS[3],
		}
		err = nil
		return result, err
	}
	// Author filter
	if args.Author != nil {
		if reflect.DeepEqual(args.Author, map[string][]int64{"$in": {2, 3}}) {
			result = []models.TaggedPostMember{
				a.mockPostDS[1],
				a.mockPostDS[2],
				a.mockPostDS[3],
			}
			err = nil
			return result, err
		}
	}
	// Type
	if args.Type != nil {
		if reflect.DeepEqual(args.Type, map[string][]int{"$in": {1, 2}}) {
			result = []models.TaggedPostMember{
				a.mockPostDS[3],
				a.mockPostDS[1],
				a.mockPostDS[0],
			}
			err = nil
			return result, err
		}
	}
	return result, err
}

// func (a *mockPostAPI) GetPosts(args models.PostArgs) ([]models.PostMember, error) {

// 	var (
// 		result    []models.PostMember
// 		author    models.Member
// 		updatedBy models.UpdatedBy
// 		err       error
// 	)
// 	if len(mockPostDS) == 0 {
// 		err = errors.New("Posts Not Found")
// 		return result, err
// 	}

// 	// Create new copy of mockPostDS in case mockPostDS order is messed up by sort
// 	sortedMockPostDS := make([]models.Post, len(mockPostDS))
// 	copy(sortedMockPostDS, mockPostDS)

// 	switch args.Sorting {
// 	// ascending
// 	case "updated_at":
// 		sort.SliceStable(sortedMockPostDS, func(i, j int) bool {
// 			return sortedMockPostDS[i].UpdatedAt.Before(sortedMockPostDS[j].UpdatedAt)
// 		})
// 	// descending, newer
// 	case "-updated_at":
// 		sort.SliceStable(sortedMockPostDS, func(i, j int) bool {
// 			return sortedMockPostDS[i].UpdatedAt.After(sortedMockPostDS[j].UpdatedAt)
// 		})
// 	}

// 	if args.MaxResult < uint8(len(sortedMockPostDS)) {
// 		sortedMockPostDS = sortedMockPostDS[(args.Page-1)*uint16(args.MaxResult) : args.Page*uint16(args.MaxResult)]
// 	}

// 	for _, sortedpost := range sortedMockPostDS {
// 		for _, member := range mockMemberDS {
// 			if sortedpost.Author.Valid && member.ID == sortedpost.Author.String {
// 				author = member
// 			}
// 			if sortedpost.UpdatedBy.Valid && member.ID == sortedpost.UpdatedBy.String {
// 				updatedBy = models.UpdatedBy(member)
// 			}
// 		}
// 		result = append(result, models.PostMember{Post: sortedpost, Member: author, UpdatedBy: updatedBy})
// 		// Clear up temp struct before next loop
// 		author = models.Member{}
// 		updatedBy = models.MemberBasic{}
// 	}

// 	if result != nil {
// 		err = nil
// 	}
// 	return result, err
// }

func (a *mockPostAPI) GetPost(id uint32) (models.TaggedPostMember, error) {
	var (
		result models.TaggedPostMember
	)
	err := errors.New("Post Not Found")
	for _, value := range a.mockPostDS {
		if value.PostMember.Post.ID == id {
			result = value
			err = nil
			break
		}
	}
	return result, err
}

func (a *mockPostAPI) InsertPost(p models.Post) (int, error) {

	var tpm models.TaggedPostMember
	var id uint32
	if len(a.mockPostDS) != 0 {
		id = a.mockPostDS[len(a.mockPostDS)-1].ID
	} else {
		id = 1
	}
	p.ID = id
	tpm.PostMember.Post = p
	a.mockPostDS = append(a.mockPostDS, tpm)
	return int(p.ID), nil
}

func (a *mockPostAPI) UpdatePost(p models.Post) (err error) {
	err = errors.New("Post Not Found")
	for index, value := range a.mockPostDS {
		if value.ID == p.ID {
			a.mockPostDS[index].LikeAmount = p.LikeAmount
			a.mockPostDS[index].Title = p.Title
			err = nil
			return err
		}
	}
	return err
}
func (a *mockPostAPI) UpdateAll(req models.PostUpdateArgs) (err error) {

	err = errors.New("Posts Not Found")

	for _, r := range req.IDs {
		for _, v := range a.mockPostDS {
			if r == int(v.Post.ID) {
				err = nil
			}
		}
	}

	return err
}

func (a *mockPostAPI) DeletePost(id uint32) (err error) {
	err = errors.New("Post Not Found")
	for index, value := range a.mockPostDS {
		if value.ID == id {
			mockPostDS[index].Active = models.NullInt{Int: int64(models.PostStatus["deactive"].(float64)), Valid: true}
			return nil
		}
	}
	return err
}

func (a *mockPostAPI) Count(req *models.PostArgs) (result int, err error) {
	result = 4
	err = nil
	// Type
	if req.Type != nil {
		if reflect.DeepEqual(req.Type, map[string][]int{"$in": {1, 2}}) {
			return 3, nil
		}
	}

	// CountAuthor
	if req.Author != nil {
		if reflect.DeepEqual(req.Author, map[string][]int64{"$nin": {0, 1}}) {
			return 2, nil
		}
	}
	if reflect.DeepEqual(req.Active, map[string][]int{"$nin": {0}}) {
		return 4, nil
	}
	// CountActive
	if reflect.DeepEqual(req.Active, map[string][]int{"$in": {0, 1}}) {
		return 2, nil
	}
	return result, err
}

func (a *mockPostAPI) Hot() (result []models.HotPost, err error) {
	return result, err
}

func (a *mockPostAPI) SchedulePublish() error {
	return nil
}

func TestRoutePost(t *testing.T) {

	var postTest mockPostAPI

	posts := []models.TaggedPostMember{
		{PostMember: models.PostMember{Post: mockPostDS[0], Member: memberToBasic(mockMembers[0]), UpdatedBy: memberToBasic(mockMembers[0])}},
		{PostMember: models.PostMember{Post: mockPostDS[1], Member: memberToBasic(mockMembers[1]), UpdatedBy: models.MemberBasic{}}},
		{PostMember: models.PostMember{Post: mockPostDS[2], Member: memberToBasic(mockMembers[2]), UpdatedBy: models.MemberBasic{}}},
		{PostMember: models.PostMember{Post: mockPostDS[3], Member: memberToBasic(mockMembers[2]), UpdatedBy: models.MemberBasic{}}},
	}

	teststep := []TestStep{
		TestStep{
			name:     "GET/posts",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"UpdatedAtDescending", "GET", `/posts`, ``, http.StatusOK,
					[]models.TaggedPostMember{
						posts[3], posts[1], posts[0], posts[2],
					}},
				genericTestcase{"UpdatedAtAscending", "GET", `/posts?sort=updated_at`, ``, http.StatusOK,
					[]models.TaggedPostMember{posts[2], posts[0], posts[1], posts[3]}},
				genericTestcase{"MaxResult", "GET", `/posts?max_result=2`, ``, http.StatusOK,
					[]models.TaggedPostMember{posts[3], posts[1]}},
				genericTestcase{"AuthorFilter", "GET", `/posts?author={"$in":[2,3]}`, ``, http.StatusOK,
					[]models.TaggedPostMember{posts[1], posts[2], posts[3]}},
				genericTestcase{"ActiveFilter", "GET", `/posts?active={"$nin":[1]}`, ``, http.StatusOK,
					[]models.TaggedPostMember{posts[1], posts[3]}},
				genericTestcase{"NotFound", "GET", `/posts?active={"$nin":[0,1]}`, ``, http.StatusOK,
					[]models.TaggedPostMember{}},
				genericTestcase{"Type", "GET", `/posts?type={"$in":[1,2]}`, ``, http.StatusOK,
					[]models.TaggedPostMember{posts[3], posts[1], posts[0]}},
			},
		},
		TestStep{
			name:     "GET/post",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"UpdatedAtDescending", "GET", `/posts`, ``, http.StatusOK,
					[]models.TaggedPostMember{
						posts[3], posts[1], posts[0], posts[2],
					}},
				genericTestcase{"Current", "GET", `/post/1`, ``, http.StatusOK,
					[]models.TaggedPostMember{posts[0]}},
				genericTestcase{"NotExisted", "GET", `/post/3`, `{"Error":"Post Not Found"}`, http.StatusNotFound,
					[]models.TaggedPostMember{}},
			},
		},
		TestStep{
			name:     "POST",
			init:     func() { postTest.setup([]models.TaggedPostMember{}) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"New", "POST", `/post`, `{"author":1,"title":"You can't save the world alone, but I can"}`, http.StatusOK, ``},
				genericTestcase{"EmptyPayload", "POST", `/post`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Post"}`},
				// post_id will not repeat now
				// genericTestcase{"Existing", "POST", "/post", `{"author":1}`, http.StatusBadRequest, `{"Error":"Post ID Already Taken"}`},
				genericTestcase{"WithTags", "POST", `/post`, `{"author":53,"title":"Why so serious?", "tags":[1,2]}`, http.StatusOK, ``},
			},
		},
		TestStep{
			name:     "PUT",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"UpdateCurrent", "PUT", `/post`, `{"id":1,"author":2}`, http.StatusOK, ``},
				genericTestcase{"NotExisted", "PUT", `/post`, `{"id":12345, "author":1}`, http.StatusBadRequest, `{"Error":"Post Not Found"}`},
				genericTestcase{"UpdateTags", "PUT", `/post`, `{"id":1, "tags":[5,3], "updated_by":1}`, http.StatusOK, ``},
				// UpdateSchedule the same with UpdateTags, need to be changed or confirmed
				genericTestcase{"UpdateSchedule", "PUT", `/post`, `{"id":1, "tags":[5,3], "updated_by":1}`, http.StatusOK, ``},
			},
		},
		TestStep{
			name:     "DELETE/posts",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"SimpleDelete", "DELETE", `/posts?ids=[1,2]`, ``, http.StatusOK, ``},
				genericTestcase{"EmptyID", "DELETE", `/posts?ids=[]`, ``, http.StatusBadRequest, `{"Error":"ID List Empty"}`},
				genericTestcase{"NotFound", "DELETE", `/posts?ids=[3,5]`, ``, http.StatusNotFound, `{"Error":"Posts Not Found"}`},
			},
		},
		TestStep{
			name:     "DELETE/post",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"SimpleDelete", "DELETE", `/post/1`, ``, http.StatusOK, ``},
				genericTestcase{"NotFound", "DELETE", `/post/12345`, ``, http.StatusNotFound, `{"Error":"Post Not Found"}`},
			},
		},
		TestStep{
			name:     "Publish",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"Posts", "PUT", `/posts`, `{"ids":[1,6]}`, http.StatusOK, ``},
				genericTestcase{"NotFound", "PUT", `/posts`, `{"ids":[3,5]}`, http.StatusNotFound, `{"Error":"Posts Not Found"}`},
				genericTestcase{"InvalidPayload", "PUT", `/posts`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Request Body"}`},
			},
		},
		TestStep{
			name:     "Count",
			init:     func() { postTest.setup(posts) },
			teardown: func() { postTest.teardown() },
			register: &postTest,
			cases: []genericTestcase{
				genericTestcase{"Posts", "GET", `/posts/count`, ``, http.StatusOK, `{"_meta":{"total":4}}`},
				genericTestcase{"Active", "GET", `/posts/count?active={"$in":[0,1]}`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
				genericTestcase{"Author", "GET", `/posts/count?author={"$nin":[0,1]}`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
				genericTestcase{"MoreThanOneActive", "GET", `/posts/count?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`},
				genericTestcase{"NotEntirelyValidActive", "GET", `/posts/count?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
				genericTestcase{"NoValidActive", "GET", `/posts/count?active={"$nin":[-3,-4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`},
				genericTestcase{"Type", "GET", `/posts/count?type={"$in":[1,2]}`, ``, http.StatusOK, `{"_meta":{"total":3}}`}},
		},
	}
	asserter := func(resp string, tc genericTestcase, t *testing.T) {

		type response struct {
			Items []models.TaggedPostMember `json:"_items"`
		}

		var Response response
		var expected = tc.resp.([]models.TaggedPostMember)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", tc.name, resp)
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		} else {
			// Exact same length
			if len(Response.Items) != 0 && len(expected) != 0 {
				for i := range expected {
					if (Response.Items[i].PostMember.Post.ID != expected[i].PostMember.Post.ID) ||
						(Response.Items[i].PostMember.Member != expected[i].PostMember.Member) ||
						(Response.Items[i].PostMember.UpdatedBy != expected[i].PostMember.UpdatedBy) {
						t.Errorf("%s, %vth round expect to get \n%v\n , but get \n%v\n", tc.name, i, expected[i], Response.Items[i])
					}
				}
			}
		}
	}
	for _, ts := range teststep {
		t.Run(ts.name, func(t *testing.T) {
			DoTest(t, ts, asserter)
		})
	}
}

// type mockPostAPI struct{}

// func initPostTest() {
// 	mockPostDSBack = mockPostDS
// }

// func clearPostTest() {
// 	mockPostDS = mockPostDSBack
// }

type ExpectResp struct {
	httpcode int
	err      string
}

func memberToBasic(m models.Member) (result models.MemberBasic) {
	result = models.MemberBasic{
		ID:           m.ID,
		Nickname:     m.Nickname,
		ProfileImage: m.ProfileImage,
		Description:  m.Description,
		Role:         m.Role,
	}
	return result
}
