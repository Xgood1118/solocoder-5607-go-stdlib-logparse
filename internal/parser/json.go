package parser

import (
	"encoding/json"
	"fmt"
	"logparse/internal/util"
	"strconv"
	"time"
)

type jsonParser struct{}

type jsonLogEntry struct {
	IP        string      `json:"ip"`
	Time      string      `json:"time"`
	Timestamp interface{} `json:"timestamp"`
	Method    string      `json:"method"`
	Path      string      `json:"path"`
	URI       string      `json:"uri"`
	Status    interface{} `json:"status"`
	Bytes     interface{} `json:"bytes"`
	Size      interface{} `json:"size"`
	Referer   string      `json:"referer"`
	UserAgent string      `json:"user_agent"`
	UserAgent2 string     `json:"userAgent"`
	RespTime  interface{} `json:"resp_time"`
	Dur       interface{} `json:"duration"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Msg       string      `json:"msg"`
	Request   string      `json:"request"`
	RemoteAddr string     `json:"remote_addr"`
	HTTPHost  string      `json:"http_host"`
}

func (p *jsonParser) Parse(line string) (*util.LogEntry, error) {
	var j jsonLogEntry
	if err := json.Unmarshal([]byte(line), &j); err != nil {
		return nil, fmt.Errorf("invalid json: %v", err)
	}

	entry := &util.LogEntry{
		Raw: line,
	}

	if j.IP != "" {
		entry.IP = j.IP
	} else if j.RemoteAddr != "" {
		entry.IP = j.RemoteAddr
	}

	if j.Time != "" {
		if t, err := util.ParseTime(j.Time); err == nil {
			entry.Time = t
		}
	} else if j.Timestamp != nil {
		if ts, ok := j.Timestamp.(float64); ok {
			entry.Time = time.Unix(int64(ts), 0)
		} else if ts, ok := j.Timestamp.(string); ok {
			if t, err := util.ParseTime(ts); err == nil {
				entry.Time = t
			}
		}
	}

	if j.Method != "" {
		entry.Method = j.Method
	}
	if j.Path != "" {
		entry.Path = j.Path
	} else if j.URI != "" {
		entry.Path = j.URI
	}
	if j.Request != "" {
		entry.Path = j.Request
	}

	if j.Status != nil {
		entry.Status = int(toFloat(j.Status))
	}
	if j.Bytes != nil {
		entry.Bytes = int64(toFloat(j.Bytes))
	} else if j.Size != nil {
		entry.Bytes = int64(toFloat(j.Size))
	}

	if j.Referer != "" {
		entry.Referer = j.Referer
	}
	if j.UserAgent != "" {
		entry.UserAgent = j.UserAgent
	} else if j.UserAgent2 != "" {
		entry.UserAgent = j.UserAgent2
	}

	if j.RespTime != nil {
		entry.RespTime = toFloat(j.RespTime)
	} else if j.Dur != nil {
		entry.RespTime = toFloat(j.Dur)
	}

	if j.Level != "" {
		entry.Level = j.Level
	}
	if j.Message != "" {
		entry.Message = j.Message
	} else if j.Msg != "" {
		entry.Message = j.Msg
	}

	return entry, nil
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}
