package util

import (
	"time"
)

var timeFormats = []string{
	"02/Jan/2006:15:04:05 -0700",
	"02/Jan/2006:15:04:05 MST",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"2006-01-02",
	"02/Jan/2006 15:04:05",
}

func ParseTime(s string) (time.Time, error) {
	for _, format := range timeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Layout: "multiple formats", Value: s}
}

func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}
	return ParseTime(s)
}

func TruncateToHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
