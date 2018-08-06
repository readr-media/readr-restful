package utils

import (
	"fmt"
	"regexp"
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
	switch resourceType {
	case "post":
		return fmt.Sprintf("/post/%d", resourceID)
	case "project":
		return fmt.Sprintf("/series/%s", slug)
	case "report":
		return fmt.Sprintf("/project/%s", slug)
	case "memo":
		return fmt.Sprintf("/series/%s/%d", slug, resourceID)
	default:
		return resourceString
	}
}
