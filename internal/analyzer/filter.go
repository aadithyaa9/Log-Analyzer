package analyzer

import (
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/models"
)

// FilterFunc is a generic filter function type
type FilterFunc func(*models.LogEntry) bool

// Filter is a generic filter that can be applied to log entries
type Filter[T any] struct {
	predicate func(T) bool
}

// NewFilter creates a new generic filter
func NewFilter[T any](predicate func(T) bool) *Filter[T] {
	return &Filter[T]{predicate: predicate}
}

// Apply applies the filter to aaan item
func (f *Filter[T]) Apply(item T) bool {
	return f.predicate(item)
}

// LevelFilter creates a filter for a specific log level
func LevelFilter(level models.LogLevel) FilterFunc {
	return func(entry *models.LogEntry) bool {
		return entry.Level == level
	}
}

// MinLevelFilter creates a filter for minimum log level
func MinLevelFilter(minLevel models.LogLevel) FilterFunc {
	return func(entry *models.LogEntry) bool {
		return entry.Level >= minLevel
	}
}

// PatternFilter creates a filter for pattern matching
func PatternFilter(pattern string) FilterFunc {
	return func(entry *models.LogEntry) bool {
		return entry.MatchesPattern(pattern)
	}
}

// TimeRangeFilter creates a filter for time range
func TimeRangeFilter(start, end time.Time) FilterFunc {
	return func(entry *models.LogEntry) bool {
		if !start.IsZero() && entry.Timestamp.Before(start) {
			return false
		}
		if !end.IsZero() && entry.Timestamp.After(end) {
			return false
		}
		return true
	}
}

// SourceFilter creates a filter for specific source
func SourceFilter(source string) FilterFunc {
	return func(entry *models.LogEntry) bool {
		return entry.Source == source
	}
}

// CombineFilters combines multiple filters with AND logic
func CombineFilters(filters ...FilterFunc) FilterFunc {
	return func(entry *models.LogEntry) bool {
		for _, filter := range filters {
			if !filter(entry) {
				return false
			}
		}
		return true
	}
}

// ApplyFilters applies all filters to an entry
func ApplyFilters(entry *models.LogEntry, filters []FilterFunc) bool {
	for _, filter := range filters {
		if !filter(entry) {
			return false
		}
	}
	return true
}
