package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	_ "path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aadithyaa9/loganalyzer/internal/analyzer"
	"github.com/aadithyaa9/loganalyzer/internal/models"
	"github.com/aadithyaa9/loganalyzer/internal/reporter"
	"github.com/aadithyaa9/loganalyzer/internal/watcher"
	"github.com/fatih/color"
)

const (
	version = "1.0.0"
	banner  = `
╔══════════════════════════════════════════════════════════╗
║          🔍 LogAnalyzer v%s                         ║
║     Fast, Concurrent Log Analysis Tool in Go            ║
╚══════════════════════════════════════════════════════════╝
`
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		handleAnalyze()
	case "watch":
		handleWatch()
	case "stats":
		handleStats()
	case "help":
		printUsage()
	case "version":
		fmt.Printf("LogAnalyzer v%s\n", version)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleAnalyze() {
	// Define flags
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	file := fs.String("file", "", "Single log file to analyze")
	dir := fs.String("dir", "", "Directory containing log files")
	level := fs.String("level", "", "Minimum log level (DEBUG, INFO, WARN, ERROR, FATAL)")
	pattern := fs.String("pattern", "", "Pattern to search for")
	workers := fs.Int("workers", 4, "Number of concurrent workers")
	format := fs.String("format", "table", "Output format (table, json)")
	output := fs.String("output", "", "Output file (default: stdout)")
	topErrors := fs.Int("top-errors", 0, "Show top N error patterns")

	fs.Parse(os.Args[2:])

	// Validate inputs
	if *file == "" && *dir == "" {
		fmt.Println("Error: Either --file or --dir must be specified")
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Print banner
	printBanner()

	// Parse log level
	var minLevel models.LogLevel
	if *level != "" {
		minLevel = models.ParseLogLevel(strings.ToUpper(*level))
		if minLevel == models.UNKNOWN {
			fmt.Printf("Warning: Unknown log level '%s', showing all levels\n", *level)
			minLevel = models.DEBUG
		}
	}

	// Create analyzer config
	config := &analyzer.Config{
		Workers:    *workers,
		Level:      minLevel,
		Pattern:    *pattern,
		AutoDetect: true,
	}

	// Create analyzer
	a := analyzer.NewAnalyzer(config)

	// Analyze
	fmt.Println("🚀 Starting analysis...")
	startTime := time.Now()

	var err error
	if *file != "" {
		err = a.AnalyzeFile(*file)
	} else {
		err = a.AnalyzeDirectory(*dir)
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Analysis complete in %s\n\n", time.Since(startTime).Round(time.Millisecond))

	// Get results
	results := a.GetResults()
	entries := results.GetEntries()
	stats := results.GetStats()

	// Determine output writer
	var writer *os.File
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Printf("❌ Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		writer = f
		fmt.Printf("📝 Writing output to %s\n\n", *output)
	} else {
		writer = os.Stdout
	}

	// Generate report
	var rep reporter.Reporter
	switch *format {
	case "json":
		rep = reporter.GetReporter(reporter.JSONFormat)
	default:
		rep = reporter.GetReporter(reporter.TableFormat)
	}

	if err := rep.Report(entries, stats, writer); err != nil {
		fmt.Printf("❌ Failed to generate report: %v\n", err)
		os.Exit(1)
	}

	// Print top errors if requested
	if *topErrors > 0 && *format == "table" {
		reporter.PrintTopErrors(stats, writer, *topErrors)
	}

	// Print source breakdown
	if *format == "table" && len(stats.SourceCounts) > 1 {
		reporter.PrintSourceBreakdown(stats, writer)
	}
}

func handleWatch() {
	// Define flags
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	file := fs.String("file", "", "Log file to watch")
	pattern := fs.String("pattern", "", "Pattern to filter for")
	level := fs.String("level", "", "Minimum log level to show")
	interval := fs.Duration("interval", 1*time.Second, "Check interval")
	showAll := fs.Bool("all", false, "Show all existing entries (not just new ones)")

	fs.Parse(os.Args[2:])

	// Validate
	if *file == "" {
		fmt.Println("Error: --file must be specified")
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Check if file exists
	if _, err := os.Stat(*file); os.IsNotExist(err) {
		fmt.Printf("❌ File does not exist: %s\n", *file)
		os.Exit(1)
	}

	printBanner()

	// Parse log level
	var minLevel models.LogLevel
	if *level != "" {
		minLevel = models.ParseLogLevel(strings.ToUpper(*level))
	}

	// Create watcher config
	config := &watcher.Config{
		FilePath: *file,
		Pattern:  *pattern,
		MinLevel: minLevel,
		Interval: *interval,
		ShowAll:  *showAll,
	}

	// Create watcher
	w := watcher.NewWatcher(config)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n👋 Stopping watcher...")
		cancel()
	}()

	// Start watching
	if err := w.Watch(ctx); err != nil {
		if err != context.Canceled {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleStats() {
	// Define flags
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	file := fs.String("file", "", "Single log file to analyze")
	dir := fs.String("dir", "", "Directory containing log files")
	workers := fs.Int("workers", 4, "Number of concurrent workers")

	fs.Parse(os.Args[2:])

	// Validate
	if *file == "" && *dir == "" {
		fmt.Println("Error: Either --file or --dir must be specified")
		fs.PrintDefaults()
		os.Exit(1)
	}

	printBanner()

	// Create analyzer config
	config := &analyzer.Config{
		Workers:    *workers,
		AutoDetect: true,
	}

	// Create analyzer
	a := analyzer.NewAnalyzer(config)

	// Analyze
	fmt.Println("📊 Gathering statistics...")

	var err error
	if *file != "" {
		err = a.AnalyzeFile(*file)
	} else {
		err = a.AnalyzeDirectory(*dir)
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Print statistics
	stats := a.GetResults().GetStats()
	fmt.Println(stats.Summary())
	reporter.PrintSourceBreakdown(stats, os.Stdout)
}

func printBanner() {
	fmt.Printf(color.CyanString(banner), version)
}

func printUsage() {
	printBanner()
	fmt.Println("Usage: loganalyzer <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  analyze    Analyze log files")
	fmt.Println("  watch      Watch a log file in real-time")
	fmt.Println("  stats      Show statistics for log files")
	fmt.Println("  help       Show this help message")
	fmt.Println("  version    Show version information")

	fmt.Println("\nAnalyze Options:")
	fmt.Println("  --file <path>        Single log file to analyze")
	fmt.Println("  --dir <path>         Directory containing log files")
	fmt.Println("  --level <level>      Minimum log level (DEBUG, INFO, WARN, ERROR, FATAL)")
	fmt.Println("  --pattern <string>   Pattern to search for")
	fmt.Println("  --workers <num>      Number of concurrent workers (default: 4)")
	fmt.Println("  --format <format>    Output format: table, json (default: table)")
	fmt.Println("  --output <path>      Output file (default: stdout)")
	fmt.Println("  --top-errors <num>   Show top N error patterns")

	fmt.Println("\nWatch Options:")
	fmt.Println("  --file <path>        Log file to watch")
	fmt.Println("  --pattern <string>   Pattern to filter for")
	fmt.Println("  --level <level>      Minimum log level to show")
	fmt.Println("  --interval <dur>     Check interval (default: 1s)")
	fmt.Println("  --all                Show all existing entries")

	fmt.Println("\nStats Options:")
	fmt.Println("  --file <path>        Single log file to analyze")
	fmt.Println("  --dir <path>         Directory containing log files")
	fmt.Println("  --workers <num>      Number of concurrent workers (default: 4)")

	fmt.Println("\nExamples:")
	fmt.Println("  # Analyze a single file for errors")
	fmt.Println("  loganalyzer analyze --file app.log --level ERROR")
	fmt.Println()
	fmt.Println("  # Analyze directory with pattern matching")
	fmt.Println("  loganalyzer analyze --dir ./logs --pattern \"database\" --workers 8")
	fmt.Println()
	fmt.Println("  # Watch file in real-time")
	fmt.Println("  loganalyzer watch --file app.log --level WARN")
	fmt.Println()
	fmt.Println("  # Generate JSON report")
	fmt.Println("  loganalyzer analyze --dir ./logs --format json --output report.json")
	fmt.Println()
	fmt.Println("  # Show statistics")
	fmt.Println("  loganalyzer stats --dir ./logs")
	fmt.Println()
}
