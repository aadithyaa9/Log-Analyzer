package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/models"
)

// PlainTextLogParser parses plain text logs
type PlainTextLogParser struct{}

// Parse parses a plain text log line
// Expected formats:
// [2024-01-20 15:04:05] ERROR: Something went wrong
// 2024-01-20 15:04:05 INFO Something happened
func (p *PlainTextLogParser) Parse(line string, source string) (*models.LogEntry, error) {
	if strings.TrimSpace(line) == "" {
		return nil, ErrEmptyLine
	}

	// Try to extract timestamp
	timestamp, rest, err := extractTimestamp(line)
	if err != nil {
		// If no timestamp found, use current time
		timestamp = time.Now()
		rest = line
	}

	// Try to extract log level
	level, message := extractLevelAndMessage(rest)

	entry := &models.LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   strings.TrimSpace(message),
		Source:    source,
		Raw:       line,
	}

	return entry, nil
}

// CanParse checks if a line can be parsed as plain text
func (p *PlainTextLogParser) CanParse(line string) bool {
	// Plain text parser is the fallback, so it can parse anything
	return len(strings.TrimSpace(line)) > 0
}

// Name returns the parser name
func (p *PlainTextLogParser) Name() string {
	return "PlainText"
}

// extractTimestamp tries to find and parse a timestamp at the start of the line
func extractTimestamp(line string) (time.Time, string, error) {
	// Remove leading bracket if present
	line = strings.TrimSpace(line)
	hasBracket := false
	if strings.HasPrefix(line, "[") {
		hasBracket = true
		line = line[1:]
	}

	// Try to find timestamp patterns
	formats := []struct {
		format string
		length int
	}{
		{"2006-01-02 15:04:05", 19},
		{"2006/01/02 15:04:05", 19},
		{"2006-01-02T15:04:05", 19},
	}

	for _, f := range formats {
		if len(line) >= f.length {
			timestampStr := line[:f.length]
			if t, err := time.Parse(f.format, timestampStr); err == nil {
				rest := line[f.length:]
				// Remove closing bracket if present
				if hasBracket && strings.HasPrefix(rest, "]") {
					rest = rest[1:]
				}
				return t, strings.TrimSpace(rest), nil
			}
		}
	}

	return time.Time{}, line, fmt.Errorf("no timestamp found")
}

// extractLevelAndMessage extracts log level and message
func extractLevelAndMessage(line string) (models.LogLevel, string) {
	line = strings.TrimSpace(line)

	// Common patterns: "ERROR:", "ERROR -", "ERROR"
	levels := []string{"FATAL", "ERROR", "WARN", "WARNING", "INFO", "DEBUG"}

	for _, lvl := range levels {
		// Check for "LEVEL:" or "LEVEL -" or "LEVEL "
		patterns := []string{
			lvl + ":",
			lvl + " -",
			lvl + " ",
		}

		for _, pattern := range patterns {
			if strings.HasPrefix(strings.ToUpper(line), pattern) {
				message := strings.TrimSpace(line[len(pattern):])
				return models.ParseLogLevel(lvl), message
			}
		}
	}

	// If no level found, check if line contains level keyword
	for _, lvl := range levels {
		if strings.Contains(strings.ToUpper(line), lvl) {
			return models.ParseLogLevel(lvl), line
		}
	}

	// Default to INFO if no level found
	return models.INFO, line
}
