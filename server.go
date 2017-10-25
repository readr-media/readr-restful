package main

import (
	"flag"
	"fmt"

	"github.com/readr-media/readr-restful/models"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	sqlUser    = flag.String("sql-user", "root", "User account to SQL server")
	sqlAddress = flag.String("sql-address", "127.0.0.1:3306", "Address to the SQL server")
	sqlAuth    = flag.String("sql-auth", "", "Password to SQL server")
)

func sqlMiddleware(connString string) gin.HandlerFunc {
	db := sqlx.MustConnect("mysql", connString)

	return func(c *gin.Context) {
		// Registered sqlx db session as "DB" in middleware
		c.Set("DB", db)
		c.Next()
	}
}

func main() {
	flag.Parse()
	fmt.Printf("sql user:%s, sql address:%s, auth:%s \n", *sqlUser, *sqlAddress, *sqlAuth)
	// db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/memberdb", *sqlUser, *sqlAuth, *sqlAddress))
	dbConn := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true", *sqlUser, *sqlAuth, *sqlAddress)

	// Start with default middleware
	router := gin.Default()

	// Plug in mySQL middleware
	router.Use(sqlMiddleware(dbConn))

	router.GET("/healthz", func(c *gin.Context) {
		c.String(200, "")
	})

	router.GET("/member/:id", models.GetMember)
	router.POST("/member", models.InsertMember)
	router.PUT("/member", models.UpdateMember)

	router.Run()
}
