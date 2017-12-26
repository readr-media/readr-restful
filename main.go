package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
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

// type Env struct {
//	db models.Datastore
// }

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
	models.Connect(dbURI)

	// Plug in mySQL middleware
	// router.Use(sqlMiddleware(dbConn))

	routes.MemberHandler.SetRoutes(router)
	routes.PostHandler.SetRoutes(router)
	routes.ProjectHandler.SetRoutes(router)
	routes.AuthHandler.SetRoutes(router)

	routes.MiscHandler.SetRoutes(router)

	router.Run()
}
