package routes

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	//"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/models"
)

var r *gin.Engine

type mockDatastore struct{}
var DS models.DatastoreInterface = new(mockDatastore)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// init mockdb according to https://github.com/jmoiron/sqlx/issues/204
	/*mockSQL, _, _ := sqlmock.New()
	defer mockSQL.Close()
	sqlDB := sqlx.NewDb(mockSQL, "sqlmock")*/

	r = gin.New()
	ArticleHandler.SetRoutes(r)
	MemberHandler.SetRoutes(r)

	models.DS = DS
	os.Exit(m.Run())
}


var memberList = []models.Member{
	models.Member{
		ID:     "TaiwanNo.1",
		Active: true,
	},
}

var articleList = []models.Article{
	models.Article{
		ID:     "3345678",
		Author: models.NullString{String: "李宥儒", Valid: true},
		Active: 1,
	},
}

// ------------------------ Implementation of Datastore interface ---------------------------
func (mdb *mockDatastore) Get(item models.TableStruct) (models.TableStruct, error) {

	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for _, value := range memberList {
			if item.ID == value.ID {
				result = value
				err = nil
			}
		}
	case models.Article:
		result = models.Member{}
		err = errors.New("Article Not Found")
		for _, value := range articleList {
			if item.ID == value.ID {
				result = value
				err = nil
			}
		}
	default:
		log.Fatal("Can't not parse model type")
	}
	return result, err
}

func (mdb *mockDatastore) Create(item models.TableStruct) (interface{}, error) {

	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		for _, member := range memberList {
			if item.ID == member.ID {
				return models.Member{}, errors.New("Duplicate entry")
			}
		}
		memberList = append(memberList, item)
		result = memberList[len(memberList)-1]
		err = nil
	case models.Article:
		for _, article := range articleList {
			if item.ID == article.ID {
				result = models.Article{}
				err = errors.New("Duplicate entry")
				return result, err
			}
		}
		articleList = append(articleList, item)
		result = articleList[len(articleList)-1]
		err = nil
	}
	return result, err
}

func (mdb *mockDatastore) Update(item models.TableStruct) (interface{}, error) {
	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for _, value := range memberList {
			if value.ID == item.ID {
				result = item
				err = nil
			}
		}
	case models.Article:
		result = models.Article{}
		err = errors.New("Article Not Found")
		for index, value := range articleList {
			if value.ID == item.ID {
				articleList[index].LikeAmount = item.LikeAmount
				articleList[index].Title = item.Title
				return articleList[index], nil
			}
		}
	}
	return result, err
}

func (mdb *mockDatastore) Delete(item models.TableStruct) (interface{}, error) {
	var (
		result models.TableStruct
		err    error
	)
	switch item := item.(type) {
	case models.Member:
		result = models.Member{}
		err = errors.New("User Not Found")
		for index, value := range memberList {
			if item.ID == value.ID {
				memberList[index].Active = false
				return memberList[index], nil
			}
		}
	case models.Article:
		result = models.Article{}
		err = errors.New("Article Not Found")
		for index, value := range articleList {
			if item.ID == value.ID {
				articleList[index].Active = 0
				return articleList[index], nil
			}
		}
	default:
		log.Fatal("Can't not parse model type")
	}
	return result, err
}

// ---------------------------------- End of Datastore implementation --------------------------------


// func getRouter() *gin.Engine {
// 	r := gin.Default()
// 	return r
// }
