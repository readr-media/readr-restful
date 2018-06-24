package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config appConfig

type appConfig struct {
	SQL struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	} `mapstructure:"sql"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		Channels struct {
			TalkComment string `mapstructure:"talk_comments"`
		} `mapstructure:"channels"`
	} `mapstructure:""`

	Crawler struct {
		Header struct {
			Accept    string `mapstructure:"Accept"`
			UserAgent string `mapstructure:"User-Agent"`
			Cookie    string `mapstructure:"Cookie"`
		} `mapstructure:"headers"`
	} `mapstructure:"cralwer"`

	Mail struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		UserName string `mapstructure:"user_name"`
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
		Members struct {
			Active   int `mapstructure:"active"`
			Deactive int `mapstructure:"deactive"`
			Delete   int `mapstructure:"delete"`
		} `mapstructure:"members"`

		Posts struct {
			Deactive int `mapstructure:"deactive"`
			Active   int `mapstructure:"active"`
		} `mapstructure:"posts"`

		PostType struct {
			Review int `mapstructure:"review"`
			News   int `mapstructure:"news"`
			Video  int `mapstructure:"video"`
			Live   int `mapstructure:"live"`
		} `mapstructure:"post_type"`

		PostPublishStatus struct {
			Unpublish int `mapstructure:"unpublish"`
			Draft     int `mapstructure:"draft"`
			Publish   int `mapstructure:"publish"`
			Schedule  int `mapstructure:"schedule"`
			Pending   int `mapstructure:"pending"`
		} `mapstructure:"post_publish_status"`

		Tags struct {
			Deactive int `mapstructure:"deactive"`
			Active   int `mapstructure:"active"`
		} `mapstructure:"tags"`

		ProjectsActive struct {
			Active   int `mapstructure:"active"`
			Deactive int `mapstructure:"deactive"`
		} `mapstructure:"projects_active"`

		ProjectsStatus struct {
			Candidate int `mapstructure:"candidate"`
			Wip       int `mapstructure:"wip"`
			Done      int `mapstructure:"done"`
		} `mapstructure:"projects_status"`

		ProjectsPublishStatus struct {
			Unpublish int `mapstructure:"unpublish"`
			Draft     int `mapstructure:"draft"`
			Publish   int `mapstructure:"publish"`
			Schedule  int `mapstructure:"schedule"`
		} `mapstructure:"projects_publish_status"`

		Memos struct {
			Active   int `mapstructure:"active"`
			Deactive int `mapstructure:"deactive"`
			Pending  int `mapstructure:"pending"`
		} `mapstructure:"memos"`

		MemosPublishStatus struct {
			Unpublish int `mapstructure:"unpublish"`
			Draft     int `mapstructure:"draft"`
			Publish   int `mapstructure:"publish"`
			Schedule  int `mapstructure:"schedule"`
		} `mapstructure:"memos_publish_status"`

		Comment struct {
			Active   int `mapstructure:"active"`
			Deactive int `mapstructure:"deactive"`
		} `mapstructure:"comment"`

		CommentStatus struct {
			Hide int `mapstructure:"hide"`
			Show int `mapstructure:"show"`
		} `mapstructure:"comment_status"`

		ReportedCommentStatus struct {
			Pending  int `mapstructure:"pending"`
			Resolved int `mapstructure:"resolved"`
		} `mapstructure:"reported_comment_status"`

		Reports struct {
			Deactive int `mapstructure:"deactive"`
			Active   int `mapstructure:"active"`
		} `mapstructure:"reports"`

		ReportsPublishStatus struct {
			Unpublish int `mapstructure:"unpublish"`
			Draft     int `mapstructure:"draft"`
			Publish   int `mapstructure:"publish"`
			Schedule  int `mapstructure:"schedule"`
		} `mapstructure:"reports_publish_status"`

		FollowingType struct {
			Member  int `mapstructure:"member"`
			Post    int `mapstructure:"post"`
			Project int `mapstructure:"project"`
			Memo    int `mapstructure:"memo"`
			Report  int `mapstructure:"report"`
		} `mapstructure:"following_type"`
	} `mapstructure:"models"`
}

func LoadConfig(configPaths ...string) error {
	// v := viper.New()

	// fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
	// for _, path := range configPaths {
	// 	// v.AddConfigPath(path)
	// 	fmt.Println(path)
	// }

	// v.AddConfigPath("./config")
	// v.SetConfigName("main")
	// v.SetConfigType("json")

	// if err := v.ReadInConfig(); err != nil {
	// 	log.Fatalf("Error reading config file, %s", err)
	// }

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("Error unmarshal config file, %s", err)
		return err
	}
	return nil
}
