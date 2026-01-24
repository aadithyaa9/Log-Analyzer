package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aadithyaa9/loganalyzer/internal/models"
	"github.com/fatih/color"
)

// TableReporter formats output as a readable table
type TableReporter struct{}

// Name returns the reporter name
func (r *TableReporter) Name() string {
	return "Table"
}

// Report generates a table report
func (r *TableReporter) Report(entries []*models.LogEntry, stats *models.Statistics, writer io.Writer) error {
	// Print statistics first
	fmt.Fprintln(writer, stats.Summary())

	// Print entries
	if len(entries) == 0 {
		fmt.Fprintln(writer, "\n✨ No entries found matching the criteria")
		return nil
	}

	fmt.Fprintln(writer, "\n📋 Log Entries")
	fmt.Fprintln(writer, strings.Repeat("─", 80))

	// Determine how many entries to show
	maxEntries := 50
	if len(entries) > maxEntries {
		fmt.Fprintf(writer, "Showing first %d of %d entries\n\n", maxEntries, len(entries))
		entries = entries[:maxEntries]
	}

	// Print each entry
	for i, entry := range entries {
		r.printEntry(writer, i+1, entry)
	}

	return nil
}

// printEntry prints a single log entry with color
func (r *TableReporter) printEntry(writer io.Writer, index int, entry *models.LogEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	// Color based on level
	var levelStr string
	switch entry.Level {
	case models.FATAL:
		levelStr = color.New(color.FgRed, color.Bold).Sprint("FATAL")
	case models.ERROR:
		levelStr = color.New(color.FgRed).Sprint("ERROR")
	case models.WARN:
		levelStr = color.New(color.FgYellow).Sprint("WARN ")
	case models.INFO:
		levelStr = color.New(color.FgGreen).Sprint("INFO ")
	case models.DEBUG:
		levelStr = color.New(color.FgCyan).Sprint("DEBUG")
	default:
		levelStr = fmt.Sprintf("%-5s", entry.Level.String())
	}

	// Truncate message if too long
	message := entry.Message
	if len(message) > 70 {
		message = message[:67] + "..."
	}

	fmt.Fprintf(writer, "%3d. [%s] %s | %s | %s\n",
		index,
		timestamp,
		levelStr,
		color.New(color.FgHiBlack).Sprintf("%-15s", truncate(entry.Source, 15)),
		message,
	)
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PrintTopErrors prints the most common error patterns
func PrintTopErrors(stats *models.Statistics, writer io.Writer, limit int) {
	fmt.Fprintln(writer, "\n🔥 Top Error Patterns")
	fmt.Fprintln(writer, strings.Repeat("─", 80))

	// Get pattern counts
	type patternCount struct {
		pattern string
		count   int
	}

	patterns := make([]patternCount, 0)
	for pattern, count := range stats.PatternCounts {
		patterns = append(patterns, patternCount{pattern, count})
	}

	// Simple bubble sort for small data
	for i := 0; i < len(patterns); i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].count > patterns[i].count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	// Print top N
	if limit > len(patterns) {
		limit = len(patterns)
	}

	for i := 0; i < limit; i++ {
		fmt.Fprintf(writer, "%2d. %s: %s\n",
			i+1,
			color.New(color.FgYellow).Sprintf("%5d occurrences", patterns[i].count),
			patterns[i].pattern,
		)
	}
}

// PrintSourceBreakdown prints breakdown by source
func PrintSourceBreakdown(stats *models.Statistics, writer io.Writer) {
	fmt.Fprintln(writer, "\n📁 Breakdown by Source")
	fmt.Fprintln(writer, strings.Repeat("─", 80))

	for source, count := range stats.SourceCounts {
		fmt.Fprintf(writer, "%-30s: %6d entries\n", source, count)
	}
}
