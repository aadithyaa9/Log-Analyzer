package reporter

import (
	"encoding/json"
	"io"
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/models"
)

// JSONReporter formats output as JSON
type JSONReporter struct{}

// Name returns the reporter name
func (r *JSONReporter) Name() string {
	return "JSON"
}

// Report generates a JSON report
func (r *JSONReporter) Report(entries []*models.LogEntry, stats *models.Statistics, writer io.Writer) error {
	report := r.buildReport(entries, stats)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	return encoder.Encode(report)
}

// JSONReport represents the JSON structure
type JSONReport struct {
	Summary    Summary     `json:"summary"`
	Statistics StatsJSON   `json:"statistics"`
	Entries    []EntryJSON `json:"entries"`
}

// Summary holds summary information
type Summary struct {
	TotalEntries   int     `json:"total_entries"`
	FilesProcessed int     `json:"files_processed"`
	BytesProcessed int64   `json:"bytes_processed"`
	ProcessingTime string  `json:"processing_time"`
	EntriesPerSec  float64 `json:"entries_per_second"`
}

// StatsJSON holds statistics in JSON format
type StatsJSON struct {
	LevelCounts  map[string]int `json:"level_counts"`
	SourceCounts map[string]int `json:"source_counts"`
	TimeRange    TimeRange      `json:"time_range"`
}

// TimeRange holds time range information
type TimeRange struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Duration string `json:"duration"`
}

// EntryJSON represents a log entry in JSON format
type EntryJSON struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}

// buildReport builds the JSON report structure
func (r *JSONReporter) buildReport(entries []*models.LogEntry, stats *models.Statistics) *JSONReport {
	// Build summary
	summary := Summary{
		TotalEntries:   stats.TotalEntries,
		FilesProcessed: stats.FilesProcessed,
		BytesProcessed: stats.BytesProcessed,
		ProcessingTime: stats.ProcessingTime.String(),
	}

	if stats.ProcessingTime.Seconds() > 0 {
		summary.EntriesPerSec = float64(stats.TotalEntries) / stats.ProcessingTime.Seconds()
	}

	// Build statistics
	levelCounts := make(map[string]int)
	for level, count := range stats.LevelCounts {
		levelCounts[level.String()] = count
	}

	statistics := StatsJSON{
		LevelCounts:  levelCounts,
		SourceCounts: stats.SourceCounts,
		TimeRange: TimeRange{
			Start:    formatTime(stats.FirstTimestamp),
			End:      formatTime(stats.LastTimestamp),
			Duration: stats.LastTimestamp.Sub(stats.FirstTimestamp).String(),
		},
	}

	// Build entries
	jsonEntries := make([]EntryJSON, len(entries))
	for i, entry := range entries {
		jsonEntries[i] = EntryJSON{
			Timestamp: entry.Timestamp.Format(time.RFC3339),
			Level:     entry.Level.String(),
			Message:   entry.Message,
			Source:    entry.Source,
		}
	}

	return &JSONReport{
		Summary:    summary,
		Statistics: statistics,
		Entries:    jsonEntries,
	}
}

// formatTime formats a time or returns empty string if zero
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
