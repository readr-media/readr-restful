package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type mockDB struct{}

var memberList = []models.Member{
	models.Member{
		ID:     "TaiwanNo.1",
		Active: true,
	},
}

var articleList = []models.Article{
	models.Article{
		ID:     "3345678",
		Author: models.NullString{String: "李宥儒", Valid: true},
		Active: 1,
	},
}
var env Env

// ------------------------ Implementation of Datastore interface ---------------------------
func (mdb *mockDB) Get(item models.TableStruct) (models.TableStruct, error) {

	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for _, value := range memberList {
			if item.ID == value.ID {
				result = value
				err = nil
			}
		}
	case models.Article:
		result = models.Member{}
		err = errors.New("Article Not Found")
		for _, value := range articleList {
			if item.ID == value.ID {
				result = value
				err = nil
			}
		}
	default:
		log.Fatal("Can't not parse model type")
	}
	return result, err
}

func (mdb *mockDB) Create(item models.TableStruct) (interface{}, error) {

	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		for _, member := range memberList {
			if item.ID == member.ID {
				return models.Member{}, errors.New("Duplicate entry")
			}
		}
		memberList = append(memberList, item)
		result = memberList[len(memberList)-1]
		err = nil
	case models.Article:
		for _, article := range articleList {
			if item.ID == article.ID {
				result = models.Article{}
				err = errors.New("Duplicate entry")
				return result, err
			}
		}
		articleList = append(articleList, item)
		result = articleList[len(articleList)-1]
		err = nil
	}
	return result, err
}

func (mdb *mockDB) Update(item models.TableStruct) (interface{}, error) {
	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for _, value := range memberList {
			if value.ID == item.ID {
				result = item
				err = nil
			}
		}
	case models.Article:
		result = models.Article{}
		err = errors.New("Article Not Found")
		for index, value := range articleList {
			if value.ID == item.ID {
				articleList[index].LikeAmount = item.LikeAmount
				articleList[index].Title = item.Title
				return articleList[index], nil
			}
		}
	}
	return result, err
}

func (mdb *mockDB) Delete(item models.TableStruct) (interface{}, error) {
	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for index, value := range memberList {
			if item.ID == value.ID {
				memberList[index].Active = false
				return memberList[index], nil
			}
		}
	case models.Article:
		result = models.Article{}
		err = errors.New("Article Not Found")
		for index, value := range articleList {
			if item.ID == value.ID {
				articleList[index].Active = 0
				return articleList[index], nil
			}
		}
	default:
		log.Fatal("Can't not parse model type")
	}
	return result, err
}

// ---------------------------------- End of Datastore implementation --------------------------------
// var r = gin.Default()
var r *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	r = gin.Default()
	r.GET("/member/:id", env.MemberGetHandler)
	r.POST("/member", env.MemberPostHandler)
	r.PUT("/member", env.MemberPutHandler)
	r.DELETE("/member/:id", env.MemberDeleteHandler)

	r.GET("/article/:id", env.ArticleGetHandler)
	r.POST("/article", env.ArticlePostHandler)
	r.PUT("/article", env.ArticlePutHandler)
	r.DELETE("/article/:id", env.ArticleDeleteHandler)

	env.db = &mockDB{}
	os.Exit(m.Run())
}

// func getRouter() *gin.Engine {
// 	r := gin.Default()
// 	return r
// }

func TestGetExistMember(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/TaiwanNo.1", nil)

	// r := getRouter()
	// r.GET("/member/:id", env.MemberGetHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	expected, _ := json.Marshal(memberList[0])
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestGetNotExistMember(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/abc123", nil)

	// r := getRouter()
	// r.GET("/member/:id", env.MemberGetHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}

	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostEmptyMember(t *testing.T) {

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/member", nil)
	// req.Header.Set("Content-Type", "application/json")
	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Invalid User"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"spaceoddity", 
		"name":"Major Tom"
	}`)
	req, _ := http.NewRequest("POST", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var (
		resp     models.Member
		expected models.Member
	)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(jsonStr, &expected); err != nil {
		log.Fatal(err)
	}
	if resp.ID != expected.ID || resp.Name != expected.Name {
		t.Fail()
	}
}

func TestPostExistedMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{"id":"TaiwanNo.1"}`)
	req, _ := http.NewRequest("POST", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"User Already Existed"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPutMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"TaiwanNo.1", 
		"name":"Major Tom"
	}`)
	req, _ := http.NewRequest("PUT", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.PUT("/member", env.MemberPutHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var (
		resp     models.Member
		expected models.Member
	)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(jsonStr, &expected); err != nil {
		log.Fatal(err)
	}
	if resp.ID != expected.ID || resp.Name != expected.Name {
		t.Fail()
	}
}

func TestPutNonExistMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
			"id":"ChinaNo.19", 
			"name":"Major Tom"
		}`)
	req, _ := http.NewRequest("PUT", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.PUT("/member", env.MemberPutHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestDeleteExistMember(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/member/TaiwanNo.1", nil)

	// r := getRouter()
	// r.DELETE("/member/:id", env.MemberDeleteHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var resp models.Member
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if resp.Active == true {
		t.Fail()
	}
}

func TestDeleteNonExistMember(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/member/ChinaNo.19", nil)

	// r := getRouter()
	// r.DELETE("/member/:id", env.MemberDeleteHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}
	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

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
