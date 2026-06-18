package util

import (
	"net"
	"strings"
)

func MatchIP(pattern, ip string) bool {
	if pattern == "" {
		return true
	}
	if pattern == ip {
		return true
	}
	if strings.Contains(pattern, "*") {
		return matchWildcard(pattern, ip)
	}
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return false
		}
		return ipNet.Contains(parsedIP)
	}
	return false
}

func matchWildcard(pattern, s string) bool {
	patternParts := strings.Split(pattern, "*")
	if len(patternParts) == 1 {
		return pattern == s
	}
	if !strings.HasPrefix(s, patternParts[0]) {
		return false
	}
	remaining := s[len(patternParts[0]):]
	for i := 1; i < len(patternParts)-1; i++ {
		idx := strings.Index(remaining, patternParts[i])
		if idx == -1 {
			return false
		}
		remaining = remaining[idx+len(patternParts[i]):]
	}
	lastPart := patternParts[len(patternParts)-1]
	if lastPart == "" {
		return true
	}
	return strings.HasSuffix(remaining, lastPart)
}
