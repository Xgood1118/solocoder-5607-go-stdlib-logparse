package parser

import (
	"logparse/internal/util"
)

type Parser interface {
	Parse(line string) (*util.LogEntry, error)
}

func NewCombinedParser() Parser {
	return &combinedParser{}
}

func NewErrorParser() Parser {
	return &errorParser{}
}

func NewJSONParser() Parser {
	return &jsonParser{}
}

func NewRegexParser(pattern string) (Parser, error) {
	return newRegexParser(pattern)
}
