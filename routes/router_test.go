package routes

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/models"
)

var r *gin.Engine

type mockDatastore struct{}

// var DS models.DatastoreInterface = new(mockDatastore)
// type mockDatastore struct{}
// var DS models.DatastoreInterface = new(mockDatastore)

var MemberAPI models.MemberInterface = new(mockMemberAPI)
var PostAPI models.PostInterface = new(mockPostAPI)
var ProjectAPI models.ProjectAPIInterface = new(mockProjectAPI)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// init mockdb according to https://github.com/jmoiron/sqlx/issues/204
	/*mockSQL, _, _ := sqlmock.New()
	defer mockSQL.Close()
	sqlDB := sqlx.NewDb(mockSQL, "sqlmock")*/

	r = gin.New()
	PostHandler.SetRoutes(r)
	MemberHandler.SetRoutes(r)
	ProjectHandler.SetRoutes(r)

	// models.DS = DS
	models.ProjectAPI = ProjectAPI
	models.MemberAPI = MemberAPI
	models.PostAPI = PostAPI

	os.Exit(m.Run())
}

var mockMemberDS = []models.Member{
	models.Member{
		ID:        "TaiwanNo.1",
		Active:    1,
		UpdatedAt: models.NullTime{Time: time.Date(2017, 6, 8, 16, 27, 52, 0, time.UTC), Valid: true},
	},
	models.Member{
		ID:        "test6743",
		Active:    1,
		Birthday:  models.NullTime{Time: time.Date(2001, 1, 3, 0, 0, 0, 0, time.UTC), Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 11, 11, 23, 11, 37, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "Lulu_Brakus@yahoo.com", Valid: true},
	},
	models.Member{
		ID:        "Barney.Corwin@hotmail.com",
		Active:    1,
		Gender:    models.NullString{String: "M", Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2017, 1, 3, 19, 32, 37, 0, time.UTC), Valid: true},
		Birthday:  models.NullTime{Time: time.Date(1939, 11, 9, 0, 0, 0, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
	},
}

var mockPostDS = []models.Post{
	models.Post{
		ID:        3345678,
		Author:    models.NullString{String: "李宥儒", Valid: true},
		Active:    1,
		UpdatedAt: models.NullTime{Time: time.Date(2017, 11, 8, 16, 27, 52, 0, time.UTC), Valid: true},
	},
	models.Post{
		ID:         9876543,
		Author:     models.NullString{String: "史學爾", Valid: true},
		Active:     1,
		Title:      models.NullString{String: "", Valid: true},
		LikeAmount: 256,
		UpdatedAt:  models.NullTime{Time: time.Date(2017, 10, 18, 23, 11, 37, 0, time.UTC), Valid: true},
	},
	models.Post{
		ID:         5535678,
		Author:     models.NullString{String: "莊勝雄", Valid: true},
		Active:     1,
		Title:      models.NullString{String: "", Valid: true},
		LikeAmount: 256,
		UpdatedAt:  models.NullTime{Time: time.Date(2017, 9, 6, 7, 55, 25, 0, time.UTC), Valid: true},
	},
}

var mockProjectDS = []models.Project{
	models.Project{
		ID:            "32767",
		Title:         models.NullString{String: "Hello", Valid: true},
		PostID:        0,
		LikeAmount:    0,
		CommentAmount: 0,
		Active:        1,
	},
}

// func getRouter() *gin.Engine {
// 	r := gin.Default()
// 	return r
// }
