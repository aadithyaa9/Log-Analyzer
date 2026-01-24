package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/models"
	"github.com/aadithyaa9/loganalyzer/internal/parser"
)

// Config holds analyzer configuration
type Config struct {
	Workers    int
	Level      models.LogLevel
	Pattern    string
	StartTime  time.Time
	EndTime    time.Time
	AutoDetect bool
	ParserType parser.ParserType
}

// Analyzer processes log files concurrently
type Analyzer struct {
	config     *Config
	aggregator *Aggregator
	parser     parser.LogParser
}

// NewAnalyzer creates a new Analyzer
func NewAnalyzer(config *Config) *Analyzer {
	// Get parser based on config
	var p parser.LogParser
	if config.AutoDetect {
		p = nil // Will auto-detect per line
	} else {
		p = parser.GetParser(config.ParserType)
	}

	return &Analyzer{
		config:     config,
		aggregator: NewAggregator(),
		parser:     p,
	}
}

// AnalyzeFile analyzes a single log file
func (a *Analyzer) AnalyzeFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get file info for stats
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineCount := 0
	entries := make([]*models.LogEntry, 0, 1000)

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		if line == "" {
			continue
		}

		// Auto-detect parser if needed
		currentParser := a.parser
		if currentParser == nil {
			currentParser = parser.DetectParser(line)
		}

		// Parse the line
		entry, err := currentParser.Parse(line, filepath.Base(filePath))
		if err != nil {
			// Skip invalid lines
			continue
		}

		// Apply filters
		if !a.shouldInclude(entry) {
			continue
		}

		entries = append(entries, entry)

		// Batch insert to reduce lock contention
		if len(entries) >= 1000 {
			a.aggregator.AddBatch(entries)
			entries = make([]*models.LogEntry, 0, 1000)
		}
	}

	// Add remaining entries
	if len(entries) > 0 {
		a.aggregator.AddBatch(entries)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Update stats
	a.aggregator.GetStats().AddFile(fileInfo.Size())

	return nil
}

// AnalyzeDirectory analyzes all log files in a directory concurrently
func (a *Analyzer) AnalyzeDirectory(dirPath string) error {
	startTime := time.Now()

	// Find all log files
	files, err := findLogFiles(dirPath)
	if err != nil {
		return fmt.Errorf("failed to find log files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no log files found in %s", dirPath)
	}

	// Create worker pool
	numWorkers := a.config.Workers
	if numWorkers <= 0 {
		numWorkers = 4 // Default
	}

	// Channels for work distribution
	fileChan := make(chan string, len(files))
	errChan := make(chan error, len(files))

	// Send files to channel
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for file := range fileChan {
				if err := a.AnalyzeFile(file); err != nil {
					errChan <- fmt.Errorf("worker %d: %w", workerID, err)
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Set processing time
	a.aggregator.GetStats().SetProcessingTime(time.Since(startTime))

	// Sort entries by time
	a.aggregator.SortByTime()

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during processing", len(errors))
	}

	return nil
}

// GetResults returns the aggregated results
func (a *Analyzer) GetResults() *Aggregator {
	return a.aggregator
}

// shouldInclude checks if an entry should be included based on filters
func (a *Analyzer) shouldInclude(entry *models.LogEntry) bool {
	// Level filter
	if a.config.Level != models.UNKNOWN && entry.Level < a.config.Level {
		return false
	}

	// Pattern filter
	if a.config.Pattern != "" && !entry.MatchesPattern(a.config.Pattern) {
		return false
	}

	// Time range filter
	if !a.config.StartTime.IsZero() && entry.Timestamp.Before(a.config.StartTime) {
		return false
	}
	if !a.config.EndTime.IsZero() && entry.Timestamp.After(a.config.EndTime) {
		return false
	}

	return true
}

// findLogFiles recursively finds all .log files in a directory
func findLogFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".log" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
