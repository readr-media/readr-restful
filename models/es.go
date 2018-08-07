package models

import (
	"log"

	"github.com/olivere/elastic"
)

func ESConn(config map[string]string) (*elastic.Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(config["url"]),
		elastic.SetSniff(false))
	if err != nil {
		log.Println("ES connection error:", err)
		return client, err
	}

	return client, nil
}
