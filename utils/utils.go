package utils

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"database/sql/driver"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/readr-media/readr-restful/config"
)

func GetResourceTableInfo(resource string) (tableName string, idName string) {

	switch resource {
	case "post":
		tableName = "posts"
		idName = "post_id"
	case "project":
		tableName = "projects"
		idName = "project_id"
	case "memo":
		tableName = "memos"
		idName = "memo_id"
	case "report":
		tableName = "memos"
		idName = "id"
	case "tag":
		tableName = "tags"
		idName = "tag_id"
	default:
		tableName = ""
		idName = ""
	}

	return tableName, idName
}

func ParseResourceInfo(resourceString string) (resourceType string, resourceID string) {
	if matched, _ := regexp.MatchString(`\/post\/[0-9]*$`, resourceString); matched {
		id := regexp.MustCompile(`\/post\/([0-9]*)$`).FindStringSubmatch(resourceString)
		return "post", id[1]
	} else if matched, _ := regexp.MatchString(`\/series\/(.*)$`, resourceString); matched {
		slug := regexp.MustCompile(`\/series\/(.*)$`).FindStringSubmatch(resourceString)
		return "project", slug[1]
	} else if matched, _ := regexp.MatchString(`\/project\/(.*)$`, resourceString); matched {
		slug := regexp.MustCompile(`\/project\/(.*)$`).FindStringSubmatch(resourceString)
		return "report", slug[1]
	} else if matched, _ := regexp.MatchString(`\/series\/.*/([0-9]*)$`, resourceString); matched {
		id := regexp.MustCompile(`\/series\/.*/([0-9]*)$`).FindStringSubmatch(resourceString)
		return "memo", id[1]
	} else {
		return resourceType, resourceID
	}
}

func GenerateResourceInfo(resourceType string, resourceID int, slug string) (resourceString string) {
	resStringPrefix := config.Config.DomainName
	switch resourceType {
	case "post":
		return fmt.Sprintf("%s/post/%d", resStringPrefix, resourceID)
	case "project":
		return fmt.Sprintf("%s/series/%s", resStringPrefix, slug)
	case "report":
		return fmt.Sprintf("%s/project/%s", resStringPrefix, slug)
	case "memo":
		return fmt.Sprintf("%s/series/%s/%d", resStringPrefix, slug, resourceID)
	default:
		return resourceString
	}
}

type Nullable interface {
	Value() (driver.Value, error)
}

func MarshalIgnoreNullNullable(variable interface{}, jsonBody map[string]interface{}) {
	rt, rv := reflect.TypeOf(variable), reflect.ValueOf(variable)
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i).Interface()
		n, ok := fv.(Nullable)
		jsonTag := rt.Field(i).Tag.Get("json")
		if jsonTag != "-" && jsonTag != "" {
			if !ok {
				if rv.Field(i).Kind() == reflect.Ptr && rv.Field(i).IsNil() {
					continue
				} else if (rv.Field(i).Kind() == reflect.Array || rv.Field(i).Kind() == reflect.Slice || rv.Field(i).Kind() == reflect.Map) && rv.Field(i).Len() == 0 {
					continue
				} else {
					jsonBody[jsonTag] = fv
				}
			} else {
				if nv, _ := n.Value(); nv != nil {
					jsonBody[jsonTag] = fv
				}
			}
		}
	}
}

func HTTPRequest(method string, url string, headers map[string]string, body []byte) (resp *http.Response, respBody []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, respBody, errors.New(fmt.Sprintf("http.NewRequest Error: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	c := &http.Client{}
	resp, err = c.Do(req)
	if err != nil {
		return resp, respBody, errors.New(fmt.Sprintf("Make HTTP Request Error: %s", err.Error()))
	}

	defer resp.Body.Close()
	respBody, _ = ioutil.ReadAll(resp.Body)
	return resp, respBody, nil

}

func CutAbstract(html string, length int64, formatter func(abstract []rune) string) (result string, err error) {
	// buf := bytes.NewBuffer(strings.NewReader(html))

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}
	content := doc.Find("p:not(:has(img))").First().Text()

	abstract := []rune(content)
	abstract = abstract[:int(math.Min(float64(len(content)), float64(length)))]
	result = formatter(abstract)
	return result, nil
}
