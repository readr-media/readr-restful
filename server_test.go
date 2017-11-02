package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockDB struct{}

func (mdb *mockDB) Get() (interface{}, error) {
	member := models.Member{ID: "jjo4iTaiwan"}
	return member, nil
}

func TestMemberGetHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/jjo4iTaiwan", nil)

}
