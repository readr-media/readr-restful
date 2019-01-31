package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config appConfig

type appConfig struct {
	SQL struct {
		Host       string                       `mapstructure:"host"`
		Port       int                          `mapstructure:"port"`
		User       string                       `mapstructure:"user"`
		Password   string                       `mapstructure:"password"`
		SchemaPath string                       `mapstructure:"schema_path"`
		TableMeta  map[string]map[string]string `mapstructure:"table_meta"`
	} `mapstructure:"sql"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		Cache    struct {
			LatestCommentCount int `mapstructure:"latest_comment_count"`
			NotificationCount  int `mapstructure:"notification_count"`
		} `mapstructure:"cache"`
	} `mapstructure:"redis"`

	ES struct {
		Url        string `mapstructure:"url"`
		LogIndices string `mapstructure:"log_indices"`
	} `mapstructure:"es"`

	Crawler struct {
		Headers map[string]string `mapstructure:"headers"`
	} `mapstructure:"crawler"`

	Mail struct {
		Host         string `mapstructure:"host"`
		Port         int    `mapstructure:"port"`
		User         string `mapstructure:"user"`
		Password     string `mapstructure:"password"`
		UserName     string `mapstructure:"user_name"`
		DevTeam      string `mapstructure:"dev_team"`
		Enable       bool   `mapstructure:"enable"`
		TemplatePath string `mapstructure:"template_path"`
	} `mapstructure:"mail"`

	SearchFeed struct {
		AppID     string `mapstructure:"app_id"`
		AppKey    string `mapstructure:"app_key"`
		IndexName string `mapstructure:"index_name"`
		MaxRetry  int    `mapstructure:"max_retry"`
	} `mapstructure:"search_feed"`

	Straats struct {
		AppID     string `mapstructure:"app_id"`
		AppKey    string `mapstructure:"app_key"`
		APIServer string `mapstructure:"api_server"`
	} `mapstructure:"straats"`

	Models struct {
		Assets                map[string]int `mapstructure:"assets"`
		Members               map[string]int `mapstructure:"members"`
		MemberDailyPush       map[string]int `mapstructure:"member_daily_push"`
		MemberPostPush        map[string]int `mapstructure:"member_post_push"`
		Posts                 map[string]int `mapstructure:"posts"`
		PostType              map[string]int `mapstructure:"post_type"`
		PostPublishStatus     map[string]int `mapstructure:"post_publish_status"`
		Tags                  map[string]int `mapstructure:"tags"`
		TaggingType           map[string]int `mapstructure:"tagging_type"`
		ProjectsActive        map[string]int `mapstructure:"projects_active"`
		ProjectsStatus        map[string]int `mapstructure:"projects_status"`
		ProjectsPublishStatus map[string]int `mapstructure:"projects_publish_status"`
		Memos                 map[string]int `mapstructure:"memos"`
		MemosPublishStatus    map[string]int `mapstructure:"memos_publish_status"`
		Comment               map[string]int `mapstructure:"comment"`
		CommentStatus         map[string]int `mapstructure:"comment_status"`
		ReportedCommentStatus map[string]int `mapstructure:"reported_comment_status"`
		Reports               map[string]int `mapstructure:"reports"`
		ReportsPublishStatus  map[string]int `mapstructure:"reports_publish_status"`
		FollowingType         map[string]int `mapstructure:"following_type"`
		Emotions              map[string]int `mapstructure:"emotions"`
		PointType             map[string]int `mapstructure:"point_type"`
		HotTagsWeight         map[string]int `mapstructure:"hot_tags_wieght"`
	} `mapstructure:"models"`

	ReadrID    int    `mapstructure:"readr_id"`
	DomainName string `mapstructure:"domain_name"`

	PaymentService struct {
		PartnerKey         string `mapstructure:"partner_key"`
		MerchantID         string `mapstructure:"merchant_id"`
		PrimeURL           string `mapstructure:"prime_url"`
		TokenURL           string `mapstructure:"token_url"`
		Currency           string `mapstructure:"currency"`
		PaymentDescription string `mapstructure:"payment_description"`
	} `mapstructure:"payment_service"`

	Slack struct {
		NotifyWebhook string `mapstructure:"notify_webhook"`
	} `mapstructure:"slack"`
}

func LoadConfig(configPath string, configName string) error {

	v := viper.New()
	v.SetConfigType("json")

	if configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		// Default path
		v.AddConfigPath("./config")
	}

	if configName != "" {
		v.SetConfigName(configName)
	} else {
		v.SetConfigName("main")
	}

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
		return err
	}
	log.Println("Using config file:", v.ConfigFileUsed())

	if err := v.Unmarshal(&Config); err != nil {
		log.Fatalf("Error unmarshal config file, %s", err)
		return err
	}
	return nil
}
