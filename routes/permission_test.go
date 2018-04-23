package routes

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/readr-media/readr-restful/models"
)

type mockPermissionAPI struct {
	models.PermissionAPIImpl
}

func (a *mockPermissionAPI) GetPermissions(ps []models.Permission) ([]models.Permission, error) {
	var permissions []models.Permission

OuterLoop:
	for _, p := range ps {
		for _, permission := range mockPermissionDS {
			if permission.Role == p.Role && permission.Object == p.Object {
				permissions = append(permissions, permission)
				continue OuterLoop
			}
		}
	}
	return permissions, nil
}

func (a *mockPermissionAPI) GetPermissionsByRole(role int) ([]models.Permission, error) {
	var permissions []models.Permission
	for _, permission := range mockPermissionDS {
		if permission.Role == role {
			permissions = append(permissions, permission)
		}
	}
	return permissions, nil
}

func (a *mockPermissionAPI) GetPermissionsAll() ([]models.Permission, error) {
	return mockPermissionDS, nil
}

func (a *mockPermissionAPI) InsertPermissions(ps []models.Permission) error {

	for _, permission := range mockPermissionDS {
		for _, p := range ps {
			if permission.Role == p.Role && permission.Object == p.Object {
				return errors.New("Duplicate Entry")
			}
		}
	}

	for _, p := range ps {
		mockPermissionDS = append(mockPermissionDS, p)
	}

	return nil
}

func (a *mockPermissionAPI) DeletePermissions(ps []models.Permission) error {
	for _, p := range ps {
		for index, permission := range mockPermissionDS {
			if permission.Role == p.Role && permission.Object == p.Object {
				mockPermissionDS = append(mockPermissionDS[:index], mockPermissionDS[index+1:]...)
			}
		}
	}
	return nil
}

var MockPermissionAPI mockPermissionAPI

func initPermissionTest() {

	var mockDefaultPermissions = []models.Permission{
		models.Permission{ID: 1, Role: 1, Object: models.NullString{"ChangePW", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 2, Role: 1, Object: models.NullString{"ChangeName", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 3, Role: 1, Object: models.NullString{"CreateAccount", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 4, Role: 1, Object: models.NullString{"AddRole", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 5, Role: 1, Object: models.NullString{"EditRole", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 6, Role: 1, Object: models.NullString{"DeleteRole", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 7, Role: 1, Object: models.NullString{"CreatePost", true}, Permission: models.NullInt{1, true}},
		models.Permission{ID: 8, Role: 1, Object: models.NullString{"ReadPost", true}, Permission: models.NullInt{1, true}},
	}

	err := models.PermissionAPI.InsertPermissions(mockDefaultPermissions)
	if err != nil {
		fmt.Errorf("Init test case fail, aborted. Error: %v", err)
		return
	}

}

func TestPermissionGet(t *testing.T) {

	initPermissionTest()

	type item struct {
		Role   int    `json:"role,omitempty"`
		Object string `json:"object,omitempty"`
	}

	type CaseIn struct {
		Query []item `json:query`
	}

	type CaseOut struct {
		httpcode int
		result   string
		Error    string
	}

	var TestRouteName = "/permission"
	var TestRouteMethod = "GET"

	var TestRoutePermissionGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{
			"GetSinglePermissionOK",
			CaseIn{[]item{item{1, "ChangePW"}}},
			CaseOut{http.StatusOK, `[{"id":1,"role":1,"object":"ChangePW","permission":1}]`, ""},
		},
		{
			"GetMultiPermissionOK",
			CaseIn{[]item{item{1, "ChangePW"}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusOK, `[{"id":1,"role":1,"object":"ChangePW","permission":1},{"id":3,"role":1,"object":"CreateAccount","permission":1}]`, ""},
		},
		{
			"GetPermissionNotGranted",
			CaseIn{[]item{item{1, "ChangePW"}, item{1, "Hackit"}}},
			CaseOut{http.StatusOK, `[{"id":1,"role":1,"object":"ChangePW","permission":1},{"id":0,"role":1,"object":"Hackit","permission":0}]`, ""},
		},
		{
			"GetPermissionMissingRole",
			CaseIn{[]item{item{Object: "ChangePW"}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusBadRequest, "", `{"Error":"Bad Request"}`},
		},
		{
			"GetPermissionMissingObject",
			CaseIn{[]item{item{Role: 5}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusBadRequest, "", `{"Error":"Bad Request"}`},
		},
	}

	for _, testcase := range TestRoutePermissionGetCases {
		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code != http.StatusOK && w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}

		if w.Code == http.StatusOK && w.Body.String() != testcase.out.result {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.result, w.Body.String(), testcase.name)
			t.Fail()
		}
	}
}

func TestPermissionGetAll(t *testing.T) {

	type CaseOut struct {
		httpcode int
		result   string
		Error    string
	}

	var TestRouteName = "/permission/all"
	var TestRouteMethod = "GET"

	var TestRoutePermissionGetCases = []struct {
		name string
		out  CaseOut
	}{
		{"GetSinglePermissionOK", CaseOut{http.StatusOK, `[{"role":1,"object":"ChangePW","permission":1},{"role":1,"object":"ChangeName","permission":1},{"role":1,"object":"CreateAccount","permission":1},{"role":1,"object":"AddRole","permission":1},{"role":1,"object":"EditRole","permission":1},{"role":1,"object":"DeleteRole","permission":1},{"role":1,"object":"CreatePost","permission":1},{"role":1,"object":"ReadPost","permission":1}]`, ""}},
	}

	for _, testcase := range TestRoutePermissionGetCases {
		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		var mockDefaultPermissions = []models.Permission{
			models.Permission{Role: 1, Object: models.NullString{"ChangePW", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"ChangeName", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"CreateAccount", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"AddRole", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"EditRole", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"DeleteRole", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"CreatePost", true}, Permission: models.NullInt{1, true}},
			models.Permission{Role: 1, Object: models.NullString{"ReadPost", true}, Permission: models.NullInt{1, true}},
		}

		var result []models.Permission
		err := json.Unmarshal([]byte(w.Body.String()), &result)
		if err != nil {
			t.Errorf("Unmarshaling server return text error, testcase %s", testcase.name)
			t.Fail()
		}

	OuterLoop:
		for _, permission := range mockDefaultPermissions {
			for _, p := range result {
				if permission.Role == p.Role && permission.Object == p.Object {
					continue OuterLoop
				}
			}
			t.Errorf("Expect get item message %v but not, testcase %s", permission, testcase.name)
			t.Fail()
		}
	}
}

func TestPermissionPost(t *testing.T) {

	type item struct {
		Role   int    `json:"role,omitempty"`
		Object string `json:"object,omitempty"`
	}

	type CaseIn struct {
		Query []item `json:query`
	}

	type CaseOut struct {
		httpcode int
		Error    string
	}

	var TestRouteName = "/permission"
	var TestRouteMethod = "POST"

	var TestRoutePermissionGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{
			"PostSinglePermissionOK",
			CaseIn{[]item{item{2, "ChangePW"}}},
			CaseOut{http.StatusOK, ""},
		},
		{
			"PostMultiPermissionOK",
			CaseIn{[]item{item{2, "ChangeName"}, item{2, "EditRole"}}},
			CaseOut{http.StatusOK, ""},
		},
		{
			"PostPermissionDuplicated",
			CaseIn{[]item{item{1, "ChangePW"}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusBadRequest, `{"Error":"Duplicate Entry"}`},
		},
		{
			"GetPermissionMissingRole",
			CaseIn{[]item{item{Object: "ChangePW"}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`},
		},
		{
			"GetPermissionMissingObject",
			CaseIn{[]item{item{Role: 1}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`},
		},
	}

	for _, testcase := range TestRoutePermissionGetCases {
		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code != http.StatusOK && w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}

		if w.Code == http.StatusOK {
			for _, item := range testcase.in.Query {
				itemquery := []models.Permission{models.Permission{Role: item.Role, Object: models.NullString{item.Object, true}}}
				p, err := models.PermissionAPI.GetPermissions(itemquery)
				if err != nil || p[0].Role != item.Role || p[0].Object.String != item.Object {
					t.Errorf("Expect get inserted item %v but didn't, testcase %s", item, testcase.name)
					t.Fail()
				}
			}
		}
	}
}

func TestPermissionDelete(t *testing.T) {

	type item struct {
		Role   int    `json:"role,omitempty"`
		Object string `json:"object,omitempty"`
	}

	type CaseIn struct {
		Query []item `json:query`
	}

	type CaseOut struct {
		httpcode int
		Error    string
	}

	var TestRouteName = "/permission"
	var TestRouteMethod = "DELETE"

	var TestRoutePermissionGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{
			"DeleteSinglePermissionOK",
			CaseIn{[]item{item{1, "ChangePW"}}},
			CaseOut{http.StatusOK, ""},
		},
		{
			"DeleteMultiPermissionOK",
			CaseIn{[]item{item{1, "ChangeName"}, item{1, "CreateAccount"}}},
			CaseOut{http.StatusOK, ""},
		},
		{
			"DeletePermissionNotExist",
			CaseIn{[]item{item{999, "Hackit"}}},
			CaseOut{http.StatusOK, ""},
		},
		{
			"DeletePermissionMissingRole",
			CaseIn{[]item{item{Object: "AddRole"}, item{1, "DeleteRole"}}},
			CaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`},
		},
		{
			"DeletePermissionMissingRole",
			CaseIn{[]item{item{Role: 1}, item{1, "CreatePost"}}},
			CaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`},
		},
	}

	for _, testcase := range TestRoutePermissionGetCases {
		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code != http.StatusOK && w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}
		if w.Code == http.StatusOK {
			for _, item := range testcase.in.Query {
				itemquery := []models.Permission{models.Permission{Role: item.Role, Object: models.NullString{item.Object, true}}}
				p, err := models.PermissionAPI.GetPermissions(itemquery)

				if err != nil || len(p) != 0 {
					t.Errorf("Expect no permission for the deleted item %v but didn't, testcase %s", item, testcase.name)
					t.Fail()
				}
			}
		}
	}
}
