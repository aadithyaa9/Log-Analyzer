package models

import (
	"fmt"
	"sync"
	"time"
)

// Statistics holds aggregated log analysis data
type Statistics struct {
	mu             sync.Mutex // Protects concurrent access
	TotalEntries   int
	LevelCounts    map[LogLevel]int
	PatternCounts  map[string]int
	SourceCounts   map[string]int
	FirstTimestamp time.Time
	LastTimestamp  time.Time
	ProcessingTime time.Duration
	FilesProcessed int
	BytesProcessed int64
}

// NewStatistics creates a new Statistics instance
func NewStatistics() *Statistics {
	return &Statistics{
		LevelCounts:   make(map[LogLevel]int),
		PatternCounts: make(map[string]int),
		SourceCounts:  make(map[string]int),
	}
}

// AddEntry updates statistics with a new log entry (thread-safe)
func (s *Statistics) AddEntry(entry *LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalEntries++
	s.LevelCounts[entry.Level]++
	s.SourceCounts[entry.Source]++

	// Track time range
	if s.FirstTimestamp.IsZero() || entry.Timestamp.Before(s.FirstTimestamp) {
		s.FirstTimestamp = entry.Timestamp
	}
	if entry.Timestamp.After(s.LastTimestamp) {
		s.LastTimestamp = entry.Timestamp
	}
}

// IncrementPattern increments the count for a specific pattern (thread-safe)
func (s *Statistics) IncrementPattern(pattern string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PatternCounts[pattern]++
}

// AddFile updates file processing stats (thread-safe)
func (s *Statistics) AddFile(bytesRead int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FilesProcessed++
	s.BytesProcessed += bytesRead
}

// SetProcessingTime sets the total processing time
func (s *Statistics) SetProcessingTime(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ProcessingTime = duration
}

// GetLevelCount returns the count for a specific level (thread-safe)
func (s *Statistics) GetLevelCount(level LogLevel) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.LevelCounts[level]
}

// Summary returns a formatted summary string
func (s *Statistics) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return fmt.Sprintf(`
╔══════════════════════════════════════════════════════════╗
║  Log Analysis Summary                                    ║
╚══════════════════════════════════════════════════════════╝

📊 Processing Stats
────────────────────────────────────────────────
Files Processed:    %d
Total Entries:      %d
Bytes Processed:    %.2f MB
Processing Time:    %s
Entries/sec:        %.0f

📈 Log Levels
────────────────────────────────────────────────
DEBUG:   %6d
INFO:    %6d
WARN:    %6d
ERROR:   %6d
FATAL:   %6d

⏰ Time Range
────────────────────────────────────────────────
First Entry:        %s
Last Entry:         %s
Duration:           %s
`,
		s.FilesProcessed,
		s.TotalEntries,
		float64(s.BytesProcessed)/(1024*1024),
		s.ProcessingTime.Round(time.Millisecond),
		float64(s.TotalEntries)/s.ProcessingTime.Seconds(),
		s.LevelCounts[DEBUG],
		s.LevelCounts[INFO],
		s.LevelCounts[WARN],
		s.LevelCounts[ERROR],
		s.LevelCounts[FATAL],
		s.FirstTimestamp.Format("2006-01-02 15:04:05"),
		s.LastTimestamp.Format("2006-01-02 15:04:05"),
		s.LastTimestamp.Sub(s.FirstTimestamp).Round(time.Second),
	)
}
