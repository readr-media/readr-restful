package routes

import (
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type miscHandler struct{}

type urlMetaInfo struct {
	URL    string         `json:"url"`
	OGInfo *models.OGInfo `json:"og_info"`
}

func (r *miscHandler) GetUrlMeta(c *gin.Context) {
	urlstr := c.Query("url")
	matchedUrls := regexp.MustCompile("https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{2,256}\\.[a-z]{2,6}([-a-zA-Z0-9@:%_\\+.~#?&\\/\\/=]*)").FindAllString(urlstr, -1)

	if len(matchedUrls) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Url String"})
		return
	}
	var result []urlMetaInfo
	for _, url := range matchedUrls {
		ogInfo, err := models.OGParser.GetOGInfoFromUrl(url)
		if err != nil {
			log.Printf("Parse url meta fail: %s, %v \n", url, err.Error())
		}
		result = append(result, urlMetaInfo{URL: url, OGInfo: ogInfo})
	}

	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *miscHandler) PublishResources(c *gin.Context) {
	models.PostAPI.SchedulePublish()
	models.ProjectAPI.SchedulePublish()
	models.MemoAPI.SchedulePublish()
	models.ReportAPI.SchedulePublish()
	c.Status(http.StatusOK)
}

func (r *miscHandler) StraatsSync(c *gin.Context) {
	models.StraatsSync.Cron()
	c.Status(http.StatusOK)
}

func (r *miscHandler) SetRoutes(router *gin.Engine) {
	router.GET("/url/meta", r.GetUrlMeta)

	router.PUT("/schedule/publish", r.PublishResources)
	router.PUT("/schedule/straats", r.StraatsSync)

	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})
}

var MiscHandler miscHandler
