package analyzer

import (
	"sort"
	"sync"

	"github.com/aadithyaa9/loganalyzer/internal/models"
)

// Aggregator aggregates log entries from multiple sources
type Aggregator struct {
	mu      sync.Mutex
	entries []*models.LogEntry
	stats   *models.Statistics
}

// NewAggregator creates a new Aggregator
func NewAggregator() *Aggregator {
	return &Aggregator{
		entries: make([]*models.LogEntry, 0),
		stats:   models.NewStatistics(),
	}
}

// Add adds a log entry to the aggregator (thread-safe)
func (a *Aggregator) Add(entry *models.LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.entries = append(a.entries, entry)
	a.stats.AddEntry(entry)
}

// AddBatch adds multiple entries at once (thread-safe)
func (a *Aggregator) AddBatch(entries []*models.LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, entry := range entries {
		a.entries = append(a.entries, entry)
		a.stats.AddEntry(entry)
	}
}

// GetEntries returns all aggregated entries (creates a copy for thread-safety)
func (a *Aggregator) GetEntries() []*models.LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create a copy to avoid race conditions
	result := make([]*models.LogEntry, len(a.entries))
	copy(result, a.entries)
	return result
}

// GetStats returns the statistics
func (a *Aggregator) GetStats() *models.Statistics {
	return a.stats
}

// Count returns the number of entries
func (a *Aggregator) Count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.entries)
}

// SortByTime sorts entries by timestamp
func (a *Aggregator) SortByTime() {
	a.mu.Lock()
	defer a.mu.Unlock()

	sort.Slice(a.entries, func(i, j int) bool {
		return a.entries[i].Timestamp.Before(a.entries[j].Timestamp)
	})
}

// GetTopN returns top N entries (after optional filtering)
func (a *Aggregator) GetTopN(n int) []*models.LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	if n > len(a.entries) {
		n = len(a.entries)
	}

	result := make([]*models.LogEntry, n)
	copy(result, a.entries[:n])
	return result
}

// FilterEntries returns entries that match the filter
func (a *Aggregator) FilterEntries(filter FilterFunc) []*models.LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make([]*models.LogEntry, 0)
	for _, entry := range a.entries {
		if filter(entry) {
			result = append(result, entry)
		}
	}
	return result
}

// Clear clears all entries
func (a *Aggregator) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.entries = make([]*models.LogEntry, 0)
	a.stats = models.NewStatistics()
}
