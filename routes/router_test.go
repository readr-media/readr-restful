package routes

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

var r *gin.Engine

type mockDatastore struct{}

func TestMain(m *testing.M) {
	viper.AddConfigPath("../config")
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	dialer := gomail.NewDialer(
		viper.Get("mail.host").(string),
		int(viper.Get("mail.port").(float64)),
		viper.Get("mail.user").(string),
		viper.Get("mail.password").(string),
	)

	gin.SetMode(gin.TestMode)

	r = gin.New()
	PostHandler.SetRoutes(r)
	MemberHandler.SetRoutes(r)
	ProjectHandler.SetRoutes(r)
	AuthHandler.SetRoutes(r)
	MiscHandler.SetRoutes(r, *dialer)

	models.ProjectAPI = new(mockProjectAPI)
	models.MemberAPI = new(mockMemberAPI)
	models.PostAPI = new(mockPostAPI)
	models.PermissionAPI = new(mockPermissionAPI)

	os.Exit(m.Run())
}

// Declare a backup struct for member test data
var mockMemberDSBack []models.Member

var mockMemberDS = []models.Member{
	models.Member{
		ID:        "superman@mirrormedia.mg",
		Active:    1,
		UpdatedAt: models.NullTime{Time: time.Date(2017, 6, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		Mail:      models.NullString{String: "superman@mirrormedia.mg", Valid: true},
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

var mockPostDSBack []models.Post

var mockPostDS = []models.Post{
	models.Post{
		ID:        1,
		Author:    models.NullString{String: "superman@mirrormedia.mg", Valid: true},
		Active:    1,
		UpdatedAt: models.NullTime{Time: time.Date(2017, 11, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		UpdatedBy: models.NullString{String: "superman@mirrormedia.mg", Valid: true},
	},
	models.Post{
		ID:         2,
		Author:     models.NullString{String: "test6743", Valid: true},
		Active:     2,
		Title:      models.NullString{String: "", Valid: true},
		LikeAmount: 256,
		UpdatedAt:  models.NullTime{Time: time.Date(2017, 12, 18, 23, 11, 37, 0, time.UTC), Valid: true},
	},
	models.Post{
		ID:         6,
		Author:     models.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
		Active:     4,
		Title:      models.NullString{String: "", Valid: true},
		LikeAmount: 256,
		UpdatedAt:  models.NullTime{Time: time.Date(2017, 10, 23, 7, 55, 25, 0, time.UTC), Valid: true},
	},
	models.Post{
		ID:        4,
		Active:    3,
		Author:    models.NullString{String: "Major.Tom@mirrormedia.mg", Valid: true},
		UpdatedAt: models.NullTime{Time: time.Date(2018, 1, 3, 12, 22, 20, 0, time.UTC), Valid: true},
		CreatedAt: models.NullTime{Time: time.Date(2017, 12, 31, 23, 59, 59, 999, time.UTC), Valid: true},
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
