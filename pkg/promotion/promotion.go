package promotion

import (
	"fmt"
	"reflect"
	"time"

	"github.com/readr-media/readr-restful/internal/rrsql"
)

// Promotion maps the schema of table 'promotions'
type Promotion struct {
	ID          uint64           `json:"id" db:"id"`
	Status      int              `json:"status" db:"status"`
	Active      int              `json:"active" db:"active"`
	Title       string           `json:"title" db:"title"`
	Description rrsql.NullString `json:"description" db:"description"`
	Image       rrsql.NullString `json:"image" db:"image"`
	Link        rrsql.NullString `json:"link" db:"link"`
	Order       rrsql.NullInt    `json:"order" db:"order"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   rrsql.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy   rrsql.NullString `json:"updated_by" db:"updated_by"`
	PublishedAt rrsql.NullTime   `json:"published_at" db:"published_at"`
}

// ListParams setup the interface that could be passed to Get() in DataLayer
type ListParams interface {
	Parse()
	Select() (string, []interface{}, error)
	Count() (string, []interface{}, error)
}

// DataLayer is the database interface that allow dependency injection for testing
//go:generate mockgen -package=mock -destination=mock/mock.go github.com/readr-media/readr-restful/pkg/promotion DataLayer
type DataLayer interface {
	Get(params ListParams) (results []Promotion, err error)
	Count(params ListParams) (count int, err error)
	Insert(p Promotion) (int, error)
	Update(p Promotion) error
	Delete(id uint64) error
}

// GetTags will parse the db tags in Promotion for Insert() and Update()
func (p Promotion) GetTags() (columns []string) {
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
		case rrsql.NullString:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullTime:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullInt:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullBool:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullIntSlice:
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
