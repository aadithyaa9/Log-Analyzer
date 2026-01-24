package parser

import (
	"errors"
	"github.com/yourusername/loganalyzer/internal/models"
)

// Common errors
var (
	ErrInvalidFormat = errors.New("invalid log format")
	ErrEmptyLine     = errors.New("empty log line")
	ErrParseFailure  = errors.New("failed to parse log line")
)

// LogParser is an interface for parsing different log formats
type LogParser interface {
	Parse(line string, source string) (*models.LogEntry, error)
	CanParse(line string) bool
	Name() string
}

// ParserType represents different parser types (enum pattern)
type ParserType int

const (
	JSONParser ParserType = iota
	PlainTextParser
	AutoDetect
)

func (p ParserType) String() string {
	switch p {
	case JSONParser:
		return "JSON"
	case PlainTextParser:
		return "PlainText"
	case AutoDetect:
		return "AutoDetect"
	default:
		return "Unknown"
	}
}
