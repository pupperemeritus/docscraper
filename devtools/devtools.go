package devtools

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"docscraper/config"
)

// DevTools provides development and debugging utilities
type DevTools struct {
	config      *config.Config
	debugMode   bool
	dryRunMode  bool
	progressBar *ProgressTracker
	profiler    *PerformanceProfiler
	validator   *ConfigValidator
	logger      *log.Logger
}

// NewDevTools creates a new DevTools instance
func NewDevTools(cfg *config.Config, debugMode, dryRunMode bool) *DevTools {
	logger := log.New(os.Stdout, "[DEVTOOLS] ", log.LstdFlags)

	return &DevTools{
		config:      cfg,
		debugMode:   debugMode,
		dryRunMode:  dryRunMode,
		progressBar: NewProgressTracker(),
		profiler:    NewPerformanceProfiler(),
		validator:   NewConfigValidator(),
		logger:      logger,
	}
}

// ValidateConfiguration validates the configuration and reports issues
func (dt *DevTools) ValidateConfiguration() error {
	dt.logger.Println("Validating configuration...")

	issues := dt.validator.ValidateConfig(dt.config)
	if len(issues) == 0 {
		dt.logger.Println("✓ Configuration validation passed")
		return nil
	}

	// Report issues
	dt.logger.Printf("⚠ Found %d configuration issues:", len(issues))
	for _, issue := range issues {
		dt.logger.Printf("  %s: %s", issue.Type, issue.Message)
	}

	// Check if any critical issues exist
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return fmt.Errorf("critical configuration issue: %s", issue.Message)
		}
	}

	return nil
}

// StartDryRun performs a dry run without actually scraping
func (dt *DevTools) StartDryRun() error {
	if !dt.dryRunMode {
		return fmt.Errorf("dry run mode not enabled")
	}

	dt.logger.Println("=== DRY RUN MODE ===")
	dt.logger.Printf("Root URL: %s", dt.config.RootURL)
	dt.logger.Printf("Max Depth: %d", dt.config.MaxDepth)
	dt.logger.Printf("Output Format: %s", dt.config.OutputFormat)
	dt.logger.Printf("Output Directory: %s", dt.config.OutputDir)

	// Simulate URL discovery
	dt.logger.Println("Simulating URL discovery...")
	estimatedURLs := dt.estimateURLCount()
	dt.logger.Printf("Estimated URLs to scrape: %d", estimatedURLs)

	// Estimate time and resources
	estimatedTime := time.Duration(estimatedURLs) * time.Second * 2 // 2 seconds per URL
	dt.logger.Printf("Estimated scraping time: %v", estimatedTime)

	return nil
}

// Debug logs debug information
func (dt *DevTools) Debug(format string, args ...interface{}) {
	if dt.debugMode {
		dt.logger.Printf("[DEBUG] "+format, args...)
	}
}

// StartProfiling begins performance profiling
func (dt *DevTools) StartProfiling() {
	dt.profiler.Start()
	dt.Debug("Performance profiling started")
}

// StopProfiling ends performance profiling and reports results
func (dt *DevTools) StopProfiling() *PerformanceReport {
	report := dt.profiler.Stop()
	dt.Debug("Performance profiling stopped")

	if dt.debugMode {
		dt.logger.Printf("Performance Report:")
		dt.logger.Printf("  Total Duration: %v", report.TotalDuration)
		dt.logger.Printf("  Pages Scraped: %d", report.PagesScraped)
		dt.logger.Printf("  Average Time per Page: %v", report.AverageTimePerPage)
		dt.logger.Printf("  Total Data Downloaded: %d bytes", report.TotalDataDownloaded)
		dt.logger.Printf("  Errors Encountered: %d", report.ErrorsEncountered)
	}

	return report
}

// UpdateProgress updates the progress tracker
func (dt *DevTools) UpdateProgress(current, total int, currentURL string) {
	dt.progressBar.Update(current, total, currentURL)
}

// estimateURLCount estimates the number of URLs that would be scraped
func (dt *DevTools) estimateURLCount() int {
	// Simple estimation based on depth and common site structures
	baseCount := 1
	for i := 0; i < dt.config.MaxDepth; i++ {
		baseCount += (i + 1) * 5 // Assume 5 links per page on average
	}
	return baseCount
}

// ConfigValidator validates configuration settings
type ConfigValidator struct{}

// NewConfigValidator creates a new config validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidationIssue represents a configuration validation issue
type ValidationIssue struct {
	Type     string `json:"type"`
	Severity string `json:"severity"` // "critical", "warning", "info"
	Message  string `json:"message"`
}

// ValidateConfig validates a configuration and returns any issues
func (cv *ConfigValidator) ValidateConfig(cfg *config.Config) []ValidationIssue {
	var issues []ValidationIssue

	// Validate root URL
	if cfg.RootURL == "" {
		issues = append(issues, ValidationIssue{
			Type:     "missing_root_url",
			Severity: "critical",
			Message:  "Root URL is required",
		})
	} else {
		if _, err := url.Parse(cfg.RootURL); err != nil {
			issues = append(issues, ValidationIssue{
				Type:     "invalid_root_url",
				Severity: "critical",
				Message:  fmt.Sprintf("Invalid root URL format: %v", err),
			})
		}
	}

	// Validate output directory
	if cfg.OutputDir == "" {
		issues = append(issues, ValidationIssue{
			Type:     "missing_output_dir",
			Severity: "critical",
			Message:  "Output directory is required",
		})
	} else {
		// Check if directory exists or can be created
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			issues = append(issues, ValidationIssue{
				Type:     "invalid_output_dir",
				Severity: "critical",
				Message:  fmt.Sprintf("Cannot create output directory: %v", err),
			})
		}
	}

	// Validate output format
	validFormats := []string{"markdown", "text", "json"}
	if !contains(validFormats, cfg.OutputFormat) {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_output_format",
			Severity: "critical",
			Message:  fmt.Sprintf("Invalid output format '%s'. Valid formats: %s", cfg.OutputFormat, strings.Join(validFormats, ", ")),
		})
	}

	// Validate output type
	validTypes := []string{"single", "per-page"}
	if !contains(validTypes, cfg.OutputType) {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_output_type",
			Severity: "critical",
			Message:  fmt.Sprintf("Invalid output type '%s'. Valid types: %s", cfg.OutputType, strings.Join(validTypes, ", ")),
		})
	}

	// Validate delays
	if cfg.MinDelay < 0 {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_min_delay",
			Severity: "warning",
			Message:  "Minimum delay should not be negative",
		})
	}

	if cfg.MaxDelay < cfg.MinDelay {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_delay_range",
			Severity: "warning",
			Message:  "Maximum delay should be greater than or equal to minimum delay",
		})
	}

	// Validate max depth
	if cfg.MaxDepth < 0 {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_max_depth",
			Severity: "warning",
			Message:  "Maximum depth should not be negative",
		})
	}

	if cfg.MaxDepth > 10 {
		issues = append(issues, ValidationIssue{
			Type:     "high_max_depth",
			Severity: "warning",
			Message:  "Maximum depth is very high, this may result in excessive scraping",
		})
	}

	// Validate user agents
	if len(cfg.UserAgents) == 0 {
		issues = append(issues, ValidationIssue{
			Type:     "no_user_agents",
			Severity: "info",
			Message:  "No user agents specified, will use defaults",
		})
	}

	// Validate proxy URLs if specified
	for i, proxy := range cfg.Proxies {
		if _, err := url.Parse(proxy); err != nil {
			issues = append(issues, ValidationIssue{
				Type:     "invalid_proxy_url",
				Severity: "warning",
				Message:  fmt.Sprintf("Invalid proxy URL at index %d: %v", i, err),
			})
		}
	}

	// Validate optional settings
	if cfg.ConcurrentRequests != nil && *cfg.ConcurrentRequests <= 0 {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_concurrent_requests",
			Severity: "warning",
			Message:  "Concurrent requests should be positive",
		})
	}

	if cfg.MinDelay < 0 {
		issues = append(issues, ValidationIssue{
			Type:     "invalid_min_delay",
			Severity: "warning",
			Message:  "Minimum delay should not be negative",
		})
	}

	return issues
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ProgressTracker tracks scraping progress
type ProgressTracker struct {
	current    int
	total      int
	currentURL string
	startTime  time.Time
	mutex      sync.RWMutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		startTime: time.Now(),
	}
}

// Update updates the progress
func (pt *ProgressTracker) Update(current, total int, currentURL string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.current = current
	pt.total = total
	pt.currentURL = currentURL

	// Calculate progress percentage
	var percentage float64
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}

	// Calculate ETA
	elapsed := time.Since(pt.startTime)
	var eta time.Duration
	if current > 0 {
		avgTimePerItem := elapsed / time.Duration(current)
		remaining := total - current
		eta = avgTimePerItem * time.Duration(remaining)
	}

	// Print progress
	fmt.Printf("\r[%d/%d] %.1f%% - ETA: %v - %s",
		current, total, percentage, eta.Round(time.Second), currentURL)

	if current == total {
		fmt.Println("\n✓ Scraping completed!")
	}
}

// GetStatus returns the current progress status
func (pt *ProgressTracker) GetStatus() (int, int, string, time.Duration) {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	elapsed := time.Since(pt.startTime)
	return pt.current, pt.total, pt.currentURL, elapsed
}

// PerformanceProfiler profiles scraping performance
type PerformanceProfiler struct {
	startTime           time.Time
	endTime             time.Time
	pagesScraped        int
	totalDataDownloaded int64
	errorsEncountered   int
	pageTimings         []time.Duration
	mutex               sync.RWMutex
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		pageTimings: make([]time.Duration, 0),
	}
}

// Start begins profiling
func (pp *PerformanceProfiler) Start() {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	pp.startTime = time.Now()
	pp.pagesScraped = 0
	pp.totalDataDownloaded = 0
	pp.errorsEncountered = 0
	pp.pageTimings = pp.pageTimings[:0]
}

// RecordPageScrape records metrics for a scraped page
func (pp *PerformanceProfiler) RecordPageScrape(duration time.Duration, dataSize int64, hadError bool) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	pp.pagesScraped++
	pp.totalDataDownloaded += dataSize
	pp.pageTimings = append(pp.pageTimings, duration)

	if hadError {
		pp.errorsEncountered++
	}
}

// Stop ends profiling and returns a performance report
func (pp *PerformanceProfiler) Stop() *PerformanceReport {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	pp.endTime = time.Now()

	var averageTimePerPage time.Duration
	if len(pp.pageTimings) > 0 {
		total := time.Duration(0)
		for _, timing := range pp.pageTimings {
			total += timing
		}
		averageTimePerPage = total / time.Duration(len(pp.pageTimings))
	}

	return &PerformanceReport{
		TotalDuration:       pp.endTime.Sub(pp.startTime),
		PagesScraped:        pp.pagesScraped,
		AverageTimePerPage:  averageTimePerPage,
		TotalDataDownloaded: pp.totalDataDownloaded,
		ErrorsEncountered:   pp.errorsEncountered,
		PageTimings:         append([]time.Duration{}, pp.pageTimings...),
	}
}

// PerformanceReport contains performance metrics
type PerformanceReport struct {
	TotalDuration       time.Duration   `json:"total_duration"`
	PagesScraped        int             `json:"pages_scraped"`
	AverageTimePerPage  time.Duration   `json:"average_time_per_page"`
	TotalDataDownloaded int64           `json:"total_data_downloaded"`
	ErrorsEncountered   int             `json:"errors_encountered"`
	PageTimings         []time.Duration `json:"page_timings"`
}

// SaveReport saves the performance report to a file
func (pr *PerformanceReport) SaveReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "Performance Report\n")
	fmt.Fprintf(file, "==================\n\n")
	fmt.Fprintf(file, "Total Duration: %v\n", pr.TotalDuration)
	fmt.Fprintf(file, "Pages Scraped: %d\n", pr.PagesScraped)
	fmt.Fprintf(file, "Average Time per Page: %v\n", pr.AverageTimePerPage)
	fmt.Fprintf(file, "Total Data Downloaded: %d bytes (%.2f MB)\n",
		pr.TotalDataDownloaded, float64(pr.TotalDataDownloaded)/1024/1024)
	fmt.Fprintf(file, "Errors Encountered: %d\n", pr.ErrorsEncountered)

	if len(pr.PageTimings) > 0 {
		fmt.Fprintf(file, "\nPage Timing Details:\n")
		for i, timing := range pr.PageTimings {
			fmt.Fprintf(file, "Page %d: %v\n", i+1, timing)
		}
	}

	return nil
}
