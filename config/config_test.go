package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid basic config",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
			},
			wantErr: false,
		},
		{
			name: "missing root URL",
			config: Config{
				OutputFormat: "markdown",
				OutputType:   "single",
			},
			wantErr: true,
			errMsg:  "root_url is required",
		},
		{
			name: "invalid root URL",
			config: Config{
				RootURL:      "not-a-valid-url",
				OutputFormat: "markdown",
				OutputType:   "single",
			},
			wantErr: true,
			errMsg:  "invalid root_url",
		},
		{
			name: "negative min delay",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				MinDelay:     -1,
				MaxDelay:     2,
			},
			wantErr: true,
			errMsg:  "min_delay cannot be negative",
		},
		{
			name: "max delay less than min delay",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				MinDelay:     5,
				MaxDelay:     3,
			},
			wantErr: true,
			errMsg:  "max_delay must be greater than or equal to min_delay",
		},
		{
			name: "negative max depth",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     -1,
			},
			wantErr: true,
			errMsg:  "max_depth cannot be negative",
		},
		{
			name: "invalid output format",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "invalid",
				OutputType:   "single",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
			},
			wantErr: true,
			errMsg:  "invalid output_format",
		},
		{
			name: "invalid output type",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "invalid",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
			},
			wantErr: true,
			errMsg:  "invalid output_type",
		},
		{
			name: "invalid proxy URL",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
				Proxies:      []string{"not-a-valid-proxy"},
			},
			wantErr: true,
			errMsg:  "invalid proxy URL",
		},
		{
			name: "valid proxy URLs",
			config: Config{
				RootURL:      "https://example.com",
				OutputFormat: "markdown",
				OutputType:   "single",
				MinDelay:     1,
				MaxDelay:     2,
				MaxDepth:     3,
				Proxies:      []string{"http://proxy.example.com:8080", "socks5://127.0.0.1:1080"},
			},
			wantErr: false,
		},
		{
			name: "invalid concurrent requests",
			config: Config{
				RootURL:            "https://example.com",
				OutputFormat:       "markdown",
				OutputType:         "single",
				MinDelay:           1,
				MaxDelay:           2,
				MaxDepth:           3,
				ConcurrentRequests: intPtr(0),
			},
			wantErr: true,
			errMsg:  "concurrent_requests must be greater than 0",
		},
		{
			name: "invalid request timeout",
			config: Config{
				RootURL:        "https://example.com",
				OutputFormat:   "markdown",
				OutputType:     "single",
				MinDelay:       1,
				MaxDelay:       2,
				MaxDepth:       3,
				RequestTimeout: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "request_timeout must be greater than 0",
		},
		{
			name: "invalid retry attempts",
			config: Config{
				RootURL:       "https://example.com",
				OutputFormat:  "markdown",
				OutputType:    "single",
				MinDelay:      1,
				MaxDelay:      2,
				MaxDepth:      3,
				RetryAttempts: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "retry_attempts cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestConfig_GetConcurrentRequests(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   int
	}{
		{
			name:   "default value when nil",
			config: Config{},
			want:   2,
		},
		{
			name: "custom value",
			config: Config{
				ConcurrentRequests: intPtr(5),
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetConcurrentRequests(); got != tt.want {
				t.Errorf("Config.GetConcurrentRequests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetRequestTimeout(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   int
	}{
		{
			name:   "default value when nil",
			config: Config{},
			want:   30,
		},
		{
			name: "custom value",
			config: Config{
				RequestTimeout: intPtr(60),
			},
			want: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetRequestTimeout(); got != tt.want {
				t.Errorf("Config.GetRequestTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetRetryAttempts(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   int
	}{
		{
			name:   "default value when nil",
			config: Config{},
			want:   0,
		},
		{
			name: "custom value",
			config: Config{
				RetryAttempts: intPtr(3),
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetRetryAttempts(); got != tt.want {
				t.Errorf("Config.GetRetryAttempts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetIgnoreSSLErrors(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "default value when nil",
			config: Config{},
			want:   false,
		},
		{
			name: "custom value true",
			config: Config{
				IgnoreSSLErrors: boolPtr(true),
			},
			want: true,
		},
		{
			name: "custom value false",
			config: Config{
				IgnoreSSLErrors: boolPtr(false),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetIgnoreSSLErrors(); got != tt.want {
				t.Errorf("Config.GetIgnoreSSLErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_HasProxies(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "no proxies",
			config: Config{},
			want:   false,
		},
		{
			name: "empty proxies slice",
			config: Config{
				Proxies: []string{},
			},
			want: false,
		},
		{
			name: "has proxies",
			config: Config{
				Proxies: []string{"http://proxy.example.com:8080"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.HasProxies(); got != tt.want {
				t.Errorf("Config.HasProxies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test YAML loading
	yamlContent := `
root_url: "https://example.com"
output_format: "markdown"
output_type: "single"
min_delay: 1
max_delay: 2
max_depth: 3
`
	yamlFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(yamlFile.Name())

	if _, err := yamlFile.WriteString(yamlContent); err != nil {
		t.Fatal(err)
	}
	yamlFile.Close()

	var cfg Config
	err = LoadConfig(yamlFile.Name(), &cfg)
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	if cfg.RootURL != "https://example.com" {
		t.Errorf("Expected RootURL = https://example.com, got %s", cfg.RootURL)
	}
	if cfg.OutputFormat != "markdown" {
		t.Errorf("Expected OutputFormat = markdown, got %s", cfg.OutputFormat)
	}

	// Test JSON loading
	jsonContent := `{
		"root_url": "https://json.example.com",
		"output_format": "json",
		"output_type": "per-page",
		"min_delay": 2,
		"max_delay": 4,
		"max_depth": 5
	}`
	jsonFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(jsonFile.Name())

	if _, err := jsonFile.WriteString(jsonContent); err != nil {
		t.Fatal(err)
	}
	jsonFile.Close()

	var cfg2 Config
	err = LoadConfig(jsonFile.Name(), &cfg2)
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	if cfg2.RootURL != "https://json.example.com" {
		t.Errorf("Expected RootURL = https://json.example.com, got %s", cfg2.RootURL)
	}
	if cfg2.OutputFormat != "json" {
		t.Errorf("Expected OutputFormat = json, got %s", cfg2.OutputFormat)
	}

	// Test non-existent file
	var cfg3 Config
	err = LoadConfig("non-existent-file.yaml", &cfg3)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: slice,
			item:  "banana",
			want:  true,
		},
		{
			name:  "item does not exist",
			slice: slice,
			item:  "grape",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "apple",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.slice, tt.item); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions for pointer creation
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
