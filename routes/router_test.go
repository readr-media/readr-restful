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
	AuthHandler.SetRoutes(r)

	models.ProjectAPI = new(mockProjectAPI)
	models.MemberAPI = new(mockMemberAPI)
	models.ArticleAPI = new(mockArticleAPI)
	models.PermissionAPI = new(mockPermissionAPI)

	os.Exit(m.Run())
}

var mockMemberDS = []models.Member{
	models.Member{
		ID:     "TaiwanNo.1",
		Active: 1,
	},
}

var mockPostDS = []models.Post{
	models.Post{
		ID:     3345678,
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
