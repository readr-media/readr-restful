package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

type mockPointsAPI struct{}

type testInterface interface {
	setup(in interface{})
	teardown()
}

var mockPointsDS = []models.PointsProject{
	models.PointsProject{
		Points: models.Points{
			PointsID:   1,
			MemberID:   1,
			ObjectType: 1,
			ObjectID:   1,
			Points:     500,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 1, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
	models.PointsProject{
		Points: models.Points{
			PointsID:   2,
			MemberID:   1,
			ObjectType: 2,
			ObjectID:   3,
			Points:     300,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 0, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
	models.PointsProject{
		Points: models.Points{
			PointsID:   3,
			MemberID:   2,
			ObjectType: 1,
			ObjectID:   1,
			Points:     500,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 1, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
	models.PointsProject{
		Points: models.Points{
			PointsID:   4,
			MemberID:   1,
			ObjectType: 2,
			ObjectID:   23,
			Points:     100,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 1, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 5, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
}

func (a *mockPointsAPI) Get(args *models.PointsArgs) (result []models.PointsProject, err error) {
	for _, v := range mockPointsDS {

		// Check member_id
		if v.MemberID == args.ID {
			// Check Object Type
			if args.ObjectType != nil {
				if v.ObjectType == int(*args.ObjectType) {
					// Check if there are object_id filter
					if args.ObjectIDs != nil {
						for _, o := range args.ObjectIDs {
							if v.ObjectID == o {
								result = append(result, v)
							}
						}
					} else {
						result = append(result, v)
					}
				}
			} else {
				result = append(result, v)
			}
		}
	}
	return result, err
}

func (a *mockPointsAPI) Insert(pts models.PointsToken) (result int, id int, err error) {

	args := models.PointsArgs{ID: pts.MemberID}
	if total, err := a.Get(&args); err == nil {
		for _, v := range total {
			result += int(v.Points.Points)
		}
		// mockPointsDS = append(a.mockPointsDS, models.PointsProject{Points: pts, Title: rrsql.NullString{"", false}})
		result -= pts.Points.Points
	}
	return result, 1, err
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
				t.Errorf("%s want HTTP code %d but get %d", tc.name, tc.httpcode, w.Code)
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

func TestRoutePoints(t *testing.T) {

	//Only Test if query parameter change would result in correct points history number change
	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		var Response = struct {
			Items []models.PointsProject `json:"_items"`
		}{}
		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v\n", tc.name, resp)
		}
		if len(Response.Items) != tc.resp.(int) {
			t.Errorf("%s, expect points history length to be %v but get %d", tc.name, tc.resp, len(Response.Items))
		}
	}
	t.Run("Get", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"BothTypePoints", "GET", `/points/1`, ``, http.StatusOK, 3},
			genericTestcase{"SingleTypePoints", "GET", `/points/1/2`, ``, http.StatusOK, 2},
			genericTestcase{"WithObjectID", "GET", `/points/1/2?object_ids=[23]`, ``, http.StatusOK, 1},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("Insert", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"Deprecated Object Type: project", "POST", `/points`, `{"member_id":1,"object_type": 1}`, http.StatusBadRequest, `{"Error":"ObjectType Deprecated"}`},
			genericTestcase{"Deprecated Object Type: topup", "POST", `/points`, `{"member_id":1,"object_type": 3}`, http.StatusBadRequest, `{"Error":"ObjectType Deprecated"}`},
			genericTestcase{"Invalid Currency Value", "POST", `/points`, `{"member_id":1,"currency": -100,"object_type":5}`, http.StatusBadRequest, `{"Error":"Invalid Payment Amount"}`},
			genericTestcase{"Invalid ObjectType For Currency", "POST", `/points`, `{"member_id":1,"object_type": 4,"currency": 100}`, http.StatusBadRequest, `{"Error":"Currency Not Supported By ObjectType"}`},
			genericTestcase{"Missing Payment Token", "POST", `/points`, `{"member_id":1,"object_type": 5,"currency": 100}`, http.StatusBadRequest, `{"Error":"Invalid Token"}`},
			genericTestcase{"InvalidMemberInfo", "POST", `/points`, `{"member_id":1,"object_type": 5,"currency": 100, "token": "token"}`, http.StatusBadRequest, `{"Error":"Invalid Payment Info"}`},
			genericTestcase{"InvalidObjectID", "POST", `/points`, `{"member_id":1,"object_type": 2,"points": 100}`, http.StatusBadRequest, `{"Error":"Invalid Object ID"}`},
			genericTestcase{"InvalidObjectID", "POST", `/points`, `{"object_type": 2,"points": 100}`, http.StatusBadRequest, `{"Error":"Invalid ObjectType With Anonymous User"}`},

			genericTestcase{"Basic Project Memo", "POST", `/points`, `{"member_id":1,"object_type": 2,"object_id": 1,"currency": 50,"points": 50,"token":"token","member_name":"name","member_phone":"phone","member_mail":"mail"}`, http.StatusOK, `{"id":1,"points":850}`},
			genericTestcase{"Basic Gift", "POST", `/points`, `{"member_id":1,"object_type": 4,"object_id": 1,"points": -50, "reason": "System"}`, http.StatusOK, `{"id":1,"points":950}`},
			genericTestcase{"Basic Donate", "POST", `/points`, `{"member_id":1,"object_type": 5,"object_id": 1,"currency": 100,"token":"token","member_name":"name","member_phone":"phone","member_mail":"mail"}`, http.StatusOK, `{"id":1,"points":900}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
}
