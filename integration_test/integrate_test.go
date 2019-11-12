// +build integration

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
)

var r *gin.Engine
var update = flag.Bool("update", false, "update .golden files")

func TestMain(m *testing.M) {
	if err := config.LoadConfig("../config", "integration_test"); err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

	// Init Sql connetions
	//dbURI := fmt.Sprintf("%s@tcp(%s:%d)/memberdb?parseTime=true", config.Config.SQL.User, config.Config.SQL.Host, config.Config.SQL.Port)
	dbURI := fmt.Sprintf(
		"%s:%s@tcp(%s)/memberdb?parseTime=true&charset=utf8mb4&multiStatements=true",
		config.Config.SQL.User,
		config.Config.SQL.Password,
		fmt.Sprintf("%s:%v", config.Config.SQL.Host, config.Config.SQL.Port),
	)
	//dbURI := "root:qwerty@tcp(127.0.0.1)/memberdb?parseTime=true"
	rrsql.Connect(dbURI)

	// Init Redis connetions
	models.RedisConn(map[string]string{
		"read_url":  fmt.Sprint(config.Config.Redis.ReadURL),
		"write_url": fmt.Sprint(config.Config.Redis.WriteURL),
		"password":  fmt.Sprint(config.Config.Redis.Password),
	})

	models.InitPostCache()

	// Init SearchFeed
	models.SearchFeed.Init(false)

	// Set gin routings
	gin.SetMode(gin.TestMode)
	r = gin.New()
	routes.SetRoutes(r)

	flushDB()
	os.Exit(m.Run())
}

type genericRequestTestcase struct {
	name     string
	method   string
	url      string
	body     interface{}
	httpcode int
	resp     string
	misc     []interface{}
}

func genericDoRequest(tc genericRequestTestcase, t *testing.T) (int, string) {
	w := httptest.NewRecorder()
	jsonStr := []byte{}
	if s, ok := tc.body.(string); ok {
		jsonStr = []byte(s)
	} else {
		p, err := json.Marshal(tc.body)
		if err != nil {
			t.Errorf("%s, Error when marshaling input parameters", tc.name)
		}
		jsonStr = p
	}
	req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(jsonStr))
	if tc.method == "GET" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	r.ServeHTTP(w, req)

	return w.Code, w.Body.String()
}

func genericDoRequestByte(tc genericRequestTestcase, t *testing.T) (int, []byte) {
	w := httptest.NewRecorder()
	jsonStr := []byte{}
	if s, ok := tc.body.(string); ok {
		jsonStr = []byte(s)
	} else {
		p, err := json.Marshal(tc.body)
		if err != nil {
			t.Errorf("%s, Error when marshaling input parameters", tc.name)
		}
		jsonStr = p
	}
	req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(jsonStr))
	if tc.method == "GET" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	r.ServeHTTP(w, req)

	return w.Code, w.Body.Bytes()
}

func flushDB() {
	var tableNames []string
	rrsql.DB.Select(&tableNames, `SELECT table_name FROM INFORMATION_SCHEMA.tables WHERE table_schema = 'memberdb'`)

	for _, tableName := range tableNames {
		if tableName != "schema_migrations" && tableName != "roles" {
			if tableName == "polls_choices" || tableName == "members" || tableName == "polls" {
				_, err := rrsql.DB.Exec(fmt.Sprintf("SET FOREIGN_KEY_CHECKS = 0;truncate table %s;SET FOREIGN_KEY_CHECKS = 1", tableName))
				if err != nil {
					print(err.Error(), tableName, "\n")
				}
			} else {
				_, err := rrsql.DB.Exec(fmt.Sprintf("truncate table %s;", tableName))
				if err != nil {
					print(err.Error(), tableName, "\n")
				}
			}
		}
	}
}
func resetRedisKeyHelper(t *testing.T, name string, keys []string) {
	conn := models.RedisHelper.WriteConn()
	defer conn.Close()
	for _, v := range keys {
		_, err := conn.Do("DEL", v)
		if err != nil {
			t.Fatalf("Error delete redis keys %s : %v", v, err.Error())
		}
	}
	return
}
func statusCodeOKHelper(code int) bool {
	if strconv.Itoa(code)[0] == "2"[0] {
		return true
	} else {
		return false
	}
}
func assertIntHelper(t *testing.T, name string, subject string, want int, get int) {
	if want != get {
		t.Errorf("%s expect %s to be %d but get %d", name, subject, want, get)
	}
}
func assertStringHelper(t *testing.T, name string, subject string, want string, get string) {
	if want != get {
		t.Errorf("%s expect %s to be %s but get %s", name, subject, want, get)
	}
}
func assertByteHelper(t *testing.T, name string, subject string, want []byte, get []byte) {
	if !bytes.Equal(want, get) {
		t.Errorf("%s expect %s to be %s but get %s", name, subject, want, get)
	}
}

type Golden struct {
	update bool
}

func (g *Golden) SetUpdate(update bool) {
	g.update = update
}
func (g *Golden) AssertOrUpdate(t *testing.T, actualData []byte) {
	if g.update {
		err := g.Update(t.Name(), actualData)
		if err != nil {
			t.Errorf("failed to update golden file: %s", err)
		}
	} else {
		err := g.Assert(t.Name(), actualData)
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}

func (g *Golden) Update(name string, actualData []byte) error {
	if err := g.ensureDir(filepath.Dir(g.goldenFileName(name))); err != nil {
		return err
	}
	return ioutil.WriteFile(g.goldenFileName(name), actualData, 0644)
}

func (g *Golden) goldenFileName(name string) string {
	return filepath.Join("testdata", filepath.FromSlash(name)+".golden")
}

func (g *Golden) ensureDir(path string) error {
	s, err := os.Stat(path)
	switch {
	case err != nil && os.IsNotExist(err):
		// the location does not exist, so make directories to there
		return os.MkdirAll(path, 0755)
	case err == nil && !s.IsDir():
		return fmt.Errorf("testdata is a file")
	}

	return err
}

func (g *Golden) Assert(name string, actualData []byte) error {
	expectedData, err := ioutil.ReadFile(g.goldenFileName(name))

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("FileNotExist")
		}

		return fmt.Errorf("ReadFileError: %s", err.Error())
	}

	if !bytes.Equal(g.trimTime(actualData), g.trimTime(expectedData)) {
		return fmt.Errorf(
			"Result did not match the golden fixture.\n"+
				"Expected: %s\n"+
				"Got: %s",
			string(expectedData),
			string(actualData))
	}

	return nil
}

func (g *Golden) trimTime(s []byte) []byte {
	var re = regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`)
	trimed := re.ReplaceAllString(string(s), "")
	return []byte(trimed)
}
