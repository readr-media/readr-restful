package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/readr-media/readr-restful/models"
)

type mockProjectAPI struct{}

/* Should implement test set for MODELS
func (a *mockProjectAPI) Init() {
	//realsql test
	dbURI := fmt.Sprintf("root:qwerty@tcp(127.0.0.1)/memberdb?parseTime=true")
	models.Connect(dbURI)
	_, _ = models.DB.Exec("truncate table project_infos;")
	_ = models.ProjectAPI.PostProject(mockProjectDS[0])
}*/

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

func (a *mockProjectAPI) GetProjects(ps ...models.Project) ([]models.Project, error) {
	return nil, nil
}

func (a *mockProjectAPI) PostProject(p models.Project) error {
	for _, project := range mockProjectDS {
		if p.ID == project.ID {
			return errors.New("Duplicate entry")
		}
	}
	mockProjectDS = append(mockProjectDS, p)

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
			mockProjectDS[index].Active = 0
			return nil
		}
	}
	return err
}

var MockProjectAPI mockProjectAPI

func TestRouteGetExsistProject(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/project/32767", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get %d but want %d", w.Code, http.StatusOK)
	}

	var resp models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Get error when Unmarshaling: %q, error: %q", w.Body.Bytes(), err)
	}
	if resp.ID != mockProjectDS[0].ID {
		t.Errorf("Get %d but want %d", resp.ID, mockProjectDS[0].ID)
	}
	if resp.Title != mockProjectDS[0].Title {
		t.Errorf("Get %d but want %d", resp.Title, mockProjectDS[0].Title)
	}
}

func TestRouteGetExsistProjects(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/project/32767", nil)

	r.ServeHTTP(w, req)

	var resp models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Get error when Unmarshaling: %q, error: %q", w.Body.Bytes(), err)
	}
	if resp.ID != mockProjectDS[0].ID {
		t.Errorf("Get %d but want %d", resp.ID, mockProjectDS[0].ID)
	}
	if resp.Title != mockProjectDS[0].Title {
		t.Errorf("Get %d but want %d", resp.Title, mockProjectDS[0].Title)
	}
}

func TestRouteGetNotExistProject(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/project/0", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get %d but want %d", w.Code, http.StatusNotFound)
	}

	expected := `{"Error":"Project Not Found"}`
	if w.Body.String() != string(expected) {
		t.Errorf("Get %q but want %q", w.Body.String(), string(expected))
	}
}

func TestRouteGetProjectsSomeNotExist(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/project/0", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get %d but want %d", w.Code, http.StatusNotFound)
	}

	expected := `{"Error":"Project Not Found"}`
	if w.Body.String() != string(expected) {
		t.Errorf("Get %q but want %q", w.Body.String(), string(expected))
	}
}

func TestRoutePostEmptyProject(t *testing.T) {

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/project", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Get %d but want %d", w.Code, http.StatusBadRequest)
	}
	expected := `{"Error":"Invalid Project"}`
	if w.Body.String() != string(expected) {
		t.Errorf("Get %q but want %q", w.Body.String(), string(expected))
	}
}

func TestRoutePostProject(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"ID":"32768",
		"Title":"OK",
		"PostID":188,
		"LikeAmount":0,
		"CommentAmount":0,
		"Active":1
	}`)
	req, _ := http.NewRequest("POST", "/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get %d but want %d", w.Code, http.StatusOK)
	}

	//No esponse to test ( for now )
	/*var (
		resp     models.Project
		expected models.Project
	)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(jsonStr, &expected); err != nil {
		log.Fatal(err)
	}
	if resp.ID != expected.ID  {
		t.Errorf("Get %d but want %d", resp.ID, expected.ID)
	}*/
}

func TestRoutePostExistedProject(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{"ID":"32767"}`)
	req, _ := http.NewRequest("POST", "/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Get %d but want %d", w.Code, http.StatusBadRequest)
	}
	expected := `{"Error":"Project Already Existed"}`
	if w.Body.String() != string(expected) {
		t.Errorf("Get %q but want %q", w.Body.String(), string(expected))
	}
}

func TestRouteUpdateProject(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
		"ID":"32767", 
		"Title":"Modified"
	}`)
	req, _ := http.NewRequest("PUT", "/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get %d but want %d", w.Code, http.StatusOK)
	}

	//No esponse to test ( for now )
	/*var (
		resp     models.Project
		expected models.Project
	)
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Get error when Unmarshaling: %q, error: %q", w.Body.Bytes(), err)
	}
	if err := json.Unmarshal(jsonStr, &expected); err != nil {
		t.Errorf("Get error when Unmarshaling: %q, error: %q", jsonStr, err)
	}
	if resp.ID != expected.ID {
		t.Errorf("Get %d but want %d", resp.ID, expected.ID)
	}*/
}

func TestRouteUpdateNonExistProject(t *testing.T) {

	w := httptest.NewRecorder()
	var jsonStr = []byte(`{
			"ID":"0", 
			"Title":"NotExist"
		}`)
	req, _ := http.NewRequest("PUT", "/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Get %d but want %d", w.Code, http.StatusBadRequest)
	}
	expected := `{"Error":"Project Not Found"}`
	if w.Body.String() != string(expected) {
		t.Errorf("Get %q but want %q", w.Body.String(), string(expected))
	}
}

func TestRouteDeleteExistProject(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/project/32767", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get %d but want %d", w.Code, http.StatusOK)
	}
	var resp models.Project
	/*if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Get error when Unmarshaling: %q, error: %q", w.Body.Bytes(), err)
	}*/
	if resp.Active > 0 {
		t.Errorf("Get %d but want %d", resp.Active, 0)
	}
}

func TestRouteDeleteNonExistProject(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/project/0", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get %d but want %d", w.Code, http.StatusOK)
	}
	expected := `{"Error":"Project Not Found"}`
	if w.Body.String() != string(expected) {
		t.Errorf("Get %q but want %q", w.Body.String(), string(expected))
	}
}
