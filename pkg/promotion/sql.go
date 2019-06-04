package promotion

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

type sqlfield struct {
	table   string
	pattern string
	fields  []string
}

func (sf sqlfield) formatter() string {
	var results []string
FieldLoop:
	for _, f := range sf.fields {
		if f == "id" {
			results = append(results, fmt.Sprintf(`IFNULL(%s.id, 0) "%s.id"`, sf.table, sf.table))
			continue FieldLoop
		}
		switch sf.pattern {
		case `%s.%s "%s.%s"`:
			results = append(results, fmt.Sprintf(sf.pattern, sf.table, f, sf.table, f))
		case `%s.%s`:
			results = append(results, fmt.Sprintf(sf.pattern, sf.table, f))
		case `%s`:
			results = append(results, fmt.Sprintf(sf.pattern, f))
		default:
			fmt.Printf("could not parse:%s\n", sf.pattern)
		}
	}
	return strings.Join(results, ", ")
}

// SQLO , not SOLO, stands for "SQL Object".
// It tries to mapping SQL statement to struct
type SQLO struct {

	// table hosts the table name for select
	table string

	// fields hosts all the fields that will appear in SELECT statement
	fields []sqlfield

	// join provide table to be joined, in string form
	join []string

	// where comprises SQL statements strings in 'WHERE' section
	where []sqlsv

	// order by maps to the ORDER BY section in SQL
	orderby string

	// pagination is the limit string, in this pattern: LIMIT [max_results] OFFSET [page-1]
	pagination string

	// args comprises all the argument corresponding to placeholders in SQL statements
	// They will be passed into sqlx functions
	args []interface{}
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
	s.orderby = fmt.Sprintf(" ORDER BY %s", strings.Join(tmp, ", "))
}

func (s *SQLO) GenFields() string {

	var results []string
	for _, field := range s.fields {
		results = append(results, field.formatter())
	}
	return strings.Join(results, ", ")
}

func (s *SQLO) SQL() (query string, args []interface{}, err error) {

	base := bytes.NewBufferString(fmt.Sprintf("SELECT %s FROM %s", s.GenFields(), s.table))
	for _, join := range s.join {
		base.WriteString(join)
	}
	if len(s.where) > 0 {
		base.WriteString(" WHERE")
		for i, condition := range s.where {
			if i != 0 {
				base.WriteString(" AND")
			}
			base.WriteString(" ")
			base.WriteString(condition.statement)
			s.args = append(s.args, condition.variable)
		}
	}
	if s.orderby != "" {
		base.WriteString(s.orderby)
	}
	if s.pagination != "" {
		base.WriteString(s.pagination)
	}
	//base.WriteString(";")
	query, args, err = sqlx.In(base.String(), s.args...)
	if err != nil {
		return "", nil, err
	}
	query = models.DB.Rebind(query)
	return query, args, err
}
