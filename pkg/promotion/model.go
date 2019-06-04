package promotion

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

// Promotion is the mapping of the schema for table 'promotions'
type Promotion struct {
	ID          uint64            `json:"id" db:"id"`
	Status      int               `json:"status" db:"status"`
	Active      int               `json:"active" db:"active"`
	Title       string            `json:"title" db:"title"`
	Description models.NullString `json:"description" db:"description"`
	Image       models.NullString `json:"image" db:"image"`
	Link        models.NullString `json:"link" db:"link"`
	Order       models.NullInt    `json:"order" db:"order"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   models.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy   models.NullString `json:"updated_by" db:"updated_by"`
	PublishedAt models.NullTime   `json:"published_at" db:"published_at"`
}

func (p Promotion) getTags() (columns []string) {
	// var columns []string

	u := reflect.ValueOf(p)
	for i := 0; i < u.NumField(); i++ {
		tag := u.Type().Field(i).Tag
		field := u.Field(i).Interface()

		switch field := field.(type) {
		case string:
			if field != "" {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullString:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullTime:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullInt:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullBool:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullIntSlice:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case time.Time:
			columns = append(columns, tag.Get("db"))
		case bool, int, uint32, int64, uint64:
			columns = append(columns, tag.Get("db"))
		default:
			fmt.Println("unrecognised format: ", u.Field(i).Type())
		}
	}
	return columns
}

// DataLayer is the database interface that allow dependency injection for testing
type DataLayer interface {
	Get(params *ListParams) (results []Promotion, err error)
	Insert(p Promotion) (int, error)
	Update(p Promotion) error
	Delete(id uint64) error
}

//dataAPI will implement DataLayer interface implicitly
type dataAPI struct{}

// DataAPI is the single created interface pointed to dataAPI
var DataAPI DataLayer = new(dataAPI)

func (a *dataAPI) Get(params *ListParams) (results []Promotion, err error) {

	// Build SELECT query from params
	query, values, err := params.parse().SQL()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Get query:%s,\nvalues:%v\n", query, values)
	// Select from db
	err = models.DB.Select(&results, query, values...)
	if err != nil {
		log.Printf("Failed to get promotions from database:%s\n", err.Error())
		return nil, err
	}
	return results, nil
}

func (a *dataAPI) Insert(p Promotion) (int, error) {

	// tags is the db tags in Promotion which is simple type or valid Nullable type
	tags := p.getTags()
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

func (a *dataAPI) Update(p Promotion) error {

	tags := p.getTags()
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
		return err
	}
	rowCnt, err := result.RowsAffected()

	// If delete more than 1 column or no column, return error
	// But this already happend in the database if it's the former situation.

	// There is no meaning in checking rowCnt > 1
	// Its only meaning is we could get an error
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Promotion Not Existing")
	}
	return nil
}
