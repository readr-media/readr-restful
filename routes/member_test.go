package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type mockMemberAPI struct{}

func initMemberTest() {
	copy(mockMemberDSBack, mockMemberDS)
}

func clearMemberTest() {
	copy(mockMemberDS, mockMemberDSBack)
}

func (a *mockMemberAPI) GetMembers(req *models.MemberArgs) (result []models.Member, err error) {

	if req.CustomEditor == true {
		result = []models.Member{mockMemberDS[0]}
		err = nil
		return result, err
	}

	if req.Role != nil {
		result = []models.Member{mockMemberDS[1]}
		err = nil
		return result, err
	}
	if len(req.Active) > 1 {
		return []models.Member{}, errors.New("Too many active lists")
	}
	for k, v := range req.Active {
		if k == "$nin" && reflect.DeepEqual(v, []int{0, -1}) {
			return []models.Member{mockMemberDS[0]}, nil
		} else if k == "$nin" && reflect.DeepEqual(v, []int{-1, 0, 1}) {
			return []models.Member{}, nil
		} else if reflect.DeepEqual(v, []int{-3, 0, 1}) {
			return []models.Member{}, errors.New("Not all active elements are valid")
		} else if reflect.DeepEqual(v, []int{3, 4}) {
			return []models.Member{}, errors.New("No valid active request")
		}
	}

	result = make([]models.Member, len(mockMemberDS))
	copy(result, mockMemberDS)
	switch req.Sorting {
	case "updated_at":
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].UpdatedAt.Before(result[j].UpdatedAt)
		})
		err = nil
	case "-updated_at":
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].UpdatedAt.After(result[j].UpdatedAt)
		})
		err = nil
	}
	if req.MaxResult == 2 {
		result = result[0:2]
	}
	return result, err
}

func (a *mockMemberAPI) GetMember(idType string, id string) (models.Member, error) {
	result := models.Member{}
	err := errors.New("User Not Found")
	for _, value := range mockMemberDS {
		if value.ID == id {
			result = value
			err = nil
		}
	}
	return result, err
}

func (a *mockMemberAPI) InsertMember(m models.Member) error {
	var err error
	for _, member := range mockMemberDS {
		if member.ID == m.ID {
			return errors.New("Duplicate entry")
		}
	}
	mockMemberDS = append(mockMemberDS, m)
	err = nil
	return err
}
func (a *mockMemberAPI) UpdateMember(m models.Member) error {

	err := errors.New("User Not Found")
	for index, member := range mockMemberDS {
		if member.ID == m.ID {
			mockMemberDS[index] = m
			err = nil
		}
	}
	return err
}

func (a *mockMemberAPI) DeleteMember(idType string, id string) error {

	err := errors.New("User Not Found")
	for index, value := range mockMemberDS {
		if id == value.ID {
			mockMemberDS[index].Active = models.NullInt{Int: int64(models.MemberStatus["delete"].(float64)), Valid: true}
			return nil
		}
	}
	return err
}

func (a *mockMemberAPI) UpdateAll(ids []string, active int) (err error) {

	result := make([]int, 0)
	for _, value := range ids {
		for i, v := range mockMemberDS {
			if v.ID == value {
				mockMemberDS[i].Active = models.NullInt{Int: int64(active), Valid: true}
				result = append(result, i)
			}
		}
	}
	if len(result) == 0 {
		err = errors.New("Members Not Found")
		return err
	}
	return err
}

func (a *mockMemberAPI) Count(req *models.MemberArgs) (result int, err error) {
	result = 0
	err = errors.New("Members Not Found")
	if req.CustomEditor == true {
		return 1, nil
	}
	if req.Role != nil && *req.Role == int64(9) {
		return 1, nil
	}
	for k, v := range req.Active {
		if k == "$in" && reflect.DeepEqual(v, []int{1, -1}) {
			return 2, nil
		}
		if k == "$nin" && reflect.DeepEqual(v, []int{-1}) {
			return 2, nil
		}
	}
	return result, err
}

func (a *mockMemberAPI) GetUUIDsByNickname(key string, roles map[string][]int) (result []models.NicknameUUID, err error) {
	for _, v := range mockMemberDS {
		if v.Nickname.Valid {
			if matched, err := regexp.MatchString(key, v.Nickname.String); err == nil && matched {
				result = append(result, models.NicknameUUID{UUID: v.UUID, Nickname: v.Nickname})
			}
		}
	}
	return result, err
}

func TestRouteGetMembers(t *testing.T) {

	initMemberTest()
	type ExpectGetsResp struct {
		ExpectResp
		resp []models.Member
	}
	testCase := []struct {
		name   string
		route  string
		expect ExpectGetsResp
	}{
		{"UpdatedAtDescending", "/members", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[1], mockMemberDS[0], mockMemberDS[2]}}},
		{"UpdatedAtAscending", "/members?sort=updated_at", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[2], mockMemberDS[0], mockMemberDS[1]}}},
		{"max_result", "/members?max_result=2", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[1], mockMemberDS[0]}}},
		{"ActiveFilter", `/members?active={"$nin":[0,-1]}`, ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[0]}}},
		{"CustomEditorFilter", `/members?custom_editor=true`, ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[0]}}},
		{"NoMatchMembers", `/members?active={"$nin":[-1,0,1]}`,
			ExpectGetsResp{ExpectResp{
				http.StatusOK, ``},
				[]models.Member{}}},
		{"MoreThanOneActive", `/members?active={"$nin":[1,0], "$in":[-1,3]}`,
			ExpectGetsResp{
				ExpectResp{http.StatusBadRequest, `{"Error":"Too many active lists"}`},
				[]models.Member{}}},
		{"NotEntirelyValidActive", `/members?active={"$in":[-3,0,1]}`,
			ExpectGetsResp{
				ExpectResp{http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
				[]models.Member{}}},
		{"NoValidActive", `/members?active={"$nin":[3,4]}`,
			ExpectGetsResp{
				ExpectResp{http.StatusBadRequest, `{"Error":"No valid active request"}`},
				[]models.Member{}}},
		{"Role", `/members?role=1`, ExpectGetsResp{ExpectResp{http.StatusOK, ``}, []models.Member{mockMemberDS[1]}}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}

			expected, _ := json.Marshal(map[string][]models.Member{"_items": tc.expect.resp})

			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Errorf("%s incorrect response.\nWant\n%s\nBut get\n%s\n", tc.name, string(expected), w.Body.String())
			}
		})
	}
	clearMemberTest()
}

func TestRouteGetMember(t *testing.T) {

	initMemberTest()

	type ExpectGetResp struct {
		ExpectResp
		resp models.Member
	}
	testCase := []struct {
		name   string
		route  string
		expect ExpectGetResp
	}{
		{"Current", "/member/superman@mirrormedia.mg", ExpectGetResp{ExpectResp{http.StatusOK, ""}, mockMemberDS[0]}},
		{"NotExisted", "/member/wonderwoman@mirrormedia.mg", ExpectGetResp{ExpectResp{http.StatusNotFound, `{"Error":"User Not Found"}`}, models.Member{}}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)

			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}

			expected, _ := json.Marshal(map[string][]models.Member{"_items": []models.Member{tc.expect.resp}})
			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Errorf("%s incorrect response", tc.name)
			}
		})
	}

	clearMemberTest()
}
func TestRoutePostMember(t *testing.T) {
	initMemberTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"New", `{"id":"spaceoddity", "name":"Major Tom"}`, ExpectResp{http.StatusOK, ""}},
		{"EmptyPayload", `{}`, ExpectResp{http.StatusBadRequest, `{"Error":"Invalid User"}`}},
		{"Existed", `{"id":"superman@mirrormedia.mg"}`, ExpectResp{http.StatusBadRequest, `{"Error":"User Already Existed"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var jsonStr = []byte(tc.payload)
			req, _ := http.NewRequest("POST", "/member", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearMemberTest()
}
func TestRoutePutMember(t *testing.T) {
	initMemberTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"Current", `{"id":"superman@mirrormedia.mg", "name":"Clark Kent"}`, ExpectResp{http.StatusOK, ""}},
		{"NotExisted", `{"id":"MajorTom@mirrormedia.mg", "name":"spaceoddity"}`, ExpectResp{http.StatusBadRequest, `{"Error":"User Not Found"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var jsonStr = []byte(tc.payload)
			req, _ := http.NewRequest("PUT", "/member", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}

	clearMemberTest()
}

func TestRouteDeleteMembers(t *testing.T) {
	initMemberTest()
	testCase := []struct {
		name   string
		route  string
		expect ExpectResp
	}{
		{"Delete", `/members?ids=["superman@mirrormedia.mg","test6743"]`, ExpectResp{http.StatusOK, ""}},
		{"Empty", `/members?ids=[]`, ExpectResp{http.StatusBadRequest, `{"Error":"ID List Empty"}`}},
		{"InvalidQueryArray", `/members?ids=["superman@mirrormedia.mg,"test6743"]`, ExpectResp{http.StatusBadRequest, `{"Error":"invalid character 't' after array element"}`}},
		{"NotFound", `/members?ids=["superman", "wonderwoman"]`, ExpectResp{http.StatusBadRequest, `{"Error":"Members Not Found"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearMemberTest()
}
func TestRouteDeleteMember(t *testing.T) {
	initMemberTest()
	testCase := []struct {
		name   string
		route  string
		expect ExpectResp
	}{
		{"Current", "/member/superman@mirrormedia.mg", ExpectResp{http.StatusOK, ""}},
		{"NonExisted", "/member/wonderwoman@mirrormedia.mg", ExpectResp{http.StatusNotFound, `{"Error":"User Not Found"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearMemberTest()
}

func TestUpdateMemberPassword(t *testing.T) {

	type ChangePWCaseIn struct {
		ID       string `json:"id,omitempty"`
		Password string `json:"password,omitempty"`
	}

	var TestRouteChangePWCases = []struct {
		name     string
		in       ChangePWCaseIn
		httpcode int
	}{
		{"ChangePWOK", ChangePWCaseIn{ID: "superman@mirrormedia.mg", Password: "angrypug"}, http.StatusOK},
		{"ChangePWFail", ChangePWCaseIn{ID: "superman@mirrormedia.mg"}, http.StatusBadRequest},
		{"ChangePWNoID", ChangePWCaseIn{Password: "angrypug"}, http.StatusBadRequest},
		{"ChangePWMemberNotFound", ChangePWCaseIn{ID: "aquaman@mirrormedia.mg", Password: "angrypug"}, http.StatusNotFound},
	}

	for _, testcase := range TestRouteChangePWCases {
		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest("PUT", "/member/password", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code == http.StatusOK {
			member, err := models.MemberAPI.GetMember("member_id", testcase.in.ID)
			if err != nil {
				t.Errorf("Cannot get user after update PW, testcase %s", testcase.name)
				t.Fail()
			}

			newPW, err := utils.CryptGenHash(testcase.in.Password, member.Salt.String)
			switch {
			case err != nil:
				t.Errorf("Error when hashing password, testcase %s", testcase.name)
				t.Fail()
			case newPW != member.Password.String:
				t.Errorf("%v", member.ID)
				t.Errorf("Password update fail Want %v but get %v, testcase %s", newPW, member.Password.String, testcase.name)
				t.Fail()
			}
		}
	}
}

func TestRouteActivateMultipleMembers(t *testing.T) {
	initMemberTest()
	testCase := []struct {
		name    string
		payload string
		expect  ExpectResp
	}{
		{"CurrentMembers", `{"ids": ["superman@mirrormedia.mg","test6743"]}`, ExpectResp{http.StatusOK, ``}},
		{"NotFound", `{"ids": ["ironman", "spiderman"]}`, ExpectResp{http.StatusNotFound, `{"Error":"Members Not Found"}`}},
		{"InvalidPayload", `{}`, ExpectResp{http.StatusBadRequest, `{"Error":"Invalid Request Body"}`}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			jsonStr := []byte(tc.payload)
			req, _ := http.NewRequest("PUT", "/members", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %s but get %s", tc.name, tc.expect.err, w.Body.String())
			}
		})
	}
	clearMemberTest()
}

func TestRouteCountMembers(t *testing.T) {
	initMemberTest()
	type ExpectCountResp struct {
		httpcode int
		resp     string
		err      string
	}
	testCase := []struct {
		name   string
		route  string
		expect ExpectCountResp
	}{
		{"SimpleCount", `/members/count`, ExpectCountResp{http.StatusOK, `{"_meta":{"total":2}}`, ``}},
		{"CountActive", `/members/count?active={"$in":[1,-1]}`, ExpectCountResp{http.StatusOK, `{"_meta":{"total":2}}`, ``}},
		{"CountCustomEditor", `/members/count?custom_editor=true`, ExpectCountResp{http.StatusOK, `{"_meta":{"total":1}}`, ``}},
		{"MoreThanOneActive", `/members/count?active={"$nin":[1,0], "$in":[-1,3]}`,
			ExpectCountResp{http.StatusBadRequest, ``,
				`{"Error":"Too many active lists"}`}},
		{"NotEntirelyValidActive", `/members/count?active={"$in":[-3,0,1]}`,
			ExpectCountResp{http.StatusBadRequest, ``,
				`{"Error":"Not all active elements are valid"}`}},
		{"NoValidActive", `/members/count?active={"$nin":[3,4]}`,
			ExpectCountResp{http.StatusBadRequest, ``,
				`{"Error":"No valid active request"}`}},
		{"Role", "/members/count?role=9", ExpectCountResp{http.StatusOK, `{"_meta":{"total":1}}`, ``}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}
			if w.Code == http.StatusOK && w.Body.String() != tc.expect.resp {
				t.Errorf("%s incorrect response.\nWant\n%s\nBut get\n%s\n", tc.name, tc.expect.resp, w.Body.String())
			}
		})
	}
	clearMemberTest()
}

func TestRouteKeyNickname(t *testing.T) {
	initMemberTest()
	type ExpectKeyResp struct {
		ExpectResp
		resp string
	}
	testCase := []struct {
		name   string
		route  string
		expect ExpectKeyResp
	}{
		{"Keyword", `/members/nickname?keyword=read`, ExpectKeyResp{ExpectResp{http.StatusOK, ``}, `{"_items":[{"uuid":"3d6512e8-3e30-11e8-b94b-cfe922eb374f","nickname":"reader"}]}`}},
		{"KeywordAndRoles", `/members/nickname?keyword=read&roles={"$in":[3,9]}`, ExpectKeyResp{ExpectResp{http.StatusOK, ``}, `{"_items":[{"uuid":"3d6512e8-3e30-11e8-b94b-cfe922eb374f","nickname":"reader"}]}`}},
		{"InvalidKeyword", `/members/nickname`, ExpectKeyResp{ExpectResp{http.StatusBadRequest, `{"Error":"Invalid keyword"}`}, ``}},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.route, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expect.httpcode {
				t.Errorf("%s expect status %d but get %d", tc.name, tc.expect.httpcode, w.Code)
			}
			if w.Code != http.StatusOK && w.Body.String() != tc.expect.err {
				t.Errorf("%s expect error message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}

			if w.Code == http.StatusOK && w.Body.String() != tc.expect.resp {
				t.Errorf("%s incorrect response.\nWant\n%s\nBut get\n%s\n", tc.name, tc.expect.resp, w.Body.String())
			}
		})
	}
	clearMemberTest()
}
