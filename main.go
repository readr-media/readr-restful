package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
	"github.com/robfig/cron"
	"gopkg.in/gomail.v2"
)

// func init() {

// 	viper.AddConfigPath("./config")
// 	viper.SetConfigName("main")

// 	if err := viper.ReadInConfig(); err != nil {
// 		log.Fatalf("Error reading config file, %s", err)
// 	}

// 	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
// }

func main() {

	var (
		configFile string
		configName string
	)
	flag.StringVar(&configFile, "path", "", "Configuration file path.")
	flag.StringVar(&configName, "file", "", "Configuration file name.")
	flag.Parse()
	fmt.Println("path:%s, file:%s\n", configFile, configName)
	if err := config.LoadConfig(configFile, configName); err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}
	fmt.Println(config.Config.Redis)
	// sqlHost := viper.Get("sql.host")
	// sqlPort := viper.GetInt("sql.port")
	// sqlUser := viper.Get("sql.user")
	// sqlPass := viper.Get("sql.password")

	// models.CommentActive = viper.GetStringMap("models.comment")
	// models.CommentStatus = viper.GetStringMap("models.comment_status")
	// models.ReportedCommentStatus = viper.GetStringMap("models.reported_comment_status")
	// models.MemberStatus = viper.GetStringMap("models.members")
	// models.MemoStatus = viper.GetStringMap("models.memos")
	// models.MemoPublishStatus = viper.GetStringMap("models.memos_publish_status")
	// models.ReportActive = viper.GetStringMap("models.reports")
	// models.ReportPublishStatus = viper.GetStringMap("models.reports_publish_status")
	// models.PostStatus = viper.GetStringMap("models.posts")
	// models.PostType = viper.GetStringMap("models.post_type")
	// models.PostPublishStatus = viper.GetStringMap("models.post_publish_status")
	// models.ProjectActive = viper.GetStringMap("models.projects_active")
	// models.ProjectStatus = viper.GetStringMap("models.projects_status")
	// models.ProjectPublishStatus = viper.GetStringMap("models.projects_publish_status")
	// models.TagStatus = viper.GetStringMap("models.tags")
	// routes.OGParserHeaders = viper.GetStringMapString("cralwer.headers")

	// waiting: allow full utf8 support by incorporating `?charset=utf8mb4` in connection
	dbURI := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true&charset=utf8mb4", config.Config.SQL.User, config.Config.SQL.Password, fmt.Sprintf("%s:%v", config.Config.SQL.Host, config.Config.SQL.Port))
	// dbURI := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true&charset=utf8mb4", sqlUser, sqlPass, fmt.Sprintf("%s:%v", sqlHost, sqlPort))

	// Start with default middleware
	router := gin.Default()

	// models.InitDB(dbURI)
	models.Connect(dbURI)

	// Init Redis connetions
	models.RedisConn(map[string]string{
		"url":      fmt.Sprint(config.Config.Redis.Host, ":", config.Config.Redis.Port),
		"password": fmt.Sprint(config.Config.Redis.Password),
	})
	// models.RedisConn(map[string]string{
	// 	"url":      fmt.Sprint(viper.Get("redis.host"), ":", viper.Get("redis.port")),
	// 	"password": fmt.Sprint(viper.Get("redis.password")),
	// })

	// init mail sender
	dialer := gomail.NewDialer(
		config.Config.Mail.Host,
		config.Config.Mail.Port,
		config.Config.Mail.User,
		config.Config.Mail.Password,
	)
	// dialer := gomail.NewDialer(
	// 	viper.Get("mail.host").(string),
	// 	int(viper.Get("mail.port").(float64)),
	// 	viper.Get("mail.user").(string),
	// 	viper.Get("mail.password").(string),
	// )

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
	routes.PubsubHandler.SetRoutes(router)
	routes.ReportHandler.SetRoutes(router)
	routes.TagHandler.SetRoutes(router)

	router.Run()
}
