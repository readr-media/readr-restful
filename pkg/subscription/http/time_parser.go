package http

import (
	"fmt"
	"time"
)

func payInterval(t time.Time) (start time.Time, end time.Time, err error) {
	truncatedT := t.UTC().Truncate(24 * time.Hour)
	year, month, day := truncatedT.Date()
	fmt.Println(year, month, day)
	switch month {
	case time.May, time.July, time.October, time.December:
		if day == 30 {
			return time.Time{}, time.Time{}, nil
		} else if day == 31 {
			return truncatedT.AddDate(0, -1, -1), truncatedT.AddDate(0, -1, 0), nil
		}
	case time.April, time.June, time.September, time.November:
		if day == 30 {
			return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 2), nil
		}
	case time.February:
		if day == 28 {
			return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 4), nil
		}
	case time.March:
		if day > 28 && day < 31 {
			return time.Time{}, time.Time{}, nil
		} else if day == 31 {
			return truncatedT.AddDate(0, -1, -3), truncatedT.AddDate(0, -1, -2), nil
		}
	default:
		return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 1), nil
	}
	return truncatedT.AddDate(0, -1, 0), truncatedT.AddDate(0, -1, 1), nil
}
