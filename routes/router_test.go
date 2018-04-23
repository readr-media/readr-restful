package routes

import (
	"bytes"
	//"fmt"
	"log"
	"os"
	"testing"
	"time"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
	"github.com/spf13/viper"
)

func init() {
	viper.AddConfigPath("../config")
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

}

var r *gin.Engine

type mockDatastore struct{}

func TestMain(m *testing.M) {
	/*
		os.Setenv("mode", "local")

		// TODO: Should implement test set for MODELS
		// Init Sql connetions
		dbURI := "root:qwerty@tcp(127.0.0.1)/memberdb?parseTime=true"
		models.Connect(dbURI)
		_, _ = models.DB.Exec("truncate table projects;")
		_, _ = models.DB.Exec("truncate table members;")
		_, _ = models.DB.Exec("truncate table permissions;")
		_, _ = models.DB.Exec("truncate table posts;")
		_, _ = models.DB.Exec("truncate table following_members;")
		_, _ = models.DB.Exec("truncate table following_posts;")
		_, _ = models.DB.Exec("truncate table following_projects;")
		_, _ = models.DB.Exec("truncate table tags;")
		_, _ = models.DB.Exec("truncate table post_tags;")
		_, _ = models.DB.Exec("truncate table memos;")

		// Init Redis connetions
		models.RedisConn(map[string]string{
			"url":      fmt.Sprint(viper.Get("redis.host"), ":", viper.Get("redis.port")),
			"password": fmt.Sprint(viper.Get("redis.password")),
		})
		models.Algolia.Init()
	*/
	gin.SetMode(gin.TestMode)

	r = gin.New()
	AuthHandler.SetRoutes(r)
	FollowingHandler.SetRoutes(r)
	MemberHandler.SetRoutes(r)
	MemoHandler.SetRoutes(r)
	PermissionHandler.SetRoutes(r)
	PostHandler.SetRoutes(r)
	ProjectHandler.SetRoutes(r)
	TagHandler.SetRoutes(r)
	MiscHandler.SetRoutes(r)
	MailHandler.SetRoutes(r, initMailDialer())
	PointsHandler.SetRoutes(r)

	models.MemberStatus = viper.GetStringMap("models.members")
	models.MemoStatus = viper.GetStringMap("models.memos")
	models.MemoPublishStatus = viper.GetStringMap("models.memos_publish_status")
	models.PostStatus = viper.GetStringMap("models.posts")
	models.PostType = viper.GetStringMap("models.post_type")
	models.PostPublishStatus = viper.GetStringMap("models.post_publish_status")
	models.ProjectActive = viper.GetStringMap("models.projects_active")
	models.ProjectStatus = viper.GetStringMap("models.projects_status")
	models.ProjectPublishStatus = viper.GetStringMap("models.projects_publish_status")
	models.TagStatus = viper.GetStringMap("models.tags")

	models.ProjectAPI = new(mockProjectAPI)
	models.MemberAPI = new(mockMemberAPI)
	models.PostAPI = new(mockPostAPI)
	models.PermissionAPI = new(mockPermissionAPI)
	models.FollowingAPI = new(mockFollowingAPI)
	models.TagAPI = new(mockTagAPI)
	models.MemoAPI = new(mockMemoAPI)
	models.MailAPI = new(mockMailAPI)

	os.Exit(m.Run())
}

type genericTestcase struct {
	name     string
	method   string
	url      string
	body     interface{}
	httpcode int
	resp     interface{}
}

func genericDoTest(tc genericTestcase, t *testing.T, function interface{}) {
	t.Run(tc.name, func(t *testing.T) {
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

		if w.Code != tc.httpcode {
			t.Errorf("%s want %d but get %d", tc.name, tc.httpcode, w.Code)
		}
		switch tc.resp.(type) {
		case string:
			if w.Body.String() != tc.resp {
				t.Errorf("%s expect (error) message %v but get %v", tc.name, tc.resp, w.Body.String())
			}
		default:
			if fn, ok := function.(func(resp string, tc genericTestcase, t *testing.T)); ok {
				fn(w.Body.String(), tc, t)
			}
		}
	})
}

// Declare a backup struct for member test data
var mockMemberDSBack []models.Member

var mockMemberDS = []models.Member{
	models.Member{
		MemberID:     "superman@mirrormedia.mg",
		UUID:         "3d64e480-3e30-11e8-b94b-cfe922eb374f",
		Nickname:     models.NullString{String: "readr", Valid: true},
		Active:       models.NullInt{Int: 1, Valid: true},
		UpdatedAt:    models.NullTime{Time: time.Date(2017, 6, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		Mail:         models.NullString{String: "superman@mirrormedia.mg", Valid: true},
		CustomEditor: models.NullBool{Bool: true, Valid: true},
		Role:         models.NullInt{Int: 9, Valid: true},
	},
	models.Member{
		MemberID:  "test6743",
		UUID:      "3d651126-3e30-11e8-b94b-cfe922eb374f",
		Nickname:  models.NullString{String: "yeahyeahyeah", Valid: true},
		Active:    models.NullInt{Int: 0, Valid: true},
		Birthday:  models.NullTime{Time: time.Date(2001, 1, 3, 0, 0, 0, 0, time.UTC), Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 11, 11, 23, 11, 37, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "Lulu_Brakus@yahoo.com", Valid: true},
		Role:      models.NullInt{Int: 3, Valid: true},
	},
	models.Member{
		MemberID:  "Barney.Corwin@hotmail.com",
		UUID:      "3d6512e8-3e30-11e8-b94b-cfe922eb374f",
		Nickname:  models.NullString{String: "reader", Valid: true},
		Active:    models.NullInt{Int: -1, Valid: true},
		Gender:    models.NullString{String: "M", Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 1, 3, 19, 32, 37, 0, time.UTC), Valid: true},
		Birthday:  models.NullTime{Time: time.Date(1939, 11, 9, 0, 0, 0, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
		Role:      models.NullInt{Int: 1, Valid: true},
	},
}

var mockPostDSBack []models.Post

var mockPostDS = []models.Post{
	models.Post{
		ID:        1,
		Author:    models.NullString{String: "superman@mirrormedia.mg", Valid: true},
		Active:    models.NullInt{Int: 1, Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 11, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		UpdatedBy: models.NullString{String: "superman@mirrormedia.mg", Valid: true},
		Type:      models.NullInt{Int: 1, Valid: true},
	},
	models.Post{
		ID:         2,
		Author:     models.NullString{String: "test6743", Valid: true},
		Active:     models.NullInt{Int: 2, Valid: true},
		Title:      models.NullString{String: "", Valid: true},
		LikeAmount: models.NullInt{Int: 256, Valid: true},
		UpdatedAt:  models.NullTime{Time: time.Date(2017, 12, 18, 23, 11, 37, 0, time.UTC), Valid: true},
		Type:       models.NullInt{Int: 2, Valid: true},
	},
	models.Post{
		ID:         6,
		Author:     models.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
		Active:     models.NullInt{Int: 4, Valid: true},
		Title:      models.NullString{String: "", Valid: true},
		LikeAmount: models.NullInt{Int: 257, Valid: true},
		UpdatedAt:  models.NullTime{Time: time.Date(2017, 10, 23, 7, 55, 25, 0, time.UTC), Valid: true},
		Type:       models.NullInt{Int: 0, Valid: true},
	},
	models.Post{
		ID:        4,
		Active:    models.NullInt{Int: 3, Valid: true},
		Author:    models.NullString{String: "Major.Tom@mirrormedia.mg", Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2018, 1, 3, 12, 22, 20, 0, time.UTC), Valid: true},
		CreatedAt: models.NullTime{Time: time.Date(2017, 12, 31, 23, 59, 59, 999, time.UTC), Valid: true},
		Type:      models.NullInt{Int: 1, Valid: true},
	},
}

var mockPermissionDS = []models.Permission{}

// func getRouter() *gin.Engine {
// 	r := gin.Default()
// 	return r
// }
