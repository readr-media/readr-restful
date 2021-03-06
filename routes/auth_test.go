package routes

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/scrypt"

	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

func initAuthTest() {

	var mockLoginMembers = []models.Member{
		models.Member{
			ID:           81,
			MemberID:     "logintest1@mirrormedia.mg",
			Password:     rrsql.NullString{"hellopassword", true},
			Role:         rrsql.NullInt{1, true},
			Active:       rrsql.NullInt{1, true},
			RegisterMode: rrsql.NullString{"ordinary", true},
			UUID:         "abc1d5b1-da54-4200-b57e-f06e59fd8467",
			Points:       rrsql.NullInt{Int: 0, Valid: true},
			Mail:         rrsql.NullString{"logintest1@mirrormedia.mg", true},
		},
		models.Member{
			ID:           82,
			MemberID:     "logintest2018",
			Password:     rrsql.NullString{"1233211234567", true},
			Role:         rrsql.NullInt{1, true},
			Active:       rrsql.NullInt{1, true},
			RegisterMode: rrsql.NullString{"oauth-fb", true},
			UUID:         "abc1d5b1-da54-4200-b67e-f06e59fd8467",
			Points:       rrsql.NullInt{Int: 0, Valid: true},
			Mail:         rrsql.NullString{"logintest2018", true},
		},
		models.Member{
			ID:           83,
			MemberID:     "logindeactived",
			Password:     rrsql.NullString{"88888888", true},
			Role:         rrsql.NullInt{1, true},
			Active:       rrsql.NullInt{0, true},
			RegisterMode: rrsql.NullString{"ordinary", true},
			UUID:         "abc1d5b1-da54-4200-b77e-f06e59fd8467",
			Points:       rrsql.NullInt{Int: 0, Valid: true},
			Mail:         rrsql.NullString{"logindeactived", true},
		}}

	for _, member := range mockLoginMembers {

		salt := make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			fmt.Errorf(err.Error())
			return
		}
		member.Salt = rrsql.NullString{string(salt), true}
		hpw, err := scrypt.Key([]byte(member.Password.String), []byte(member.Salt.String), 32768, 8, 1, 64)
		member.Password = rrsql.NullString{string(hpw), true}
		_, err = models.MemberAPI.InsertMember(member)
		if err != nil {
			log.Println(err)
			fmt.Errorf("Init auth test case fail, aborted. Error: %v", err)
			return
		}
	}
}

func TestRouteLogin(t *testing.T) {

	initAuthTest()

	type userInfoResponse struct {
		Member      models.Member `json:"member"`
		Permissions []string      `json:"permissions"`
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
		{"LoginPW", LoginCaseIn{"logintest1@mirrormedia.mg", "hellopassword", "ordinary"}, LoginCaseOut{http.StatusOK, userInfoResponse{models.Member{MemberID: "logintest1@mirrormedia.mg"}, []string{"ReadPost"}}, ""}},
		{"LoginFB", LoginCaseIn{"logintest2018", "", "oauth-fb"}, LoginCaseOut{http.StatusOK, userInfoResponse{models.Member{MemberID: "logintest2018"}, []string{"ReadPost"}}, ""}},
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
			jsonStrPerp = jsonStrPerp + `"register_mode":"` + testcase.in.mode + `",`
		}
		jsonStr := []byte(jsonStrPerp[0:len(jsonStrPerp)-1] + `}`)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code != http.StatusOK && w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}

		resp := userInfoResponse{}
		json.Unmarshal([]byte(w.Body.String()), &resp)

		if w.Code == http.StatusOK && (testcase.out.resp.Member.MemberID != resp.Member.MemberID) {
			t.Errorf("Expect get user id %s but get %s, testcase %s", testcase.out.resp.Member.MemberID, resp.Member.MemberID, testcase.name)
			t.Fail()
		}
	}
}

func TestRouteRegister(t *testing.T) {

	initAuthTest()

	type RegisterCaseIn struct {
		MemberID string `json:"member_id,omitempty"`
		Password string `json:"password,omitempty"`
		Mail     string `json:"mail,omitempty"`
		SocialID string `json:"social_id,omitempty"`
		Mode     string `json:"register_mode,omitempty"`
		Nickname string `json:"nickname,omitempty"`
		Gender   string `json:"gender,omitempty"`
	}

	type RegisterCaseOut struct {
		httpcode int
		Error    string
	}

	var TestRouteRegisterCases = []struct {
		name string
		in   RegisterCaseIn
		out  RegisterCaseOut
	}{
		{"RegisterOK", RegisterCaseIn{
			Password: "mir",
			Mail:     "registertest1@mirrormedia.mg",
			Mode:     "ordinary"}, RegisterCaseOut{http.StatusOK, `ok`}},
		{"RegisterNoPassword", RegisterCaseIn{
			Password: "",
			Mail:     "logintest1@mirrormedia.mg",
			Mode:     "ordinary"}, RegisterCaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`}},
		{"RegisterNoMail", RegisterCaseIn{
			Password: "mir",
			Mail:     "",
			Mode:     "ordinary"}, RegisterCaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`}},
		{"RegisterNoMode", RegisterCaseIn{
			Password: "mirr",
			Mail:     "logintest1@mirrormedia.mg",
			Mode:     ""}, RegisterCaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`}},
		{"RegisterSocialOK", RegisterCaseIn{
			Password: "mir",
			Mail:     "logintest1@mirrormedia.mg",
			Mode:     "oauth-fb",
			SocialID: "112233445566"}, RegisterCaseOut{http.StatusOK, `ok`}},
		{"RegisterNoSocialID", RegisterCaseIn{
			Password: "mir",
			Mail:     "logintest1@mirrormedia.mg",
			Mode:     "oauth-fb",
			SocialID: ""}, RegisterCaseOut{http.StatusBadRequest, `{"Error":"Bad Request"}`}},
		{"RegisterUserDupe", RegisterCaseIn{
			Password: "1233211234567",
			Mail:     "registerdupeuser@mirrormedia.mg",
			Mode:     "ordinary"}, RegisterCaseOut{http.StatusBadRequest, `{"Error":"User Duplicated","Mode":"ordinary"}`}},
	}

	for _, testcase := range TestRouteRegisterCases {

		w := httptest.NewRecorder()

		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code != http.StatusOK && w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}

		resp := RegisterCaseOut{}
		json.Unmarshal([]byte(w.Body.String()), &resp)

		if w.Code == http.StatusOK {
			member, _ := models.MemberAPI.GetMember(models.GetMemberArgs{
				ID:     testcase.in.MemberID,
				IDType: "member_id",
			})
			if testcase.in.MemberID != member.MemberID {
				t.Errorf("Expect get user id %s but get %s, testcase %s", testcase.in.MemberID, member.MemberID, testcase.name)
				t.Fail()
			}
		}

	}
}
