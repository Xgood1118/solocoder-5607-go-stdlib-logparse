package parser

import (
	"logparse/internal/util"
	"strings"
	"sync"
)

type autoParser struct {
	mu            sync.Mutex
	combined      *combinedParser
	errorp        *errorParser
	jsonp         *jsonParser
	consecutiveFails int
	maxFails      int
}

func NewAutoParser() Parser {
	return &autoParser{
		combined:  &combinedParser{},
		errorp:   &errorParser{},
		jsonp:    &jsonParser{},
		maxFails: 5,
	}
}

func (p *autoParser) Parse(line string) (*util.LogEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, nil
	}

	if entry, err := p.jsonp.Parse(line); err == nil && entry != nil {
		p.consecutiveFails = 0
		return entry, nil
	}

	if entry, err := p.combined.Parse(line); err == nil && entry != nil {
		p.consecutiveFails = 0
		return entry, nil
	}

	if entry, err := p.errorp.Parse(line); err == nil && entry != nil {
		p.consecutiveFails = 0
		return entry, nil
	}

	p.consecutiveFails++
	if p.consecutiveFails >= p.maxFails {
		p.consecutiveFails = 0
	}

	return nil, nil
}
