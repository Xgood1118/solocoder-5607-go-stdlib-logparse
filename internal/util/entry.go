package util

import "time"

type LogEntry struct {
	IP         string
	Time       time.Time
	Method     string
	Path       string
	Protocol   string
	Status     int
	Bytes      int64
	Referer    string
	UserAgent  string
	Raw        string
	RespTime   float64
	Level      string
	Message    string
}

func IsErrorStatus(status int) bool {
	return status >= 400
}
