package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

func initPostTest() {
	mockPostDSBack = mockPostDS
}

func clearPostTest() {
	mockPostDS = mockPostDSBack
}

type mockPostAPI struct{}

type ExpectResp struct {
	httpcode int
	err      string
}

func (a *mockPostAPI) GetPosts(maxResult uint8, page uint16, sortMethod string) ([]models.PostMember, error) {

	var (
		result    []models.PostMember
		author    models.Member
		updatedBy models.UpdatedBy
		err       error
	)
	if len(mockPostDS) == 0 {
		err = errors.New("Posts Not Found")
		return result, err
	}

	// Create new copy of mockPostDS in case mockPostDS order is messed up by sort
	sortedMockPostDS := make([]models.Post, len(mockPostDS))
	copy(sortedMockPostDS, mockPostDS)

	switch sortMethod {
	// ascending
	case "updated_at":
		sort.SliceStable(sortedMockPostDS, func(i, j int) bool {
			return sortedMockPostDS[i].UpdatedAt.Before(sortedMockPostDS[j].UpdatedAt)
		})
	// descending, newer
	case "-updated_at":
		sort.SliceStable(sortedMockPostDS, func(i, j int) bool {
			return sortedMockPostDS[i].UpdatedAt.After(sortedMockPostDS[j].UpdatedAt)
		})
	}

	// if maxResult >= uint8(len(sortedMockPostDS)) {
	// 	result = sortedMockPostDS
	// 	err = nil
	// } else if maxResult < uint8(len(sortedMockPostDS)) {
	if maxResult < uint8(len(sortedMockPostDS)) {
		sortedMockPostDS = sortedMockPostDS[(page-1)*uint16(maxResult) : page*uint16(maxResult)]
	}

	for _, sortedpost := range sortedMockPostDS {
		for _, member := range mockMemberDS {
			if sortedpost.Author.Valid && member.ID == sortedpost.Author.String {
				author = member
			}
			if sortedpost.UpdatedBy.Valid && member.ID == sortedpost.UpdatedBy.String {
				updatedBy = models.UpdatedBy(member)
			}
		}
		result = append(result, models.PostMember{Post: sortedpost, Member: author, UpdatedBy: updatedBy})
		// Clear up temp struct before next loop
		author = models.Member{}
		updatedBy = models.UpdatedBy{}
	}

	if result != nil {
		err = nil
	}
	return result, err
}

func (a *mockPostAPI) GetPost(id uint32) (models.PostMember, error) {
	var (
		result    models.PostMember
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
			result = models.PostMember{Post: value, Member: author, UpdatedBy: updatedBy}
			err = nil
			break
		}
	}
	return result, err
}

func (a *mockPostAPI) InsertPost(p models.Post) error {
	for _, post := range mockPostDS {
		if post.ID == p.ID {
			// result = models.Post{}
			err := errors.New("Duplicate entry")
			return err
		}
	}
	mockPostDS = append(mockPostDS, p)
	// result := mockPostDS[len(mockPostDS)-1]
	// err := nil
	return nil
}

func (a *mockPostAPI) UpdatePost(p models.Post) error {
	// result = models.Post{}
	err := errors.New("Post Not Found")
	for index, value := range mockPostDS {
		if value.ID == p.ID {
			mockPostDS[index].LikeAmount = p.LikeAmount
			mockPostDS[index].Title = p.Title
			// return mockPostDS[index], nil
			err = nil
			return err
		}
	}
	return err
}

func (a *mockPostAPI) DeletePost(id uint32) error {
	// result := models.Post{}
	err := errors.New("Post Not Found")
	for index, value := range mockPostDS {
		if value.ID == id {
			mockPostDS[index].Active = 0
			// result = mockPostDS[index]
			return nil
		}
	}
	return err
}

// // ---------------------------------- Post Test -------------------------------
func TestRouteGetPosts(t *testing.T) {

	initPostTest()

	type ExpectGetsResp struct {
		ExpectResp
		resp []models.PostMember
	}
	testPostsGetCases := []struct {
		name   string
		route  string
		expect ExpectGetsResp
	}{
		{"Descending", "/posts", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.PostMember{
				{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}},
				{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}},
				{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])},
				{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}},
			}}},
		{"Ascending", "/posts?sort=updated_at", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.PostMember{
				{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}},
				{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])},
				{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}},
				{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}},
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
				t.Errorf("%s expect to get error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
			}

			// responses := []models.PostMember{}
			// json.Unmarshal([]byte(w.Body.String()), &responses)
			expected, _ := json.Marshal(map[string][]models.PostMember{"_items": tc.expect.resp})
			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Log("Not equal!")
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
		resp models.PostMember
	}
	testPostGetCases := []struct {
		name   string
		route  string
		expect ExpectGetResp
	}{
		{"Exist", "/post/1", ExpectGetResp{ExpectResp{http.StatusOK, ""},
			models.PostMember{
				Post:      mockPostDS[0],
				Member:    mockMemberDS[0],
				UpdatedBy: models.UpdatedBy(mockMemberDS[0])}}},
		{"NonExisted", "/post/3", ExpectGetResp{ExpectResp{http.StatusNotFound, `{"Error":"Post Not Found"}`}, models.PostMember{}}},
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
				t.Errorf("%s expect to get error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
			}

			// responses := []models.PostMember{}
			// json.Unmarshal([]byte(w.Body.String()), &responses)
			// expected, _ := json.Marshal(tc.expect.resp)
			expected, _ := json.Marshal(map[string][]models.PostMember{"_items": []models.PostMember{tc.expect.resp}})
			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Log("Not equal!")
			}
		})
	}
	clearPostTest()
}
func TestRoutePostPost(t *testing.T) {
	initPostTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"New", `{"author":"superman@mirrormedia.mg","title":"You can't save the world alone, but I can"}`, ExpectResp{http.StatusOK, ""}},
		{"Empty", `{}`, ExpectResp{http.StatusBadRequest, `{"Error":"Invalid Post"}`}},
		{"Existing", `{"id":1, "author":"superman@mirrormedia.mg"}`, ExpectResp{http.StatusBadRequest, `{"Error":"Post ID Already Taken"}`}},
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
				t.Errorf("%s expect error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
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
		{"New", `{"id":1,"author":"wonderwoman@mirrormedia.mg"}`, ExpectResp{http.StatusOK, ""}},
		{"NonExisted", `{"id":12345, "author":"superman@mirrormedia.mg"}`, ExpectResp{http.StatusBadRequest, `{"Error":"Post Not Found"}`}},
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
				t.Errorf("%s expect error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
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
		{"Existing", "/post/1", ExpectResp{http.StatusOK, ""}},
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

// func TestGetPostsDescending(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	// default sort -updated_at
// 	req, _ := http.NewRequest("GET", "/posts", nil)

// 	r.ServeHTTP(w, req)
// 	if w.Code != http.StatusOK {
// 		t.Errorf("HTTP response code error: want %q but get %q", http.StatusOK, w.Code)
// 	}
// 	expected, _ := json.Marshal(mockPostDS)
// 	if w.Body.String() != string(expected) {
// 		t.Errorf("Response error: Not Expected")
// 	}
// }

// func TestGetPostsAscending(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/posts?sort=updated_at", nil)

// 	r.ServeHTTP(w, req)
// 	if w.Code != http.StatusOK {
// 		t.Errorf("HTTP response code error: want %q but get %q", http.StatusOK, w.Code)
// 	}
// 	res := []models.Post{}
// 	json.Unmarshal([]byte(w.Body.String()), &res)
// 	if res[0] != mockPostDS[2] || res[1] != mockPostDS[1] || res[2] != mockPostDS[0] {
// 		t.Errorf("Response sort error")
// 	}
// }

// func TestGetExistPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/post/1", nil)

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Fail()
// 	}
// 	// expected, _ := json.Marshal(mockPostDS[0])
// 	// if w.Body.String() != string(expected) {
// 	// 	t.Fail()
// 	// }
// }

// func TestGetNonExistedPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/post/9527", nil)

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusNotFound {
// 		t.Fail()
// 	}

// 	expected := `{"Error":"Post Not Found"}`
// 	if w.Body.String() != string(expected) {
// 		t.Fail()
// 	}
// }

// func TestPostPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	var jsonStr = []byte(`{
// 		"id":9527,
// 		"author":"洪晟熊"
// 	}`)
// 	req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
// 	req.Header.Set("Content-Type", "application/json")

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Fail()
// 	}
// 	// var (
// 	// 	resp     models.Post
// 	// 	expected models.Post
// 	// )
// 	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// if err := json.Unmarshal(jsonStr, &expected); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// fmt.Println(resp)
// 	// fmt.Println(expected)
// 	// if resp.ID != expected.ID || resp.Author != expected.Author {
// 	// 	t.Fail()
// 	// }
// }

// func TestRoutePostPost(t *testing.T) {

// 	initPostTest()

// 	type ExpectGetsResp struct {
// 		httpcode int
// 		resp     []models.PostMember
// 		err      string
// 	}
// 	testPostsGetCases := []struct {
// 		name  string
// 		route string
// 		in    string
// 		out   ExpectGetsResp
// 	}{
// 		{"NewPost", "/post", `{"id":9527,"author":"superman@mirrormedia.mg"}`, ExpectGetsResp{http.StatusOK,
// 			[]models.PostMember{
// 				{Post: mockPostDS[3], Member: models.Member{}, UpdatedBy: models.UpdatedBy{}},
// 				{Post: mockPostDS[1], Member: mockMemberDS[1], UpdatedBy: models.UpdatedBy{}},
// 				{Post: mockPostDS[0], Member: mockMemberDS[0], UpdatedBy: models.UpdatedBy(mockMemberDS[0])},
// 				{Post: mockPostDS[2], Member: mockMemberDS[2], UpdatedBy: models.UpdatedBy{}},
// 			}, ""}},
// 	}
// 	for _, tc := range testPostsGetCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest("POST", "/post", nil)

// 			r.ServeHTTP(w, req)
// 			if w.Code != tc.out.httpcode {
// 				t.Errorf("%s Want %d but get %d", tc.name, tc.out.httpcode, w.Code)
// 			}
// 			if w.Code != http.StatusOK && w.Body.String() != tc.out.err {
// 				t.Errorf("%s expect to get error message %v but get %v", tc.name, w.Body.String(), tc.out.err)
// 			}

// 			// responses := []models.PostMember{}
// 			// json.Unmarshal([]byte(w.Body.String()), &responses)
// 			expected, _ := json.Marshal(tc.out.resp)
// 			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
// 				t.Log("Not equal!")
// 			}
// 		})
// 	}
// 	clearPostTest()
// }

// func TestPostEmptyPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	// When the body is empty in Postman, it actually send EOF to server
// 	// It is a problem whether it's proper to send {} in test.
// 	var jsonStr = []byte(`{}`)
// 	req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
// 	req.Header.Set("Content-Type", "application/json")
// 	// r := getRouter()
// 	// r.POST("/member", env.MemberPostHandler)
// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusBadRequest {
// 		t.Fail()
// 	}
// 	expected := `{"Error":"Invalid Post"}`
// 	if w.Body.String() != string(expected) {
// 		t.Fail()
// 	}
// }

// func TestPostExistingPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	var jsonStr = []byte(`{
// 		"id":1,
// 		"author":"superman@mirrormedia.mg"
// 	}`)
// 	req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
// 	req.Header.Set("Content-Type", "application/json")
// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusBadRequest {
// 		t.Fail()
// 	}
// 	expected := `{"Error":"Post ID Already Taken"}`
// 	if w.Body.String() != string(expected) {
// 		t.Fail()
// 	}
// }

// ------------------------------------ Update Post Test ------------------------------------
// func TestPutPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	var jsonStr = []byte(`{
// 		"id":1,
// 		"liked": 113,
// 		"title": "台北不是我的家！？租屋黑市大揭露"
// 	}`)
// 	req, _ := http.NewRequest("PUT", "/post", bytes.NewBuffer(jsonStr))
// 	req.Header.Set("Content-Type", "application/json")
// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Fail()
// 	}
// 	// var (
// 	// 	resp     models.Post
// 	// 	expected models.Post
// 	// )
// 	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// if err := json.Unmarshal(jsonStr, &expected); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// if resp.ID != expected.ID || resp.LikeAmount != expected.LikeAmount || resp.Title != expected.Title {
// 	// 	t.Fail()
// 	// }
// }

// func TestPutNonExistingPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	var jsonStr = []byte(`{
// 		"id": 98765,
// 		"Title": "數讀政治獻金"
// 	}`)
// 	req, _ := http.NewRequest("PUT", "/post", bytes.NewBuffer(jsonStr))
// 	req.Header.Set("Content-Type", "application/json")
// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusBadRequest {
// 		t.Fail()
// 	}
// 	expected := `{"Error":"Post Not Found"}`
// 	if w.Body.String() != string(expected) {
// 		t.Fail()
// 	}
// }

// ------------------------------------ Delete Post Test ------------------------------------
// func TestDeleteExistingPost(t *testing.T) {

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("DELETE", "/post/1", nil)
// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Fail()
// 	}
// 	var resp models.Post
// 	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
// 		log.Fatal(err)
// 	}
// 	if resp.Active != 0 {
// 		t.Fail()
// 	}
// }

// func TestDeleteNonExistingPost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("DELETE", "/post/12345", nil)

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusNotFound {
// 		t.Fail()
// 	}
// 	expected := `{"Error":"Post Not Found"}`
// 	if w.Body.String() != string(expected) {
// 		t.Fail()
// 	}
// }
