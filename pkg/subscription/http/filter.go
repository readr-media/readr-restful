package http

import (
	"fmt"
	"strings"
	"time"

	"github.com/readr-media/readr-restful/internal/rrsql"
)

// ListRequest holds the query parameters for list request
type ListRequest struct {
	Status     int                  `form:"status"`
	LastPaidAt map[string]time.Time `form:"last_paid_at"`
}

// Select generates the query statement and arguments for MySQL SELECT
func (r *ListRequest) Select() (query string, values []interface{}, err error) {
	fields := []string{"subscriptions.*"}

	// ws = where string, wv = where values
	ws := make([]string, 0)
	wv := make([]interface{}, 0)
	var where string
	// r.Status
	ws = append(ws, fmt.Sprintf("%s %s (?)", "subscriptions.status", "="))
	wv = append(wv, r.Status)
	// r.LastPaidAt
	if r.LastPaidAt != nil {
		fmt.Printf("LastPayAt:%v\n", r.LastPaidAt)
		for o, v := range r.LastPaidAt {
			ops, err := rrsql.OperatorCoverter(o)
			if err != nil {
				return "", nil, err
			}
			ws = append(ws, fmt.Sprintf("%s %s (?)", "subscriptions.last_paid_at", ops))
			wv = append(wv, v)
		}
	}
	if len(ws) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(ws, " AND "))
	} else if len(ws) == 0 {
		where = ""
	}
	query = fmt.Sprintf(`
	SELECT %s FROM subscriptions %s `,
		strings.Join(fields, ","),
		where,
	)
	values = append(values, wv...)
	return query, values, nil
}

// NewListRequest accepts option to modify the initial content of ListRequest, and return a pointer of ListRequest
func NewListRequest(options ...func(*ListRequest)) *ListRequest {
	var params ListRequest

	for _, f := range options {
		f(&params)
	}
	return &params

}
