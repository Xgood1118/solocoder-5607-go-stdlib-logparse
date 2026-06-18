package parser

import (
	"logparse/internal/util"
	"strings"
	"sync"
)

type autoParser struct {
	detected Parser
	mu       sync.Mutex
	combined *combinedParser
	errorp   *errorParser
	jsonp    *jsonParser
}

func (p *autoParser) Parse(line string) (*util.LogEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.detected != nil {
		return p.detected.Parse(line)
	}

	if p.combined == nil {
		p.combined = &combinedParser{}
		p.errorp = &errorParser{}
		p.jsonp = &jsonParser{}
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, nil
	}

	if strings.HasPrefix(trimmed, "{") {
		if entry, err := p.jsonp.Parse(line); err == nil {
			p.detected = p.jsonp
			return entry, nil
		}
	}

	if entry, err := p.combined.Parse(line); err == nil && entry.IP != "" {
		p.detected = p.combined
		return entry, nil
	}

	if entry, err := p.errorp.Parse(line); err == nil {
		p.detected = p.errorp
		return entry, nil
	}

	if entry, err := p.jsonp.Parse(line); err == nil {
		p.detected = p.jsonp
		return entry, nil
	}

	return nil, nil
}
