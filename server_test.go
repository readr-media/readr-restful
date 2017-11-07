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

var memberlist = []models.Member{
	models.Member{
		ID:     "TaiwanNo.1",
		Active: true,
	},
}

func (mdb *mockDB) Get(item models.TableStruct) (models.TableStruct, error) {

	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		for _, value := range memberlist {
			if item.ID == value.ID {
				result = value
				err = nil
			} else {
				result = models.Member{}
				err = errors.New("User Not Found")
			}
		}
	}
	return result, err
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
	req, _ := http.NewRequest("GET", "/member/TaiwanNo.1", nil)

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
	expected, _ := json.Marshal(models.Member{ID: "TaiwanNo.1", Active: true})
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
