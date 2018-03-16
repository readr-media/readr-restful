package routes

import (
	"errors"
	"log"
	"testing"

	"encoding/json"
	"net/http"

	"github.com/readr-media/readr-restful/models"
)

var mockProjectDS = []models.Project{}

type mockProjectAPI struct{}

func (a *mockProjectAPI) GetProject(p models.Project) (result models.Project, err error) {
	return models.Project{ID: p.ID}, err
}

func (a *mockProjectAPI) GetProjects(args models.GetProjectArgs) (result []models.Project, err error) {
	if args.Keyword == "%no%" {
		return []models.Project{
			models.Project{ID: 32769, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{50470, true}, Description: models.NullString{"id not provided", true}},
			models.Project{ID: 32234, Title: models.NullString{"nonActive", true}, Active: models.NullInt{0, true}, Order: models.NullInt{60, true}},
		}, nil
	}
	if args.Sorting == "project_id" {
		return []models.Project{
			models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
			models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
			models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
			models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
			models.Project{ID: 32769, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{50470, true}, Description: models.NullString{"id not provided", true}},
		}, nil
	}
	if len(args.Status) == 1 {
		return []models.Project{
			models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
		}, nil
	}
	if len(args.Slugs) == 1 {
		return []models.Project{
			models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
		}, nil
	} else if len(args.Slugs) == 2 {
		return []models.Project{
			models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
			models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
		}, nil
	}
	if len(args.IDs) == 2 {
		return []models.Project{
			models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
			models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
		}, nil
	} else if len(args.IDs) == 1 {
		return []models.Project{}, nil
	}
	if args.MaxResult == 1 && args.Page == 2 {
		return []models.Project{
			models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
		}, nil
	}
	if args.MaxResult == 1 {
		return []models.Project{
			models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
		}, nil
	}
	return []models.Project{
		models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
		models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
		models.Project{ID: 32769, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{50470, true}, Description: models.NullString{"id not provided", true}},
		models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
		models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
	}, nil
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
				resp.Order != expected[i].Order ||
				resp.Slug != expected[i].Slug ||
				resp.Status != expected[i].Status {
				t.Errorf("%s, expect %v, but get %v ", tc.name, expected[i], resp)
			}
		}
	}

	t.Run("PostProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"PostProjectOK", "POST", "/project", `{"id":32768,"title":"OK","post_id":188,"like_amount":0,"comment_amount":0,"active":1,"project_order":60229,"slug":"sampleslug0001"}`, http.StatusOK, ``},
			genericTestcase{"PostProjectSlug", "POST", "/project", `{"id":32233,"title":"OK","post_id":188,"active":1,"project_order":61,"slug":"sampleslug0002","status":2}`, http.StatusOK, ``},
			genericTestcase{"PostProjectNonActive", "POST", "/project", `{"id":32234,"title":"nonActive","post_id":188,"active":0,"project_order":60}`, http.StatusOK, ``},
			genericTestcase{"PostProjectNoID", "POST", "/project", `{"title":"OK","post_id":188,"description":"id not provided", "like_amount":0,"comment_amount":0,"active":1,"project_order":50470}`, http.StatusOK, ``},
			genericTestcase{"PostProjectEmptyBody", "POST", "/project", ``, http.StatusBadRequest, `{"Error":"Invalid Project"}`},
			genericTestcase{"PostProjectDupe", "POST", "/project", `{"id":32767, "title":"Dupe"}`, http.StatusBadRequest, `{"Error":"Project Already Existed"}`},
			genericTestcase{"PostProjectInvalidActive", "POST", "/project", `{"id":11493, "title":"InvActive", "active":3}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("PutProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"UpdateProjectOK", "PUT", "/project", `{"id":32767,"title":"Modified","active":1,"project_order":99999}`, http.StatusOK, ``},
			genericTestcase{"UpdateProjectNotExist", "PUT", "/project", `{"id":11493,"title":"NotExist"}`, http.StatusBadRequest, `{"Error":"Project Not Found"}`},
			genericTestcase{"UpdateProjectInvalidActive", "PUT", "/project", `{"id":32767,"active":3}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`},
			genericTestcase{"UpdatePublishProjectWithNoSlug", "PUT", "/project", `{"id":32769,"status":2}`, http.StatusBadRequest, `{"Error":"Must Have Slug Before Publish"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("GetProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"GetProjectBasicOK", "GET", "/project/list", ``, http.StatusOK, []models.Project{
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
				models.Project{ID: 32769, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{50470, true}, Description: models.NullString{"id not provided", true}},
				models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
				models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetProjectMaxResultOK", "GET", "/project/list?max_result=1", ``, http.StatusOK, []models.Project{
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
			}},
			genericTestcase{"GetProjectOffsetOK", "GET", "/project/list?max_result=1&page=2", ``, http.StatusOK, []models.Project{
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
			}},
			genericTestcase{"GetProjectWithIDsOK", "GET", `/project/list?ids=[1,32767]`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
				models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
			}},
			genericTestcase{"GetProjectWithIDsNotFound", "GET", "/project/list?ids=[9527]", ``, http.StatusOK, `{"_items":[]}`},
			genericTestcase{"GetProjectWithSlugs", "GET", `/project/list?slugs=["sampleslug0001"]`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
			}},
			genericTestcase{"GetProjectWithMultipleSlugs", "GET", `/project/list?slugs=["sampleslug0001","sampleslug0002"]`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
				models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
			}},
			genericTestcase{"GetProjectWithStatus", "GET", `/project/list?status={"$in":[2]}`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
			}},
			genericTestcase{"GetProjectWithSorting", "GET", `/project/list?sort=project_id`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 1, Title: models.NullString{"Alpha", true}, Active: models.NullInt{1, true}},
				models.Project{ID: 32233, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{61, true}, Slug: models.NullString{"sampleslug0002", true}, Status: models.NullInt{2, true}},
				models.Project{ID: 32767, Title: models.NullString{"Modified", true}, Active: models.NullInt{1, true}, Order: models.NullInt{99999, true}},
				models.Project{ID: 32768, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{60229, true}, Slug: models.NullString{"sampleslug0001", true}},
				models.Project{ID: 32769, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{50470, true}, Description: models.NullString{"id not provided", true}},
			}},
			genericTestcase{"GetProjectWithSearchKey", "GET", `/project/list?keyword=no&active={"$in":[0,1]}`, ``, http.StatusOK, []models.Project{
				models.Project{ID: 32769, Title: models.NullString{"OK", true}, Active: models.NullInt{1, true}, Order: models.NullInt{50470, true}, Description: models.NullString{"id not provided", true}},
				models.Project{ID: 32234, Title: models.NullString{"nonActive", true}, Active: models.NullInt{0, true}, Order: models.NullInt{60, true}},
			}},
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
