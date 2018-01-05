package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type mockMemberAPI struct{}

func initMemberTest() {
	mockMemberDSBack = mockMemberDS
}

func clearMemberTest() {
	mockMemberDS = mockMemberDSBack
}

func (mapi *mockMemberAPI) GetMembers(maxResult uint8, page uint16, sortMethod string) ([]models.Member, error) {

	var (
		result []models.Member
		err    error
	)
	if len(mockMemberDS) == 0 {
		err = errors.New("Members Not Found")
		return result, err
	}

	sortedMockMemberDS := make([]models.Member, len(mockMemberDS))
	copy(sortedMockMemberDS, mockMemberDS)

	switch sortMethod {
	case "updated_at":
		sort.SliceStable(sortedMockMemberDS, func(i, j int) bool {
			return sortedMockMemberDS[i].UpdatedAt.Before(sortedMockMemberDS[j].UpdatedAt)
		})
	case "-updated_at":
		sort.SliceStable(sortedMockMemberDS, func(i, j int) bool {
			return sortedMockMemberDS[i].UpdatedAt.After(sortedMockMemberDS[j].UpdatedAt)
		})
	}

	if maxResult >= uint8(len(sortedMockMemberDS)) {
		result = sortedMockMemberDS
		err = nil
	} else if maxResult < uint8(len(sortedMockMemberDS)) {
		result = sortedMockMemberDS[(page-1)*uint16(maxResult) : page*uint16(maxResult)]
		err = nil
	}
	return result, err
}

func (mapi *mockMemberAPI) GetMember(id string) (models.Member, error) {
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

func (mapi *mockMemberAPI) InsertMember(m models.Member) error {
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
func (mapi *mockMemberAPI) UpdateMember(m models.Member) error {

	err := errors.New("User Not Found")
	for index, member := range mockMemberDS {
		if member.ID == m.ID {
			mockMemberDS[index] = m
			err = nil
		}
	}
	return err
}

func (mapi *mockMemberAPI) DeleteMember(id string) error {

	// result := models.Member{}
	err := errors.New("User Not Found")
	for index, value := range mockMemberDS {
		if id == value.ID {
			mockMemberDS[index].Active = models.NullInt{0, true}
			return mockMemberDS[index], nil
		}
	}
	return err
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
		{"Descending", "/members", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[1], mockMemberDS[0], mockMemberDS[2]}}},
		{"Ascending", "/members?sort=updated_at", ExpectGetsResp{ExpectResp{http.StatusOK, ""},
			[]models.Member{mockMemberDS[2], mockMemberDS[0], mockMemberDS[1]}}},
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
				t.Errorf("%s expect errro message %v but get %v", tc.name, tc.expect.err, w.Body.String())
			}

			expected, _ := json.Marshal(map[string][]models.Member{"_items": tc.expect.resp})
			if w.Code == http.StatusOK && w.Body.String() != string(expected) {
				t.Errorf("%s incorrect response", tc.name)
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
				t.Errorf("%s expect error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
			}

			// expected, _ := json.Marshal(tc.expect.resp)
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
				t.Errorf("%s expect error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
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
		{"NotExisted", `{"id":"spaceoddity", "name":"Major Tom"}`, ExpectResp{http.StatusBadRequest, `{"Error":"User Not Found"}`}},
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
				t.Errorf("%s expect error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
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
				t.Errorf("%s expect error message %v but get %v", tc.name, w.Body.String(), tc.expect.err)
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
			member, err := models.MemberAPI.GetMember(testcase.in.ID)
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
