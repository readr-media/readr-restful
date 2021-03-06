package routes

import (
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"testing"

	"encoding/json"
	"net/http"

	"github.com/readr-media/readr-restful/internal/args"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

var mockProjectDS = []models.Project{}

var mockProjectAuthors = []models.Stunt{}

type mockProjectAPI struct{}

func (a *mockProjectAPI) CountProjects(arg args.ArgsParser) (result int, err error) {
	return 5, err
}

func (a *mockProjectAPI) GetProject(p models.Project) (result models.Project, err error) {
	if p.ID == 32768 {
		return models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}, err
	} else {
		return models.Project{ID: p.ID}, err
	}
}

func (a *mockProjectAPI) GetProjects(args models.GetProjectArgs) (result []models.ProjectAuthors, err error) {
	if args.Keyword == "no" {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{0, true}, Order: rrsql.NullInt{60, true}}},
		}, nil
	}
	if args.Keyword == "327" {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
			{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
			{Project: models.Project{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{50470, true}, Description: rrsql.NullString{"id not provided", true}}},
		}, nil
	}
	if args.Sorting == "project_id" {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}}},
			{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
			{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
			{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
			{Project: models.Project{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{50470, true}, Description: rrsql.NullString{"id not provided", true}}},
		}, nil
	}
	if len(args.Status) == 1 {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
			{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
		}, nil
	}
	if len(args.Slugs) == 1 {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
		}, nil
	} else if len(args.Slugs) == 2 {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
			{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
		}, nil
	}
	if len(args.IDs) == 2 {
		if reflect.DeepEqual([]string(args.Fields), args.FullAuthorTags()) {
			return []models.ProjectAuthors{
				models.ProjectAuthors{
					Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}},
					Authors: []models.Stunt{mockProjectAuthors[0]},
				},
				models.ProjectAuthors{
					Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{mockProjectAuthors[0], mockProjectAuthors[1]},
				}}, nil
		} else if reflect.DeepEqual([]string(args.Fields), []string{"id", "nickname"}) {
			return []models.ProjectAuthors{
				models.ProjectAuthors{
					Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}},
					Authors: []models.Stunt{models.Stunt{ID: mockProjectAuthors[0].ID, Nickname: mockProjectAuthors[0].Nickname}},
				},
				models.ProjectAuthors{
					Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{models.Stunt{ID: mockProjectAuthors[0].ID, Nickname: mockProjectAuthors[0].Nickname}, models.Stunt{ID: mockProjectAuthors[1].ID, Nickname: mockProjectAuthors[1].Nickname}},
				}}, nil
		}
		return []models.ProjectAuthors{
			models.ProjectAuthors{
				Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}},
				Authors: []models.Stunt{models.Stunt{Nickname: mockProjectAuthors[0].Nickname}},
			},
			models.ProjectAuthors{
				Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}},
				Authors: []models.Stunt{models.Stunt{Nickname: mockProjectAuthors[0].Nickname}, models.Stunt{Nickname: mockProjectAuthors[1].Nickname}},
			},
		}, nil
	} else if len(args.IDs) == 1 {
		return []models.ProjectAuthors{}, nil
	}
	if len(args.PublishStatus) == 1 {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
		}, nil
	}
	if args.MaxResult == 1 && args.Page == 2 {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
		}, nil
	}
	if args.MaxResult == 1 {
		return []models.ProjectAuthors{
			{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
		}, nil
	}
	return []models.ProjectAuthors{
		{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
		{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
		{Project: models.Project{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{50470, true}, Description: rrsql.NullString{"id not provided", true}}},
		{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
		{Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}}},
	}, nil
}

func (a *mockProjectAPI) GetContents(id int, args models.GetProjectArgs) (result []interface{}, err error) {
	return nil, err
}

func (a *mockProjectAPI) FilterProjects(args *models.FilterProjectArgs) (result []interface{}, err error) {
	return result, err
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

func (a *mockProjectAPI) SchedulePublish() error {
	return nil
}

var MockProjectAPI mockProjectAPI

func TestRouteProjects(t *testing.T) {

	// Clear Data Stores for tests
	mockProjectDSBack := mockProjectDS
	mockProjectDS = []models.Project{}
	if rrsql.DB.DB != nil {
		_, _ = rrsql.DB.Exec("truncate table projects;")
	}

	// Insert test data
	for _, params := range []models.Project{
		models.Project{Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Alpha", true}, PublishStatus: rrsql.NullInt{1, true}, Progress: rrsql.NullFloat{99.87, true}},
		models.Project{ID: 32767, Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Omega", true}, Order: rrsql.NullInt{99999, true}},
	} {
		err := models.ProjectAPI.InsertProject(params)
		if err != nil {
			log.Printf("Insert project fail when init test case. Error: %v", err)
		}
	}

	// Get test author data
	a, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+"_authors.golden"))
	if err != nil {
		t.Fatalf("failed reading .golden: %s", err)
	}
	if err = json.Unmarshal(a, &mockProjectAuthors); err != nil {
		t.Errorf("failed unmarshalling author data: %s", err)
	}

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		type response struct {
			Items []models.ProjectAuthors `json:"_items"`
		}

		var Response response
		var expected = tc.resp.([]models.ProjectAuthors)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", tc.name, resp)
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
			// Check return authors
			if resp.Authors != nil && expected[i].Authors != nil {
				if !reflect.DeepEqual(resp.Authors, expected[i].Authors) {
					t.Errorf("%s, expect authors %v, but get authors %v\n", tc.name, expected[i].Authors, resp.Authors)
				}
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
			genericTestcase{"UpdateProjectStatusOK", "PUT", "/project", `{"id":32768,"status":2}`, http.StatusOK, ``},
			genericTestcase{"UpdateProjectProgressOK", "PUT", "/project", `{"id":32768,"progress":99}`, http.StatusOK, ``},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("GetProject", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"GetProjectBasicOK", "GET", "/project/list", ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{50470, true}, Description: rrsql.NullString{"id not provided", true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}}},
			}},
			genericTestcase{"GetProjectMaxResultOK", "GET", "/project/list?max_result=1", ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
			}},
			genericTestcase{"GetProjectOffsetOK", "GET", "/project/list?max_result=1&page=2", ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
			}},
			genericTestcase{"GetProjectWithIDsOK", "GET", `/project/list?ids=[1,32767]`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{
					Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}},
					Authors: []models.Stunt{models.Stunt{Nickname: mockProjectAuthors[0].Nickname}},
				},
				models.ProjectAuthors{
					Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{models.Stunt{Nickname: mockProjectAuthors[0].Nickname}, models.Stunt{Nickname: mockProjectAuthors[1].Nickname}},
				},
			}},
			genericTestcase{"GetProjectWithIDsNotFound", "GET", "/project/list?ids=[9527]", ``, http.StatusOK, `{"_items":[]}`},
			genericTestcase{"GetProjectWithSlugs", "GET", `/project/list?slugs=["sampleslug0001"]`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
			}},
			genericTestcase{"GetProjectWithMultipleSlugs", "GET", `/project/list?slugs=["sampleslug0001","sampleslug0002"]`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
			}},
			genericTestcase{"GetProjectWithStatus", "GET", `/project/list?status={"$in":[2]}`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
			}},
			genericTestcase{"GetProjectWithPublishStatus", "GET", `/project/list?publish_status={"$in":[1]}`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
			}},
			genericTestcase{"GetProjectWithSorting", "GET", `/project/list?sort=project_id`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{61, true}, Slug: rrsql.NullString{"sampleslug0002", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{50470, true}, Description: rrsql.NullString{"id not provided", true}}},
			}},
			genericTestcase{"GetProjectKeywordMatchTitle", "GET", `/project/list?keyword=no&active={"$in":[0,1]}`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{0, true}, Order: rrsql.NullInt{60, true}}},
			}},
			genericTestcase{"GetProjectKeywordMatchID", "GET", `/project/list?keyword=327`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{60229, true}, Slug: rrsql.NullString{"sampleslug0001", true}, Status: rrsql.NullInt{2, true}}},
				models.ProjectAuthors{Project: models.Project{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{50470, true}, Description: rrsql.NullString{"id not provided", true}}},
			}},
			genericTestcase{"GetProjectCount", "GET", `/project/count`, ``, http.StatusOK, `{"_meta":{"total":5}}`},
			genericTestcase{"GetProjectWithAuthorsFieldsSet", "GET", `/project/list?ids=[1,32767]&fields=["id","nickname"]`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{
					Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}},
					Authors: []models.Stunt{models.Stunt{ID: mockProjectAuthors[0].ID, Nickname: mockProjectAuthors[0].Nickname}},
				},
				models.ProjectAuthors{
					Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{models.Stunt{ID: mockProjectAuthors[0].ID, Nickname: mockProjectAuthors[0].Nickname}, models.Stunt{ID: mockProjectAuthors[1].ID, Nickname: mockProjectAuthors[1].Nickname}},
				},
			}},
			genericTestcase{"GetProjectWithAuthorsFull", "GET", `/project/list?ids=[1,32767]&mode=full`, ``, http.StatusOK, []models.ProjectAuthors{
				models.ProjectAuthors{
					Project: models.Project{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}, Order: rrsql.NullInt{99999, true}},
					Authors: []models.Stunt{mockProjectAuthors[0]},
				},
				models.ProjectAuthors{
					Project: models.Project{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{mockProjectAuthors[0], mockProjectAuthors[1]},
				},
			}},
			genericTestcase{"GetProjectWithAuthorsInvalidFields", "GET", `/project/list?fields=["cat"]`, ``, http.StatusBadRequest, `{"Error":"Invalid Fields"}`},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}
	})
	t.Run("GetProjectContents", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"GetContents", "GET", `/project/contents/unknown`, ``, http.StatusBadRequest, `{"Error":"ID Must Be Integer"}`},
			genericTestcase{"GetContentsWithMemberID", "GET", `/project/contents/1000020?member_id=101`, ``, http.StatusOK, `{"_items":null}`},
			genericTestcase{"GetContentsWithPageAndMaxresult", "GET", `/project/contents/1000020?max_result=10&page=2`, ``, http.StatusOK, `{"_items":null}`},
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
