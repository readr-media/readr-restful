package models

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/readr-media/readr-restful/config"
)

type algolia struct {
	client      algoliasearch.Client
	index       algoliasearch.Index
	objectTypes map[string]string
}

var Algolia algolia = algolia{objectTypes: map[string]string{
	"post":    "post",
	"project": "project",
}}

func (a *algolia) Init() {
	// app_id := viper.Get("search_feed.app_id").(string)
	// app_key := viper.Get("search_feed.app_key").(string)
	// index_name := viper.Get("search_feed.index_name").(string)
	app_id := config.Config.SearchFeed.AppID
	app_key := config.Config.SearchFeed.AppKey
	index_name := config.Config.SearchFeed.IndexName

	a.client = algoliasearch.NewClient(app_id, app_key)
	a.index = a.client.InitIndex(index_name)
}

func (a *algolia) insert(objects []algoliasearch.Object) error {
	if os.Getenv("mode") == "local" {
		return nil
	}
	var (
		retry     int   = 0
		max_retry int   = config.Config.SearchFeed.MaxRetry
		err       error = nil
		// max_retry int = int(viper.Get("search_feed.max_retry").(float64))
	)
	for retry < max_retry {
		_, err = a.index.AddObjects(objects)
		if err != nil {
			retry += 1
			time.Sleep(time.Second * 10 * time.Duration(retry))
		} else {
			return nil
		}
	}
	err = fmt.Errorf("Search feed insert fail: %s, retry limit exceeded", err.Error())
	return err
}

func (a *algolia) delete(ids []string) error {
	if os.Getenv("mode") == "local" {
		return nil
	}
	var (
		retry     int   = 0
		max_retry int   = config.Config.SearchFeed.MaxRetry
		err       error = nil
		// max_retry int   = int(viper.Get("search_feed.max_retry").(float64))
	)
	for retry < max_retry {
		_, err = a.index.DeleteObjects(ids)
		if err != nil {
			retry++
			time.Sleep(time.Second * 10 * time.Duration(retry))
		} else {
			return nil
		}
	}
	err = fmt.Errorf("Search feed delete fail: %s, retry limit exceeded", err.Error())
	return err
}

func (a *algolia) insertResource(tpmsi interface{}, resource_name string) (err error) {
	objects := []algoliasearch.Object{}
	switch resource_name {
	case "post":
		tpms, ok := tpmsi.([]TaggedPostMember)
		if !ok {
			return errors.New("Invalid Data Format")
		}
		for _, tpm := range tpms {
			if tpm.Post.ID == 0 {
				return errors.New("Post Has No ID")
			}
			o := a.extractStruct(tpm.Post)
			o["objectID"] = fmt.Sprintf("%s_%s", a.objectTypes[resource_name], fmt.Sprint(tpm.Post.ID))
			o["objectType"] = a.objectTypes[resource_name]
			o["author"] = a.extractStruct(tpm.Member)
			o["updated_by"] = a.extractStruct(tpm.UpdatedBy)
			if tpm.Tags.Valid {
				tags := []int{}
				for _, tag := range strings.Split(tpm.Tags.String, ",") {
					tag_id, err := strconv.Atoi(strings.Split(tag, ":")[0])
					if err != nil {
						return errors.New("Unexpected Non-Integer Tag ID")
					}
					tags = append(tags, tag_id)
				}
				o["tags"] = tags
			}
			objects = append(objects, o)
		}
	case "project":
		tpms, ok := tpmsi.([]ProjectAuthors)
		if !ok {
			return errors.New("Invalid Data Format")
		}
		for _, tpm := range tpms {
			if tpm.ID == 0 {
				return errors.New("Project Has No ID")
			}
			o := a.extractStruct(tpm)
			o["objectID"] = fmt.Sprintf("%s_%s", a.objectTypes[resource_name], fmt.Sprint(tpm.ID))
			o["objectType"] = a.objectTypes[resource_name]
			objects = append(objects, o)
		}
	case "report":
		tpms, ok := tpmsi.([]ReportAuthors)
		if !ok {
			return errors.New("Invalid Data Format")
		}
		for _, tpm := range tpms {
			if tpm.ID == 0 {
				return errors.New("Project Has No ID")
			}
			o := a.extractStruct(tpm)
			o["objectID"] = fmt.Sprintf("%s_%s", a.objectTypes[resource_name], fmt.Sprint(tpm.ID))
			o["objectType"] = a.objectTypes[resource_name]
			objects = append(objects, o)
		}
	default:
		err = errors.New("Resource Not Supported")
	}

	err = a.insert(objects)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (a *algolia) deleteResource(post_ids []int, resource_name string) (err error) {
	ids := []string{}
	for _, i := range post_ids {
		ids = append(ids, fmt.Sprintf("%s_%s", a.objectTypes[resource_name], fmt.Sprint(i)))
	}

	err = a.delete(ids)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (a *algolia) extractStruct(input interface{}) (o algoliasearch.Object) {
	o = algoliasearch.Object{}
	tpm_value := reflect.ValueOf(input)
	if tpm_value.Kind() == reflect.Ptr {
		tpm_value = tpm_value.Elem()
	}
	for i := 0; i < tpm_value.NumField(); i++ {
		name := tpm_value.Type().Field(i).Tag.Get("json")
		if name == "-" {
			continue
		}
		switch field := tpm_value.Field(i).Interface().(type) {
		case int, uint32, string:
			o[name] = field
		case NullString:
			o[name], _ = field.Value()
		case NullBool:
			o[name], _ = field.Value()
		case NullInt:
			o[name], _ = field.Value()
		case NullTime:
			o[name], _ = field.Value()
		}
	}
	return o
}

func (a *algolia) InsertPost(input []TaggedPostMember) error {
	return a.insertResource(input, "post")
}

func (a *algolia) InsertProject(input []ProjectAuthors) error {
	return a.insertResource(input, "project")
}

func (a *algolia) InsertReport(input []ReportAuthors) error {
	return a.insertResource(input, "report")
}

func (a *algolia) DeletePost(ids []int) error {
	return a.deleteResource(ids, "post")
}

func (a *algolia) DeleteProject(ids []int) error {
	return a.deleteResource(ids, "project")
}

func (a *algolia) DeleteReport(ids []int) error {
	return a.deleteResource(ids, "report")
}
