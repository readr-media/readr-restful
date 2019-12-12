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
		{"random-normal-day", time.Date(2019, 12, 10, 2, 3, 4, 5, time.UTC), time.Date(2019, 11, 10, 0, 0, 0, 0, time.UTC), time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)},
		{"end-of-big-month", time.Date(2019, 5, 31, 3, 4, 5, 6, time.UTC), time.Date(2019, 4, 30, 0, 0, 0, 0, time.UTC), time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC)},
		{"last-two-day-from-big-end", time.Date(2019, 5, 30, 1, 2, 2, 3, time.UTC), time.Time{}, time.Time{}},
		{"end-of-small-month", time.Date(2019, 4, 30, 1, 2, 3, 4, time.UTC), time.Date(2019, 3, 30, 0, 0, 0, 0, time.UTC), time.Date(2019, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"last-two-day-from-small-end", time.Date(2019, 4, 29, 1, 2, 3, 4, time.UTC), time.Date(2019, 3, 29, 0, 0, 0, 0, time.UTC), time.Date(2019, 3, 30, 0, 0, 0, 0, time.UTC)},
		{"cross-year", time.Date(2019, 1, 31, 1, 2, 3, 4, time.UTC), time.Date(2018, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"end-of-Feb", time.Date(2019, 2, 28, 1, 2, 3, 0, time.UTC), time.Date(2019, 1, 28, 0, 0, 0, 0, time.UTC), time.Date(2019, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"last-days-in-Mar", time.Date(2019, 3, 30, 1, 2, 3, 4, time.UTC), time.Time{}, time.Time{}},
		{"last-day-of-Mar", time.Date(2019, 3, 31, 1, 2, 3, 4, time.UTC), time.Date(2019, 2, 28, 0, 0, 0, 0, time.UTC), time.Date(2019, 3, 1, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end, _ := payInterval(tc.input)
			assert.Equal(t, tc.expStart, start, "start date should be the same")
			assert.Equal(t, tc.expEnd, end, "end date should be the same")
		})
	}
}
