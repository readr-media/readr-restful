package http

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/readr-media/readr-restful/pkg/promotion/mysql"
)

// ListParams used to bind query parameters when listing promotion
type ListParams struct {
	MaxResult int    `form:"max_result"`
	Page      int    `form:"page"`
	Sort      string `form:"sort"`

	// Status string
	// Active *int64

	// Embedded SQLO in the struct
	// ListParams could be pass through interface because of decoupled Parse()
	o *mysql.SQLO
}

// NewListParams create a new ListParams struct, and modify it with input functions,
// then return the result. Leaving the input empty, the return will be empty ListParams
func NewListParams(options ...func(*ListParams) (err error)) (*ListParams, error) {

	// var params ListParams
	params := ListParams{}

	for _, option := range options {
		if err := option(&params); err != nil {
			return nil, err
		}
	}
	return &params, nil
}

func (p *ListParams) validate() (err error) {

	// Validate sort fields
	for _, v := range strings.Split(p.Sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|created_at|id|order)", v); err != nil || !matched {
			return errors.New("invalid sort")
		}
	}
	return nil
}

// Parse will populate the SQLO in ListParams
func (p *ListParams) Parse() {

	// Set table name = "promotions"
	// fields = promotions.*
	p.o = mysql.NewSQLO(func(s *mysql.SQLO) {
		s.Table = "promotions"
		s.Fields = append(s.Fields, mysql.Sqlfield{Table: "promotions", Pattern: `%s.%s`, Fields: []string{"*"}})
	})
	if p.MaxResult != 0 && p.Page > 0 {
		p.o.Pagination = fmt.Sprintf(" LIMIT %d OFFSET %d", p.MaxResult, (p.Page-1)*p.MaxResult)
	}
	if p.Sort != "" {
		p.o.FormatOrderBy(p.Sort)
	}
}

// Select is a wrap for SQLO's Select()
func (p *ListParams) Select() (query string, args []interface{}, err error) {
	query, args, err = p.o.Select()
	return query, args, err
}
