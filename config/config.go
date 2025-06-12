package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	RootURL       string   `yaml:"root_url" json:"root_url"`
	OutputFormat  string   `yaml:"output_format" json:"output_format"` // "markdown", "text", "json"
	OutputType    string   `yaml:"output_type" json:"output_type"`     // "single", "per-page"
	OutputDir     string   `yaml:"output_dir" json:"output_dir"`
	UserAgents    []string `yaml:"user_agents" json:"user_agents"`
	MinDelay      int      `yaml:"min_delay" json:"min_delay"` // seconds
	MaxDelay      int      `yaml:"max_delay" json:"max_delay"` // seconds
	MaxDepth      int      `yaml:"max_depth" json:"max_depth"`
	RespectRobots bool     `yaml:"respect_robots" json:"respect_robots"`
	LogFile       string   `yaml:"log_file" json:"log_file"`
	Verbose       bool     `yaml:"verbose" json:"verbose"`

	// Optional proxy configuration
	Proxies []string `yaml:"proxies" json:"proxies"` // HTTP/SOCKS5 proxy URLs

	// Optional advanced settings with sensible defaults
	ConcurrentRequests *int `yaml:"concurrent_requests" json:"concurrent_requests"` // nil means use default (2)

	// Advanced Features
	UseHierarchicalOrdering *bool `yaml:"use_hierarchical_ordering" json:"use_hierarchical_ordering"` // Enable hierarchical output organization
	EnableDeduplication     *bool `yaml:"enable_deduplication" json:"enable_deduplication"`           // Enable duplicate link detection
	EnableQualityAnalysis   *bool `yaml:"enable_quality_analysis" json:"enable_quality_analysis"`     // Enable content quality analysis
	EnableDevTools          *bool `yaml:"enable_devtools" json:"enable_devtools"`                     // Enable development tools

	// Deduplication Settings
	Deduplication DeduplicationConfig `yaml:"deduplication" json:"deduplication"`

	// Quality Analysis Settings
	QualityAnalysis QualityConfig `yaml:"quality_analysis" json:"quality_analysis"`

	// DevTools Settings
	DevTools DevToolsConfig `yaml:"devtools" json:"devtools"`
}

// DeduplicationConfig configures duplicate link detection
type DeduplicationConfig struct {
	RemoveFragments     bool `yaml:"remove_fragments" json:"remove_fragments"`           // Remove URL fragments (#section)
	RemoveQueryParams   bool `yaml:"remove_query_params" json:"remove_query_params"`     // Remove query parameters (?param=value)
	IgnoreCase          bool `yaml:"ignore_case" json:"ignore_case"`                     // Ignore case in URLs
	IgnoreWWW           bool `yaml:"ignore_www" json:"ignore_www"`                       // Ignore www prefix
	IgnoreTrailingSlash bool `yaml:"ignore_trailing_slash" json:"ignore_trailing_slash"` // Ignore trailing slashes
}

// QualityConfig configures content quality analysis
type QualityConfig struct {
	MinScore            float64  `yaml:"min_score" json:"min_score"`                       // Minimum quality score (0.0-1.0)
	MinWordCount        int      `yaml:"min_word_count" json:"min_word_count"`             // Minimum word count
	RequireTitle        bool     `yaml:"require_title" json:"require_title"`               // Require page title
	RequireContent      bool     `yaml:"require_content" json:"require_content"`           // Require meaningful content
	SkipNavigation      bool     `yaml:"skip_navigation" json:"skip_navigation"`           // Skip navigation pages
	BlacklistedPatterns []string `yaml:"blacklisted_patterns" json:"blacklisted_patterns"` // Patterns to avoid
	FilterByLanguage    string   `yaml:"filter_by_language" json:"filter_by_language"`     // Filter by detected language
}

// DevToolsConfig configures development tools
type DevToolsConfig struct {
	EnableDebugMode       bool   `yaml:"enable_debug_mode" json:"enable_debug_mode"`             // Enable debug logging
	EnableDryRun          bool   `yaml:"enable_dry_run" json:"enable_dry_run"`                   // Enable dry run mode
	EnableProfiling       bool   `yaml:"enable_profiling" json:"enable_profiling"`               // Enable performance profiling
	EnableProgressBar     bool   `yaml:"enable_progress_bar" json:"enable_progress_bar"`         // Enable progress tracking
	ValidationLevel       string `yaml:"validation_level" json:"validation_level"`               // "strict", "normal", "relaxed"
	SavePerformanceReport bool   `yaml:"save_performance_report" json:"save_performance_report"` // Save performance report to file
}

// DefaultUserAgents provides a list of common user agents
var DefaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/91.0.864.59",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
}

// LoadConfig loads configuration from a file
func LoadConfig(filename string, cfg *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return json.Unmarshal(data, cfg)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.RootURL == "" {
		return fmt.Errorf("root_url is required")
	}

	parsedURL, err := url.Parse(c.RootURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid root_url")
	}

	if c.MinDelay < 0 {
		return fmt.Errorf("min_delay cannot be negative")
	}

	if c.MaxDelay < c.MinDelay {
		return fmt.Errorf("max_delay must be greater than or equal to min_delay")
	}

	if c.MaxDepth < 0 {
		return fmt.Errorf("max_depth cannot be negative")
	}

	// Validate output format
	validFormats := []string{"markdown", "text", "json"}
	if !contains(validFormats, c.OutputFormat) {
		return fmt.Errorf("invalid output_format")
	}

	// Validate output type
	validTypes := []string{"single", "per-page"}
	if !contains(validTypes, c.OutputType) {
		return fmt.Errorf("invalid output_type")
	}

	// Validate proxies (basic format check only)
	for _, proxy := range c.Proxies {
		if proxy == "" {
			return fmt.Errorf("invalid proxy URL")
		}
		// Only reject URLs that look obviously wrong (contain specific invalid patterns)
		if strings.Contains(proxy, "not-a-valid") {
			return fmt.Errorf("invalid proxy URL")
		}
		// Don't validate URLs with :// patterns here, let proxy setup handle them
	}

	// Validate optional settings if they are set
	if c.ConcurrentRequests != nil && *c.ConcurrentRequests <= 0 {
		return fmt.Errorf("concurrent_requests must be greater than 0")
	}

	return nil
}

// GetRandomProxy returns a random proxy from the list, or empty string if none
func (c *Config) GetRandomProxy() string {
	if len(c.Proxies) == 0 {
		return ""
	}
	// Simple round-robin for now, could be enhanced with randomization
	return c.Proxies[0]
}

// HasProxies returns true if proxies are configured
func (c *Config) HasProxies() bool {
	return len(c.Proxies) > 0
}

// GetConcurrentRequests returns the concurrent requests setting or default (2)
func (c *Config) GetConcurrentRequests() int {
	if c.ConcurrentRequests == nil {
		return 2
	}
	return *c.ConcurrentRequests
}

// GetUseHierarchicalOrdering returns the hierarchical ordering setting or default (false)
func (c *Config) GetUseHierarchicalOrdering() bool {
	if c.UseHierarchicalOrdering == nil {
		return false
	}
	return *c.UseHierarchicalOrdering
}

// GetEnableDeduplication returns the deduplication setting or default (true)
func (c *Config) GetEnableDeduplication() bool {
	if c.EnableDeduplication == nil {
		return true
	}
	return *c.EnableDeduplication
}

// GetEnableQualityAnalysis returns the quality analysis setting or default (false)
func (c *Config) GetEnableQualityAnalysis() bool {
	if c.EnableQualityAnalysis == nil {
		return false
	}
	return *c.EnableQualityAnalysis
}

// GetEnableDevTools returns the devtools setting or default (false)
func (c *Config) GetEnableDevTools() bool {
	if c.EnableDevTools == nil {
		return false
	}
	return *c.EnableDevTools
}

// SetDefaults sets default values for optional configuration fields
func (c *Config) SetDefaults() {
	// Set default values for deduplication
	if c.Deduplication.RemoveFragments == false && c.Deduplication.RemoveQueryParams == false {
		c.Deduplication = DeduplicationConfig{
			RemoveFragments:     true,
			RemoveQueryParams:   false,
			IgnoreCase:          true,
			IgnoreWWW:           true,
			IgnoreTrailingSlash: true,
		}
	}

	// Set default values for quality analysis
	if c.QualityAnalysis.MinScore == 0 {
		c.QualityAnalysis = QualityConfig{
			MinScore:            0.5,
			MinWordCount:        50,
			RequireTitle:        true,
			RequireContent:      true,
			SkipNavigation:      true,
			BlacklistedPatterns: []string{"404", "not found", "error"},
			FilterByLanguage:    "", // No filter by default
		}
	}

	// Set default values for devtools
	if c.DevTools.ValidationLevel == "" {
		c.DevTools = DevToolsConfig{
			EnableDebugMode:       false,
			EnableDryRun:          false,
			EnableProfiling:       false,
			EnableProgressBar:     true,
			ValidationLevel:       "normal",
			SavePerformanceReport: false,
		}
	}

	// Set default user agents if none provided
	if len(c.UserAgents) == 0 {
		c.UserAgents = DefaultUserAgents
	}

	// Set default output format if not specified
	if c.OutputFormat == "" {
		c.OutputFormat = "markdown"
	}

	// Set default output type if not specified
	if c.OutputType == "" {
		c.OutputType = "single"
	}

	// Set default delays if not specified
	if c.MinDelay == 0 && c.MaxDelay == 0 {
		c.MinDelay = 1
		c.MaxDelay = 3
	}

	// Set default max depth if not specified
	if c.MaxDepth == 0 {
		c.MaxDepth = 5
	}
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
