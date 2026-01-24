package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/models"
)

// JSONLogParser parses JSON-formatted logs
type JSONLogParser struct{}

// jsonLogEntry represents the JSON structure
type jsonLogEntry struct {
	Timestamp string `json:"timestamp"`
	Time      string `json:"time"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Msg       string `json:"msg"`
}

// Parse parses a JSON log line
func (p *JSONLogParser) Parse(line string, source string) (*models.LogEntry, error) {
	if strings.TrimSpace(line) == "" {
		return nil, ErrEmptyLine
	}

	var jsonEntry jsonLogEntry
	if err := json.Unmarshal([]byte(line), &jsonEntry); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFormat, err)
	}

	// Handle different timestamp fields
	timestampStr := jsonEntry.Timestamp
	if timestampStr == "" {
		timestampStr = jsonEntry.Time
	}

	// Parse timestamp (try multiple formats)
	timestamp, err := parseTimestamp(timestampStr)
	if err != nil {
		timestamp = time.Now() // Fallback to now
	}

	// Handle different message fields
	message := jsonEntry.Message
	if message == "" {
		message = jsonEntry.Msg
	}

	entry := &models.LogEntry{
		Timestamp: timestamp,
		Level:     models.ParseLogLevel(strings.ToUpper(jsonEntry.Level)),
		Message:   message,
		Source:    source,
		Raw:       line,
	}

	return entry, nil
}

// CanParse checks if a line is valid JSON
func (p *JSONLogParser) CanParse(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	return strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}")
}

// Name returns the parser name
func (p *JSONLogParser) Name() string {
	return "JSON"
}

// parseTimestamp tries to parse timestamp in multiple formats
func parseTimestamp(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", s)
}
