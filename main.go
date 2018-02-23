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
	models.MemoStatus = viper.GetStringMap("models.memos")
	models.PostStatus = viper.GetStringMap("models.posts")
	models.PostType = viper.GetStringMap("models.post_type")
	models.ProjectStatus = viper.GetStringMap("models.projects")
	models.TagStatus = viper.GetStringMap("models.tags")

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

	// Init Mongodb connection
	models.MongoConn(fmt.Sprint("mongodb://", viper.Get("mongo.talk.host"), ":", viper.Get("mongo.talk.port"), "/talk"))

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

	models.InitPostCache()

	routes.AuthHandler.SetRoutes(router)
	routes.FollowingHandler.SetRoutes(router)
	routes.MemberHandler.SetRoutes(router)
	routes.MemoHandler.SetRoutes(router)
	routes.PermissionHandler.SetRoutes(router)
	routes.PostHandler.SetRoutes(router)
	routes.ProjectHandler.SetRoutes(router)
	routes.TagHandler.SetRoutes(router)

	routes.MiscHandler.SetRoutes(router, *dialer)

	router.Run()
}
