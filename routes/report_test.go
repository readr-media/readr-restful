package routes

import (
	"errors"
	// "io/ioutil"
	"log"
	// "os"
	// "path/filepath"
	"reflect"
	// "testing"

	// "encoding/json"
	// "net/http"

	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

var mockReportDS = []models.Report{}

var mockReportAuthors = []models.Stunt{}

type mockReportAPI struct{}

func (a *mockReportAPI) CountReports(arg models.GetReportArgs) (result int, err error) {
	return 6, err
}

func (a *mockReportAPI) GetReport(p models.Report) (result models.Report, err error) {
	if p.ID == 32768 {
		return models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}}, err
	} else {
		return models.Report{ID: p.ID}, err
	}
}

func (a *mockReportAPI) GetReports(args models.GetReportArgs) (result []models.ReportAuthors, err error) {
	if args.MaxResult == 1 && len(args.Slugs) > 0 {
		return []models.ReportAuthors{}, errors.New("Not Found")
	}
	if args.Keyword == "no" {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	}
	if args.Keyword == "327" {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
			{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
			{Report: models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}}},
		}, nil
	}
	log.Println(args.Sorting)
	if args.Sorting == "-slug,post_id" {
		return []models.ReportAuthors{

			{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
			{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
			{Report: models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
			{Report: models.Report{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{1, true}}},
			{Report: models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}}},
			{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
		}, nil
	}
	if len(args.Slugs) == 1 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
		}, nil
	} else if len(args.Slugs) == 2 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
			{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
		}, nil
	}
	if len(args.ProjectSlugs) == 1 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
		}, nil
	}
	if len(args.IDs) == 2 {
		if reflect.DeepEqual([]string(args.Fields), args.FullAuthorTags()) {
			return []models.ReportAuthors{
				models.ReportAuthors{
					Report:  models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{mockReportAuthors[0]},
				},
				models.ReportAuthors{
					Report:  models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{mockReportAuthors[0], mockReportAuthors[1]},
				}}, nil
		} else if reflect.DeepEqual([]string(args.Fields), []string{"id", "nickname"}) {
			return []models.ReportAuthors{
				models.ReportAuthors{
					Report:  models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{models.Stunt{ID: mockReportAuthors[0].ID, Nickname: mockReportAuthors[0].Nickname}},
				},
				models.ReportAuthors{
					Report:  models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}},
					Authors: []models.Stunt{models.Stunt{ID: mockReportAuthors[0].ID, Nickname: mockReportAuthors[0].Nickname}, models.Stunt{ID: mockReportAuthors[1].ID, Nickname: mockReportAuthors[1].Nickname}},
				}}, nil
		}
		return []models.ReportAuthors{
			models.ReportAuthors{
				Report:  models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}},
				Authors: []models.Stunt{models.Stunt{Nickname: mockReportAuthors[0].Nickname}},
			},
			models.ReportAuthors{
				Report:  models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}},
				Authors: []models.Stunt{models.Stunt{Nickname: mockReportAuthors[0].Nickname}, models.Stunt{Nickname: mockReportAuthors[1].Nickname}},
			},
		}, nil
	} else if len(args.IDs) == 1 {
		return []models.ReportAuthors{}, nil
	}
	if len(args.ReportPublishStatus) == 1 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
		}, nil
	}
	if len(args.ProjectPublishStatus) == 1 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
		}, nil
	}
	if args.MaxResult == 1 && args.Page == 2 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
		}, nil
	}
	if args.MaxResult == 1 {
		return []models.ReportAuthors{
			{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
		}, nil
	}
	return []models.ReportAuthors{
		{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
		{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
		{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
		{Report: models.Report{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{1, true}}},
		{Report: models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}}},
		{Report: models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
	}, nil
}

func (a *mockReportAPI) InsertReport(p models.Report) (int, error) {
	for _, report := range mockReportDS {
		if p.ID == report.ID {
			return 0, errors.New("Duplicate entry")
		}
	}

	if p.ID == 0 {
		p.ID = 32769
	}

	mockReportDS = append(mockReportDS, p)
	return int(p.ID), nil
}

func (a *mockReportAPI) UpdateReport(p models.Report) error {
	err := errors.New("Report Not Found")
	for index, value := range mockReportDS {
		if value.ID == p.ID {
			mockReportDS[index] = p
			err = nil
			break
		}
	}
	return err
}

func (a *mockReportAPI) DeleteReport(p models.Report) error {
	err := errors.New("Report Not Found")
	for index, value := range mockReportDS {
		if p.ID == value.ID {
			mockReportDS[index].Active.Int = 0
			return nil
		}
	}
	return err
}
func (a *mockReportAPI) InsertAuthors(reportID int, authorIDs []int) (err error) {
	return err
}

func (a *mockReportAPI) UpdateAuthors(reportID int, authorIDs []int) (err error) {
	return err
}
func (m *mockReportAPI) SchedulePublish() ([]int, error)                { return []int{}, nil }
func (m *mockReportAPI) PublishHandler(ids []int) error                 { return nil }
func (m *mockReportAPI) UpdateHandler(ids []int, params ...int64) error { return nil }

var MockReportAPI mockReportAPI

// func TestRouteReports(t *testing.T) {

// 	// Clear Data Stores for tests
// 	mockReportDSBack := mockReportDS
// 	mockReportDS = []models.Report{}
// 	if models.DB.DB != nil {
// 		_, _ = models.DB.Exec("truncate table reports;")
// 	}

// 	if os.Getenv("db_driver") == "mysql" {
// 		_, _ = models.DB.Exec("truncate table projects;")
// 	} else {
// 		mockProjectDS = []models.Project{}
// 	}

// 	// Insert test data
// 	for _, params := range []models.Report{
// 		models.Report{Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Alpha", true}, PublishStatus: rrsql.NullInt{1, true}},
// 		models.Report{ID: 32767, Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Omega", true}},
// 	} {
// 		_, err := models.ReportAPI.InsertReport(params)
// 		if err != nil {
// 			log.Printf("Insert report fail when init test case. Error: %v", err)
// 		}
// 	}

// 	for _, params := range []models.Project{
// 		models.Project{ID: 420, Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Test project for memo", true}, Slug: rrsql.NullString{"testproject", true}},
// 		models.Project{ID: 421, Active: rrsql.NullInt{1, true}, Title: rrsql.NullString{"Test project2 for memo", true}, Slug: rrsql.NullString{"testproject2", true}, PublishStatus: rrsql.NullInt{1, true}},
// 	} {
// 		err := models.ProjectAPI.InsertProject(params)
// 		if err != nil {
// 			log.Printf("Insert project fail when init test case. Error: %v", err)
// 		}
// 	}

// 	// Get test author data
// 	a, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+"_authors.golden"))
// 	if err != nil {
// 		t.Fatalf("failed reading .golden: %s", err)
// 	}
// 	if err = json.Unmarshal(a, &mockReportAuthors); err != nil {
// 		t.Errorf("failed unmarshalling author data: %s", err)
// 	}

// 	asserter := func(resp string, tc genericTestcase, t *testing.T) {
// 		type response struct {
// 			Items []models.ReportAuthors `json:"_items"`
// 		}

// 		var Response response
// 		var expected = tc.resp.([]models.ReportAuthors)

// 		err := json.Unmarshal([]byte(resp), &Response)
// 		if err != nil {
// 			t.Errorf("%s, Unexpected result body: %v", tc.name, resp)
// 		}
// 		if len(Response.Items) != len(expected) {
// 			t.Errorf("%s, expect tag length to be %v but get %v", tc.name, len(expected), len(Response.Items))
// 			return
// 		}

// 		for i, resp := range Response.Items {
// 			if resp.ID != expected[i].ID ||
// 				resp.Title != expected[i].Title ||
// 				resp.Active != expected[i].Active ||
// 				resp.Slug != expected[i].Slug ||
// 				resp.PublishStatus != expected[i].PublishStatus {
// 				t.Errorf("%s, expect %v, but get %v ", tc.name, expected[i], resp)
// 			}
// 			// Check return authors
// 			if resp.Authors != nil && expected[i].Authors != nil {
// 				if !reflect.DeepEqual(resp.Authors, expected[i].Authors) {
// 					t.Errorf("%s, expect authors %v, but get authors %v\n", tc.name, expected[i].Authors, resp.Authors)
// 				}
// 			}
// 		}
// 	}
// 	t.Run("PostReport", func(t *testing.T) {
// 		testcases := []genericTestcase{
// 			genericTestcase{"PostReportOK", "POST", "/report", `{"id":32768,"title":"OK","post_id":188,"like_amount":0,"comment_amount":0,"active":1,"slug":"sampleslug0001", "project_id":421}`, http.StatusOK, `{"_items":{"last_id":32768}}`},
// 			genericTestcase{"PostReportSlug", "POST", "/report", `{"id":32233,"title":"OK","post_id":188,"active":1,"slug":"sampleslug0002", "project_id":420}`, http.StatusOK, `{"_items":{"last_id":32233}}`},
// 			genericTestcase{"PostReportNonActive", "POST", "/report", `{"id":32234,"title":"nonActive","post_id":188}`, http.StatusOK, `{"_items":{"last_id":32234}}`},
// 			genericTestcase{"PostReportNoID", "POST", "/report", `{"title":"OK","post_id":188,"content":"id not provided", "like_amount":0,"comment_amount":0,"active":1}`, http.StatusOK, `{"_items":{"last_id":32769}}`},
// 			genericTestcase{"PostReportEmptyBody", "POST", "/report", ``, http.StatusBadRequest, `{"Error":"Invalid Report"}`},
// 			genericTestcase{"PostReportDupe", "POST", "/report", `{"id":32767, "title":"Dupe"}`, http.StatusBadRequest, `{"Error":"Report Already Existed"}`},
// 		}
// 		for _, tc := range testcases {
// 			genericDoTest(tc, t, asserter)
// 		}
// 	})
// 	t.Run("PutReport", func(t *testing.T) {
// 		testcases := []genericTestcase{
// 			genericTestcase{"UpdateReportOK", "PUT", "/report", `{"id":32767,"title":"Modified","active":1}`, http.StatusOK, ``},
// 			genericTestcase{"UpdateReportNotExist", "PUT", "/report", `{"id":11493,"title":"NotExist"}`, http.StatusBadRequest, `{"Error":"Report Not Found"}`},
// 			genericTestcase{"UpdateReportInvalidActive", "PUT", "/report", `{"id":32767,"active":3}`, http.StatusBadRequest, `{"Error":"Invalid Parameter"}`},
// 			genericTestcase{"UpdatePublishReportWithNoSlug", "PUT", "/report", `{"id":32769,"publish_status":2}`, http.StatusBadRequest, `{"Error":"Must Have Slug Before Publish"}`},
// 			genericTestcase{"UpdateReportStatusOK", "PUT", "/report", `{"id":32768,"publish_status":2}`, http.StatusOK, ``},
// 		}
// 		for _, tc := range testcases {
// 			genericDoTest(tc, t, asserter)
// 		}
// 	})
// 	t.Run("GetReport", func(t *testing.T) {
// 		testcases := []genericTestcase{
// 			genericTestcase{"GetReportBasicOK", "GET", "/report/list", ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{1, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
// 			}},
// 			genericTestcase{"GetReportMaxResultOK", "GET", "/report/list?max_result=1", ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 			}},
// 			genericTestcase{"GetReportOffsetOK", "GET", "/report/list?max_result=1&page=2", ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
// 			}},
// 			genericTestcase{"GetReportWithIDsOK", "GET", `/report/list?ids=[1,32767]`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{
// 					Report:  models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}},
// 					Authors: []models.Stunt{models.Stunt{Nickname: mockReportAuthors[0].Nickname}},
// 				},
// 				models.ReportAuthors{
// 					Report:  models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}},
// 					Authors: []models.Stunt{models.Stunt{Nickname: mockReportAuthors[0].Nickname}, models.Stunt{Nickname: mockReportAuthors[1].Nickname}},
// 				},
// 			}},
// 			genericTestcase{"GetReportWithIDsNotFound", "GET", "/report/list?ids=[9527]", ``, http.StatusOK, `{"_items":[]}`},
// 			genericTestcase{"GetReportWithSlugs", "GET", `/report/list?report_slugs=["sampleslug0001"]`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 			}},
// 			genericTestcase{"GetReportWithMultipleSlugs", "GET", `/report/list?report_slugs=["sampleslug0001","sampleslug0002"]`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 			}},
// 			genericTestcase{"GetReportWithProjectSlugs", "GET", `/report/list?project_slugs=["testproject"]`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
// 			}},
// 			genericTestcase{"GetReportWithReportPublishStatus", "GET", `/report/list?report_publish_status={"$in":[1]}`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
// 			}},
// 			genericTestcase{"GetReportWithProjectPublishStatus", "GET", `/report/list?project_publish_status={"$in":[1]}`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 			}},
// 			genericTestcase{"GetReportWithMultipleSorting", "GET", `/report/list?sort=-slug,id`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32233, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0002", true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{1, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
// 			}},
// 			genericTestcase{"GetReportKeywordMatchTitle", "GET", `/report/list?keyword=no&active={"$in":[0,1]}`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32234, Title: rrsql.NullString{"nonActive", true}, Active: rrsql.NullInt{1, true}}},
// 			}},
// 			genericTestcase{"GetReportKeywordMatchID", "GET", `/report/list?keyword=327`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{Report: models.Report{ID: 32768, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Slug: rrsql.NullString{"sampleslug0001", true}, PublishStatus: rrsql.NullInt{2, true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32769, Title: rrsql.NullString{"OK", true}, Active: rrsql.NullInt{1, true}, Content: rrsql.NullString{"id not provided", true}}},
// 				models.ReportAuthors{Report: models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}}},
// 			}},
// 			genericTestcase{"GetReportCount", "GET", `/report/count`, ``, http.StatusOK, `{"_meta":{"total":6}}`},
// 			genericTestcase{"GetReportWithAuthorsFieldsSet", "GET", `/report/list?ids=[1,32767]&fields=["id","nickname"]`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{
// 					Report:  models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}},
// 					Authors: []models.Stunt{models.Stunt{ID: mockReportAuthors[0].ID, Nickname: mockReportAuthors[0].Nickname}},
// 				},
// 				models.ReportAuthors{
// 					Report:  models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}},
// 					Authors: []models.Stunt{models.Stunt{ID: mockReportAuthors[0].ID, Nickname: mockReportAuthors[0].Nickname}, models.Stunt{ID: mockReportAuthors[1].ID, Nickname: mockReportAuthors[1].Nickname}},
// 				},
// 			}},
// 			genericTestcase{"GetReportWithAuthorsFull", "GET", `/report/list?ids=[1,32767]&mode=full`, ``, http.StatusOK, []models.ReportAuthors{
// 				models.ReportAuthors{
// 					Report:  models.Report{ID: 32767, Title: rrsql.NullString{"Modified", true}, Active: rrsql.NullInt{1, true}},
// 					Authors: []models.Stunt{mockReportAuthors[0]},
// 				},
// 				models.ReportAuthors{
// 					Report:  models.Report{ID: 1, Title: rrsql.NullString{"Alpha", true}, Active: rrsql.NullInt{1, true}, PublishStatus: rrsql.NullInt{1, true}},
// 					Authors: []models.Stunt{mockReportAuthors[0], mockReportAuthors[1]},
// 				},
// 			}},
// 			genericTestcase{"GetReportWithAuthorsInvalidFields", "GET", `/report/list?fields=["cat"]`, ``, http.StatusBadRequest, `{"Error":"Invalid Fields"}`},
// 		}
// 		for _, tc := range testcases {
// 			genericDoTest(tc, t, asserter)
// 		}
// 	})
// 	t.Run("DeleteReport", func(t *testing.T) {
// 		testcases := []genericTestcase{
// 			genericTestcase{"DeleteReportOK", "DELETE", "/report/32767", ``, http.StatusOK, ``},
// 			genericTestcase{"DeleteReportNotExist", "DELETE", "/report/0", ``, http.StatusNotFound, `{"Error":"Report Not Found"}`},
// 		}
// 		for _, tc := range testcases {
// 			genericDoTest(tc, t, asserter)
// 		}
// 	})
// 	t.Run("PostReportAuthors", func(t *testing.T) {
// 		testcases := []genericTestcase{
// 			genericTestcase{"PostReportAuthorsOK", "POST", `/report/author`, `{"report_id":1000010, "author_ids":[1]}`, http.StatusOK, ``},
// 			genericTestcase{"PostReportAuthorsInvalidParameters", "POST", "/report/author", `{"report_id":1000010}`, http.StatusBadRequest, `{"Error":"Insufficient Parameters"}`},
// 		}
// 		for _, tc := range testcases {
// 			genericDoTest(tc, t, asserter)
// 		}
// 	})
// 	t.Run("PutReportAuthors", func(t *testing.T) {
// 		testcases := []genericTestcase{
// 			genericTestcase{"PutReportAuthorsOK", "PUT", `/report/author`, `{"report_id":1000010, "author_ids":[1]}`, http.StatusOK, ``},
// 			genericTestcase{"PutReportAuthorsInvalidParameters", "PUT", "/report/author", `{"author_ids":[1000010]}`, http.StatusBadRequest, `{"Error":"Insufficient Parameters"}`},
// 		}
// 		for _, tc := range testcases {
// 			genericDoTest(tc, t, asserter)
// 		}
// 	})

// 	//Restore backuped data store
// 	mockReportDS = mockReportDSBack

// 	if os.Getenv("db_driver") == "mysql" {
// 		_, _ = models.DB.Exec("truncate table projects;")
// 	} else {
// 		mockProjectDS = []models.Project{}
// 	}

// }
