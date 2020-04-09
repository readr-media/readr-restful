package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPayInterval(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    time.Time
		expStart time.Time
		expEnd   time.Time
	}{
		{"random-normal-day", time.Date(2019, 12, 10, 2, 3, 4, 5, time.UTC), getBeginTimeOf(2019, 11, 10), getBeginTimeOf(2019, 11, 11)},
		{"end-of-big-month", time.Date(2019, 5, 31, 3, 4, 5, 6, time.UTC), time.Time{}, time.Time{}},
		{"end-of-small-month", time.Date(2019, 4, 30, 1, 2, 3, 4, time.UTC), getBeginTimeOf(2019, 3, 30), getBeginTimeOf(2019, 4, 1)},
		{"end-of-feb", time.Date(2019, 2, 28, 1, 2, 3, 4, time.UTC), getBeginTimeOf(2019, 1, 28), getBeginTimeOf(2019, 2, 1)},
		{"cross-year", time.Date(2019, 1, 31, 1, 2, 3, 4, time.UTC), getBeginTimeOf(2018, 12, 31), getBeginTimeOf(2019, 1, 1)},
		{"last-two-day-from-small-month", time.Date(2019, 4, 29, 1, 2, 3, 4, time.UTC), getBeginTimeOf(2019, 3, 29), getBeginTimeOf(2019, 3, 30)},
		{"leap-year-29th-in-Mar", time.Date(2020, 3, 29, 1, 2, 3, 4, time.UTC), getBeginTimeOf(2020, 2, 29), getBeginTimeOf(2020, 3, 1)},
		{"non-leap-year-29th-in-Mar", time.Date(2019, 3, 29, 1, 2, 3, 4, time.UTC), time.Time{}, time.Time{}},
		{"non-leap-year-28th-in-Mar", time.Date(2019, 3, 28, 1, 2, 3, 4, time.UTC), getBeginTimeOf(2019, 2, 28), getBeginTimeOf(2019, 3, 1)},
		{"30th-in-Mar", time.Date(2019, 3, 30, 1, 2, 3, 4, time.UTC), time.Time{}, time.Time{}},
		{"last-day-of-Mar", time.Date(2019, 3, 31, 1, 2, 3, 4, time.UTC), time.Time{}, time.Time{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end, _ := payInterval(tc.input)
			assert.Equal(t, tc.expStart, start, "start date should be the same")
			assert.Equal(t, tc.expEnd, end, "end date should be the same")
		})
	}
}

func getBeginTimeOf(year int, month time.Month, day int) (t time.Time) {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
