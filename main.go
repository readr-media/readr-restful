package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
	"github.com/spf13/viper"
)

// var (
// 	sqlUser    = flag.String("sql-user", "root", "User account to SQL server")
// 	sqlAddress = flag.String("sql-address", "127.0.0.1:3306", "Address to the SQL server")
// 	sqlAuth    = flag.String("sql-auth", "", "Password to SQL server")
// )

// func sqlMiddleware(connString string) gin.HandlerFunc {
// 	db := sqlx.MustConnect("mysql", connString)

// 	return func(c *gin.Context) {
// 		// Registered sqlx db session as "DB" in middleware
// 		c.Set("DB", db)
// 		c.Next()
// 	}
// }

type Env struct {
	db models.Datastore
}

func (env *Env) MemberGetHandler(c *gin.Context) {

	input := models.Member{ID: c.Param("id")}
	member, err := env.db.Get(input)

	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, member)
}

func (env *Env) MemberPostHandler(c *gin.Context) {

	member := models.Member{}
	c.Bind(&member)

	// Pre-request test
	if member.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid User"})
		return
	}
	if !member.CreateTime.Valid {
		member.CreateTime.Time = time.Now()
		member.CreateTime.Valid = true
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}

	result, err := env.db.Create(member)
	// var req models.Databox = &member
	// result, err := req.Create()
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (env *Env) MemberPutHandler(c *gin.Context) {

	member := models.Member{}
	c.Bind(&member)
	// Use id field to check if Member Struct was binded successfully
	// If the binding failed, id would be emtpy string
	if member.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Member Data"})
		return
	}
	if member.CreateTime.Valid {
		member.CreateTime.Time = time.Time{}
		member.CreateTime.Valid = false
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}
	// var req models.Databox = &member
	// result, err := req.Update()
	result, err := env.db.Update(member)
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (env *Env) MemberDeleteHandler(c *gin.Context) {

	input := models.Member{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	member, err := env.db.Delete(input)

	// member, err := req.Delete()
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, member)
}

func (env *Env) ArticleGetHandler(c *gin.Context) {

	input := models.Article{ID: c.Param("id")}
	article, err := env.db.Get(input)

	if err != nil {
		switch err.Error() {
		case "Article Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Article Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, article)
}

func (env *Env) ArticlePostHandler(c *gin.Context) {

	article := models.Article{}
	err := c.Bind(&article)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if article.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Article ID"})
		return
	}
	if !article.CreateTime.Valid {
		article.CreateTime.Time = time.Now()
		article.CreateTime.Valid = true
	}
	if !article.UpdatedAt.Valid {
		article.UpdatedAt.Time = time.Now()
		article.UpdatedAt.Valid = true
	}
	if article.Active != 1 {
		article.Active = 1
	}
	result, err := env.db.Create(article)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Article ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (env *Env) ArticlePutHandler(c *gin.Context) {

	article := models.Article{}
	c.Bind(&article)
	// Check if article struct was binded successfully
	if article.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Article Data"})
		return
	}
	if article.CreateTime.Valid {
		article.CreateTime.Time = time.Time{}
		article.CreateTime.Valid = false
	}
	if !article.UpdatedAt.Valid {
		article.UpdatedAt.Time = time.Now()
		article.UpdatedAt.Valid = true
	}
	result, err := env.db.Update(article)
	if err != nil {
		switch err.Error() {
		case "Article Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Article Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (env *Env) ArticleDeleteHandler(c *gin.Context) {

	input := models.Article{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	article, err := env.db.Delete(input)

	// member, err := req.Delete()
	if err != nil {
		switch err.Error() {
		case "Article Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Article Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, article)
}

func main() {
	viper.AddConfigPath("./config")
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())

	sqlHost := viper.Get("sql.host")
	sqlPort := viper.GetInt("sql.port")
	sqlUser := viper.Get("sql.user")
	sqlPass := viper.Get("sql.password")
	// flag.Parse()
	// fmt.Printf("sql user:%s, sql address:%s, auth:%s \n", *sqlUser, *sqlAddress, *sqlAuth)
	// fmt.Println(sqlPort)
	fmt.Printf("sql user:%s, sql address:%s, auth:%s \n", sqlUser, fmt.Sprintf("%s:%v", sqlHost, sqlPort), sqlPass)

	// db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/memberdb", *sqlUser, *sqlAuth, *sqlAddress))
	// dbURI := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true", *sqlUser, *sqlAuth, *sqlAddress)
	dbURI := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true", sqlUser, sqlPass, fmt.Sprintf("%s:%v", sqlHost, sqlPort))
	// Start with default middleware
	router := gin.Default()

	// models.InitDB(dbURI)
	db, err := models.NewDB(dbURI)
	if err != nil {
		log.Panic(err)
	}
	env := &Env{db}
	// Plug in mySQL middleware
	// router.Use(sqlMiddleware(dbConn))

	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	router.GET("/member/:id", env.MemberGetHandler)
	router.POST("/member", env.MemberPostHandler)
	router.PUT("/member", env.MemberPutHandler)
	router.DELETE("/member/:id", env.MemberDeleteHandler)

	router.GET("/article/:id", env.ArticleGetHandler)
	router.POST("/article", env.ArticlePostHandler)
	router.PUT("/article", env.ArticlePutHandler)
	router.DELETE("/article/:id", env.ArticleDeleteHandler)

	router.Run()
}
