package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/routes"
	//"github.com/robfig/cron"
)

func main() {

	var (
		configFile string
		configName string
	)
	flag.StringVar(&configFile, "path", "", "Configuration file path.")
	flag.StringVar(&configName, "file", "", "Configuration file name.")
	flag.Parse()

	if err := config.LoadConfig(configFile, configName); err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

	// Include multiStatements=True for migration usage
	dbURI := fmt.Sprintf("%s:%s@tcp(%s)/memberdb?parseTime=true&charset=utf8mb4&multiStatements=true", config.Config.SQL.User, config.Config.SQL.Password, fmt.Sprintf("%s:%v", config.Config.SQL.Host, config.Config.SQL.Port))

	// Start with default middleware
	router := gin.Default()

	// Init Mysql connections
	models.Connect(dbURI)

	// Init Redis connections
	models.RedisConn(map[string]string{
		"url":      fmt.Sprint(config.Config.Redis.Host, ":", config.Config.Redis.Port),
		"password": fmt.Sprint(config.Config.Redis.Password),
	})

	// Init Straats API
	models.StraatsHelper.Init()

	//Init cron jobs
	//c := cron.New()
	//c.AddFunc("@hourly", func() { models.PostCache.SyncFromDataStorage() })
	//c.AddFunc("@every 30m", func() { models.StraatsSync.Cron() })
	//c.Start()

	// Set postcache settings
	models.InitPostCache()

	// Init Algolia
	models.Algolia.Init()

	// Set gin routings
	routes.SetRoutes(router)

	// Implemented Prometheus metrics
	router.GET("/metrics", func() gin.HandlerFunc {
		return func(c *gin.Context) {
			promhttp.Handler().ServeHTTP(c.Writer, c.Request)
		}
	}())

	router.Run()
}
