package routes

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/mail"
)

// func init() {
// 	viper.AddConfigPath("../config")
// 	viper.SetConfigName("main")

// 	if err := viper.ReadInConfig(); err != nil {
// 		log.Fatalf("Error reading config file, %s", err)
// 	}

// }

var r *gin.Engine

type mockDatastore struct{}

func TestMain(m *testing.M) {
	os.Setenv("mode", "local")
	os.Setenv("db_driver", "mock")
	/*
		os.Setenv("db_driver", "mysql")
		// TODO: Should implement test set for MODELS
		// Init Sql connetions
		dbURI := "root:qwerty@tcp(127.0.0.1)/memberdb?parseTime=true"
		models.Connect(dbURI)
		_, _ = models.DB.Exec("truncate table projects;")
		_, _ = models.DB.Exec("truncate table members;")
		_, _ = models.DB.Exec("truncate table permissions;")
		_, _ = models.DB.Exec("truncate table posts;")
		_, _ = models.DB.Exec("truncate table following;")
		_, _ = models.DB.Exec("truncate table tags;")
		_, _ = models.DB.Exec("truncate table post_tags;")
		_, _ = models.DB.Exec("truncate table memos;")
		_, _ = models.DB.Exec("truncate table comments;")
		_, _ = models.DB.Exec("truncate table comments_reported;")
		_, _ = models.DB.Exec("truncate table reports;")
		_, _ = models.DB.Exec("truncate table report_authors;")
	*/

	if err := config.LoadConfig("../config", ""); err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

	// Init Redis connetions
	models.RedisConn(map[string]string{
		"read_url":  fmt.Sprint(config.Config.Redis.ReadURL),
		"write_url": fmt.Sprint(config.Config.Redis.WriteURL),
		"password":  fmt.Sprint(config.Config.Redis.Password),
	})

	models.SearchFeed.Init(false)

	gin.SetMode(gin.TestMode)

	r = gin.New()
	SetRoutes(r)

	models.CommentAPI = new(mockCommentAPI)
	models.FollowingAPI = new(mockFollowingAPI)
	models.ProjectAPI = new(mockProjectAPI)
	models.MemberAPI = new(mockMemberAPI)
	models.PostAPI = new(mockPostAPI)
	models.PermissionAPI = new(mockPermissionAPI)
	models.TagAPI = new(mockTagAPI)
	models.MemoAPI = new(mockMemoAPI)
	mail.MailAPI = new(mockMailAPI)
	models.ReportAPI = new(mockReportAPI)
	models.PointsAPI = new(mockPointsAPI)
	models.NotificationGen = new(mockNotificationGenerator)

	models.FollowCache = new(mockFollowCache)
	models.CommentCache = new(mockCommentCache)

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
// var mockMemberDSBack []models.Member

// var mockMemberDS = []models.Member{
// 	models.Member{
// 		ID:           1,
// 		MemberID:     "superman@mirrormedia.mg",
// 		UUID:         "3d64e480-3e30-11e8-b94b-cfe922eb374f",
// 		Nickname:     rrsql.NullString{String: "readr", Valid: true},
// 		Active:       rrsql.NullInt{Int: 1, Valid: true},
// 		UpdatedAt:    rrsql.NullTime{Time: time.Date(2017, 6, 8, 16, 27, 52, 0, time.UTC), Valid: true},
// 		Mail:         rrsql.NullString{String: "superman@mirrormedia.mg", Valid: true},
// 		CustomEditor: rrsql.NullBool{Bool: true, Valid: true},
// 		Role:         rrsql.NullInt{Int: 9, Valid: true},
// 	},
// 	models.Member{
// 		ID:        2,
// 		MemberID:  "test6743@test.test",
// 		UUID:      "3d651126-3e30-11e8-b94b-cfe922eb374f",
// 		Nickname:  rrsql.NullString{String: "yeahyeahyeah", Valid: true},
// 		Active:    rrsql.NullInt{Int: 0, Valid: true},
// 		Birthday:  rrsql.NullTime{Time: time.Date(2001, 1, 3, 0, 0, 0, 0, time.UTC), Valid: true},
// 		UpdatedAt: rrsql.NullTime{Time: time.Date(2017, 11, 11, 23, 11, 37, 0, time.UTC), Valid: true},
// 		Mail:      rrsql.NullString{String: "Lulu_Brakus@yahoo.com", Valid: true},
// 		Role:      rrsql.NullInt{Int: 3, Valid: true},
// 	},
// 	models.Member{
// 		ID:        3,
// 		MemberID:  "Barney.Corwin@hotmail.com",
// 		UUID:      "3d6512e8-3e30-11e8-b94b-cfe922eb374f",
// 		Nickname:  rrsql.NullString{String: "reader", Valid: true},
// 		Active:    rrsql.NullInt{Int: -1, Valid: true},
// 		Gender:    rrsql.NullString{String: "M", Valid: true},
// 		UpdatedAt: rrsql.NullTime{Time: time.Date(2017, 1, 3, 19, 32, 37, 0, time.UTC), Valid: true},
// 		Birthday:  rrsql.NullTime{Time: time.Date(1939, 11, 9, 0, 0, 0, 0, time.UTC), Valid: true},
// 		Mail:      rrsql.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
// 		Role:      rrsql.NullInt{Int: 1, Valid: true},
// 	},
// }

var mockPostDSBack []models.Post

var mockPostDS = []models.Post{
	models.Post{
		ID:        1,
		Author:    rrsql.NullInt{Int: 1, Valid: true},
		Active:    rrsql.NullInt{Int: 1, Valid: true},
		UpdatedAt: rrsql.NullTime{Time: time.Date(2017, 11, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		UpdatedBy: rrsql.NullInt{Int: 1, Valid: true},
		Type:      rrsql.NullInt{Int: 1, Valid: true},
		Slug:      rrsql.NullString{String: "slug", Valid: true},
	},
	models.Post{
		ID:         2,
		Author:     rrsql.NullInt{Int: 2, Valid: true},
		Active:     rrsql.NullInt{Int: 2, Valid: true},
		Title:      rrsql.NullString{String: "", Valid: true},
		LikeAmount: rrsql.NullInt{Int: 256, Valid: true},
		UpdatedAt:  rrsql.NullTime{Time: time.Date(2017, 12, 18, 23, 11, 37, 0, time.UTC), Valid: true},
		Type:       rrsql.NullInt{Int: 2, Valid: true},
		ProjectID:  rrsql.NullInt{Int: 11000, Valid: true},
	},
	models.Post{
		ID:         6,
		Author:     rrsql.NullInt{Int: 3, Valid: true},
		Active:     rrsql.NullInt{Int: 4, Valid: true},
		Title:      rrsql.NullString{String: "", Valid: true},
		LikeAmount: rrsql.NullInt{Int: 257, Valid: true},
		UpdatedAt:  rrsql.NullTime{Time: time.Date(2017, 10, 23, 7, 55, 25, 0, time.UTC), Valid: true},
		Type:       rrsql.NullInt{Int: 0, Valid: true},
		ProjectID:  rrsql.NullInt{Int: 11000, Valid: true},
	},
	models.Post{
		ID:        4,
		Active:    rrsql.NullInt{Int: 3, Valid: true},
		Author:    rrsql.NullInt{Int: 4, Valid: true},
		UpdatedAt: rrsql.NullTime{Time: time.Date(2018, 1, 3, 12, 22, 20, 0, time.UTC), Valid: true},
		CreatedAt: rrsql.NullTime{Time: time.Date(2017, 12, 31, 23, 59, 59, 999, time.UTC), Valid: true},
		Type:      rrsql.NullInt{Int: 1, Valid: true},
	},
}

var mockPermissionDS = []models.Permission{}

// Mocks Objects for External Service Controllers
type mockNotificationGenerator struct{}

func (m mockNotificationGenerator) GenerateCommentNotifications(comment models.InsertCommentArgs) (err error) {
	return nil
}
func (m mockNotificationGenerator) GenerateProjectNotifications(resource interface{}, resourceTyep string) (err error) {
	return nil
}
func (m mockNotificationGenerator) GeneratePostNotifications(p models.TaggedPostMember) (err error) {
	return nil
}

type mockMailAPI struct{}

func (m *mockMailAPI) Send(args mail.MailArgs) (err error)                                { return nil }
func (m *mockMailAPI) SendUpdateNote(args models.GetFollowMapArgs) (err error)            { return nil }
func (m *mockMailAPI) SendUpdateNoteAllResource(args models.GetFollowMapArgs) (err error) { return nil }
func (m *mockMailAPI) GenDailyDigest() (err error)                                        { return err }
func (m *mockMailAPI) SendDailyDigest(s []string) (err error)                             { return err }
func (m *mockMailAPI) SendProjectUpdateMail(resource interface{}, resourceTyep string) (err error) {
	return err
}
func (m *mockMailAPI) SendCECommentNotify(tmp models.TaggedPostMember) (err error)   { return nil }
func (m *mockMailAPI) SendReportPublishMail(report models.ReportAuthors) (err error) { return nil }
func (m *mockMailAPI) SendMemoPublishMail(memo models.MemoDetail) (err error)        { return nil }
func (m *mockMailAPI) SendFollowProjectMail(args models.FollowArgs) (err error)      { return nil }

// func getRouter() *gin.Engine {
// 	r := gin.Default()
// 	return r
// }
