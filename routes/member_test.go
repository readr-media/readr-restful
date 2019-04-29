package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type mockMemberAPI struct{}

// Declare a backup struct for member test data
var mockMembers = []models.Member{
	models.Member{
		ID:           1,
		MemberID:     "superman@mirrormedia.mg",
		UUID:         "3d64e480-3e30-11e8-b94b-cfe922eb374f",
		Nickname:     models.NullString{String: "readr", Valid: true},
		Active:       models.NullInt{Int: 1, Valid: true},
		UpdatedAt:    models.NullTime{Time: time.Date(2017, 6, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		Mail:         models.NullString{String: "superman@mirrormedia.mg", Valid: true},
		CustomEditor: models.NullBool{Bool: true, Valid: true},
		Role:         models.NullInt{Int: 9, Valid: true},
		Points:       models.NullInt{Int: 0, Valid: true},
	},
	models.Member{
		ID:        2,
		MemberID:  "test6743@test.test",
		UUID:      "3d651126-3e30-11e8-b94b-cfe922eb374f",
		Nickname:  models.NullString{String: "yeahyeahyeah", Valid: true},
		Active:    models.NullInt{Int: 0, Valid: true},
		Birthday:  models.NullTime{Time: time.Date(2001, 1, 3, 0, 0, 0, 0, time.UTC), Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 11, 11, 23, 11, 37, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "Lulu_Brakus@yahoo.com", Valid: true},
		Role:      models.NullInt{Int: 3, Valid: true},
		Points:    models.NullInt{Int: 0, Valid: true},
	},
	models.Member{
		ID:        3,
		MemberID:  "Barney.Corwin@hotmail.com",
		UUID:      "3d6512e8-3e30-11e8-b94b-cfe922eb374f",
		Nickname:  models.NullString{String: "reader", Valid: true},
		Active:    models.NullInt{Int: 0, Valid: true},
		Gender:    models.NullString{String: "M", Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 1, 3, 19, 32, 37, 0, time.UTC), Valid: true},
		Birthday:  models.NullTime{Time: time.Date(1939, 11, 9, 0, 0, 0, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
		Role:      models.NullInt{Int: 1, Valid: true},
		Points:    models.NullInt{Int: 0, Valid: true},
	},
}

var mockMemberDS = []models.Member{}

func (a *mockMemberAPI) GetMembers(req *models.MemberArgs) (result []models.Member, err error) {

	if req.CustomEditor == true {
		result = []models.Member{mockMemberDS[0]}
		err = nil
		return result, err
	}

	if req.Role != nil {
		result = []models.Member{mockMemberDS[2]}
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

func (a *mockMemberAPI) GetMember(idType string, id string) (result models.Member, err error) {
	intID, _ := strconv.Atoi(id)
	for _, value := range mockMemberDS {
		if idType == "id" && value.ID == int64(intID) {
			return value, nil
		} else if idType == "member_id" && value.MemberID == id {
			return value, nil
		} else if idType == "mail" && id == "registerdupeuser@mirrormedia.mg" {
			return models.Member{RegisterMode: models.NullString{"ordinary", true}}, nil
		}
	}
	err = errors.New("User Not Found")
	return result, err
}

func (a *mockMemberAPI) InsertMember(m models.Member) (id int, err error) {
	for _, member := range mockMemberDS {
		if member.MemberID == m.MemberID {
			return 0, errors.New("Duplicate entry")
		}
	}
	m.ID = int64(len(mockMemberDS) + 1)
	mockMemberDS = append(mockMemberDS, m)
	err = nil
	return int(m.ID), err
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
	intID, _ := strconv.Atoi(id)
	for index, value := range mockMemberDS {
		if int64(intID) == value.ID {
			// mockMemberDS[index].Active = models.NullInt{Int: int64(models.MemberStatus["delete"].(float64)), Valid: true}
			mockMemberDS[index].Active = models.NullInt{Int: int64(config.Config.Models.Members["delete"]), Valid: true}
			return nil
		}
	}
	return err
}

func (a *mockMemberAPI) UpdateAll(ids []int64, active int) (err error) {

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
			return 4, nil
		}
	}
	return result, err
}

func (a *mockMemberAPI) GetIDsByNickname(params models.GetMembersKeywordsArgs) (result []models.Stunt, err error) {
	if params.Keywords == "readr" {
		if params.Roles != nil {
			result = append(result, models.Stunt{ID: &(mockMemberDS[0].ID), Nickname: &(mockMemberDS[0].Nickname)})
			return result, err
		}
		result = append(result, models.Stunt{ID: &(mockMemberDS[0].ID), Nickname: &(mockMemberDS[0].Nickname)})
		return result, err

	}
	return result, err
}

func TestRouteMembers(t *testing.T) {
	if os.Getenv("db_driver") == "mysql" {
		_, _ = models.DB.Exec("truncate table members;")
	} else {
		mockMemberDS = []models.Member{}
	}

	for _, m := range mockMembers {
		_, err := models.MemberAPI.InsertMember(m)
		if err != nil {
			log.Printf("Init member test fail %s", err.Error())
		}
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.Member `json:"_items"`
		}

		var Response response
		var expected []models.Member = tc.resp.([]models.Member)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", resp, err.Error())
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect member length to be %v but get %v", tc.name, len(expected), len(Response.Items))
		}

		for i, resp := range Response.Items {
			exp := expected[i]
			if resp.ID == exp.ID &&
				resp.Active == exp.Active &&
				resp.UpdatedAt == exp.UpdatedAt &&
				resp.Role == exp.Role {
				continue
			}
			t.Errorf("%s, expect to get %v, but %v ", tc.name, exp, resp)
		}
	}

	t.Run("GetMembers", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"UpdatedAtDescending", "GET", "/members", ``, http.StatusOK, []models.Member{mockMembers[1], mockMembers[0], mockMembers[2]}},
			genericTestcase{"UpdatedAtAscending", "GET", "/members?sort=updated_at", ``, http.StatusOK, []models.Member{mockMembers[2], mockMembers[0], mockMembers[1]}},
			genericTestcase{"max_result", "GET", "/members?max_result=2", ``, http.StatusOK, []models.Member{mockMembers[1], mockMembers[0]}},
			genericTestcase{"ActiveFilter", "GET", `/members?active={"$nin":[0,-1]}`, ``, http.StatusOK, []models.Member{mockMembers[0]}},
			genericTestcase{"CustomEditorFilter", "GET", `/members?custom_editor=true`, ``, http.StatusOK, []models.Member{mockMembers[0]}},
			genericTestcase{"NoMatchMembers", "GET", `/members?active={"$nin":[-1,0,1]}`, ``, http.StatusOK, `{"_items":[]}`},
			genericTestcase{"MoreThanOneActive", "GET", `/members?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`},
			genericTestcase{"NotEntirelyValidActive", "GET", `/members?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
			genericTestcase{"NoValidActive", "GET", `/members?active={"$nin":[3,4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`},
			genericTestcase{"Role", "GET", `/members?role=1`, ``, http.StatusOK, []models.Member{mockMembers[2]}},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMember", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"Current", "GET", "/member/1", ``, http.StatusOK, []models.Member{mockMembers[0]}},
			genericTestcase{"NotExisted", "GET", "/member/24601", ``, http.StatusNotFound, `{"Error":"User Not Found"}`},
			genericTestcase{"NotExisted", "GET", "/member/superman@mirrormedia.mg", ``, http.StatusOK, []models.Member{mockMembers[0]}},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PostMember", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"New", "POST", "/member", `{"member_id":"spaceoddity", "name":"Major Tom", "mail":"spaceoddity"}`, http.StatusOK, `{"_items":{"last_id":4}}`},
			//genericTestcase{"EmptyPayload", "POST", "/member", `{}`, http.StatusBadRequest, `{"Error":"Invalid User"}`},
			//genericTestcase{"Existed", "POST", "/member", `{"id": 1, "member_id":"superman@mirrormedia.mg"}`, http.StatusBadRequest, `{"Error":"User Already Existed"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("CountMembers", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"SimpleCount", "GET", "/members/count", ``, http.StatusOK, `{"_meta":{"total":4}}`},
			genericTestcase{"CountActive", "GET", `/members/count?active={"$in":[1,-1]}`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
			genericTestcase{"CountCustomEditor", "GET", `/members/count?custom_editor=true`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
			genericTestcase{"MoreThanOneActive", "GET", `/members/count?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`},
			genericTestcase{"NotEntirelyValidActive", "GET", `/members/count?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
			genericTestcase{"NoValidActive", "GET", `/members/count?active={"$nin":[3,4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`},
			genericTestcase{"Role", "GET", "/members/count?role=9", ``, http.StatusOK, `{"_meta":{"total":1}}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("KeyNickname", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"Keyword", "GET", `/members/nickname?keyword=readr`, ``, http.StatusOK, `{"_items":[{"id":1,"nickname":"readr"}]}`},
			genericTestcase{"KeywordAndRoles", "GET", `/members/nickname?keyword=readr&roles={"$in":[3,9]}`, ``, http.StatusOK, `{"_items":[{"id":1,"nickname":"readr"}]}`},
			genericTestcase{"InvalidKeyword", "GET", `/members/nickname`, ``, http.StatusBadRequest, `{"Error":"Invalid keyword"}`},
			genericTestcase{"InvalidFields", "GET", `/members/nickname?keyword=readr&fields=["line"]`, ``, http.StatusBadRequest, `{"Error":"Invalid fields: line"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PutMember", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"New", "PUT", "/member", `{"id":1, "name":"Clark Kent"}`, http.StatusOK, ``},
			genericTestcase{"UpdateDailyPush", "PUT", "/member", `{"id":1, "daily_push":true}`, http.StatusOK, ``},
			genericTestcase{"NotExisted", "PUT", "/member", `{"id":24601, "name":"spaceoddity"}`, http.StatusBadRequest, `{"Error":"User Not Found"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})

	t.Run("DeleteMembers", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"Delete", "DELETE", `/members?ids=[2]`, ``, http.StatusOK, ``},
			genericTestcase{"Empty", "DELETE", `/members?ids=[]`, ``, http.StatusBadRequest, `{"Error":"ID List Empty"}`},
			//genericTestcase{"InvalidQueryArray", "DELETE", `/members?ids=["superman@mirrormedia.mg,"test6743"]`, ``, http.StatusBadRequest, `{"Error":"invalid character 't' after array element"}`},
			genericTestcase{"NotFound", "DELETE", `/members?ids=[24601, 24602]`, ``, http.StatusBadRequest, `{"Error":"Members Not Found"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("DeleteMember", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"Current", "DELETE", `/member/3`, ``, http.StatusOK, ``},
			genericTestcase{"NonExisted", "DELETE", `/member/24601`, ``, http.StatusNotFound, `{"Error":"User Not Found"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})
	t.Run("ActivateMultipleMembers", func(t *testing.T) {
		for _, testcase := range []genericTestcase{
			genericTestcase{"CurrentMembers", "PUT", `/members`, `{"ids": [1,2]}`, http.StatusOK, ``},
			genericTestcase{"NotFound", "PUT", `/members`, `{"ids": [24601, 24602]}`, http.StatusNotFound, `{"Error":"Members Not Found"}`},
			genericTestcase{"InvalidPayload", "PUT", `/members`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Request Body"}`},
		} {
			genericDoTest(testcase, t, asserter)
		}
	})

}

func TestRouteMemberUpdatePassword(t *testing.T) {

	type ChangePWCaseIn struct {
		ID       string `json:"id,omitempty"`
		Password string `json:"password,omitempty"`
	}

	var TestRouteChangePWCases = []struct {
		name     string
		in       ChangePWCaseIn
		httpcode int
	}{
		{"ChangePWOK", ChangePWCaseIn{ID: "1", Password: "angrypug"}, http.StatusOK},
		{"ChangePWFail", ChangePWCaseIn{ID: "1"}, http.StatusBadRequest},
		{"ChangePWNoID", ChangePWCaseIn{Password: "angrypug"}, http.StatusBadRequest},
		{"ChangePWMemberNotFound", ChangePWCaseIn{ID: "24601", Password: "angrypug"}, http.StatusNotFound},
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
			member, err := models.MemberAPI.GetMember("id", testcase.in.ID)
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
