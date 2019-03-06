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
	// router := gin.Default()
	router := gin.New()
	router.Use(gin.Recovery())

	// Set customed logger, specify routes skiped from logged
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/metrics"))

	// Init Mysql connections
	models.Connect(dbURI)

	// Init Redis connections
	redisConfig := map[string]string{
		"read_url":  fmt.Sprint(config.Config.Redis.ReadURL),
		"write_url": fmt.Sprint(config.Config.Redis.WriteURL),
		"password":  fmt.Sprint(config.Config.Redis.Password),
	}
	models.RedisConn(redisConfig)
	fmt.Println("Init Redis Connection :", redisConfig)

	// Init Straats API
	models.StraatsHelper.Init()

	// Set postcache settings
	models.InitPostCache()

	// Init SearchFeed
	models.SearchFeed.Init(true)

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
