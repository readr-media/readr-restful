package mysql

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/promotion"
)

//dataAPI will implement DataLayer interface implicitly
type dataAPI struct{}

// DataAPI is the single created interface pointed to dataAPI
var DataAPI promotion.DataLayer = new(dataAPI)

func (a *dataAPI) Get(params promotion.ListParams) (results []promotion.Promotion, err error) {

	// Build SELECT query from params
	params.Parse()
	query, values, err := params.Select()
	if err != nil {
		return nil, err
	}
	// Select from db
	err = models.DB.Select(&results, query, values...)
	if err != nil {
		log.Printf("Failed to get promotions from database:%s\n", err.Error())
		return nil, err
	}
	return results, nil
}

func (a *dataAPI) Count(params promotion.ListParams) (count int, err error) {

	params.Parse()
	query, values, err := params.Count()
	if err != nil {
		return 0, err
	}

	// Only select a row for count
	err = models.DB.QueryRow(query, values...).Scan(&count)
	if err != nil {
		log.Printf("Failed to count promotions:%s\n", err.Error())
		return count, err
	}
	return count, nil
}

func (a *dataAPI) Insert(p promotion.Promotion) (int, error) {

	// tags is the db tags in Promotion which is simple type or valid Nullable type
	tags := p.GetTags()
	query := fmt.Sprintf(`INSERT INTO promotions (%s) VALUES (:%s)`, strings.Join(tags, ","), strings.Join(tags, ",:"))

	results, err := models.DB.NamedExec(query, p)
	if err != nil {
		return 0, err
	}

	rowCnt, err := results.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		err := errors.New("More Than One Rows Affected")
		// Handled by log
		log.Fatal(err.Error())
		return 0, err
	} else if rowCnt == 0 {
		return 0, errors.New("Promotion Insert fail")
	}
	// Get last insert row id
	lastID, err := results.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last inserted ID:%v\n", err)
		return 0, err
	}
	return int(lastID), nil
}

func (a *dataAPI) Update(p promotion.Promotion) error {

	tags := p.GetTags()
	// For tags like 'title', create a string 'title = :title'
	// This is used for UPDATE fields in NamedExec for sqlx
	fields := func(pattern string, tags []string) string {
		var results []string
		for _, value := range tags {
			results = append(results, fmt.Sprintf(pattern, value, value))
		}
		return strings.Join(results, " ,")
	}(`%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE promotions SET %s WHERE id = :id`, fields)

	_, err := models.DB.NamedExec(query, p)
	if err != nil {
		return err
	}
	return nil
}

func (a *dataAPI) Delete(id uint64) error {

	result, err := models.DB.Exec(fmt.Sprintf("UPDATE promotions SET active = %d WHERE id = ?", config.Config.Models.Promotions["deactive"]), id)
	if err != nil {
		log.Printf("error deleting promotions in MySQL:%s\n", err.Error())
		return err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Printf("error getting affected rows after deleting promotions:%s\n", err.Error())
	}
	// If delete more than 1 column or no column, return error
	// But this already happend in the database if it's the former situation.

	// There is no meaning in checking rowCnt > 1
	// Its only meaning is we could get an error
	if rowCnt > 1 {
		return errors.New("more than one rows affected")
	} else if rowCnt == 0 {
		return errors.New("no promotion deleted")
	}
	return nil
}
