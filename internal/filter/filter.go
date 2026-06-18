package filter

import (
	"logparse/internal/util"
	"time"
)

type Filter struct {
	From        time.Time
	To          time.Time
	StatusCodes []string
	PathPattern string
	IPPattern   string
	Method      string
	UserAgent   string
}

func (f *Filter) Match(entry *util.LogEntry) bool {
	if !f.From.IsZero() && entry.Time.Before(f.From) {
		return false
	}
	if !f.To.IsZero() && entry.Time.After(f.To) {
		return false
	}
	if len(f.StatusCodes) > 0 && entry.Status > 0 {
		if !util.MatchStatus(f.StatusCodes, entry.Status) {
			return false
		}
	}
	if f.PathPattern != "" && entry.Path != "" {
		if !util.MatchPath(f.PathPattern, entry.Path) {
			return false
		}
	}
	if f.IPPattern != "" && entry.IP != "" {
		if !util.MatchIP(f.IPPattern, entry.IP) {
			return false
		}
	}
	if f.Method != "" && entry.Method != "" {
		if f.Method != entry.Method {
			return false
		}
	}
	if f.UserAgent != "" && entry.UserAgent != "" {
		if !util.Contains(entry.UserAgent, f.UserAgent) {
			return false
		}
	}
	return true
}
