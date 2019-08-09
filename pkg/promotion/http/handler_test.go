package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/promotion"
	"github.com/readr-media/readr-restful/pkg/promotion/mock"
	"github.com/readr-media/readr-restful/pkg/promotion/mysql"
)

func TestPromotionHandlerList(t *testing.T) {
	// Setting up mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// inject database dependency with mock
	mockData := mock.NewMockDataLayer(ctrl)
	mysql.DataAPI = mockData

	gin.SetMode(gin.TestMode)
	r := gin.New()

	Router.SetRoutes(r)

	// table-driven testcases
	for _, tc := range []struct {
		name     string
		httpcode int
		path     string
		params   *ListParams
		err      string
	}{
		// Should we use gomock.Any() to replace &ListParams{}?
		{"default-params", http.StatusOK, `/promotions`, &ListParams{MaxResult: 15, Page: 1, Sort: "-created_at", Active: map[string][]int{"$in": []int{1}}}, ``},
		{"max-result", http.StatusOK, `/promotions?max_result=25`, &ListParams{MaxResult: 25, Page: 1, Sort: "-created_at", Active: map[string][]int{"$in": []int{1}}}, ``},
		{"page", http.StatusOK, `/promotions?page=7`, &ListParams{MaxResult: 15, Page: 7, Sort: "-created_at", Active: map[string][]int{"$in": []int{1}}}, ``},
		{"modify-invalid-sort", http.StatusOK, `/promotions?sort=updated_by`, &ListParams{MaxResult: 15, Page: 1, Sort: "-created_at", Active: map[string][]int{"$in": []int{1}}}, ``},
		{"active", http.StatusOK, `/promotions?active=$in:0,1`, &ListParams{MaxResult: 15, Page: 1, Sort: "-created_at", Active: map[string][]int{"$in": []int{0, 1}}}, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)

			// expect database service Get to be called once
			// gomock could validate the input argument
			if tc.httpcode == http.StatusOK {
				mockData.EXPECT().Get(tc.params).Times(1)
			}

			r.ServeHTTP(w, req)
			// Check return http status code
			assert.Equal(t, w.Code, tc.httpcode)
			// Check return error if http status is not 200
			if tc.httpcode != http.StatusOK && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}

func TestPromotionHandlerPost(t *testing.T) {
	// Setting up mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// inject database dependency with mock
	mockData := mock.NewMockDataLayer(ctrl)
	mysql.DataAPI = mockData

	gin.SetMode(gin.TestMode)
	r := gin.New()

	Router.SetRoutes(r)

	// table-driven testcases
	for _, tc := range []struct {
		name     string
		httpcode int
		body     promotion.Promotion
		err      string
	}{
		{"empty-payload", http.StatusBadRequest, promotion.Promotion{}, `{"error":"null promotion payload"}`},
		{"empty-title", http.StatusBadRequest, promotion.Promotion{Status: 1}, `{"error":"invalid title"}`},
		{"default-promo", http.StatusCreated, promotion.Promotion{Status: 1, Active: 1, Title: "test", CreatedAt: time.Now().Local(), UpdatedAt: models.NullTime{Time: time.Now().Local(), Valid: true}}, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// marshaling insert payload
			p, err := json.Marshal(tc.body)
			if err != nil {
				t.Errorf("%s, error when marshaling input parameters", tc.name)
			}
			req, _ := http.NewRequest("POST", `/promotions`, bytes.NewBuffer(p))
			// Set request header or the payload could not be binded
			req.Header.Set("Content-Type", "application/json")

			// if the expected http status is 201,
			// expect database service Insert to be called once
			if tc.httpcode == http.StatusCreated {
				mockData.EXPECT().Insert(gomock.Any()).Times(1)
			}

			r.ServeHTTP(w, req)
			// Check return http status code
			assert.Equal(t, w.Code, tc.httpcode)
			// Check return error if http status is not 200
			if tc.httpcode != http.StatusCreated && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}

func TestPromotionHandlerPut(t *testing.T) {
	// Setting up mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// inject database dependency with mock
	mockData := mock.NewMockDataLayer(ctrl)
	mysql.DataAPI = mockData

	gin.SetMode(gin.TestMode)
	r := gin.New()

	Router.SetRoutes(r)

	// table-driven testcases
	for _, tc := range []struct {
		name     string
		httpcode int
		body     promotion.Promotion
		err      string
	}{
		{"empty-payload", http.StatusBadRequest, promotion.Promotion{}, `{"error":"null promotion payload"}`},
		{"zero-id", http.StatusBadRequest, promotion.Promotion{ID: 0, Status: 1}, `{"error":"invalid promotion id"}`},
		{"default-promo", http.StatusNoContent, promotion.Promotion{ID: 128, Status: 1, Active: 0, CreatedAt: time.Now().Local(), UpdatedAt: models.NullTime{Time: time.Now().Local(), Valid: true}}, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// marshaling insert payload
			p, err := json.Marshal(tc.body)
			if err != nil {
				t.Errorf("%s, error when marshaling input parameters", tc.name)
			}
			req, _ := http.NewRequest("PUT", `/promotions`, bytes.NewBuffer(p))
			// Set request header or the payload could not be binded
			req.Header.Set("Content-Type", "application/json")

			// if the expected http status is 204,
			// expect database service Insert to be called once
			if tc.httpcode == http.StatusNoContent {
				mockData.EXPECT().Update(gomock.Any()).Times(1)
			}

			r.ServeHTTP(w, req)
			// Check return http status code
			assert.Equal(t, w.Code, tc.httpcode)
			// Check return error if http status is not 200
			if tc.httpcode != http.StatusNoContent && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}
func TestPromotionHandlerDelete(t *testing.T) {
	// Setting up mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// inject database dependency with mock
	mockData := mock.NewMockDataLayer(ctrl)
	mysql.DataAPI = mockData

	gin.SetMode(gin.TestMode)
	r := gin.New()

	Router.SetRoutes(r)

	// table-driven testcases
	for _, tc := range []struct {
		name     string
		httpcode int
		path     string
		id       uint64
		err      string
	}{
		{"absent-id", http.StatusNotFound, `/promotions`, 0, ``},
		{"invalid-id", http.StatusBadRequest, `/promotions/foo`, 0, `{"error":"unable to parse id:foo"}`},
		{"simple-delete", http.StatusOK, `/promotions/1`, 1, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", tc.path, nil)

			// expect database service Get to be called once
			// gomock could validate the input argument
			if tc.httpcode == http.StatusOK {
				mockData.EXPECT().Delete(tc.id).Times(1)
			}

			r.ServeHTTP(w, req)
			// Check return http status code
			assert.Equal(t, w.Code, tc.httpcode)
			// Check return error if http status is not 200
			if tc.httpcode != http.StatusOK && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}

func TestMain(m *testing.M) {
	if err := config.LoadConfig("../../../config", ""); err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}
	os.Exit(m.Run())
}
