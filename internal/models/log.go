package models

import (
	"fmt"
	"time"
)

// LogLevel represents the severity of a log entry (enum pattern)
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
	UNKNOWN
)

// String implements the Stringer interface for LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(s string) LogLevel {
	switch s {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return UNKNOWN
	}
}

// LogEntry represents a parsed log line
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Source    string
	Raw       string
}

// String implements the Stringer interface for LogEntry
func (e LogEntry) String() string {
	return fmt.Sprintf("[%s] %s: %s (from: %s)",
		e.Timestamp.Format("2006-01-02 15:04:05"),
		e.Level,
		e.Message,
		e.Source,
	)
}

// MatchesLevel checks if the entry matches the given level
func (e *LogEntry) MatchesLevel(level LogLevel) bool {
	return e.Level == level
}

// MatchesPattern checks if the entry contains the pattern
func (e *LogEntry) MatchesPattern(pattern string) bool {
	if pattern == "" {
		return true
	}
	// Simple substring match (you can enhance with regex)
	return contains(e.Message, pattern) || contains(e.Raw, pattern)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
