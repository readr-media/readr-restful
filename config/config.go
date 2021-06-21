package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config appConfig

type appConfig struct {
	SQL struct {
		Host                    string                       `mapstructure:"host"`
		Port                    int                          `mapstructure:"port"`
		User                    string                       `mapstructure:"user"`
		Password                string                       `mapstructure:"password"`
		SchemaPath              string                       `mapstructure:"schema_path"`
		TableMeta               map[string]map[string]string `mapstructure:"table_meta"`
		TrasactionIDPlaceholder string                       `mapstructure:"trasaction_id_placeholder"`
	} `mapstructure:"sql"`

	Redis struct {
		ReadURL  string `mapstructure:"read_url"`
		WriteURL string `mapstructure:"write_url"`
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
		Host      string `mapstructure:"host"`
		IndexName string `mapstructure:"index_name"`
		MaxRetry  int    `mapstructure:"max_retry"`
	} `mapstructure:"search_feed"`

	Models struct {
		Assets                map[string]int `mapstructure:"assets"`
		AssetType             map[string]int `mapstructure:"asset_type"`
		AssetCopyright        map[string]int `mapstructure:"asset_copyright"`
		Cards                 map[string]int `mapstructure:"cards"`
		CardStatus            map[string]int `mapstructure:"card_status"`
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
		PointStatus           map[string]int `mapstructure:"point_status"`
		HotTagsWeight         map[string]int `mapstructure:"hot_tags_wieght"`
		Promotions            map[string]int `mapstructure:"promotions"`
	} `mapstructure:"models"`

	ReadrID      int    `mapstructure:"readr_id"`
	DefaultOrder int    `mapstructure:"default_order"`
	DomainName   string `mapstructure:"domain_name"`
	TokenSecret  string `mapstructure:"token_secret"`

	PaymentService struct {
		PartnerKey          string `mapstructure:"partner_key"`
		MerchantID          string `mapstructure:"merchant_id"`
		PrimeURL            string `mapstructure:"prime_url"`
		TokenURL            string `mapstructure:"token_url"`
		Currency            string `mapstructure:"currency"`
		PaymentDescription  string `mapstructure:"payment_description"`
		FrontendRedirectUrl string `mapstructure:"frontend_redirect_url"`
		BackendNotifyUrl    string `mapstructure:"backend_notify_url"`
	} `mapstructure:"payment_service"`

	InvoiceService struct {
		MerchantID string `mapstructure:"merchant_id"`
		URL        string `mapstructure:"url"`
		APIVersion string `mapstructure:"api_version"`
		Key        string `mapstructure:"key"`
		IV         string `mapstructure:"iv"`
	} `mapstructure:"invoice_service"`

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
