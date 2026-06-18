package parser

import (
	"fmt"
	"logparse/internal/util"
	"regexp"
	"strings"
	"time"
)

type errorParser struct{}

var errorHeaderRegex = regexp.MustCompile(
	`^(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+\d+#\d+:\s+\*?\d*\s*(.*)$`,
)

func (p *errorParser) Parse(line string) (*util.LogEntry, error) {
	matches := errorHeaderRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid error format")
	}

	t, err := time.Parse("2006/01/02 15:04:05", matches[1])
	if err != nil {
		t = time.Time{}
	}

	entry := &util.LogEntry{
		Time:    t,
		Level:   strings.ToUpper(matches[2]),
		Raw:     line,
	}

	rest := matches[3]
	entry.Message = rest

	if idx := strings.Index(rest, ", client: "); idx != -1 {
		after := rest[idx+len(", client: "):]
		if commaIdx := strings.Index(after, ", "); commaIdx != -1 {
			entry.IP = after[:commaIdx]
		} else {
			entry.IP = after
		}
		if idx > 0 {
			entry.Message = rest[:idx]
		}
	}

	if idx := strings.Index(rest, ", request: \""); idx != -1 {
		after := rest[idx+len(", request: \""):]
		if endIdx := strings.Index(after, "\""); endIdx != -1 {
			reqStr := after[:endIdx]
			reqParts := strings.Fields(reqStr)
			if len(reqParts) >= 1 {
				entry.Method = reqParts[0]
			}
			if len(reqParts) >= 2 {
				entry.Path = reqParts[1]
			}
			if len(reqParts) >= 3 {
				entry.Protocol = reqParts[2]
			}
		}
	}

	if entry.Level == "ERROR" || entry.Level == "CRIT" || entry.Level == "ALERT" || entry.Level == "EMERG" {
		entry.Status = 500
	} else if entry.Level == "WARN" {
		entry.Status = 400
	}

	return entry, nil
}
