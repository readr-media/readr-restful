package http

import (
	"time"
)

func payInterval(current time.Time) (start time.Time, end time.Time, err error) {
	truncatedCurrent := current.UTC().Truncate(24 * time.Hour)
	year, month, day := truncatedCurrent.Date()

	switch month {
	case time.May, time.July, time.October, time.December:
		if day == 31 {
			return time.Time{}, time.Time{}, nil
		}
	case time.April, time.June, time.September, time.November:
		if day == 30 {
			return truncatedCurrent.AddDate(0, -1, 0), truncatedCurrent.AddDate(0, -1, 2), nil
		}
	case time.February:
		if day == 28 {
			return truncatedCurrent.AddDate(0, -1, 0), truncatedCurrent.AddDate(0, -1, 4), nil
		}
	case time.March:
		if day > 29 || (day == 29 && !isLeapYear(year)) {
			return time.Time{}, time.Time{}, nil
		}
	default:
		return truncatedCurrent.AddDate(0, -1, 0), truncatedCurrent.AddDate(0, -1, 1), nil
	}
	return truncatedCurrent.AddDate(0, -1, 0), truncatedCurrent.AddDate(0, -1, 1), nil
}

func isLeapYear(year int) bool {
	switch {
	case year%400 == 0:
		return true
	case year%100 == 0:
		return false
	case year%4 == 0:
		return true
	default:
		return false
	}
}
