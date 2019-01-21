// +build integration

package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
)

var r *gin.Engine

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
	models.Connect(dbURI)

	// Init Redis connetions
	models.RedisConn(map[string]string{
		"url":      fmt.Sprint(config.Config.Redis.Host, ":", config.Config.Redis.Port),
		"password": fmt.Sprint(config.Config.Redis.Password),
	})

	models.Algolia.Init()
	models.InitPostCache()

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

func flushDB() {
	var tableNames []string
	models.DB.Select(&tableNames, `SELECT table_name FROM INFORMATION_SCHEMA.tables WHERE table_schema = 'memberdb'`)

	for _, tableName := range tableNames {
		if tableName != "schema_migrations" && tableName != "roles" {
			if tableName == "polls_choices" || tableName == "members" || tableName == "polls" {
				_, err := models.DB.Exec(fmt.Sprintf("SET FOREIGN_KEY_CHECKS = 0;truncate table %s;SET FOREIGN_KEY_CHECKS = 1", tableName))
				if err != nil {
					print(err.Error(), tableName, "\n")
				}
			} else {
				_, err := models.DB.Exec(fmt.Sprintf("truncate table %s;", tableName))
				if err != nil {
					print(err.Error(), tableName, "\n")
				}
			}
		}
	}
}
func resetRedisKeyHelper(t *testing.T, name string, keys []string) {
	conn := models.RedisHelper.Conn()
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
