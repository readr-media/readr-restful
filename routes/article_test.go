package routes

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

// ---------------------------------- Article Test -------------------------------

func TestGetExistArticle(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/article/3345678", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	expected, _ := json.Marshal(articleList[0])
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestGetNonExistedArticle(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/article/9527", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}

	expected := `{"Error":"Article Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostArticle(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"9527",
		"author":"洪晟熊"
	}`)
	req, _ := http.NewRequest("POST", "/article", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var (
		resp     models.Article
		expected models.Article
	)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(jsonStr, &expected); err != nil {
		log.Fatal(err)
	}
	if resp.ID != expected.ID || resp.Author != expected.Author {
		t.Fail()
	}
}

func TestPostEmptyArticle(t *testing.T) {
	w := httptest.NewRecorder()
	// When the body is empty in Postman, it actually send EOF to server
	// It is a problem whether it's proper to send {} in test.
	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("POST", "/article", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Invalid Article ID"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostExistingArticle(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"3345678",
		"author":"李宥儒"
	}`)
	req, _ := http.NewRequest("POST", "/article", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Article ID Already Taken"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

// ------------------------------------ Update Article Test ------------------------------------
func TestPutArticle(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"3345678",
		"liked": 113,
		"title": "台北不是我的家！？租屋黑市大揭露"
	}`)
	req, _ := http.NewRequest("PUT", "/article", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var (
		resp     models.Article
		expected models.Article
	)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(jsonStr, &expected); err != nil {
		log.Fatal(err)
	}
	if resp.ID != expected.ID || resp.LikeAmount != expected.LikeAmount || resp.Title != expected.Title {
		t.Fail()
	}
}

func TestPutNonExistingArticle(t *testing.T) {
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id": "98765",
		"Title": "數讀政治獻金"
	}`)
	req, _ := http.NewRequest("PUT", "/article", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Article Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

// ------------------------------------ Delete Article Test ------------------------------------
func TestDeleteExistingArticle(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/article/3345678", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var resp models.Article
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if resp.Active != 0 {
		t.Fail()
	}
}

func TestDeleteNonExistingArticle(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/article/12345", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}
	expected := `{"Error":"Article Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}
