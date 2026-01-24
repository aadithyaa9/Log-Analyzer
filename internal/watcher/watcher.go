package watcher

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/models"
	"github.com/aadithyaa9/loganalyzer/internal/parser"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

// Config holds watcher configuration
type Config struct {
	FilePath string
	Pattern  string
	MinLevel models.LogLevel
	Interval time.Duration
	ShowAll  bool
}

// Watcher watches a log file for changes in real-time
type Watcher struct {
	config     *Config
	parser     parser.LogParser
	file       *os.File
	lastOffset int64
}

// NewWatcher creates a new file watcher
func NewWatcher(config *Config) *Watcher {
	return &Watcher{
		config: config,
		parser: parser.DetectParser(""), // Will auto-detect
	}
}

// Watch starts watching the file for changes
func (w *Watcher) Watch(ctx context.Context) error {
	// Open file
	file, err := os.Open(w.config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	w.file = file

	// Seek to end of file (like tail -f)
	if !w.config.ShowAll {
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}
		w.lastOffset = fileInfo.Size()
		if _, err := file.Seek(w.lastOffset, 0); err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
	}

	// Create file system watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer fsWatcher.Close()

	// Add file to watcher
	if err := fsWatcher.Add(w.config.FilePath); err != nil {
		return fmt.Errorf("failed to watch file: %w", err)
	}

	fmt.Printf("🔍 Watching %s for changes...\n", w.config.FilePath)
	if w.config.Pattern != "" {
		fmt.Printf("🎯 Filtering for pattern: %s\n", w.config.Pattern)
	}
	if w.config.MinLevel != models.UNKNOWN {
		fmt.Printf("📊 Minimum level: %s\n", w.config.MinLevel)
	}
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println(color.New(color.FgCyan).Sprint("───────────────────────────────────────────────"))

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case event := <-fsWatcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				w.readNewLines()
			}

		case err := <-fsWatcher.Errors:
			return fmt.Errorf("watcher error: %w", err)

		case <-ticker.C:
			// Periodic check (fallback in case events are missed)
			w.readNewLines()
		}
	}
}

// readNewLines reads new lines added to the file
func (w *Watcher) readNewLines() {
	// Get current file size
	fileInfo, err := w.file.Stat()
	if err != nil {
		return
	}

	currentSize := fileInfo.Size()

	// Check if file was truncated (log rotation)
	if currentSize < w.lastOffset {
		w.lastOffset = 0
		w.file.Seek(0, 0)
	}

	// Read new content
	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		line := scanner.Text()
		w.processLine(line)
	}

	// Update offset
	w.lastOffset = currentSize
}

// processLine processes a single log line
func (w *Watcher) processLine(line string) {
	if line == "" {
		return
	}

	// Auto-detect parser
	currentParser := parser.DetectParser(line)

	// Parse line
	entry, err := currentParser.Parse(line, w.config.FilePath)
	if err != nil {
		return
	}

	// Apply filters
	if w.config.MinLevel != models.UNKNOWN && entry.Level < w.config.MinLevel {
		return
	}

	if w.config.Pattern != "" && !entry.MatchesPattern(w.config.Pattern) {
		return
	}

	// Display the entry with color
	w.displayEntry(entry)
}

// displayEntry displays a log entry with color coding
func (w *Watcher) displayEntry(entry *models.LogEntry) {
	timestamp := entry.Timestamp.Format("15:04:05")

	var levelColor *color.Color
	switch entry.Level {
	case models.FATAL:
		levelColor = color.New(color.FgRed, color.Bold)
	case models.ERROR:
		levelColor = color.New(color.FgRed)
	case models.WARN:
		levelColor = color.New(color.FgYellow)
	case models.INFO:
		levelColor = color.New(color.FgGreen)
	case models.DEBUG:
		levelColor = color.New(color.FgCyan)
	default:
		levelColor = color.New(color.FgWhite)
	}

	fmt.Printf("%s %s %s\n",
		color.New(color.FgHiBlack).Sprintf("[%s]", timestamp),
		levelColor.Sprintf("%-5s", entry.Level),
		entry.Message,
	)
}

// WatchWithStats watches file and displays periodic statistics
func (w *Watcher) WatchWithStats(ctx context.Context, statsInterval time.Duration) error {
	stats := models.NewStatistics()

	// Channel for log entries
	entryChan := make(chan *models.LogEntry, 100)

	// Start stats printer goroutine
	go func() {
		ticker := time.NewTicker(statsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-entryChan:
				stats.AddEntry(entry)
			case <-ticker.C:
				w.printStats(stats)
			}
		}
	}()

	// Start watching (simplified version)
	file, err := os.Open(w.config.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			line := scanner.Text()
			if entry, err := parser.DetectParser(line).Parse(line, w.config.FilePath); err == nil {
				entryChan <- entry
				w.displayEntry(entry)
			}
		}
	}

	return scanner.Err()
}

// printStats prints current statistics
func (w *Watcher) printStats(stats *models.Statistics) {
	fmt.Println(color.New(color.FgCyan).Sprint("\n───────────────────────────────────────────────"))
	fmt.Printf("📊 Stats (last %d entries)\n", stats.TotalEntries)
	fmt.Printf("   ERROR: %d | WARN: %d | INFO: %d\n",
		stats.GetLevelCount(models.ERROR),
		stats.GetLevelCount(models.WARN),
		stats.GetLevelCount(models.INFO),
	)
	fmt.Println(color.New(color.FgCyan).Sprint("───────────────────────────────────────────────\n"))
}
