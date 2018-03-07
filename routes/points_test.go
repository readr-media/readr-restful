package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/readr-media/readr-restful/models"
)

type mockPointsAPI struct {
	mockPointsDS []models.Points
}
type testInterface interface {
	setup(in interface{})
	teardown()
}

func (a *mockPointsAPI) setup(in interface{}) {
	a.mockPointsDS = make([]models.Points, len(in.([]models.Points)))
	copy(a.mockPointsDS, in.([]models.Points))
	models.PointsAPI = a
}

func (a *mockPointsAPI) teardown() {
	a.mockPointsDS = nil
}

func (a *mockPointsAPI) Get(id string, objType *int64) (result []models.Points, err error) {
	for _, v := range a.mockPointsDS {
		if v.MemberID == id {
			if objType != nil {
				if v.ObjectType == int(*objType) {
					result = append(result, v)
				}
			} else {
				result = append(result, v)
			}
		}
	}
	if len(result) == 0 {
		err = errors.New("Points Not Found")
	}
	return result, err
}

func (a *mockPointsAPI) Insert(pts models.Points) (result int, err error) {

	if total, err := a.Get(pts.MemberID, nil); err == nil {
		for _, v := range total {
			if v.ObjectType != pts.ObjectType {
				result += int(v.Points)
			} else {
				return 0, errors.New("Duplicate entry")
			}
		}
		a.mockPointsDS = append(a.mockPointsDS, pts)
		result += pts.Points
	} else if err.Error() == "Points Not Found" {
		a.mockPointsDS = append(a.mockPointsDS, pts)
		result = pts.Points
	}
	return result, err
}

func (a *mockPointsAPI) Update(pts models.Points) (result int, err error) {

	if total, err := a.Get(pts.MemberID, nil); err == nil {
		for k, v := range total {
			if v.ObjectType != pts.ObjectType {
				result += int(v.Points)
			} else {
				a.mockPointsDS[k] = pts
				result += pts.Points
			}
		}
	} else if err.Error() == "Points Not Found" {
		return result, errors.New("Points Not Found")
	}
	return result, err
}

func TestRoutePoints(t *testing.T) {

	var pointTest mockPointsAPI

	points := []models.Points{
		models.Points{
			MemberID:   "superman@superman.com",
			ObjectType: 1,
			ObjectID:   1,
			Points:     500,
			CreatedAt:  models.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  models.NullString{String: "wonderwomen@wonderwomen.com", Valid: true},
			UpdatedAt:  models.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
		models.Points{
			MemberID:   "superman@superman.com",
			ObjectType: 2,
			ObjectID:   3,
			Points:     300,
			CreatedAt:  models.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  models.NullString{String: "superman@superman.com", Valid: true},
			UpdatedAt:  models.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
		models.Points{
			MemberID:   "wonderwomen@wonderwomen.com",
			ObjectType: 1,
			ObjectID:   1,
			Points:     500,
			CreatedAt:  models.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  models.NullString{String: "wonderwomen@wonderwomen.com", Valid: true},
			UpdatedAt:  models.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	}

	teststep := []TestStep{
		TestStep{
			name:     "GET",
			init:     func() { pointTest.setup(points) },
			teardown: func() { pointTest.teardown() },
			register: &pointTest,
			cases: []genericTestcase{
				genericTestcase{"BothTypePoints", "GET", `/points/superman@superman.com`, ``, http.StatusOK, []models.Points{points[0], points[1]}},
				genericTestcase{"SingleTypePoints", "GET", `/points/superman@superman.com/2`, ``, http.StatusOK, []models.Points{points[1]}},
				genericTestcase{"PointsNotFound", "GET", `/points/wonderwomen@wonderwomen.com/2`, ``, http.StatusNotFound, `{"Error":"Points Not Found"}`},
			},
		},
		TestStep{
			name:     "POST",
			init:     func() { pointTest.setup(points) },
			teardown: func() { pointTest.teardown() },
			register: &pointTest,
			cases: []genericTestcase{
				genericTestcase{"BasicPoints", "POST", `/points`, `{"member_id":"wonderwomen@wonderwomen.com","object_type": 2,"object_id": 1,"points": 100}`, http.StatusOK, `{"points":600}`},
				genericTestcase{"DuplicatePoints", "POST", `/points`, `{"member_id":"superman@superman.com","object_type": 2,"object_id": 1,"points": 100}`, http.StatusBadRequest, `{"Error":"Already exists"}`},
			},
		},
		TestStep{
			name:     "PUT",
			init:     func() { pointTest.setup(points) },
			teardown: func() { pointTest.teardown() },
			register: &pointTest,
			cases: []genericTestcase{
				genericTestcase{"SinglePoints", "PUT", `/points`, `{"member_id":"superman@superman.com","object_type":1,"object_id":1,"points":100}`, http.StatusOK, `{"points":400}`},
				genericTestcase{"NotExistedPoints", "PUT", `/points`, `{"member_id":"flash@flash.com","object_type":1,"object_id":1,"points":100}`, http.StatusBadRequest, `{"Error":"Points Not Found"}`},
			},
		},
	}
	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.Points `json:"_items"`
		}

		var Response response
		var expected = tc.resp.([]models.Points)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", tc.name, resp)
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		}

	OuterLoop:
		for _, respPoints := range Response.Items {
			for _, expPoints := range expected {
				if reflect.DeepEqual(respPoints, expPoints) {
					continue OuterLoop
				}
			}
			t.Errorf("%s, Not expect to get %v ", tc.name, respPoints)
		}
	}
	for _, ts := range teststep {
		t.Run(ts.name, func(t *testing.T) {
			DoTest(t, ts, asserter)
		})
	}
}

type TestStep struct {
	init     func()
	teardown func()
	register testInterface
	name     string
	cases    []genericTestcase
}

func DoTest(t *testing.T, ts TestStep, function interface{}) {
	ts.init()
	for _, tc := range ts.cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var jsonStr = []byte{}
			if s, ok := tc.body.(string); ok {
				jsonStr = []byte(s)
			} else {
				p, err := json.Marshal(tc.body)
				if err != nil {
					t.Errorf("%s, Error when marshaling input parameters", tc.name)
				}
				jsonStr = p
			}
			req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tc.httpcode {
				t.Errorf("%s want %d but get %d", tc.name, tc.httpcode, w.Code)
			}
			switch tc.resp.(type) {
			case string:
				if w.Body.String() != tc.resp {
					t.Errorf("%s expect (error) message %v but get %v", tc.name, tc.resp, w.Body.String())
				}
			default:
				if fn, ok := function.(func(resp string, tc genericTestcase, t *testing.T)); ok {
					fn(w.Body.String(), tc, t)
				}
			}
		})
	}
	ts.teardown()
}
