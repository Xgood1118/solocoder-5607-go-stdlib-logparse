package parser

import (
	"fmt"
	"logparse/internal/util"
	"regexp"
	"strconv"
	"time"
)

type combinedParser struct{}

var combinedRegex = regexp.MustCompile(
	`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)"(?:\s+([\d.]+))?$`,
)

func (p *combinedParser) Parse(line string) (*util.LogEntry, error) {
	matches := combinedRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid combined format")
	}

	status, _ := strconv.Atoi(matches[6])

	bytes := int64(0)
	if matches[7] != "-" {
		if b, err := strconv.ParseInt(matches[7], 10, 64); err == nil {
			bytes = b
		}
	}

	var respTime float64
	if len(matches) > 10 && matches[10] != "" {
		respTime, _ = strconv.ParseFloat(matches[10], 64)
	}

	t, err := util.ParseTime(matches[2])
	if err != nil {
		t = time.Time{}
	}

	return &util.LogEntry{
		IP:        matches[1],
		Time:      t,
		Method:    matches[3],
		Path:      matches[4],
		Protocol:  matches[5],
		Status:    status,
		Bytes:     bytes,
		Referer:   matches[8],
		UserAgent: matches[9],
		Raw:       line,
		RespTime:  respTime,
	}, nil
}
