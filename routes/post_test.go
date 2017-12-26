package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockPostAPI struct{}

func (a *mockPostAPI) GetPosts(maxResult uint8, page uint16, sortMethod string) ([]models.Post, error) {

	var (
		result []models.Post
		err    error
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

	if maxResult >= uint8(len(sortedMockPostDS)) {
		result = sortedMockPostDS
		err = nil
	} else if maxResult < uint8(len(sortedMockPostDS)) {
		result = sortedMockPostDS[(page-1)*uint16(maxResult) : page*uint16(maxResult)]
		err = nil
	}
	return result, err
}

func (a *mockPostAPI) GetPost(id uint32) (models.Post, error) {
	result := models.Post{}
	err := errors.New("Post Not Found")
	for _, value := range mockPostDS {
		if value.ID == id {
			result = value
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

func (a *mockPostAPI) DeletePost(id uint32) (models.Post, error) {
	result := models.Post{}
	err := errors.New("Post Not Found")
	for index, value := range mockPostDS {
		if value.ID == id {
			mockPostDS[index].Active = 0
			result = mockPostDS[index]
			err = nil
			return result, err
		}
	}
	return result, err
}

// // ---------------------------------- Post Test -------------------------------
func TestGetPostsDescending(t *testing.T) {
	w := httptest.NewRecorder()
	// default sort -updated_at
	req, _ := http.NewRequest("GET", "/posts", nil)

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("HTTP response code error: want %q but get %q", http.StatusOK, w.Code)
	}
	expected, _ := json.Marshal(mockPostDS)
	if w.Body.String() != string(expected) {
		t.Errorf("Response error: Not Expected")
	}
}

func TestGetPostsAscending(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts?sort=updated_at", nil)

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("HTTP response code error: want %q but get %q", http.StatusOK, w.Code)
	}
	res := []models.Post{}
	json.Unmarshal([]byte(w.Body.String()), &res)
	if res[0] != mockPostDS[2] || res[1] != mockPostDS[1] || res[2] != mockPostDS[0] {
		t.Errorf("Response sort error")
	}
}

func TestGetExistPost(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/post/3345678", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	expected, _ := json.Marshal(mockPostDS[0])
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestGetNonExistedPost(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/post/9527", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}

	expected := `{"Error":"Post Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostPost(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":9527,
		"author":"洪晟熊"
	}`)
	req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	// var (
	// 	resp     models.Post
	// 	expected models.Post
	// )
	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
	// 	log.Fatal(err)
	// }
	// if err := json.Unmarshal(jsonStr, &expected); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(resp)
	// fmt.Println(expected)
	// if resp.ID != expected.ID || resp.Author != expected.Author {
	// 	t.Fail()
	// }
}

func TestPostEmptyPost(t *testing.T) {
	w := httptest.NewRecorder()
	// When the body is empty in Postman, it actually send EOF to server
	// It is a problem whether it's proper to send {} in test.
	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Invalid Post"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostExistingPost(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":3345678,
		"author":"李宥儒"
	}`)
	req, _ := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Post ID Already Taken"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

// ------------------------------------ Update Post Test ------------------------------------
func TestPutPost(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":3345678,
		"liked": 113,
		"title": "台北不是我的家！？租屋黑市大揭露"
	}`)
	req, _ := http.NewRequest("PUT", "/post", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	// var (
	// 	resp     models.Post
	// 	expected models.Post
	// )
	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
	// 	log.Fatal(err)
	// }
	// if err := json.Unmarshal(jsonStr, &expected); err != nil {
	// 	log.Fatal(err)
	// }
	// if resp.ID != expected.ID || resp.LikeAmount != expected.LikeAmount || resp.Title != expected.Title {
	// 	t.Fail()
	// }
}

func TestPutNonExistingPost(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id": 98765,
		"Title": "數讀政治獻金"
	}`)
	req, _ := http.NewRequest("PUT", "/post", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Post Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

// ------------------------------------ Delete Post Test ------------------------------------
func TestDeleteExistingPost(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/post/3345678", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var resp models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if resp.Active != 0 {
		t.Fail()
	}
}

func TestDeleteNonExistingPost(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/post/12345", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}
	expected := `{"Error":"Post Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}
