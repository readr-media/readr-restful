package utils

import (
	"fmt"
	"regexp"

	"github.com/readr-media/readr-restful/config"
)

func ValidateUserID(id string) bool {
	result := true
	if id == "" {
		result = false
	}
	return result
}

func ValidatePassword(id string) bool {
	// to be implemented ...
	if id == "" {
		return false
	}
	return true
}

func ValidateTaggingType(id int) bool {
	for _, v := range config.Config.Models.TaggingType {
		if id == v {
			return true
		}
	}
	return false
}

func ValidateStringArgs(target, pattern string) bool {
	if matched, err := regexp.MatchString(pattern, target); err != nil || !matched {
		fmt.Println(matched)
		return false
	}
	return true
}
