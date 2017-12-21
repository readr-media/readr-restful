package routes

import (
	"os"
	"testing"

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
var ArticleAPI models.ArticleInterface = new(mockArticleAPI)
var ProjectAPI models.ProjectAPIInterface = new(mockProjectAPI)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// init mockdb according to https://github.com/jmoiron/sqlx/issues/204
	/*mockSQL, _, _ := sqlmock.New()
	defer mockSQL.Close()
	sqlDB := sqlx.NewDb(mockSQL, "sqlmock")*/

	r = gin.New()
	ArticleHandler.SetRoutes(r)
	MemberHandler.SetRoutes(r)
	ProjectHandler.SetRoutes(r)

	// models.DS = DS
	models.ProjectAPI = ProjectAPI
	models.MemberAPI = MemberAPI
	models.ArticleAPI = ArticleAPI

	os.Exit(m.Run())
}

var mockMemberDS = []models.Member{
	models.Member{
		ID:     "TaiwanNo.1",
		Active: true,
	},
}

var mockArticleDS = []models.Article{
	models.Article{
		ID:     "3345678",
		Author: models.NullString{String: "李宥儒", Valid: true},
		Active: 1,
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
