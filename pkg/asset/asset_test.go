package asset

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

type mockAssetAPI struct {
	apiBackup AssetInterface
}

func (a *mockAssetAPI) setup() {
	a.apiBackup = AssetAPI
	AssetAPI = a
}

func (a *mockAssetAPI) teardown() {
	AssetAPI = a.apiBackup
}

func (a *mockAssetAPI) Count(args *GetAssetArgs) (count int, err error) {
	return 1, nil
}

func (a *mockAssetAPI) Delete(ids []int) (err error) {
	for _, v := range ids {
		if v == 5 {
			return errors.New("Assets Not Found")
		}
	}
	return nil
}

func (a *mockAssetAPI) FilterAssets(args *GetAssetArgs) (result []FilteredAsset, err error) {
	return result, err
}

func (a *mockAssetAPI) GetAssets(args *GetAssetArgs) (result []Asset, err error) {
	return result, err
}

func (a *mockAssetAPI) Insert(asset Asset) (lastID int64, err error) {
	return lastID, err
}

func (a *mockAssetAPI) Update(asset Asset) (err error) {
	if asset.ID == 12345 {
		return errors.New("Assets Not Found")
	}
	return nil
}

func TestRouteAsset(t *testing.T) {

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

	var assetTest mockAssetAPI
	teststep := []TestStep{
		TestStep{
			name:     "Count",
			init:     func() { assetTest.setup() },
			teardown: func() { assetTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"Posts", "GET", `/asset/count`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
				genericTestcase{"Active", "GET", `/asset/count?active={"$in":[0,1]}`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
				genericTestcase{"Author", "GET", `/asset/count?author={"$nin":[0,1]}`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
				genericTestcase{"MoreThanOneActive", "GET", `/asset/count?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`},
				genericTestcase{"NotEntirelyValidActive", "GET", `/asset/count?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
				genericTestcase{"NoValidActive", "GET", `/asset/count?active={"$nin":[-3,-4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`},
				genericTestcase{"Type", "GET", `/asset/count?type={"$in":[1,2]}`, ``, http.StatusOK, `{"_meta":{"total":1}}`}},
		},
		TestStep{
			name:     "DELETE",
			init:     func() { assetTest.setup() },
			teardown: func() { assetTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"SimpleDelete", "DELETE", `/asset?ids=[1,2]`, ``, http.StatusOK, ``},
				genericTestcase{"EmptyID", "DELETE", `/asset?ids=[]`, ``, http.StatusBadRequest, `{"Error":"ID List Empty"}`},
				genericTestcase{"NotFound", "DELETE", `/asset?ids=[3,5]`, ``, http.StatusNotFound, `{"Error":"Assets Not Found"}`},
			},
		},
		TestStep{
			name:     "GET",
			init:     func() { assetTest.setup() },
			teardown: func() { assetTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"UpdatedAtDescending", "GET", `/asset`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"UpdatedAtAscending", "GET", `/asset?sort=updated_at`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"MaxResult", "GET", `/asset?max_result=2`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"AuthorFilter", "GET", `/asset?author={"$in":[2,3]}`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"ActiveFilter", "GET", `/asset?active={"$nin":[1]}`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"NotFound", "GET", `/asset?active={"$nin":[0,1]}`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"Type", "GET", `/asset?type={"$in":[1,2]}`, ``, http.StatusOK, []Asset{}},
				genericTestcase{"ShowDetails", "GET", `/asset?show_author=true&show_updater=true&show_tag=true&show_comment=true`, ``, http.StatusOK, []Asset{}},
			},
		},
		TestStep{
			name:     "POST",
			init:     func() { assetTest.setup() },
			teardown: func() { assetTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"InsertOK", "POST", `/asset`, `{"id":1,"content_type":"image","asset_type":1,"destination":"url"}`, http.StatusOK, ``},
				genericTestcase{"MissingDestination", "POST", `/asset`, `{"id":1,"asset_type":1}`, http.StatusBadRequest, `{"Error":"Missing Destination"}`},
				genericTestcase{"MissingAssetType", "POST", `/asset`, `{"id":1,"destination":"url"}`, http.StatusBadRequest, `{"Error":"Missing AssetType"}`},
				genericTestcase{"InvalidAssetType", "POST", `/asset`, `{"id":1,"destination":"url","asset_type":4}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`},
				genericTestcase{"InvalidCopyright", "POST", `/asset`, `{"id":1,"asset_type":1,"destination":"url","copyright":9}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`},
			},
		},
		TestStep{
			name:     "PUT",
			init:     func() { assetTest.setup() },
			teardown: func() { assetTest.teardown() },
			cases: []genericTestcase{
				genericTestcase{"UpdateOK", "PUT", `/asset`, `{"id":1,"content_type":"text","updated_by":1}`, http.StatusOK, ``},
				genericTestcase{"UpdateWithouotUpdater", "PUT", `/asset`, `{"id":1,"content_type":"text"}`, http.StatusBadRequest, `{"Error":"Neither updated_by or author is valid"}`},
				genericTestcase{"NotFound", "PUT", `/asset`, `{"id":12345,"content_type":"text","updated_by":1}`, http.StatusNotFound, `{"Error":"Assets Not Found"}`},
				genericTestcase{"InvalidAssetType", "POST", `/asset`, `{"id":1,"asset_type":4}`, http.StatusBadRequest, `{"Error":"Missing Destination"}`},
				genericTestcase{"InvalidCopyright", "POST", `/asset`, `{"id":1,"asset_copyright":4}`, http.StatusBadRequest, `{"Error":"Missing Destination"}`},
			},
		},
	}
	asserter := func(resp string, tc genericTestcase, t *testing.T) {

		type response struct {
			Items []Asset `json:"_items"`
		}

		var Response response
		var expected = tc.resp.([]Asset)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", tc.name, resp)
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		}
	}
	for _, ts := range teststep {
		t.Run(ts.name, func(t *testing.T) {
			DoTest(server, t, ts, asserter)
		})
	}
}

// type mockPostAPI struct{}

// func initPostTest() {
// 	mockPostDSBack = mockPostDS
// }

// func clearPostTest() {
// 	mockPostDS = mockPostDSBack
// }

type ExpectResp struct {
	httpcode int
	err      string
}

func memberToBasic(m models.Member) (result *models.MemberBasic) {
	result = &models.MemberBasic{
		ID:           m.ID,
		Nickname:     m.Nickname,
		ProfileImage: m.ProfileImage,
		Description:  m.Description,
		Role:         m.Role,
	}
	return result
}
