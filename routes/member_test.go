package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type mockMemberAPI struct{}

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
	// result := MemberList[len(MemberList)-1]
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

func (mapi *mockMemberAPI) DeleteMember(id string) (models.Member, error) {

	result := models.Member{}
	err := errors.New("User Not Found")
	for index, value := range mockMemberDS {
		if id == value.ID {
			mockMemberDS[index].Active = 0
			return mockMemberDS[index], nil
		}
	}
	return result, err
}

func TestGetMembersDescending(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/members", nil)

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("HTTP response code error: want %q but get %q", http.StatusOK, w.Code)
	}
	res := []models.Member{}
	json.Unmarshal([]byte(w.Body.String()), &res)
	if res[0] != mockMemberDS[1] || res[1] != mockMemberDS[0] || res[2] != mockMemberDS[2] {
		t.Errorf("Response sort error")
	}
}

func TestGetMembersAscending(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/members?sort=updated_at", nil)

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("HTTP response code error: want %q but get %q", http.StatusOK, w.Code)
	}
	res := []models.Member{}
	json.Unmarshal([]byte(w.Body.String()), &res)
	if res[0] != mockMemberDS[2] || res[1] != mockMemberDS[0] || res[2] != mockMemberDS[1] {
		t.Errorf("Response sort error")
	}
}

func TestGetExistMember(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/TaiwanNo.1", nil)

	// r := getRouter()
	// r.GET("/member/:id", env.MemberGetHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	expected, _ := json.Marshal(mockMemberDS[0])
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestGetNotExistMember(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/member/abc123", nil)

	// r := getRouter()
	// r.GET("/member/:id", env.MemberGetHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}

	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostEmptyMember(t *testing.T) {

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/member", nil)
	// req.Header.Set("Content-Type", "application/json")
	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"Invalid User"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPostMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"spaceoddity", 
		"name":"Major Tom"
	}`)
	req, _ := http.NewRequest("POST", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	// var (
	// 	resp     models.Member
	// 	expected models.Member
	// )
	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
	// 	log.Fatal(err)
	// }
	// if err := json.Unmarshal(jsonStr, &expected); err != nil {
	// 	log.Fatal(err)
	// }
	// if resp.ID != expected.ID || resp.Name != expected.Name {
	// 	t.Fail()
	// }
}

func TestPostExistedMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{"id":"TaiwanNo.1"}`)
	req, _ := http.NewRequest("POST", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.POST("/member", env.MemberPostHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"User Already Existed"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestPutMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"id":"TaiwanNo.1", 
		"name":"Major Tom"
	}`)
	req, _ := http.NewRequest("PUT", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.PUT("/member", env.MemberPutHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	// var (
	// 	resp     models.Member
	// 	expected models.Member
	// )
	// if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
	// 	log.Fatal(err)
	// }
	// if err := json.Unmarshal(jsonStr, &expected); err != nil {
	// 	log.Fatal(err)
	// }
	// if resp.ID != expected.ID || resp.Name != expected.Name {
	// 	t.Fail()
	// }
}

func TestPutNonExistMember(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
			"id":"ChinaNo.19", 
			"name":"Major Tom"
		}`)
	req, _ := http.NewRequest("PUT", "/member", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// r := getRouter()
	// r.PUT("/member", env.MemberPutHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fail()
	}
	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestDeleteExistMember(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/member/TaiwanNo.1", nil)

	// r := getRouter()
	// r.DELETE("/member/:id", env.MemberDeleteHandler)
	// Start gin server
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}
	var resp models.Member
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if resp.Active == 1 {
		t.Fail()
	}
}

func TestDeleteNonExistMember(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/member/ChinaNo.19", nil)

	// r := getRouter()
	// r.DELETE("/member/:id", env.MemberDeleteHandler)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}
	expected := `{"Error":"User Not Found"}`
	if w.Body.String() != string(expected) {
		t.Fail()
	}
}

func TestUpdateMemberPassword(t *testing.T) {

	type ChangePWCaseIn struct {
		ID       string `json:id,omitempty`
		Password string `json:"password,omitempty"`
	}

	var TestRouteChangePWCases = []struct {
		name     string
		in       ChangePWCaseIn
		httpcode int
	}{
		{"ChangePWOK", ChangePWCaseIn{ID: "TaiwanNo.1", Password: "angrypug"}, http.StatusOK},
		{"ChangePWFail", ChangePWCaseIn{ID: "TaiwanNo.1"}, http.StatusBadRequest},
		{"ChangePWNoID", ChangePWCaseIn{Password: "angrypug"}, http.StatusBadRequest},
		{"ChangePWMemberNotFound", ChangePWCaseIn{ID: "TaiwanNo.9527", Password: "angrypug"}, http.StatusNotFound},
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
