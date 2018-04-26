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
	models.MemoPublishStatus = viper.GetStringMap("models.memos_publish_status")
	models.PostStatus = viper.GetStringMap("models.posts")
	models.PostType = viper.GetStringMap("models.post_type")
	models.PostPublishStatus = viper.GetStringMap("models.post_publish_status")
	models.ProjectActive = viper.GetStringMap("models.projects_active")
	models.ProjectStatus = viper.GetStringMap("models.projects_status")
	models.ProjectPublishStatus = viper.GetStringMap("models.projects_publish_status")
	models.TagStatus = viper.GetStringMap("models.tags")

	// waiting: allow full utf8 support by incorporating `?charset=utf8mb4` in connection
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
	models.MongoConn(fmt.Sprint(viper.Get("mongo.talk.host")))

	// init mail sender
	dialer := gomail.NewDialer(
		viper.Get("mail.host").(string),
		int(viper.Get("mail.port").(float64)),
		viper.Get("mail.user").(string),
		viper.Get("mail.password").(string),
	)

	// Init Straats API
	models.StraatsHelper.Init()

	//Init cron jobs
	c := cron.New()
	c.AddFunc("@hourly", func() { models.PostCache.SyncFromDataStorage() })
	c.AddFunc("@every 30m", func() { models.StraatsSync.Cron() })

	c.Start()

	models.InitPostCache()

	models.Algolia.Init()

	routes.AuthHandler.SetRoutes(router)
	routes.CommentsHandler.SetRoutes(router)
	routes.FollowingHandler.SetRoutes(router)
	routes.MailHandler.SetRoutes(router, *dialer)
	routes.MemberHandler.SetRoutes(router)
	routes.MemoHandler.SetRoutes(router)
	routes.MiscHandler.SetRoutes(router)
	routes.NotificationHandler.SetRoutes(router)
	routes.PermissionHandler.SetRoutes(router)
	routes.PointsHandler.SetRoutes(router)
	routes.PostHandler.SetRoutes(router)
	routes.ProjectHandler.SetRoutes(router)
	routes.TagHandler.SetRoutes(router)

	router.Run()
}
