package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockPostAPI struct{}

func initPostTest() {
	mockPostDSBack = mockPostDS
}

func clearPostTest() {
	mockPostDS = mockPostDSBack
}

type ExpectResp struct {
	httpcode int
	err      string
}

func (a *mockPostAPI) GetPosts(args *models.PostArgs) (result []models.TaggedPostMember, err error) {
	result = []models.TaggedPostMember{
		{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
		{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
		{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
		{PostMember: models.PostMember{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}}},
	}
	err = nil

	if args.Sorting == "updated_at" {
		result = []models.TaggedPostMember{
			{PostMember: models.PostMember{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}}},
			{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
			{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
			{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
		}
		err = nil
	}
	if args.MaxResult == 2 {
		result = result[:2]
	}

	if reflect.DeepEqual(args.Active, map[string][]int{"$nin": {0, 1, 2, 3, 4}}) {
		return []models.TaggedPostMember{}, nil
	}
	// Active filter
	if reflect.DeepEqual(args.Active, map[string][]int{"$nin": {1, 4}}) {
		result = []models.TaggedPostMember{
			{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
			{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
		}
		err = nil
		return result, err
	}
	// Author filter
	if args.Author != nil {
		if reflect.DeepEqual(args.Author, map[string][]string{"$in": {"superman@mirrormedia.mg", "Major.Tom@mirrormedia.mg"}}) {
			result = []models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
			}
			err = nil
			return result, err
		}
	}
	// Type
	if args.Type != nil {
		if reflect.DeepEqual(args.Type, map[string][]int{"$in": {1, 2}}) {
			result = []models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
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
// 		updatedBy = models.UpdatedBy{}
// 	}

// 	if result != nil {
// 		err = nil
// 	}
// 	return result, err
// }

func (a *mockPostAPI) GetPost(id uint32) (models.TaggedPostMember, error) {
	var (
		result    models.TaggedPostMember
		author    models.Member
		updatedBy models.UpdatedBy
	)
	err := errors.New("Post Not Found")
	for _, value := range mockPostDS {
		if value.ID == id {
			for _, member := range mockMemberDS {
				if value.Author.Valid && member.ID == value.Author.String {
					author = member
				}
				if value.UpdatedBy.Valid && member.ID == value.UpdatedBy.String {
					updatedBy = models.UpdatedBy(member)
				}
			}
			result = models.TaggedPostMember{PostMember: models.PostMember{Post: value, Member: author, UpdatedBy: updatedBy}}
			err = nil
			break
		}
	}
	return result, err
}

func (a *mockPostAPI) InsertPost(p models.Post) (int, error) {
	for _, post := range mockPostDS {
		if post.ID == p.ID {
			err := errors.New("Duplicate entry")
			return 0, err
		}
	}
	mockPostDS = append(mockPostDS, p)
	return len(mockPostDS) - 1, nil
}

func (a *mockPostAPI) UpdatePost(p models.Post) error {
	err := errors.New("Post Not Found")
	for index, value := range mockPostDS {
		if value.ID == p.ID {
			mockPostDS[index].LikeAmount = p.LikeAmount
			mockPostDS[index].Title = p.Title
			err = nil
			return err
		}
	}
	return err
}
func (a *mockPostAPI) UpdateAll(req models.PostUpdateArgs) (err error) {

	for _, v := range req.IDs {
		if v == 1 || v == 2 || v == 4 || v == 6 {
			err = nil
		} else {
			err = errors.New("Posts Not Found")
		}
	}
	return err
}

// func (a *mockPostAPI) UpdateAll(req models.PostUpdateArgs) (err error) {

// 	result := make([]int, 0)
// 	for _, value := range ids {
// 		for i, v := range mockPostDS {
// 			if v.ID == value {
// 				mockPostDS[i].Active = models.NullInt{Int: int64(models.PostStatus["active"].(float64)), Valid: true}
// 				result = append(result, i)
// 			}
// 		}
// 	}
// 	if len(result) == 0 {
// 		err = errors.New("Posts Not Found")
// 		return err
// 	}
// 	return err
// }

func (a *mockPostAPI) DeletePost(id uint32) error {
	// result := models.Post{}
	err := errors.New("Post Not Found")
	for index, value := range mockPostDS {
		if value.ID == id {
			mockPostDS[index].Active = models.NullInt{Int: int64(models.PostStatus["deactive"].(float64)), Valid: true}
			// result = mockPostDS[index]
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
		if reflect.DeepEqual(req.Author, map[string][]string{"$nin": {"superman@mirrormedia.mg", "Major.Tom@mirrormedia.mg"}}) {
			return 3, nil
		}
	}
	if reflect.DeepEqual(req.Active, map[string][]int{"$nin": {0}}) {
		return 4, nil
	}
	// CountActive
	if reflect.DeepEqual(req.Active, map[string][]int{"$in": {2, 3}}) {
		return 2, nil
	}
	return result, err
}

// // ---------------------------------- Post Test -------------------------------
func TestRouteGetPosts(t *testing.T) {

	initPostTest()

	type ExpectGetsResp struct {
		ExpectResp
		resp []models.TaggedPostMember
	}
	testPostsGetCases := []struct {
		name   string
		route  string
		expect ExpectGetsResp
	}{
		{"UpdatedAtDescending", "/posts", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
				{PostMember: models.PostMember{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}}},
			}}},
		{"UpdatedAtAscending", "/posts?sort=updated_at", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
				{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
			}}},
		{"max_result", "/posts?max_result=2", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
			}}},
		{"AuthorFilter", `/posts?author={"$in":["superman@mirrormedia.mg", "Major.Tom@mirrormedia.mg"]}`, ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
			}}},
		{"ActiveFilter", `/posts?active={"$nin":[1,4]}`, ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
			}}},
		{"NotFound", `/posts?active={"$nin":[0,1,2,3,4]}`, ExpectGetsResp{ExpectResp{http.StatusOK, ``},
			[]models.TaggedPostMember{}}},
		{"Type", `/posts?type={"$in":[1,2]}`, ExpectGetsResp{ExpectResp{http.StatusOK, ``},
			[]models.TaggedPostMember{
				{PostMember: models.PostMember{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}}},
				{PostMember: models.PostMember{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])}},
			}}},
	}
	for _, tc := range testPostsGetCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)

			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s Want %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect to get error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
			expected, _ := json.Marshal(map[string][]models.TaggedPostMember{"_items": tc.expect.resp})
			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Errorf("%s response want\n%s\nbut get\n%s", tc.name, string(expected), w.Body.String())
			}
			// Do we have to test Active for testRouteDelete ?
		})
	}
	clearPostTest()
}
func TestRouteGetPost(t *testing.T) {
	initPostTest()

	type ExpectGetResp struct {
		ExpectResp
		resp models.TaggedPostMember
	}
	testPostGetCases := []struct {
		name   string
		route  string
		expect ExpectGetResp
	}{
		{"Current", "/post/1", ExpectGetResp{ExpectResp{http.StatusOK, ""},
			models.TaggedPostMember{
				PostMember: models.PostMember{
					Post:      mockPostDS[0],
					Member:    mockMemberDS[0],
					UpdatedBy: models.UpdatedBy(mockMemberDS[0])},
				Tags: models.NullString{}}}},
		{"NotExisted", "/post/3", ExpectGetResp{ExpectResp{http.StatusNotFound, `{"Error":"Post Not Found"}`}, models.TaggedPostMember{}}},
	}
	for _, tc := range testPostGetCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)

			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s Want %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect to get error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}

			expected, _ := json.Marshal(map[string][]models.TaggedPostMember{"_items": []models.TaggedPostMember{tc.expect.resp}})
			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Errorf("%s incorrect response", tc.name)
			}
		})
	}
	clearPostTest()
}
func TestRouteInsertPost(t *testing.T) {
	initPostTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"New", `{"author":"superman@mirrormedia.mg","title":"You can't save the world alone, but I can"}`, ExpectResp{http.StatusOK, ""}},
		{"EmptyPayload", `{}`, ExpectResp{http.StatusBadRequest, `{"Error":"Invalid Post"}`}},
		{"Existing", `{"id":1, "author":"superman@mirrormedia.mg"}`, ExpectResp{http.StatusBadRequest, `{"Error":"Post ID Already Taken"}`}},
		{"WithTags", `{"id":53,"author":"Joker@mirrormedia.mg","title":"Why so serious?", "tags":[1,2]}`, ExpectResp{http.StatusOK, ""}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var jsonStr = []byte(tc.payload)
			req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s want %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearPostTest()
}

func TestRoutePutPost(t *testing.T) {
	initPostTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"Current", `{"id":1,"author":"wonderwoman@mirrormedia.mg"}`, ExpectResp{http.StatusOK, ""}},
		{"NotExisted", `{"id":12345, "author":"superman@mirrormedia.mg"}`, ExpectResp{http.StatusBadRequest, `{"Error":"Post Not Found"}`}},
		{"UpdateTags", `{"id":1, "tags":[5,3]}`, ExpectResp{http.StatusOK, ``}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var jsonStr = []byte(tc.payload)
			req, _ := http.NewRequest("PUT", "/post", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s want %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearPostTest()
}

func TestRouteDeleteMultiplePosts(t *testing.T) {
	initPostTest()
	testCase := []struct {
		name   string
		route  string
		expect ExpectResp
	}{
		{"Delete", `/posts?ids=[1, 2]`, ExpectResp{http.StatusOK, ``}},
		{"Empty", `/posts?ids=[]`, ExpectResp{http.StatusBadRequest, `{"Error":"ID List Empty"}`}},
		{"NotFound", `/posts?ids=[3, 5]`, ExpectResp{http.StatusNotFound, `{"Error":"Posts Not Found"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearPostTest()
}
func TestRouteDeletePost(t *testing.T) {
	initPostTest()
	testCase := []struct {
		name   string
		route  string
		expect ExpectResp
	}{
		{"Current", "/post/1", ExpectResp{http.StatusOK, ""}},
		{"NotFound", "/post/12345", ExpectResp{http.StatusNotFound, `{"Error":"Post Not Found"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", tc.route, nil)
			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s Want %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect to get error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearPostTest()
}

func TestRoutePublishMultiplePosts(t *testing.T) {
	initPostTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"CurrentPost", `{"ids": [1,6]}`, ExpectResp{http.StatusOK, ``}},
		{"NotFound", `{"ids": [3,5]}`, ExpectResp{http.StatusNotFound, `{"Error":"Posts Not Found"}`}},
		{"InvalidPayload", `{}`, ExpectResp{http.StatusBadRequest, `{"Error":"Invalid Request Body"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			jsonStr := []byte(tc.payload)
			req, _ := http.NewRequest("PUT", "/posts", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %s but get %s", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearPostTest()
}

func TestRouteCountPosts(t *testing.T) {
	initPostTest()
	type ExpectCountResp struct {
		httpcode int
		resp     string
		err      string
	}
	testCase := []struct {
		name   string
		route  string
		expect ExpectCountResp
	}{
		{"SimpleCount", `/posts/count`, ExpectCountResp{http.StatusOK, `{"_meta":{"total":4}}`, ``}},
		{"CountActive", `/posts/count?active={"$in":[2,3]}`, ExpectCountResp{http.StatusOK, `{"_meta":{"total":2}}`, ``}},
		{"CountAuthor", `/posts/count?author={"$nin":["superman@mirrormedia.mg", "Major.Tom@mirrormedia.mg"]}`, ExpectCountResp{http.StatusOK, `{"_meta":{"total":3}}`, ``}},
		{"MoreThanOneActive", `/posts/count?active={"$nin":[1,0], "$in":[-1,3]}`,
			ExpectCountResp{http.StatusBadRequest, ``,
				`{"Error":"Too many active lists"}`}},
		{"NotEntirelyValidActive", `/posts/count?active={"$in":[-3,0,1]}`,
			ExpectCountResp{http.StatusBadRequest, ``,
				`{"Error":"Not all active elements are valid"}`}},
		{"NoValidActive", `/posts/count?active={"$nin":[-3,-4]}`,
			ExpectCountResp{http.StatusBadRequest, ``,
				`{"Error":"No valid active request"}`}},
		{"Type", `/posts/count?type={"$in":[1,2]}`,
			ExpectCountResp{http.StatusOK, `{"_meta":{"total":3}}`, ``}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
			if w.Code == http.StatusOK && w.Body.String() != tc.expect.resp {
				t.Errorf("%s incorrect response.\nWant\n%s\nBut get\n%s\n", tc.name, tc.expect.resp, w.Body.String())
			}
		})
	}
	clearPostTest()
}
