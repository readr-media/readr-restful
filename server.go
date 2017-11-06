package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
)

var (
	sqlUser    = flag.String("sql-user", "root", "User account to SQL server")
	sqlAddress = flag.String("sql-address", "127.0.0.1:3306", "Address to the SQL server")
	sqlAuth    = flag.String("sql-auth", "", "Password to SQL server")
)

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

	member, err := env.db.Get(c.Param("id"))

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, member)
}

func (env *Env) MemberPostHandler(c *gin.Context) {
	member := models.Member{}
	c.Bind(&member)

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
		c.String(500, "Internal Server Error")
		return
	}
	c.JSON(200, result)
}

func (env *Env) MemberPutHandler(c *gin.Context) {

	member := models.Member{}
	c.Bind(&member)

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
		c.String(500, "Internal Server Error")
		return
	}
	c.JSON(200, result)
}

func (env *Env) MemberDeleteHandler(c *gin.Context) {

	member, err := env.db.Get(c.Param("id"))
	// var req models.Databox = &models.Member{ID: userID}

	// member, err := req.Delete()
	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
	c.JSON(200, member)
}

func main() {
	flag.Parse()
	fmt.Printf("sql user:%s, sql address:%s, auth:%s \n", *sqlUser, *sqlAddress, *sqlAuth)
	// db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/memberdb", *sqlUser, *sqlAuth, *sqlAddress))
	dbURI := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true", *sqlUser, *sqlAuth, *sqlAddress)
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
		c.String(200, "")
	})

	router.GET("/member/:id", env.MemberGetHandler)
	router.POST("/member", env.MemberPostHandler)
	router.PUT("/member", env.MemberPutHandler)
	router.DELETE("/member/:id", env.MemberDeleteHandler)

	router.Run()
}
