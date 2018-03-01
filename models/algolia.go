package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/spf13/viper"
)

var objectTypes = map[string]string{
	"post": "post",
}

type algolia struct {
	client algoliasearch.Client
	index  algoliasearch.Index
}

var Algolia algolia

func (a *algolia) Init() {
	app_id := viper.Get("search_feed.app_id").(string)
	app_key := viper.Get("search_feed.app_key").(string)
	index_name := viper.Get("search_feed.index_name").(string)

	a.client = algoliasearch.NewClient(app_id, app_key)
	a.index = a.client.InitIndex(index_name)
}

func (a *algolia) insert(objects []algoliasearch.Object) error {
	var (
		retry     int   = 0
		max_retry int   = int(viper.Get("search_feed.max_retry").(float64))
		err       error = nil
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
	err = fmt.Errorf("Search feed insert fail: %s, retry limit exceeded.", err.Error())
	return err
}

func (a *algolia) delete(ids []string) error {
	var (
		retry     int   = 0
		max_retry int   = int(viper.Get("search_feed.max_retry").(float64))
		err       error = nil
	)
	for retry < max_retry {
		_, err = a.index.DeleteObjects(ids)
		if err != nil {
			retry += 1
			time.Sleep(time.Second * 10 * time.Duration(retry))
		} else {
			return nil
		}
	}
	err = fmt.Errorf("Search feed delete fail: %s, retry limit exceeded.", err.Error())
	return err
}

func (a *algolia) InsertPost(tpms []TaggedPostMember) error {
	objects := []algoliasearch.Object{}
	for _, tpm := range tpms {
		if tpm.PostMember.Post.ID == 0 {
			return errors.New("No ID")
		}
		o := algoliasearch.Object{}
		o["objectID"] = fmt.Sprintf("%s_%s", objectTypes["post"], fmt.Sprint(tpm.PostMember.Post.ID))
		if tpm.Author.Valid {
			o["author"] = tpm.Author.String
		}
		if tpm.Title.Valid {
			o["title"] = tpm.Title.String
		}
		if tpm.Content.Valid {
			o["content"] = tpm.Content.String
		}
		if tpm.Type.Valid {
			o["type"] = tpm.Type.Int
		}
		if tpm.PublishedAt.Valid {
			o["published_at"] = tpm.PublishedAt.Time
		}
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
		o["objectType"] = objectTypes["post"]
		objects = append(objects, o)
	}

	err := a.insert(objects)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (a *algolia) DeletePost(post_ids []int) error {
	ids := []string{}
	for _, i := range post_ids {
		ids = append(ids, fmt.Sprintf("%s_%s", objectTypes["post"], fmt.Sprint(i)))
	}
	err := a.delete(ids)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}