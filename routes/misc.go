package routes

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type miscHandler struct{}

type urlMetaInfo struct {
	URL    string         `json:"url"`
	OGInfo *models.OGInfo `json:"og_info"`
}

// parseFilter parse the comma-seperate filters like filter=pnr:published_at<2018-06-26T08:07:27Z
func parseFilter(filter string) (results map[string][]string) {

	results = make(map[string][]string)

	slicedFilters := strings.Split(filter, ",")
	for _, v := range slicedFilters {
		matchedGroup := regexp.MustCompile(`([A-Za-z]+):([A-Za-z_]+)(<|<=|>|>=|==|!=)([A-Za-z0-9:-]+)`).FindAllStringSubmatch(v, -1)
		if len(matchedGroup) > 0 {
			results[matchedGroup[0][1]] = []string{matchedGroup[0][2], matchedGroup[0][3], matchedGroup[0][4]}
		}
	}
	return results
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
	postIDs, err := models.PostAPI.SchedulePublish()
	if err != nil {
		log.Println(err.Error())
	} else {
		PostHandler.PublishPipeline(postIDs)
	}
	models.ProjectAPI.SchedulePublish()

	memoIDs, err := models.MemoAPI.SchedulePublish()
	if err != nil {
		log.Println(err.Error())
	} else {
		MemoHandler.PublishHandler(memoIDs)
		MemoHandler.UpdateHandler(memoIDs)
	}

	reportIDs, err := models.ReportAPI.SchedulePublish()
	if err != nil {
		log.Println(err.Error())
	} else {
		MemoHandler.PublishHandler(reportIDs)
		MemoHandler.UpdateHandler(reportIDs)
	}

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
