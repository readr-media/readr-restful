package poll

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetPollsFilter is used to store GET filters
type GetPollsFilter struct {
	MaxResults int    `form:"max_results"`
	Page       int    `form:"page"`
	Sort       string `form:"sort"`
	Embed      string `form:"embed"`
}

func (f *GetPollsFilter) validate() bool {
	// TODO: validator for each filters
	return true
}

// Parse could return a function to modify SQLO object with fields in GetPollsFilter
func (f *GetPollsFilter) Parse() func(s *SQLO) {
	return func(s *SQLO) {
		if f.Embed != "" {
			embedded := strings.Split(f.Embed, ",")

			for _, field := range embedded {
				if field == "choices" {

					s.join = append(s.join, " LEFT JOIN polls_choices AS choice ON polls.id = choice.poll_id")
					s.fields = append(s.fields, sqlfield{table: "choice", pattern: `%s.%s "%s.%s"`, fields: GetStructTags("full", "db", Choice{})})
				}
			}
		}
		if f.MaxResults != 0 && f.Page > 0 {
			s.pagination = fmt.Sprintf(" LIMIT %d OFFSET %d", f.MaxResults, f.Page-1)
		}
		if f.Sort != "" {
			s.FormatOrderBy(f.Sort)
		}
	}
}

// SetGetPollsFilter is the constructor for GetPollsFilter, exploited in router
// Default values are  MaxResults = 20, Page = 1, OrderBy = -updated_at
func SetGetPollsFilter(options ...func(*GetPollsFilter) (err error)) (*GetPollsFilter, error) {

	args := GetPollsFilter{MaxResults: 20, Page: 1, Sort: "-created_at"}

	for _, option := range options {
		if err := option(&args); err != nil {
			return nil, err
		}
	}
	return &args, nil
}

// BindQuery is used in NewRouterFilter, so it has corresponding variable fingerprints.
// Router binds all query parameters here, including all the customized parameter forms.
func BindQuery(c *gin.Context) func(*GetPollsFilter) error {
	return func(f *GetPollsFilter) (err error) {
		if err = c.ShouldBindQuery(f); err != nil {
			fmt.Println(err.Error())
		}
		return err
	}
}
