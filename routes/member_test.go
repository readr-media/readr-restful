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
