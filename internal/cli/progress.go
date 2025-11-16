package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressReporter handles progress reporting during data generation
type ProgressReporter struct {
	enabled     bool
	quiet       bool
	totalTables int
	currentBar  *progressbar.ProgressBar
	startTime   time.Time
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(cfg *Config, totalTables int) *ProgressReporter {
	return &ProgressReporter{
		enabled:     cfg.ProgressBar && !cfg.QuietMode,
		quiet:       cfg.QuietMode,
		totalTables: totalTables,
		startTime:   time.Now(),
	}
}

// StartTable starts progress tracking for a table
func (p *ProgressReporter) StartTable(tableName string, rowCount int) {
	if p.quiet {
		return
	}

	if p.enabled {
		// Create a new progress bar for this table
		p.currentBar = progressbar.NewOptions(rowCount,
			progressbar.OptionSetDescription(fmt.Sprintf("[cyan]Generating %s[reset]", tableName)),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(40),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprintf(os.Stderr, "\n")
			}),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
	} else {
		// Simple text output without progress bar
		fmt.Fprintf(os.Stderr, "Generating %s (%d rows)...\n", tableName, rowCount)
	}
}

// UpdateProgress updates the progress bar
func (p *ProgressReporter) UpdateProgress(rowsGenerated int) {
	if p.enabled && p.currentBar != nil {
		p.currentBar.Add(rowsGenerated)
	}
}

// CompleteTable marks a table as complete
func (p *ProgressReporter) CompleteTable(tableName string, rowCount int, duration time.Duration) {
	if p.quiet {
		return
	}

	if p.enabled && p.currentBar != nil {
		p.currentBar.Finish()
		p.currentBar = nil
	}

	// Log table completion
	LogDataGeneration(tableName, rowCount, duration)
}

// PrintSummary prints the final summary
func (p *ProgressReporter) PrintSummary(totalRows int, totalTables int) {
	if p.quiet {
		return
	}

	elapsed := time.Since(p.startTime)

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "✓ Generation complete\n")
	fmt.Fprintf(os.Stderr, "  Tables:   %d\n", totalTables)
	fmt.Fprintf(os.Stderr, "  Rows:     %d\n", totalRows)
	fmt.Fprintf(os.Stderr, "  Duration: %s\n", elapsed.Round(time.Millisecond))

	if elapsed.Seconds() > 0 {
		rowsPerSec := float64(totalRows) / elapsed.Seconds()
		fmt.Fprintf(os.Stderr, "  Speed:    %.0f rows/sec\n", rowsPerSec)
	}
}

// PrintError prints an error message
func (p *ProgressReporter) PrintError(msg string, err error) {
	if p.quiet {
		return
	}

	if p.currentBar != nil {
		p.currentBar.Finish()
		p.currentBar = nil
	}

	fmt.Fprintf(os.Stderr, "\n✗ Error: %s\n", msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
	}
}

// PrintWarning prints a warning message
func (p *ProgressReporter) PrintWarning(msg string) {
	if p.quiet {
		return
	}

	fmt.Fprintf(os.Stderr, "⚠ Warning: %s\n", msg)
}

// PrintInfo prints an informational message
func (p *ProgressReporter) PrintInfo(msg string) {
	if p.quiet {
		return
	}

	fmt.Fprintf(os.Stderr, "ℹ %s\n", msg)
}

// SimpleProgressReporter is a basic progress reporter without bars
type SimpleProgressReporter struct {
	quiet     bool
	output    io.Writer
	startTime time.Time
}

// NewSimpleProgressReporter creates a simple progress reporter (no progress bars)
func NewSimpleProgressReporter(output io.Writer, quiet bool) *SimpleProgressReporter {
	if output == nil {
		output = os.Stderr
	}
	return &SimpleProgressReporter{
		quiet:     quiet,
		output:    output,
		startTime: time.Now(),
	}
}

// ReportTableStart reports the start of table generation
func (s *SimpleProgressReporter) ReportTableStart(tableName string, rowCount int) {
	if s.quiet {
		return
	}
	fmt.Fprintf(s.output, "Generating table '%s' (%d rows)...\n", tableName, rowCount)
}

// ReportTableComplete reports table generation completion
func (s *SimpleProgressReporter) ReportTableComplete(tableName string, rowCount int, duration time.Duration) {
	if s.quiet {
		return
	}
	fmt.Fprintf(s.output, "✓ Completed '%s' (%d rows) in %s\n", tableName, rowCount, duration.Round(time.Millisecond))
}

// ReportSummary reports final summary
func (s *SimpleProgressReporter) ReportSummary(totalRows int, totalTables int) {
	if s.quiet {
		return
	}

	elapsed := time.Since(s.startTime)
	fmt.Fprintf(s.output, "\n")
	fmt.Fprintf(s.output, "✓ Generation complete\n")
	fmt.Fprintf(s.output, "  Tables:   %d\n", totalTables)
	fmt.Fprintf(s.output, "  Rows:     %d\n", totalRows)
	fmt.Fprintf(s.output, "  Duration: %s\n", elapsed.Round(time.Millisecond))

	if elapsed.Seconds() > 0 {
		rowsPerSec := float64(totalRows) / elapsed.Seconds()
		fmt.Fprintf(s.output, "  Speed:    %.0f rows/sec\n", rowsPerSec)
	}
}

// ReportError reports an error
func (s *SimpleProgressReporter) ReportError(msg string, err error) {
	if s.quiet {
		return
	}
	fmt.Fprintf(s.output, "✗ Error: %s\n", msg)
	if err != nil {
		fmt.Fprintf(s.output, "  %v\n", err)
	}
}

// Spinner provides a simple text spinner for indeterminate progress
type Spinner struct {
	message string
	frames  []string
	current int
	ticker  *time.Ticker
	done    chan bool
	active  bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:    make(chan bool),
	}
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	if s.active {
		return
	}

	s.active = true
	s.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				frame := s.frames[s.current%len(s.frames)]
				fmt.Fprintf(os.Stderr, "\r%s %s", frame, s.message)
				s.current++
			case <-s.done:
				fmt.Fprintf(os.Stderr, "\r")
				return
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !s.active {
		return
	}

	s.ticker.Stop()
	s.done <- true
	s.active = false

	// Clear the line
	fmt.Fprintf(os.Stderr, "\r\033[K")
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}
