package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockPostAPI struct{}

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

// ---------------------------------- Post Test -------------------------------

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
