package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

var mockProjectDS = []models.Project{}

type mockProjectAPI struct{}

func (a *mockProjectAPI) GetProject(p models.Project) (models.Project, error) {
	result := models.Project{}
	err := errors.New("Project Not Found")
	for _, value := range mockProjectDS {
		if p.ID == value.ID {
			result = value
			err = nil
		}
	}
	return result, err
}

func (a *mockProjectAPI) GetProjects(args models.GetProjectArgs) ([]models.Project, error) {

	var result = []models.Project{}
	var offset = int(args.Page-1) * int(args.MaxResult)

	if len(args.IDs) > 0 {
		for _, project := range mockProjectDS {
			for _, id := range args.IDs {
				if project.ID == id {
					result = append(result, project)
				}
			}
		}
	} else {
		result = mockProjectDS
	}

	if offset > len(mockProjectDS) {
		result = []models.Project{}
	} else {
		result = result[offset:]
	}

	if len(mockProjectDS) > int(args.MaxResult) {
		result = result[:args.MaxResult]
	}

	sort.Slice(result, func(i, j int) bool {
		switch {
		case result[j].Order.Valid == false:
			return true
		case result[i].Order.Valid == false:
			return false
		default:
			return result[i].Order.Int > result[j].Order.Int
		}
	})

	return result, nil
}

func (a *mockProjectAPI) InsertProject(p models.Project) error {
	for _, project := range mockProjectDS {
		if p.ID == project.ID {
			return errors.New("Duplicate entry")
		}
	}

	mockProjectDS = append(mockProjectDS, p)
	if p.ID == 0 {
		lastIndex := len(mockProjectDS)
		mockProjectDS[lastIndex-1].ID = lastIndex
	}

	return nil
}

func (a *mockProjectAPI) UpdateProjects(p models.Project) error {
	err := errors.New("Project Not Found")
	for index, value := range mockProjectDS {
		if value.ID == p.ID {
			mockProjectDS[index] = p
			err = nil
			break
		}
	}
	return err
}

func (a *mockProjectAPI) DeleteProjects(p models.Project) error {
	err := errors.New("Project Not Found")
	for index, value := range mockProjectDS {
		if p.ID == value.ID {
			mockProjectDS[index].Active.Int = 0
			return nil
		}
	}
	return err
}

var MockProjectAPI mockProjectAPI

func TestRouteProjects(t *testing.T) {

	// Clear Data Stores for tests
	mockProjectDSBack := mockProjectDS
	mockProjectDS = []models.Project{}
	if models.DB.DB != nil {
		_, _ = models.DB.Exec("truncate table projects;")
	}

	// Insert test data
	for _, params := range []models.Project{
		models.Project{Active: models.NullInt{1, true}, Title: models.NullString{"Alpha", true}},
		models.Project{ID: 32767, Active: models.NullInt{1, true}, Title: models.NullString{"Omega", true}, Order: models.NullInt{99999, true}},
	} {
		err := models.ProjectAPI.InsertProject(params)
		if err != nil {
			log.Printf("Insert project fail when init test case. Error: %v", err)
		}
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.Project `json:"_items"`
		}

		var Response response
		var expected []models.Project = tc.resp.([]models.Project)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", resp)
		}
		if len(Response.Items) != len(expected) {
			t.Errorf("%s, expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
			return
		}

		for i, resp := range Response.Items {
			if resp.ID != expected[i].ID ||
				resp.Title != expected[i].Title ||
				resp.Active != expected[i].Active ||
				resp.Order != expected[i].Order {
				t.Errorf("%s, expect %v, but get %v ", tc.name, expected[i], resp)
			}
		}
	}

	t.Run("PostProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"PostProjectOK", "POST", "/project", `{"id":32768,"title":"OK","post_id":188,"like_amount":0,"comment_amount":0,"active":1,"order":60229}`, http.StatusOK, ``},
			genericTestcase{"PostProjectEmptyBody", "POST", "/project", ``, http.StatusBadRequest, `{"Error":"Invalid Project"}`},
			genericTestcase{"PostProjectDupe", "POST", "/project", `{"id":32767}`, http.StatusBadRequest, `{"Error":"Project Already Existed"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("PutProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"UpdateProjectOK", "PUT", "/project", `{"ID":32767,"Title":"Modified","active":1,"order":99999}`, http.StatusOK, ``},
			genericTestcase{"UpdateProjectNotExist", "PUT", "/project", `{"ID":11493,"Title":"NotExist"}`, http.StatusBadRequest, `{"Error":"Project Not Found"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("GetProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"GetProjectBasicOK", "GET", "/project/list", ``, http.StatusOK, []models.Project{
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}},
				models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetProjectBasicMaxResultOK", "GET", "/project/list?max_result=1", ``, http.StatusOK, []models.Project{
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
			}},
			genericTestcase{"GetProjectBasicOffsetOK", "GET", "/project/list?max_result=1&page=2", ``, http.StatusOK, []models.Project{
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}},
			}},
			genericTestcase{"GetProjectBasicWithIDsOK", "GET", `/project/list?ids=1&ids=32767`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
				models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetProjectBasicWithIDsNotFound", "GET", "/project/list?ids=9527", ``, http.StatusOK, `{"_items":[]}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("DeleteProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"DeleteProjectOK", "DELETE", "/project/32767", ``, http.StatusOK, ``},
			genericTestcase{"DeleteProjectNotExist", "DELETE", "/project/0", ``, http.StatusNotFound, `{"Error":"Project Not Found"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})

	//Restore backuped data store
	mockProjectDS = mockProjectDSBack
}
