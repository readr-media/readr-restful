package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/readr-media/readr-restful/pkg/subscription"
	"github.com/readr-media/readr-restful/pkg/subscription/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionsHandlerGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockData := mock.NewMockSubscriber(ctrl)
	Router.Service = mockData

	r := gin.New()
	Router.SetRoutes(r)

	for _, tc := range []struct {
		name     string
		httpcode int
		path     string
		params   *ListRequest
		err      string
	}{
		{"default-params", http.StatusOK, `/subscriptions`, &ListRequest{}, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)

			if tc.httpcode == http.StatusOK {
				mockData.EXPECT().GetSubscriptions(tc.params).Times(1)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.httpcode, w.Code)

			if tc.httpcode != http.StatusOK && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}

func TestSubscriptionsHandlerPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockData := mock.NewMockSubscriber(ctrl)
	Router.Service = mockData

	r := gin.New()
	Router.SetRoutes(r)
	for _, tc := range []struct {
		name     string
		httpcode int
		body     subscription.Subscription
		err      string
	}{
		{"default", http.StatusCreated, subscription.Subscription{ID: 1, MemberID: 648, Amount: 100}, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			p, err := json.Marshal(tc.body)
			if err != nil {
				t.Errorf("%s, error when marshaling input parameters", tc.name)
			}
			req, _ := http.NewRequest("POST", `/subscriptions`, bytes.NewBuffer(p))

			req.Header.Set("Content-Type", "application/json")

			if tc.httpcode == http.StatusCreated {
				mockData.EXPECT().CreateSubscription(gomock.Any()).Times(1)
			}

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.httpcode, w.Code)

			if tc.httpcode != http.StatusCreated && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}
func TestSubscriptionsHandlerPut(t *testing.T) {

}

func TestSubscriptionsHandlerRecurringPay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockData := mock.NewMockSubscriber(ctrl)
	Router.Service = mockData

	r := gin.New()
	Router.SetRoutes(r)

	for _, tc := range []struct {
		name     string
		httpcode int
		err      string
	}{
		{"default", http.StatusAccepted, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", `/subscriptions/recurring`, nil)

			mockData.EXPECT().GetSubscriptions(gomock.Any()).Times(1)
			mockData.EXPECT().RoutinePay(gomock.Any()).Times(1)

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.httpcode, w.Code)
			if tc.httpcode != w.Code && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}
