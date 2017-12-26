// +build f1
package routes

import (
	"bytes"
	"golang.org/x/crypto/scrypt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockPermissionAPI struct {
	models.PermissionAPIImpl
}

func (a *mockPermissionAPI) InsertPermission(p models.Permission) (models.Permission, error) {
	return p, nil
}

func (a *mockPermissionAPI) GetPermissionsByRole(role int) ([]models.Permission, error) {
	var permissions = []models.Permission{
		models.Permission{
			Role:       models.NullString{"1", true},
			Object:     models.NullString{"ReadPost", true},
			Permission: 1,
		},
	}
	return permissions, nil
}

var MockPermissionAPI mockPermissionAPI

var mockLoginPermissions = []models.Permission{
	models.Permission{
		Role:       models.NullString{"1", true},
		Object:     models.NullString{"ReadPost", true},
		Permission: 1,
	},
	models.Permission{
		Role:       models.NullString{"1", true},
		Object:     models.NullString{"ChangeName", true},
		Permission: 1,
	}}

var mockLoginMembers = []models.Member{
	models.Member{
		ID:           "logintest1@mirrormedia.mg",
		Password:     models.NullString{"hellopassword", true},
		Salt:         models.NullString{"12345567890129375", true},
		Role:         1,
		Active:       1,
		RegisterMode: models.NullString{"ordinary", true},
	},
	models.Member{
		ID:           "logintest2018",
		Password:     models.NullString{"1233211234567", true},
		Salt:         models.NullString{"lIl11llIII1Il1I", true},
		Role:         1,
		Active:       1,
		RegisterMode: models.NullString{"oauth-fb", true},
	},
	models.Member{
		ID:           "logindeactived",
		Password:     models.NullString{"88888888", true},
		Salt:         models.NullString{"1", true},
		Role:         1,
		Active:       0,
		RegisterMode: models.NullString{"ordinary", true},
	}}

func TestRouteLogin(t *testing.T) {

	//Init
	/*dbURI := "root:qwerty@tcp(127.0.0.1)/memberdb?parseTime=true"
	models.Connect(dbURI)
	_, _ = models.DB.Exec("truncate table members;")
	_, _ = models.DB.Exec("truncate table permissions;")*/

	for _, member := range mockLoginMembers {
		hpw, err := scrypt.Key([]byte(member.Password.String), []byte(member.Salt.String), 32768, 8, 1, 64)
		member.Password = models.NullString{string(hpw), true}
		err = models.MemberAPI.InsertMember(member)
		if err != nil {
			t.Fatalf("Init test case fail, aborted. Error: %v", err)
			return
		}
	}
	for _, permission := range mockLoginPermissions {
		_, err := models.PermissionAPI.InsertPermission(permission)
		if err != nil {
			t.Fatalf("Init test case fail, aborted. Error: %v", err)
			return
		}
	}

	type LoginCaseIn struct {
		id   string
		pw   string
		mode string
	}

	type LoginCaseOut struct {
		httpcode int
		resp     userInfoResponse
		Error    string
	}

	var TestRouteLoginCases = []struct {
		name string
		in   LoginCaseIn
		out  LoginCaseOut
	}{
		{"LoginPW", LoginCaseIn{"logintest1@mirrormedia.mg", "hellopassword", "ordinary"}, LoginCaseOut{http.StatusOK, userInfoResponse{models.Member{ID: "logintest1@mirrormedia.mg"}, []string{"ReadPost"}}, ""}},
		{"LoginFB", LoginCaseIn{"logintest2018", "", "oauth-fb"}, LoginCaseOut{http.StatusOK, userInfoResponse{models.Member{ID: "logintest2018"}, []string{"ReadPost"}}, ""}},
		{"LoginNoID", LoginCaseIn{"", "password", "ordinary"}, LoginCaseOut{http.StatusBadRequest, userInfoResponse{}, `{"Error":"Bad Request"}`}},
		{"LoginWorngMode1", LoginCaseIn{"", "password", "wrongmode"}, LoginCaseOut{http.StatusBadRequest, userInfoResponse{}, `{"Error":"Bad Request"}`}},
		{"LoginWrongMode2", LoginCaseIn{"logintest1@mirrormedia.mg", "hellopassword", "oauth-fb"}, LoginCaseOut{http.StatusBadRequest, userInfoResponse{}, `{"Error":"Bad Request"}`}},
		{"LoginNotFound", LoginCaseIn{"Nobody", "password", "ordinary"}, LoginCaseOut{http.StatusNotFound, userInfoResponse{}, `{"Error":"User Not Found"}`}},
		{"LoginNotActive", LoginCaseIn{"logindeactived", "88888888", "ordinary"}, LoginCaseOut{http.StatusUnauthorized, userInfoResponse{}, `{"Error":"User Not Activated"}`}},
		{"LoginWrongPW", LoginCaseIn{"logintest1@mirrormedia.mg", "guesswho", "ordinary"}, LoginCaseOut{http.StatusUnauthorized, userInfoResponse{}, `{"Error":"Login Fail"}`}},
	}

	for _, testcase := range TestRouteLoginCases {

		w := httptest.NewRecorder()

		jsonStrPerp := `{`
		if testcase.in.id != "" {
			jsonStrPerp = jsonStrPerp + `"id":"` + testcase.in.id + `",`
		}
		if testcase.in.pw != "" {
			jsonStrPerp = jsonStrPerp + `"password":"` + testcase.in.pw + `",`
		}
		if testcase.in.mode != "" {
			jsonStrPerp = jsonStrPerp + `"mode":"` + testcase.in.mode + `",`
		}
		jsonStr := []byte(jsonStrPerp[0:len(jsonStrPerp)-1] + `}`)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code == http.StatusOK && (testcase.out.resp.member.ID != testcase.in.id) {
			t.Errorf("Expect get user id %s but get %d, testcase %s", testcase.in.id, testcase.out.resp.member.ID, testcase.name)
			t.Fail()
		}

		if w.Code != http.StatusOK && w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}

	}

}
