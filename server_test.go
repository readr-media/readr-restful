package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type mockDB struct{}

func (mdb *mockDB) Get(id string) (interface{}, error) {
	if id == "jjo4iTaiwan" {
		member := models.Member{ID: "jjo4iTaiwan", Active: true}
		return member, nil
	}
	return models.Member{}, errors.New("User Not Found")
}

func (mdb *mockDB) Create(interface{}) (interface{}, error) {
	return models.Member{}, nil
}

func (mdb *mockDB) Update(interface{}) (interface{}, error) {
	return models.Member{}, nil
}

func (mdb *mockDB) Delete(id string) (interface{}, error) {
	return models.Member{}, nil
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

func getRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestGetExistMember(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/jjo4iTaiwan", nil)

	env := Env{db: &mockDB{}}

	// r := gin.Default()
	r := getRouter()
	r.GET("/member/:id", env.MemberGetHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	// fmt.Println(w.Code)
	if w.Code != http.StatusOK {
		t.Fail()
	}
	expected, _ := json.Marshal(models.Member{ID: "jjo4iTaiwan", Active: true})
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestGetNotExistMember(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/abc123", nil)

	env := Env{db: &mockDB{}}

	// r := gin.Default()
	r := getRouter()
	r.GET("/member/:id", env.MemberGetHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}

	expected := ""
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}
