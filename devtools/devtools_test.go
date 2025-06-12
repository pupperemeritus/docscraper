package devtools

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"docscraper/config"
)

func TestConfigValidator(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name           string
		config         *config.Config
		expectIssues   int
		expectCritical bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "/tmp/test",
				MinDelay:     1,
				MaxDelay:     3,
				MaxDepth:     5,
				UserAgents:   []string{"test-agent"},
			},
			expectIssues:   0,
			expectCritical: false,
		},
		{
			name: "missing root URL",
			config: &config.Config{
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "/tmp/test",
				UserAgents:   []string{"test-agent"}, // Add to avoid no_user_agents issue
			},
			expectIssues:   1,
			expectCritical: true,
		},
		{
			name: "invalid output format",
			config: &config.Config{
				RootURL:      "https://example.com",
				OutputFormat: "invalid",
				OutputType:   "single",
				OutputDir:    "/tmp/test",
				UserAgents:   []string{"test-agent"}, // Add to avoid no_user_agents issue
			},
			expectIssues:   1,
			expectCritical: true,
		},
		{
			name: "invalid delay range",
			config: &config.Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "/tmp/test",
				MinDelay:     5,
				MaxDelay:     2,
				UserAgents:   []string{"test-agent"}, // Add to avoid no_user_agents issue
			},
			expectIssues:   1,
			expectCritical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.ValidateConfig(tt.config)

			if len(issues) != tt.expectIssues {
				t.Errorf("Expected %d issues, got %d", tt.expectIssues, len(issues))
			}

			hasCritical := false
			for _, issue := range issues {
				if issue.Severity == "critical" {
					hasCritical = true
					break
				}
			}

			if hasCritical != tt.expectCritical {
				t.Errorf("Expected critical issues: %v, got: %v", tt.expectCritical, hasCritical)
			}
		})
	}
}

func TestProgressTracker(t *testing.T) {
	tracker := NewProgressTracker()

	// Test initial state
	current, total, url, _ := tracker.GetStatus()
	if current != 0 || total != 0 || url != "" {
		t.Errorf("Expected initial state to be zero values")
	}

	// Test update
	tracker.Update(5, 10, "https://example.com")
	current, total, url, elapsed := tracker.GetStatus()

	if current != 5 || total != 10 || url != "https://example.com" {
		t.Errorf("Progress not updated correctly: %d/%d, %s", current, total, url)
	}

	if elapsed <= 0 {
		t.Errorf("Elapsed time should be positive")
	}
}

func TestPerformanceProfiler(t *testing.T) {
	profiler := NewPerformanceProfiler()

	// Start profiling
	profiler.Start()

	// Simulate some page scrapes
	profiler.RecordPageScrape(100*time.Millisecond, 1024, false)
	profiler.RecordPageScrape(200*time.Millisecond, 2048, true)
	profiler.RecordPageScrape(150*time.Millisecond, 1536, false)

	// Stop and get report
	report := profiler.Stop()

	if report.PagesScraped != 3 {
		t.Errorf("Expected 3 pages scraped, got %d", report.PagesScraped)
	}

	if report.TotalDataDownloaded != 4608 {
		t.Errorf("Expected 4608 bytes downloaded, got %d", report.TotalDataDownloaded)
	}

	if report.ErrorsEncountered != 1 {
		t.Errorf("Expected 1 error, got %d", report.ErrorsEncountered)
	}

	if len(report.PageTimings) != 3 {
		t.Errorf("Expected 3 page timings, got %d", len(report.PageTimings))
	}

	if report.TotalDuration <= 0 {
		t.Errorf("Total duration should be positive")
	}
}

func TestDevToolsValidation(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "devtools_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    tempDir,
		MinDelay:     1,
		MaxDelay:     3,
		MaxDepth:     5,
	}

	devtools := NewDevTools(cfg, true, false)

	err = devtools.ValidateConfiguration()
	if err != nil {
		t.Errorf("Validation failed for valid config: %v", err)
	}
}

func TestDevToolsDryRun(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/test",
		MaxDepth:     3,
	}

	devtools := NewDevTools(cfg, false, true)

	err := devtools.StartDryRun()
	if err != nil {
		t.Errorf("Dry run failed: %v", err)
	}
}

func TestPerformanceReportSave(t *testing.T) {
	// Create temporary file
	tempDir, err := os.MkdirTemp("", "perf_report_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	report := &PerformanceReport{
		TotalDuration:       10 * time.Second,
		PagesScraped:        5,
		AverageTimePerPage:  2 * time.Second,
		TotalDataDownloaded: 10240,
		ErrorsEncountered:   1,
		PageTimings:         []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
	}

	filename := filepath.Join(tempDir, "report.txt")
	err = report.SaveReport(filename)
	if err != nil {
		t.Errorf("Failed to save report: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Report file was not created")
	}
}

func TestDevToolsDebugMode(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/test",
	}

	// Test debug mode enabled
	devtoolsDebug := NewDevTools(cfg, true, false)
	devtoolsDebug.Debug("This is a debug message")

	// Test debug mode disabled
	devtoolsNoDebug := NewDevTools(cfg, false, false)
	devtoolsNoDebug.Debug("This should not be printed")

	// No assertions needed - just ensure no panics
}

func TestDevToolsProfiling(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/test",
	}

	devtools := NewDevTools(cfg, true, false)

	// Start profiling
	devtools.StartProfiling()

	// Simulate some work
	devtools.profiler.RecordPageScrape(100*time.Millisecond, 1024, false)

	// Stop profiling
	report := devtools.StopProfiling()

	if report == nil {
		t.Errorf("Expected performance report, got nil")
	}

	if report.PagesScraped != 1 {
		t.Errorf("Expected 1 page scraped, got %d", report.PagesScraped)
	}
}
