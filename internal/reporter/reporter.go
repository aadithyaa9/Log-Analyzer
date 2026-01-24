package reporter

import (
	"io"

	"github.com/aadithyaa9/loganalyzer/internal/models"
)

// OutputFormat represents different output formats (enum pattern)
type OutputFormat int

const (
	TableFormat OutputFormat = iota
	JSONFormat
	CSVFormat
)

func (o OutputFormat) String() string {
	switch o {
	case TableFormat:
		return "table"
	case JSONFormat:
		return "json"
	case CSVFormat:
		return "csv"
	default:
		return "unknown"
	}
}

// Reporter is an interface for different output formatters
type Reporter interface {
	Report(entries []*models.LogEntry, stats *models.Statistics, writer io.Writer) error
	Name() string
}

// GetReporter returns a reporter based on format
func GetReporter(format OutputFormat) Reporter {
	switch format {
	case JSONFormat:
		return &JSONReporter{}
	case TableFormat:
		return &TableReporter{}
	default:
		return &TableReporter{}
	}
}
