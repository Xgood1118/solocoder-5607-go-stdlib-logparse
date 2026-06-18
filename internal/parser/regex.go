package parser

import (
	"fmt"
	"logparse/internal/util"
	"regexp"
	"strconv"
)

type regexParser struct {
	re     *regexp.Regexp
	fields map[string]int
}

func newRegexParser(pattern string) (Parser, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %v", err)
	}

	fields := make(map[string]int)
	names := re.SubexpNames()
	for i, name := range names {
		if name != "" {
			fields[name] = i
		}
	}

	return &regexParser{re: re, fields: fields}, nil
}

func (p *regexParser) Parse(line string) (*util.LogEntry, error) {
	matches := p.re.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("no match")
	}

	entry := &util.LogEntry{Raw: line}

	getField := func(name string) string {
		if idx, ok := p.fields[name]; ok && idx < len(matches) {
			return matches[idx]
		}
		return ""
	}

	if ip := getField("ip"); ip != "" {
		entry.IP = ip
	}
	if ts := getField("time"); ts != "" {
		if t, err := util.ParseTime(ts); err == nil {
			entry.Time = t
		}
	}
	if method := getField("method"); method != "" {
		entry.Method = method
	}
	if path := getField("path"); path != "" {
		entry.Path = path
	} else if uri := getField("uri"); uri != "" {
		entry.Path = uri
	}
	if statusStr := getField("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			entry.Status = status
		}
	}
	if bytesStr := getField("bytes"); bytesStr != "" {
		if b, err := strconv.ParseInt(bytesStr, 10, 64); err == nil {
			entry.Bytes = b
		}
	}
	if referer := getField("referer"); referer != "" {
		entry.Referer = referer
	}
	if ua := getField("user_agent"); ua != "" {
		entry.UserAgent = ua
	} else if ua := getField("ua"); ua != "" {
		entry.UserAgent = ua
	}
	if rt := getField("resp_time"); rt != "" {
		if f, err := strconv.ParseFloat(rt, 64); err == nil {
			entry.RespTime = f
		}
	}
	if level := getField("level"); level != "" {
		entry.Level = level
	}
	if msg := getField("message"); msg != "" {
		entry.Message = msg
	} else if msg := getField("msg"); msg != "" {
		entry.Message = msg
	}

	return entry, nil
}
