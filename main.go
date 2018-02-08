package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func init() {

	viper.AddConfigPath("./config")
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
}

func main() {

	sqlHost := viper.Get("sql.host")
	sqlPort := viper.GetInt("sql.port")
	sqlUser := viper.Get("sql.user")
	sqlPass := viper.Get("sql.password")

	models.MemberStatus = viper.GetStringMap("models.members")
	models.PostStatus = viper.GetStringMap("models.posts")
	models.PostType = viper.GetStringMap("models.post_type")
	models.TagStatus = viper.GetStringMap("models.tags")
	models.ProjectStatus = viper.GetStringMap("models.projects")

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

	// Init Redis connetions
	models.RedisConn(map[string]string{
		"url":      fmt.Sprint(viper.Get("redis.host"), ":", viper.Get("redis.port")),
		"password": fmt.Sprint(viper.Get("redis.password")),
	})

	// init mail sender
	dialer := gomail.NewDialer(
		viper.Get("mail.host").(string),
		int(viper.Get("mail.port").(float64)),
		viper.Get("mail.user").(string),
		viper.Get("mail.password").(string),
	)

	//Init cron jobs
	c := cron.New()
	c.AddFunc("@hourly", func() { models.PostCache.SyncFromDataStorage() })
	c.Start()

	models.PostCache.SyncFromDataStorage()

	routes.AuthHandler.SetRoutes(router)
	routes.FollowingHandler.SetRoutes(router)
	routes.MemberHandler.SetRoutes(router)
	routes.PermissionHandler.SetRoutes(router)
	routes.PostHandler.SetRoutes(router)
	routes.ProjectHandler.SetRoutes(router)
	routes.TagHandler.SetRoutes(router)

	routes.MiscHandler.SetRoutes(router, *dialer)

	router.Run()
}
