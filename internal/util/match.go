package util

import (
	"path/filepath"
	"strconv"
	"strings"
)

func MatchPath(pattern, path string) bool {
	if pattern == "" {
		return true
	}
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		return pattern == path
	}
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	return matched
}

func MatchStatus(patterns []string, status int) bool {
	if len(patterns) == 0 {
		return true
	}
	statusStr := strconv.Itoa(status)
	for _, pattern := range patterns {
		if matchStatusPattern(pattern, statusStr) {
			return true
		}
	}
	return false
}

func matchStatusPattern(pattern, status string) bool {
	if len(pattern) != 3 {
		return pattern == status
	}
	for i := 0; i < 3; i++ {
		if pattern[i] == 'x' || pattern[i] == 'X' {
			continue
		}
		if i >= len(status) || pattern[i] != status[i] {
			return false
		}
	}
	return true
}

func Contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
