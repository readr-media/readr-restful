package main

import (
	"flag"
	"fmt"
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

func MemberGetHandler(c *gin.Context) {

	userID := c.Param("id")
	var req models.Databox = &models.Member{ID: userID}
	member, err := req.Get()

	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
	c.JSON(200, member)
}

func MemberPostHandler(c *gin.Context) {

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

	var req models.Databox = &member
	result, err := req.Create()
	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
	c.JSON(200, result)
}

func MemberPutHandler(c *gin.Context) {

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
	var req models.Databox = &member
	result, err := req.Update()
	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
	c.JSON(200, result)
}

func MemberDeleteHandler(c *gin.Context) {

	userID := c.Param("id")
	var req models.Databox = &models.Member{ID: userID}

	member, err := req.Delete()
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

	models.InitDB(dbURI)
	// Plug in mySQL middleware
	// router.Use(sqlMiddleware(dbConn))

	router.GET("/healthz", func(c *gin.Context) {
		c.String(200, "")
	})

	router.GET("/member/:id", MemberGetHandler)
	router.POST("/member", MemberPostHandler)
	router.PUT("/member", MemberPutHandler)
	router.DELETE("/member/:id", MemberDeleteHandler)

	router.Run()
}
