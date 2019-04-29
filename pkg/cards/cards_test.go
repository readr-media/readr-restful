package cards

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type TestStep struct {
	init     func()
	teardown func()
	name     string
	cases    []genericTestcase
}

type genericTestcase struct {
	name     string
	method   string
	url      string
	body     interface{}
	httpcode int
	resp     interface{}
}

func DoTest(r *gin.Engine, t *testing.T, ts TestStep, function interface{}) {
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

type mockNewsCardAPI struct {
	mockCardDS []NewsCard
	apiBackup  NewsCardInterface
}

func (a *mockNewsCardAPI) setup(in interface{}) {
	a.mockCardDS = make([]NewsCard, len(in.([]NewsCard)))
	copy(a.mockCardDS, in.([]NewsCard))
	a.apiBackup = NewsCardAPI
	NewsCardAPI = a
}

func (a *mockNewsCardAPI) teardown() {
	NewsCardAPI = a.apiBackup
	a.apiBackup = nil
}

func (a *mockNewsCardAPI) DeleteCard(id uint32) (err error) {
	fmt.Println(id)
	err = errors.New("Card Not Found")
	for index, value := range a.mockCardDS {
		if value.ID == id {
			// mockCardDS[index].Active = models.NullInt{Int: int64(models.PostStatus["deactive"].(float64)), Valid: true}
			a.mockCardDS[index].Active = models.NullInt{Int: int64(config.Config.Models.Posts["deactive"]), Valid: true}
			return nil
		}
	}
	return err
}

func (a *mockNewsCardAPI) GetCards(args *NewsCardArgs) (result []NewsCard, err error) {

	result = []NewsCard{a.mockCardDS[3], a.mockCardDS[1], a.mockCardDS[0]}

	err = nil

	if args.Sorting == "order" {
		result = []NewsCard{a.mockCardDS[0], a.mockCardDS[1], a.mockCardDS[3]}
	} else if args.MaxResult != 0 && args.MaxResult != 15 {
		if args.Page != 0 && args.Page != 1 {
			result = []NewsCard{a.mockCardDS[1]}
		} else {
			result = []NewsCard{a.mockCardDS[3], a.mockCardDS[1]}
		}
	} else if len(args.IDs) == 1 && args.IDs[0] == 1 {
		result = []NewsCard{a.mockCardDS[0]}
	} else if args.Active != nil && len(args.Active["$in"]) != 0 && args.Status != nil {
		result = []NewsCard{a.mockCardDS[3], a.mockCardDS[1]}
	} else if args.Active != nil && len(args.Active["$in"]) != 0 {
		result = []NewsCard{a.mockCardDS[2]}
	} else if args.Status != nil {
		result = []NewsCard{a.mockCardDS[2], a.mockCardDS[0]}
	}
	return result, err
}

func (a *mockNewsCardAPI) InsertCard(c NewsCard) (int, error) {

	var id uint32
	if c.ID != 0 && a.mockCardDS[c.ID-1].ID == c.ID {
		return 0, errors.New("Duplicate entry")
	}
	if len(a.mockCardDS) != 0 {
		id = a.mockCardDS[len(a.mockCardDS)-1].ID + 1
	} else {
		id = 1
	}
	c.ID = id
	a.mockCardDS = append(a.mockCardDS, c)
	return int(c.ID), nil
}

func (a *mockNewsCardAPI) UpdateCard(c NewsCard) (err error) {
	err = errors.New("Card Not Found")
	for index, value := range a.mockCardDS {
		if value.ID == c.ID {
			a.mockCardDS[index].Title = c.Title
			err = nil
			return err
		}
	}
	return err
}

func TestRouteCards(t *testing.T) {

	var (
		configFile string
		configName string
	)
	flag.StringVar(&configFile, "path", "../../config", "Configuration file path.")
	flag.StringVar(&configName, "file", "", "Configuration file name.")
	flag.Parse()

	if err := config.LoadConfig(configFile, configName); err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	Router.SetRoutes(server)

	var cardTest mockNewsCardAPI

	cards := []NewsCard{
		{ID: 1, PostID: 1, Title: models.NullString{"Test title 01", true}, Active: models.NullInt{1, true}, Status: models.NullInt{1, true}, CreatedAt: models.NullTime{time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), true}, Order: models.NullInt{1, true}},
		{ID: 2, PostID: 2, Title: models.NullString{"Test title 02", true}, Active: models.NullInt{1, true}, Status: models.NullInt{0, true}, CreatedAt: models.NullTime{time.Date(2000, 1, 2, 3, 4, 5, 7, time.UTC), true}, Order: models.NullInt{2, true}},
		{ID: 3, PostID: 3, Title: models.NullString{"Test title 03", true}, Active: models.NullInt{0, true}, Status: models.NullInt{1, true}, CreatedAt: models.NullTime{time.Date(2000, 1, 2, 3, 4, 5, 8, time.UTC), true}, Order: models.NullInt{3, true}},
		{ID: 4, PostID: 4, Title: models.NullString{"Test title 04", true}, Active: models.NullInt{1, true}, Status: models.NullInt{0, true}, CreatedAt: models.NullTime{time.Date(2000, 1, 2, 3, 4, 5, 9, time.UTC), true}, Order: models.NullInt{4, true}},
	}

	teststep := []TestStep{
		TestStep{
			name:     "GET/cards",
			init:     func() { cardTest.setup(cards) },
			teardown: func() { cardTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"GetCards", "GET", `/cards`, ``, http.StatusOK,
					[]NewsCard{cards[3], cards[1], cards[0]}},
				genericTestcase{"GetCardSorting", "GET", `/cards?sorting=order`, ``, http.StatusOK,
					[]NewsCard{cards[0], cards[1], cards[3]}},
				genericTestcase{"GetCardSortingFailed", "GET", `/cards?sorting=image`, ``, http.StatusBadRequest, `{"Error":"Invalid Sorting Value"}`},
				genericTestcase{"GetCardsMaxResult", "GET", `/cards?max_result=2`, ``, http.StatusOK,
					[]NewsCard{cards[3], cards[1]}},
				genericTestcase{"GetCardsMaxResultAndPaging", "GET", `/cards?max_result=1&page=2`, ``, http.StatusOK,
					[]NewsCard{cards[1]}},
				genericTestcase{"GetCardsIDs", "GET", `/cards?ids=[1]`, ``, http.StatusOK,
					[]NewsCard{cards[0]}},
				genericTestcase{"GetCardsActive", "GET", `/cards?active={"$in":[0]}`, ``, http.StatusOK,
					[]NewsCard{cards[2]}},
				genericTestcase{"GetCardsStatus", "GET", `/cards?status={"$in":[1]}`, ``, http.StatusOK,
					[]NewsCard{cards[2], cards[0]}},
				genericTestcase{"GetCardsActiveAndStatus", "GET", `/cards?active={"$in":[1]}&status={"$in":[0]}`, ``, http.StatusOK,
					[]NewsCard{cards[3], cards[1]}},
				genericTestcase{"GetCardWithParam", "GET", `/cards/1`, ``, http.StatusOK,
					[]NewsCard{cards[0]}},
			},
		},
		TestStep{
			name:     "POST",
			init:     func() { cardTest.setup(cards) },
			teardown: func() { cardTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"New", "POST", `/cards`, `{"post_id":1, "title":"Test inserted card 01"}`, http.StatusOK, ``},
				genericTestcase{"EmptyPayload", "POST", `/cards`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Card"}`},
				genericTestcase{"WithoutTitle", "POST", `/cards`, `{"post_id":1}`, http.StatusBadRequest, `{"Error":"Invalid Title or CardID"}`},
				genericTestcase{"WithoutPostID", "POST", `/cards`, `{"title":"Test inserted card 03"}`, http.StatusBadRequest, `{"Error":"Invalid Title or CardID"}`},
				genericTestcase{"DuplicatePostID", "POST", `/cards`, `{"id": 2, "post_id": 2, "title": "Test inserted card 04"}`, http.StatusBadRequest, `{"Error":"Card ID Already Taken"}`},
			},
		},
		TestStep{
			name:     "PUT",
			init:     func() { cardTest.setup(cards) },
			teardown: func() { cardTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"UpdateOK", "PUT", `/cards`, `{"id":1,"title": "altered card title 01"}`, http.StatusOK, ``},
				genericTestcase{"NotExisted", "PUT", `/cards`, `{"id":12345, "title":"This should not be updated"}`, http.StatusBadRequest, `{"Error":"Card Not Found"}`},
			},
		},
		TestStep{
			name:     "DELETE",
			init:     func() { cardTest.setup(cards) },
			teardown: func() { cardTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"SimpleDelete", "DELETE", `/cards/1`, ``, http.StatusOK, ``},
				genericTestcase{"IDNotExist", "DELETE", `/cards/100`, ``, http.StatusNotFound, `{"Error":"Card Not Found"}`},
			},
		},
	}
	asserter := func(resp string, tc genericTestcase, t *testing.T) {

		type response struct {
			Items []NewsCard `json:"_items"`
		}

		var Response response
		var expected = tc.resp.([]NewsCard)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", tc.name, resp)
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect result length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		} else {
			// Exact same length
			if len(Response.Items) != 0 && len(expected) != 0 {
				for i := range expected {
					if (Response.Items[i].ID != expected[i].ID) ||
						(Response.Items[i].Title.String != expected[i].Title.String) {
						t.Errorf("%s, %vth round expect to get \n%v\n , but get \n%v\n", tc.name, i, expected[i], Response.Items[i])
					}
				}
			}
		}
	}
	for _, ts := range teststep {
		t.Run(ts.name, func(t *testing.T) {
			DoTest(server, t, ts, asserter)
		})
	}
}
