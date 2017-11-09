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
		for _, value := range memberList {
			if item.ID == value.ID {
				return models.Member{}, errors.New("Duplicate User")
			}
		}
		memberList = append(memberList, item)
		result = memberList[len(memberList)-1]
		err = nil
	}

	return result, err
}

func (mdb *mockDB) Update(item models.TableStruct) (interface{}, error) {
	var (
		result models.Member
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
	}
	return result, err
}

func (mdb *mockDB) Delete(item models.TableStruct) (interface{}, error) {
	var (
		result models.Member
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for index, value := range memberList {
			if item.ID == value.ID {
				memberList[index].Active = false
				result = memberList[index]
				err = nil
			}
		}
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

	// fmt.Println(w.Code)
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

	// fmt.Println(w.Code)
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

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}
