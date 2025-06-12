package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docscraper/config"
	"docscraper/utils"
)

func TestConfigurationLoading(t *testing.T) {
	// Test YAML configuration loading
	yamlContent := `
root_url: "https://test.example.com"
output_format: "json"
output_type: "single"
output_dir: "test-output"
min_delay: 2
max_delay: 5
max_depth: 4
verbose: true
proxies:
  - "http://proxy.example.com:8080"
`
	tmpDir, err := os.MkdirTemp("", "test-config-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "test-config.yaml")
	err = os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		UserAgents: config.DefaultUserAgents,
	}

	err = config.LoadConfig(configFile, cfg)
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// Verify loaded values
	if cfg.RootURL != "https://test.example.com" {
		t.Errorf("Expected RootURL = https://test.example.com, got %s", cfg.RootURL)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("Expected OutputFormat = json, got %s", cfg.OutputFormat)
	}
	if cfg.MinDelay != 2 {
		t.Errorf("Expected MinDelay = 2, got %d", cfg.MinDelay)
	}
	if cfg.MaxDelay != 5 {
		t.Errorf("Expected MaxDelay = 5, got %d", cfg.MaxDelay)
	}
	if cfg.MaxDepth != 4 {
		t.Errorf("Expected MaxDepth = 4, got %d", cfg.MaxDepth)
	}
	if !cfg.Verbose {
		t.Error("Expected Verbose = true")
	}
	if len(cfg.Proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(cfg.Proxies))
	}
}

func TestUserAgentFileLoading(t *testing.T) {
	// Create test user agent file
	userAgents := []string{
		"Mozilla/5.0 (Test Browser 1)",
		"Mozilla/5.0 (Test Browser 2)",
		"Mozilla/5.0 (Test Browser 3)",
	}

	tmpDir, err := os.MkdirTemp("", "test-ua-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	uaFile := filepath.Join(tmpDir, "user-agents.txt")
	err = utils.WriteLinesToFile(uaFile, userAgents)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading
	loadedUAs, err := utils.LoadFileLines(uaFile)
	if err != nil {
		t.Errorf("LoadFileLines() error = %v", err)
	}

	if len(loadedUAs) != len(userAgents) {
		t.Errorf("Expected %d user agents, got %d", len(userAgents), len(loadedUAs))
	}

	for i, ua := range loadedUAs {
		if ua != userAgents[i] {
			t.Errorf("User agent %d: expected %s, got %s", i, userAgents[i], ua)
		}
	}
}

func TestDefaultConfiguration(t *testing.T) {
	// Test that default configuration has sensible values
	cfg := &config.Config{
		RootURL:       "https://example.com",
		OutputFormat:  "markdown",
		OutputType:    "per-page",
		OutputDir:     "output",
		MinDelay:      1,
		MaxDelay:      3,
		MaxDepth:      3,
		RespectRobots: true,
		LogFile:       "scraper.log",
		Verbose:       false,
		UserAgents:    config.DefaultUserAgents,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Default configuration should be valid: %v", err)
	}

	// Check default user agents
	if len(cfg.UserAgents) == 0 {
		t.Error("Default configuration should have user agents")
	}

	// Check default values for optional settings
	if cfg.GetConcurrentRequests() != 2 {
		t.Errorf("Expected default concurrent requests = 2, got %d", cfg.GetConcurrentRequests())
	}
}

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name          string
		modifyConfig  func(*config.Config)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config",
			modifyConfig: func(cfg *config.Config) {
				// No modifications - should be valid
			},
			expectError: false,
		},
		{
			name: "missing root URL",
			modifyConfig: func(cfg *config.Config) {
				cfg.RootURL = ""
			},
			expectError:   true,
			errorContains: "root_url is required",
		},
		{
			name: "invalid output format",
			modifyConfig: func(cfg *config.Config) {
				cfg.OutputFormat = "invalid"
			},
			expectError:   true,
			errorContains: "invalid output_format",
		},
		{
			name: "negative delay",
			modifyConfig: func(cfg *config.Config) {
				cfg.MinDelay = -1
			},
			expectError:   true,
			errorContains: "min_delay cannot be negative",
		},
		{
			name: "invalid delay range",
			modifyConfig: func(cfg *config.Config) {
				cfg.MinDelay = 10
				cfg.MaxDelay = 5
			},
			expectError:   true,
			errorContains: "max_delay must be greater than or equal to min_delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				RootURL:       "https://example.com",
				OutputFormat:  "markdown",
				OutputType:    "single",
				OutputDir:     "output",
				MinDelay:      1,
				MaxDelay:      3,
				MaxDepth:      3,
				RespectRobots: true,
				LogFile:       "scraper.log",
				Verbose:       false,
				UserAgents:    config.DefaultUserAgents,
			}

			tt.modifyConfig(cfg)

			err := cfg.Validate()
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestAdvancedConfigurationOptions(t *testing.T) {
	// Test advanced configuration with all optional features
	yamlContent := `
root_url: "https://advanced.example.com"
output_format: "markdown"
output_type: "single"
output_dir: "advanced-output"
min_delay: 1
max_delay: 2
max_depth: 5
verbose: true
proxies:
  - "http://proxy1.example.com:8080"
  - "http://proxy2.example.com:8080"
concurrent_requests: 5
request_timeout: 60
retry_attempts: 3
ignore_ssl_errors: true
user_agents:
  - "Custom User Agent 1"
  - "Custom User Agent 2"
`
	tmpDir, err := os.MkdirTemp("", "test-advanced-config-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "advanced-config.yaml")
	err = os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	err = config.LoadConfig(configFile, cfg)
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// Validate advanced options
	if !cfg.HasProxies() {
		t.Error("Config should have proxies")
	}
	if len(cfg.Proxies) != 2 {
		t.Errorf("Expected 2 proxies, got %d", len(cfg.Proxies))
	}
	if cfg.GetConcurrentRequests() != 5 {
		t.Errorf("Expected concurrent requests = 5, got %d", cfg.GetConcurrentRequests())
	}
	if len(cfg.UserAgents) != 2 {
		t.Errorf("Expected 2 user agents, got %d", len(cfg.UserAgents))
	}

	// Validate the configuration
	err = cfg.Validate()
	if err != nil {
		t.Errorf("Advanced configuration should be valid: %v", err)
	}
}
