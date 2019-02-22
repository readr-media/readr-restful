package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/olivere/elastic"
	"github.com/readr-media/readr-restful/config"
)

type searchEngine struct {
	searchEnabled bool
	metaIndex     string
	indices       map[string]string
	client        *elastic.Client
}

var SearchFeed searchEngine = searchEngine{
	searchEnabled: true,
	metaIndex:     "",
	indices: map[string]string{
		"post":    "post",
		"project": "project",
	}}

func (s *searchEngine) Init() {
	client, err := elastic.NewClient(
		elastic.SetURL(config.Config.SearchFeed.Host),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.Println("Warning: Unable to connect to ES.")
		s.searchEnabled = false
	}

	s.client = client

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(config.Config.SearchFeed.IndexName).Do(context.Background())
	if err != nil {
		log.Println("Warning: Fail to check index.")
		s.searchEnabled = false
	}
	if !exists {
		_, err := client.CreateIndex(config.Config.SearchFeed.IndexName).Do(context.Background())
		if err != nil {
			log.Println("Warning: Fail to create index.")
			s.searchEnabled = false
		}
	}

}

func (s *searchEngine) insert(id int, payload []byte, objectType string) (err error) {
	if s.searchEnabled {
		retry := 0
		for retry < config.Config.SearchFeed.MaxRetry {
			ctx := context.Background()
			_, err := s.client.Index().
				Index(config.Config.SearchFeed.IndexName).
				Type("_doc").
				Id(fmt.Sprintf("%s_%d", objectType, id)).
				BodyString(string(payload)).
				Do(ctx)
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
	return nil
}

func (s *searchEngine) delete(id int, objectType string) (err error) {
	if s.searchEnabled {
		retry := 0
		for retry < config.Config.SearchFeed.MaxRetry {
			ctx := context.Background()
			_, err := s.client.Delete().
				Index(config.Config.SearchFeed.IndexName).
				Type("_doc").
				Id(fmt.Sprintf("%s_%d", objectType, id)).
				Do(ctx)
			if err != nil {
				retry += 1
				time.Sleep(time.Second * 10 * time.Duration(retry))
			} else {
				return nil
			}
		}
		err = fmt.Errorf("Search feed delete fail: %s, retry limit exceeded", err.Error())
		return err
	}
	return nil
}

type searchObject struct {
	TaggedPostMember
	ObjectType string `json:"objectType"`
}

func (a *searchEngine) InsertPost(input []TaggedPostMember) error {
	type searchObject struct {
		TaggedPostMember
		ObjectType string `json:"objectType"`
	}
	for _, tpm := range input {
		typedObject := searchObject{tpm, "post"}
		typedString, err := json.Marshal(typedObject)
		if err != nil {
			log.Println("Marshal post error:", err.Error())
			return err
		}
		a.insert(int(tpm.ID), typedString, "post")
	}
	return nil
}

func (a *searchEngine) InsertProject(input []ProjectAuthors) error {
	type searchObject struct {
		ProjectAuthors
		ObjectType string `json:"objectType"`
	}
	for _, tpm := range input {
		typedObject := searchObject{tpm, "project"}
		typedString, err := json.Marshal(typedObject)
		if err != nil {
			log.Println("Marshal post error:", err.Error())
			return err
		}
		a.insert(int(tpm.Project.ID), typedString, "project")
	}
	return nil
}

func (a *searchEngine) DeletePost(ids []int) error {
	for _, id := range ids {
		a.delete(id, "post")
	}
	return nil
}

func (a *searchEngine) DeleteProject(ids []int) error {
	for _, id := range ids {
		a.delete(id, "project")
	}
	return nil
}
