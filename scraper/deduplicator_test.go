package scraper

import (
	"testing"
)

func TestNewLinkDeduplicator(t *testing.T) {
	config := URLNormalizer{
		RemoveFragment:  true,
		RemoveQuery:     false,
		RemoveTrailing:  true,
		LowerCase:       true,
		RemoveWWW:       true,
		SortQueryParams: true,
	}

	dedup := NewLinkDeduplicator(config)

	if dedup == nil {
		t.Fatal("NewLinkDeduplicator() returned nil")
	}

	if dedup.GetDuplicateCount() != 0 {
		t.Errorf("Expected duplicate count to be 0, got %d", dedup.GetDuplicateCount())
	}

	if dedup.GetSeenURLsCount() != 0 {
		t.Errorf("Expected seen URLs count to be 0, got %d", dedup.GetSeenURLsCount())
	}
}

func TestLinkDeduplicator_NormalizeURL(t *testing.T) {
	tests := []struct {
		name       string
		config     URLNormalizer
		input      string
		expected   string
		shouldFail bool
	}{
		{
			name: "remove fragment",
			config: URLNormalizer{
				RemoveFragment: true,
			},
			input:    "https://example.com/page#section",
			expected: "https://example.com/page",
		},
		{
			name: "remove query",
			config: URLNormalizer{
				RemoveQuery: true,
			},
			input:    "https://example.com/page?param=value",
			expected: "https://example.com/page",
		},
		{
			name: "remove trailing slash",
			config: URLNormalizer{
				RemoveTrailing: true,
			},
			input:    "https://example.com/page/",
			expected: "https://example.com/page",
		},
		{
			name: "keep root slash",
			config: URLNormalizer{
				RemoveTrailing: true,
			},
			input:    "https://example.com/",
			expected: "https://example.com/",
		},
		{
			name: "lowercase",
			config: URLNormalizer{
				LowerCase: true,
			},
			input:    "https://Example.Com/PAGE",
			expected: "https://example.com/page",
		},
		{
			name: "remove www",
			config: URLNormalizer{
				RemoveWWW: true,
			},
			input:    "https://www.example.com/page",
			expected: "https://example.com/page",
		},
		{
			name: "sort query params",
			config: URLNormalizer{
				SortQueryParams: true,
			},
			input:    "https://example.com/page?z=1&a=2&m=3",
			expected: "https://example.com/page?a=2&m=3&z=1",
		},
		{
			name: "all normalizations",
			config: URLNormalizer{
				RemoveFragment:  true,
				RemoveQuery:     false,
				RemoveTrailing:  true,
				LowerCase:       true,
				RemoveWWW:       true,
				SortQueryParams: true,
			},
			input:    "https://WWW.Example.Com/Page/?z=1&a=2#section",
			expected: "https://example.com/page?a=2&z=1",
		},
		{
			name: "invalid URL",
			config: URLNormalizer{
				LowerCase: true,
			},
			input:      "://invalid",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dedup := NewLinkDeduplicator(tt.config)
			result, err := dedup.NormalizeURL(tt.input)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestLinkDeduplicator_AddURL(t *testing.T) {
	config := URLNormalizer{
		RemoveFragment: true,
		LowerCase:      true,
	}
	dedup := NewLinkDeduplicator(config)

	// First addition should succeed
	if !dedup.AddURL("https://example.com/page#section") {
		t.Error("First AddURL() should return true")
	}

	if dedup.GetSeenURLsCount() != 1 {
		t.Errorf("Expected 1 seen URL, got %d", dedup.GetSeenURLsCount())
	}

	if dedup.GetDuplicateCount() != 0 {
		t.Errorf("Expected 0 duplicates, got %d", dedup.GetDuplicateCount())
	}

	// Second addition of normalized equivalent should fail
	if dedup.AddURL("https://EXAMPLE.com/page") {
		t.Error("Second AddURL() of equivalent URL should return false")
	}

	if dedup.GetSeenURLsCount() != 1 {
		t.Errorf("Expected 1 seen URL after duplicate, got %d", dedup.GetSeenURLsCount())
	}

	if dedup.GetDuplicateCount() != 1 {
		t.Errorf("Expected 1 duplicate, got %d", dedup.GetDuplicateCount())
	}

	// Different URL should succeed
	if !dedup.AddURL("https://example.com/other") {
		t.Error("AddURL() of different URL should return true")
	}

	if dedup.GetSeenURLsCount() != 2 {
		t.Errorf("Expected 2 seen URLs, got %d", dedup.GetSeenURLsCount())
	}
}

func TestLinkDeduplicator_IsDuplicate(t *testing.T) {
	config := URLNormalizer{
		RemoveFragment: true,
		LowerCase:      true,
	}
	dedup := NewLinkDeduplicator(config)

	// Initially should not be duplicate
	if dedup.IsDuplicate("https://example.com/page") {
		t.Error("IsDuplicate() should return false for unseen URL")
	}

	// Add URL
	dedup.AddURL("https://example.com/page#section")

	// Should now be duplicate (normalized equivalent)
	if !dedup.IsDuplicate("https://EXAMPLE.com/page") {
		t.Error("IsDuplicate() should return true for normalized equivalent")
	}

	// Different URL should not be duplicate
	if dedup.IsDuplicate("https://example.com/other") {
		t.Error("IsDuplicate() should return false for different URL")
	}
}

func TestLinkDeduplicator_GetCanonicalURL(t *testing.T) {
	config := URLNormalizer{
		RemoveFragment: true,
		LowerCase:      true,
	}
	dedup := NewLinkDeduplicator(config)

	originalURL := "https://EXAMPLE.com/page#section"
	expectedCanonical := "https://example.com/page"

	// Add URL
	dedup.AddURL(originalURL)

	// Get canonical version
	canonical := dedup.GetCanonicalURL(originalURL)
	if canonical != expectedCanonical {
		t.Errorf("GetCanonicalURL(%q) = %q, want %q", originalURL, canonical, expectedCanonical)
	}

	// Test URL not in map
	unknownURL := "https://other.com/page"
	canonical = dedup.GetCanonicalURL(unknownURL)
	expectedCanonical = "https://other.com/page"
	if canonical != expectedCanonical {
		t.Errorf("GetCanonicalURL(%q) = %q, want %q", unknownURL, canonical, expectedCanonical)
	}
}

func TestLinkDeduplicator_Reset(t *testing.T) {
	config := URLNormalizer{LowerCase: true}
	dedup := NewLinkDeduplicator(config)

	// Add some URLs
	dedup.AddURL("https://example.com/page1")
	dedup.AddURL("https://example.com/page1") // duplicate
	dedup.AddURL("https://example.com/page2")

	// Verify state before reset
	if dedup.GetSeenURLsCount() != 2 {
		t.Errorf("Expected 2 seen URLs before reset, got %d", dedup.GetSeenURLsCount())
	}
	if dedup.GetDuplicateCount() != 1 {
		t.Errorf("Expected 1 duplicate before reset, got %d", dedup.GetDuplicateCount())
	}

	// Reset
	dedup.Reset()

	// Verify state after reset
	if dedup.GetSeenURLsCount() != 0 {
		t.Errorf("Expected 0 seen URLs after reset, got %d", dedup.GetSeenURLsCount())
	}
	if dedup.GetDuplicateCount() != 0 {
		t.Errorf("Expected 0 duplicates after reset, got %d", dedup.GetDuplicateCount())
	}

	// Should be able to add the same URL again
	if !dedup.AddURL("https://example.com/page1") {
		t.Error("AddURL() should succeed after reset")
	}
}
