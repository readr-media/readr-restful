package promotion

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ListParams used to bind query parameters when listing promotion
type ListParams struct {
	MaxResult int    `form:"max_result"`
	Page      int    `form:"page"`
	Sort      string `form:"sort"`

	Status string
	Active *int64
}

// NewListParams create a new ListParams struct, and modify it with input functions,
// then return the result. Leaving the input empty, the return will be empty ListParams
func NewListParams(options ...func(*ListParams) (err error)) (*ListParams, error) {

	var params ListParams

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
			return errors.New("Invalid Sort")
		}
	}
	return nil
}

func (p *ListParams) parse() *SQLO {

	// Set table name = "promotions"
	// fields = promotions.*
	s := NewSQLO(func(s *SQLO) {
		s.table = "promotions"
		s.fields = append(s.fields, sqlfield{table: "promotions", pattern: `%s.%s`, fields: []string{"*"}})
	})
	if p.MaxResult != 0 && p.Page > 0 {
		s.pagination = fmt.Sprintf(" LIMIT %d OFFSET %d", p.MaxResult, (p.Page-1)*p.MaxResult)
	}
	if p.Sort != "" {
		s.FormatOrderBy(p.Sort)
	}
	return s
}
