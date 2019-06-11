package mysql

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/models"
)

type sqlsv struct {
	statement string
	variable  interface{}
}

type Sqlfield struct {
	Table   string
	Pattern string
	Fields  []string
}

func (sf Sqlfield) formatter() string {
	var results []string
FieldLoop:
	for _, f := range sf.Fields {
		if f == "id" {
			results = append(results, fmt.Sprintf(`IFNULL(%s.id, 0) "%s.id"`, sf.Table, sf.Table))
			continue FieldLoop
		}
		switch sf.Pattern {
		case `%s.%s "%s.%s"`:
			results = append(results, fmt.Sprintf(sf.Pattern, sf.Table, f, sf.Table, f))
		case `%s.%s`:
			results = append(results, fmt.Sprintf(sf.Pattern, sf.Table, f))
		case `%s`:
			results = append(results, fmt.Sprintf(sf.Pattern, f))
		default:
			fmt.Printf("could not parse:%s\n", sf.Pattern)
		}
	}
	return strings.Join(results, ", ")
}

// SQLO , not SOLO, stands for "SQL Object".
// It tries to mapping SQL statement to struct
type SQLO struct {

	// Table hosts the table name for select
	Table string

	// Fields hosts all the fields that will appear in SELECT statement
	Fields []Sqlfield

	// Join provide table to be joined, in string form
	Join []string

	// Where comprises SQL statements strings in 'WHERE' section
	Where []sqlsv

	// Order by maps to the ORDER BY section in SQL
	Orderby string

	// Pagination is the limit string, in this pattern: LIMIT [max_results] OFFSET [page-1]
	Pagination string

	// Args comprises all the argument corresponding to placeholders in SQL statements
	// They will be passed into sqlx functions
	Args []interface{}
}

func NewSQLO(options ...func(*SQLO)) *SQLO {
	so := SQLO{}
	for _, option := range options {
		option(&so)
	}
	return &so
}

func (s *SQLO) FormatOrderBy(orderby string) {

	tmp := strings.Split(orderby, ",")
	for i, v := range tmp {
		if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
			tmp[i] = v[1:] + " DESC"
		} else {
			tmp[i] = v
		}
	}
	s.Orderby = fmt.Sprintf(" ORDER BY %s", strings.Join(tmp, ", "))
}

func (s *SQLO) GenFields() string {

	var results []string
	for _, field := range s.Fields {
		results = append(results, field.formatter())
	}
	return strings.Join(results, ", ")
}

func (s *SQLO) Select() (query string, args []interface{}, err error) {

	base := bytes.NewBufferString(fmt.Sprintf("SELECT %s FROM %s", s.GenFields(), s.Table))
	for _, join := range s.Join {
		base.WriteString(join)
	}
	if len(s.Where) > 0 {
		base.WriteString(" WHERE")
		for i, condition := range s.Where {
			if i != 0 {
				base.WriteString(" AND")
			}
			base.WriteString(" ")
			base.WriteString(condition.statement)
			s.Args = append(s.Args, condition.variable)
		}
	}
	if s.Orderby != "" {
		base.WriteString(s.Orderby)
	}
	if s.Pagination != "" {
		base.WriteString(s.Pagination)
	}
	//base.WriteString(";")
	query, args, err = sqlx.In(base.String(), s.Args...)
	if err != nil {
		return "", nil, err
	}
	query = models.DB.Rebind(query)
	return query, args, err
}
