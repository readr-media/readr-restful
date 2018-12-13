package poll

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func OperatorParser(operator string) (result string) {
	switch operator {
	case "gte":
		result = `>=`
	case "gt":
		result = `>`
	case "lte":
		result = `<=`
	case "lt":
		result = `<`
	case "neq":
		result = `!=`
	case "eq":
		result = `=`
	case "in":
		result = `IN`
	case "nin":
		result = `NOT IN`
	default:
		result = ``
	}
	return result
}

func ParamsParser(rhs string, mode string) (operators []string, values []interface{}, err error) {

	var pattern string
	switch mode {
	case "comma-separated":
		pattern = `\s*(?P<operator>gte|gt|lte|lt|eq|neq|in|nin)\s*:\s*(?P<data>[[0-9-,]+])`
	case "datetime":
		pattern = `\s*(?P<operator>gte|gt|lte|lt|eq|neq|in|nin)\s*:\s*(?P<data>[0-9-:TZ]+)`
	}
	re := regexp.MustCompile(pattern)
	cutted := re.FindAllStringSubmatch(rhs, -1)
	for _, filter := range cutted {
		operators = append(operators, filter[1])
		switch mode {
		case "comma-separated":
			var csv []int
			if err := json.Unmarshal([]byte(filter[2]), &csv); err != nil {
				return nil, nil, err
			}
			values = append(values, csv)
		case "datetime":
			t, err := time.Parse(time.RFC3339, filter[2])
			if err != nil {
				return nil, nil, err
			}
			values = append(values, t)
		}
	}
	return operators, values, nil
}

// ListPollsFilter is used to store GET filters
type ListPollsFilter struct {
	MaxResult int    `form:"max_result"`
	Page      int    `form:"page"`
	Sort      string `form:"sort"`
	Embed     string `form:"embed"`

	Status  string `form:"status"`
	Active  *int64 `form:"active,omitempty"`
	StartAt string `form:"start_at"`
	IDS     string `form:"ids"`
}

func (f *ListPollsFilter) validate() bool {
	// TODO: validator for each filters
	return true
}

// Parse could return a function to modify SQLO object with fields in ListPollsFilter
func (f *ListPollsFilter) Parse() func(s *SQLO) {
	return func(s *SQLO) {
		if f.IDS != "" {
			operator, value, err := ParamsParser(f.IDS, "comma-separated")
			if err != nil {
				log.Println(err.Error())
				return
			}
			i := 0
			for i < len(operator) {
				s.where = append(s.where, sqlsv{statement: fmt.Sprintf("polls.id %s (?)", OperatorParser(operator[i])), variable: value[i]})
				i++
			}
		}
		// active should be labeled with "polls.active", since all three tables have the column "active"
		// It could be ambiguous for MySQL for placing simple active here
		if f.Active != nil {
			s.where = append(s.where, sqlsv{statement: "polls.active = ?", variable: *f.Active})
		}
		if f.Embed != "" {
			embedded := strings.Split(f.Embed, ",")
			var embedCreator bool

			for _, field := range embedded {
				if field == "choices" {

					s.join = append(s.join, " LEFT JOIN polls_choices AS choice ON polls.id = choice.poll_id")
					s.fields = append(s.fields, sqlfield{table: "choice", pattern: `%s.%s "%s.%s"`, fields: GetStructTags("full", "db", Choice{})})
				}
				if field == "created_by" {
					s.join = append(s.join, " LEFT JOIN members AS created_by ON polls.created_by = created_by.id")
					s.fields = append(s.fields, sqlfield{table: "created_by", pattern: `%s.%s "%s.%s"`, fields: []string{"nickname"}})
					embedCreator = true
				}
			}
			// If created_by is not denoted as embedded explicitly, manually select created_by id to get valid created_by id
			if !embedCreator {
				s.join = append(s.join, " LEFT JOIN members AS created_by ON polls.created_by = created_by.id")
				s.fields = append(s.fields, sqlfield{table: "created_by", pattern: `%s.%s "%s.%s"`, fields: []string{"id"}})
			}
		}
		if f.MaxResult != 0 && f.Page > 0 {
			s.pagination = fmt.Sprintf(" LIMIT %d OFFSET %d", f.MaxResult, f.Page-1)
		}
		if f.Sort != "" {
			s.FormatOrderBy(f.Sort)
		}
		// f.Status=in:[0,1],nin:[3]
		if f.Status != "" {
			condition, value, err := ParamsParser(f.Status, "comma-separated")
			if err != nil {
				log.Printf(err.Error())
				return
			}
			i := 0
			for i < len(condition) {
				s.where = append(s.where, sqlsv{statement: fmt.Sprintf("status %s (?)", OperatorParser(condition[i])), variable: value[i]})
				i++
			}
		}
		if f.StartAt != "" {
			condition, value, err := ParamsParser(f.StartAt, "datetime")
			if err != nil {
				log.Printf(err.Error())
				return
			}
			i := 0
			for i < len(condition) {
				s.where = append(s.where, sqlsv{statement: fmt.Sprintf("start_at %s ?", OperatorParser(condition[i])), variable: value[i]})
				i++
			}
		}
	}
}

// SetListPollsFilter is the constructor for ListPollsFilter, exploited in router
// Default values are MaxResult = 20, Page = 1, OrderBy = -updated_at
func SetListPollsFilter(options ...func(*ListPollsFilter) (err error)) (*ListPollsFilter, error) {

	//var defaultActive int64 = 1
	//args := ListPollsFilter{MaxResult: 20, Page: 1, Sort: "-created_at", Active: &defaultActive}
	var args ListPollsFilter

	for _, option := range options {
		if err := option(&args); err != nil {
			return nil, err
		}
	}
	return &args, nil
}

// BindQuery is used in NewRouterFilter, so it has corresponding variable fingerprints.
// Router binds all query parameters here, including all the customized parameter forms.
func BindListPollsFilter(c *gin.Context) func(*ListPollsFilter) error {
	return func(f *ListPollsFilter) (err error) {
		if err = c.ShouldBindQuery(f); err != nil {
			fmt.Println(err.Error())
		}
		return err
	}
}

type ListPicksFilter struct {
	PollID    int64
	Member    int64  `form:"member"`
	Active    int64  `form:"active"`
	CreatedAt string `form:"created_at"`
}

func SetListPicksFilter(options ...func(*ListPicksFilter) (err error)) (*ListPicksFilter, error) {
	args := ListPicksFilter{}
	for _, option := range options {
		if err := option(&args); err != nil {
			return nil, err
		}
	}
	return &args, nil
}

func BindListPicksFilter(c *gin.Context) func(*ListPicksFilter) error {
	return func(f *ListPicksFilter) (err error) {
		if err = c.ShouldBindQuery(f); err != nil {
			fmt.Println(err.Error())
		}
		if c.Param("id") != "" {
			f.PollID, _ = strconv.ParseInt(c.Param("id"), 10, 64)
		}
		return err
	}
}

func (f *ListPicksFilter) Parse() func(s *SQLO) {
	return func(s *SQLO) {
		s.where = append(s.where, sqlsv{statement: "active = ?", variable: f.Active})

		if f.PollID != 0 {
			s.where = append(s.where, sqlsv{statement: "poll_id = ?", variable: f.PollID})
		}

		if f.Member != 0 {
			s.where = append(s.where, sqlsv{statement: "member_id = ?", variable: f.Member})
		}
		// Parse format like: created_at=gte:2018-11-22, lt:2018-11-24
		if f.CreatedAt != "" {
			cutter := regexp.MustCompile(`\s*(?P<operator>gte|gt|lte|lt|eq|neq)\s*:(?P<data>[0-9-]+)`)
			cutted := cutter.FindAllStringSubmatch(f.CreatedAt, -1)
			for _, filter := range cutted {
				s.where = append(s.where, sqlsv{statement: fmt.Sprintf("created_at %s ?", OperatorParser(filter[1])), variable: filter[2]})
			}
		}
	}
}
