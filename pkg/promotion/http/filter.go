package http

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/pkg/promotion/mysql"
)

// ListParams used to bind query parameters when listing promotion
type ListParams struct {
	MaxResult int    `form:"max_result"`
	Page      int    `form:"page"`
	Sort      string `form:"sort"`
	Total     bool   `form:"total"`

	// Status string
	Active map[string][]int

	// Embedded SQLO in the struct
	// ListParams could be pass through interface because of decoupled Parse()
	o *mysql.SQLO
}

func OperatorParser(operator string) (result string) {
	switch operator {
	case "$gte":
		result = `>=`
	case "$gt":
		result = `>`
	case "$lte":
		result = `<=`
	case "$lt":
		result = `<`
	case "$neq":
		result = `!=`
	case "$eq":
		result = `=`
	case "$in":
		result = `IN`
	case "$nin":
		result = `NOT IN`
	default:
		result = ``
	}
	return result
}

func isValidActive(option int) bool {
	for _, c := range config.Config.Models.Promotions {
		if option == c {
			return true
		}
	}
	return false
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
		if matched, e := regexp.MatchString("-?(updated_at|created_at|id|order)", v); e != nil || !matched {
			err = errors.New("invalid sort")
		}
	}
	// If Active is empty, set default to active before parsing
	if len(p.Active) == 0 {
		p.Active = map[string][]int{"$in": []int{config.Config.Models.Promotions["active"]}}
	}
	return err
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
	if len(p.Active) != 0 {
		// append condition to where statements
		for operator, values := range p.Active {
			if len(values) > 0 {
				p.o.Where = append(p.o.Where, mysql.SQLsv{Statement: fmt.Sprintf("active %s (?)", OperatorParser(operator)), Variable: values})
			}
		}
	}
}

// Select is a wrap for SQLO's Select()
func (p *ListParams) Select() (query string, args []interface{}, err error) {
	query, args, err = p.o.Select()
	return query, args, err
}

// Count is a wrap for SQLO's Count()
func (p *ListParams) Count() (query string, args []interface{}, err error) {
	query, args, err = p.o.Count()
	return query, args, err
}
