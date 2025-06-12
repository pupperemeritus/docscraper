package scraper

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"docscraper/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &config.Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "test-output",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
				LogFile:      "test.log",
			},
			wantErr: false,
		},
		{
			name: "invalid root URL",
			config: &config.Config{
				RootURL:      "not-a-valid-url",
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "test-output",
				LogFile:      "test.log",
			},
			wantErr: true,
			errMsg:  "invalid root_url",
		},
		{
			name: "config with proxies",
			config: &config.Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "test-output",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
				LogFile:      "test.log",
				Proxies:      []string{"http://proxy.example.com:8080"},
			},
			wantErr: false,
		},
		{
			name: "config with invalid proxies",
			config: &config.Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				OutputDir:    "test-output",
				LogFile:      "test.log",
				Proxies:      []string{"://invalid"},
			},
			wantErr: true,
			errMsg:  "failed to setup proxy switcher",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary log file
			logFile, err := os.CreateTemp("", "test-log-*.log")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(logFile.Name())
			logFile.Close()

			tt.config.LogFile = logFile.Name()

			scraper, err := New(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("New() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					if scraper == nil {
						t.Error("New() returned nil scraper")
					}
					if scraper.config != tt.config {
						t.Error("Scraper config not set correctly")
					}
					if scraper.collector == nil {
						t.Error("Scraper collector not initialized")
					}
					if scraper.extractor == nil {
						t.Error("Scraper extractor not initialized")
					}
					if scraper.logger == nil {
						t.Error("Scraper logger not initialized")
					}
				}
			}
		})
	}
}

func TestScraper_shouldFollowLink(t *testing.T) {
	// Create a test scraper
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "test-output",
		LogFile:      "test.log",
	}

	logFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(logFile.Name())
	logFile.Close()
	cfg.LogFile = logFile.Name()

	scraper, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	baseURL, _ := url.Parse("https://example.com/docs/")

	tests := []struct {
		name     string
		link     string
		baseURL  *url.URL
		expected bool
	}{
		{
			name:     "relative link same domain",
			link:     "/docs/page1",
			baseURL:  baseURL,
			expected: true,
		},
		{
			name:     "absolute link same domain",
			link:     "https://example.com/docs/page2",
			baseURL:  baseURL,
			expected: true,
		},
		{
			name:     "external domain",
			link:     "https://external.com/page",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "PDF file",
			link:     "/docs/file.pdf",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "image file",
			link:     "/images/photo.jpg",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "fragment only",
			link:     "#section1",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "login page",
			link:     "/login",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "API endpoint",
			link:     "/api/users",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "admin section",
			link:     "/admin/dashboard",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "search page",
			link:     "/search?q=test",
			baseURL:  baseURL,
			expected: false,
		},
		{
			name:     "valid documentation link",
			link:     "/docs/guide/getting-started",
			baseURL:  baseURL,
			expected: true,
		},
		{
			name:     "invalid link format",
			link:     "not-a-valid-url",
			baseURL:  baseURL,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.shouldFollowLink(tt.link, tt.baseURL)
			if result != tt.expected {
				t.Errorf("shouldFollowLink(%q) = %v, want %v", tt.link, result, tt.expected)
			}
		})
	}
}

func TestScraper_checkRobotsTxt(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "test-output",
		LogFile:      "test.log",
	}

	logFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(logFile.Name())
	logFile.Close()
	cfg.LogFile = logFile.Name()

	scraper, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		rootURL  string
		expected bool
		wantErr  bool
	}{
		{
			name:     "invalid URL",
			rootURL:  "not-a-valid-url",
			expected: false,
			wantErr:  true,
		},
		{
			name:     "non-existent domain",
			rootURL:  "https://this-domain-should-not-exist-12345.com",
			expected: true, // Should return true if robots.txt doesn't exist
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := scraper.checkRobotsTxt(tt.rootURL)

			if tt.wantErr {
				if err == nil {
					t.Error("checkRobotsTxt() expected error but got none")
				}
			} else {
				if err != nil {
					// For network errors, we expect the function to return true (allowed)
					if allowed != true {
						t.Errorf("checkRobotsTxt() on network error should return true, got %v", allowed)
					}
				} else {
					if allowed != tt.expected {
						t.Errorf("checkRobotsTxt() = %v, want %v", allowed, tt.expected)
					}
				}
			}
		})
	}
}

func TestScraper_GetPageCount(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "test-output",
		LogFile:      "test.log",
	}

	logFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(logFile.Name())
	logFile.Close()
	cfg.LogFile = logFile.Name()

	scraper, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Initially should be 0
	if count := scraper.GetPageCount(); count != 0 {
		t.Errorf("GetPageCount() = %d, want 0", count)
	}

	// Add some test pages
	testPages := []PageData{
		{
			Title:     "Page 1",
			URL:       "https://example.com/page1",
			Content:   "Content 1",
			Timestamp: time.Now(),
			Depth:     1,
		},
		{
			Title:     "Page 2",
			URL:       "https://example.com/page2",
			Content:   "Content 2",
			Timestamp: time.Now(),
			Depth:     2,
		},
	}

	scraper.pages = testPages

	if count := scraper.GetPageCount(); count != 2 {
		t.Errorf("GetPageCount() = %d, want 2", count)
	}
}

func TestScraper_GetPages(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "test-output",
		LogFile:      "test.log",
	}

	logFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(logFile.Name())
	logFile.Close()
	cfg.LogFile = logFile.Name()

	scraper, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Initially should be empty
	pages := scraper.GetPages()
	if len(pages) != 0 {
		t.Errorf("GetPages() length = %d, want 0", len(pages))
	}

	// Add test pages
	testPages := []PageData{
		{
			Title:     "Test Page",
			URL:       "https://example.com/test",
			Content:   "Test Content",
			Timestamp: time.Now(),
			Depth:     1,
		},
	}

	scraper.pages = testPages
	pages = scraper.GetPages()

	if len(pages) != 1 {
		t.Errorf("GetPages() length = %d, want 1", len(pages))
	}

	if pages[0].Title != "Test Page" {
		t.Errorf("GetPages()[0].Title = %s, want Test Page", pages[0].Title)
	}
}

func TestPageData(t *testing.T) {
	now := time.Now()
	page := PageData{
		Title:     "Test Title",
		URL:       "https://example.com/test",
		Content:   "Test content here",
		Timestamp: now,
		Depth:     2,
	}

	if page.Title != "Test Title" {
		t.Errorf("PageData.Title = %s, want Test Title", page.Title)
	}
	if page.URL != "https://example.com/test" {
		t.Errorf("PageData.URL = %s, want https://example.com/test", page.URL)
	}
	if page.Content != "Test content here" {
		t.Errorf("PageData.Content = %s, want Test content here", page.Content)
	}
	if page.Timestamp != now {
		t.Errorf("PageData.Timestamp = %v, want %v", page.Timestamp, now)
	}
	if page.Depth != 2 {
		t.Errorf("PageData.Depth = %d, want 2", page.Depth)
	}
}

// Integration test helper
func createTestConfig() *config.Config {
	return &config.Config{
		RootURL:      "https://httpbin.org", // Public testing service
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "test-output",
		MinDelay:     1,
		MaxDelay:     2,
		MaxDepth:     1,
		LogFile:      "test.log",
		Verbose:      false,
	}
}

// Note: Integration tests that make real HTTP requests would go here
// but are commented out to avoid network dependencies in unit tests
/*
func TestScraper_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := createTestConfig()

	logFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(logFile.Name())
	logFile.Close()
	cfg.LogFile = logFile.Name()

	scraper, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// This would test actual scraping but requires network
	// err = scraper.Scrape()
	// if err != nil {
	//     t.Errorf("Scrape() error = %v", err)
	// }
}
*/
