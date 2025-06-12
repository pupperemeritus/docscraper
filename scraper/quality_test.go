package scraper

import (
	"strings"
	"testing"
	"time"
)

func TestNewContentQualityAnalyzer(t *testing.T) {
	config := QualityConfig{
		MinWordCount:        50,
		MinCodeBlockCount:   1,
		MaxEmptyLineRatio:   0.5,
		RequireTitle:        true,
		RequireContent:      true,
		SkipNavigationPages: true,
		MinContentRatio:     0.3,
		BlacklistPatterns:   []string{"error", "404"},
		WhitelistPatterns:   []string{"documentation", "guide"},
	}

	analyzer := NewContentQualityAnalyzer(config)
	if analyzer == nil {
		t.Fatal("NewContentQualityAnalyzer() returned nil")
	}

	if analyzer.config.MinWordCount != 50 {
		t.Errorf("Expected MinWordCount 50, got %d", analyzer.config.MinWordCount)
	}

	if analyzer.scorer.weights.WordCount != 0.3 {
		t.Errorf("Expected WordCount weight 0.3, got %f", analyzer.scorer.weights.WordCount)
	}
}

func TestContentQualityAnalyzer_AnalyzeContent(t *testing.T) {
	config := QualityConfig{
		MinWordCount:      50,
		RequireTitle:      true,
		RequireContent:    true,
		BlacklistPatterns: []string{"error"},
	}

	analyzer := NewContentQualityAnalyzer(config)

	// Test high-quality content
	goodContent := ScrapedContent{
		URL:   "https://example.com/docs/guide",
		Title: "Complete Guide to API Usage",
		Content: `# API Guide

This is a comprehensive guide to using our API. It contains detailed explanations,
code examples, and best practices for developers.

## Getting Started

To get started with our API, you'll need to obtain an API key from your dashboard.
The API key should be included in all requests as a header.

## Code Example

Here's a simple example of making a request:

` + "```" + `javascript
const response = await fetch('https://api.example.com/users', {
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  }
});
const data = await response.json();
console.log(data);
` + "```" + `

## Error Handling

Always handle errors appropriately in your applications. The API returns
standard HTTP status codes and error messages in JSON format.

## Best Practices

1. Always validate input data
2. Implement proper error handling
3. Use appropriate HTTP methods
4. Cache responses when possible
5. Respect rate limits

This guide covers the most important aspects of API integration and should
help you get started quickly with your development.`,
		Metadata: NodeMetadata{
			LastModified: time.Now(),
			ContentType:  "text/html",
		},
	}

	quality := analyzer.AnalyzeContent(goodContent)

	// Test quality metrics
	if quality.WordCount < 50 {
		t.Errorf("Expected word count >= 50, got %d", quality.WordCount)
	}

	if quality.CodeBlockCount < 1 {
		t.Errorf("Expected at least 1 code block, got %d", quality.CodeBlockCount)
	}

	if !quality.HasTitle {
		t.Error("Expected HasTitle to be true")
	}

	if !quality.HasHeaders {
		t.Error("Expected HasHeaders to be true")
	}

	if quality.Score < 0.5 {
		t.Errorf("Expected quality score >= 0.5 for good content, got %f", quality.Score)
	}

	if quality.Language != "en" {
		t.Errorf("Expected language 'en', got %s", quality.Language)
	}

	// Test low-quality content
	badContent := ScrapedContent{
		URL:     "https://example.com/empty",
		Title:   "",
		Content: "Error 404",
		Metadata: NodeMetadata{
			LastModified: time.Now(),
			ContentType:  "text/html",
		},
	}

	badQuality := analyzer.AnalyzeContent(badContent)

	if badQuality.Score > 0.3 {
		t.Errorf("Expected low quality score for bad content, got %f", badQuality.Score)
	}

	if len(badQuality.Issues) == 0 {
		t.Error("Expected quality issues for bad content")
	}
}

func TestContentQualityAnalyzer_countWords(t *testing.T) {
	analyzer := NewContentQualityAnalyzer(QualityConfig{})

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "simple text",
			text:     "Hello world this is a test",
			expected: 6,
		},
		{
			name:     "text with punctuation",
			text:     "Hello, world! This is a test.",
			expected: 6,
		},
		{
			name:     "empty text",
			text:     "",
			expected: 0,
		},
		{
			name:     "text with extra spaces",
			text:     "  Hello   world  this   is  a  test  ",
			expected: 6,
		},
		{
			name:     "text with newlines",
			text:     "Hello world\nthis is\na test",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.countWords(tt.text)
			if result != tt.expected {
				t.Errorf("countWords(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestContentQualityAnalyzer_countCodeBlocks(t *testing.T) {
	analyzer := NewContentQualityAnalyzer(QualityConfig{})

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single code block",
			text:     "Here's some code:\n```\nfunction test() {\n  return true;\n}\n```",
			expected: 1,
		},
		{
			name:     "multiple code blocks",
			text:     "```javascript\nconst a = 1;\n```\n\nAnd also:\n```python\nprint('hello')\n```",
			expected: 2,
		},
		{
			name:     "inline code",
			text:     "Use `console.log()` and `alert()` and `document.write()` for output",
			expected: 1, // 3 inline codes = 1 block equivalent
		},
		{
			name:     "no code",
			text:     "This is just regular text with no code",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.countCodeBlocks(tt.text)
			if result != tt.expected {
				t.Errorf("countCodeBlocks(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestContentQualityAnalyzer_countHeaders(t *testing.T) {
	analyzer := NewContentQualityAnalyzer(QualityConfig{})

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single header",
			text:     "# Main Title\nSome content here",
			expected: 1,
		},
		{
			name:     "multiple headers",
			text:     "# Title\n## Subtitle\n### Sub-subtitle\nContent",
			expected: 3,
		},
		{
			name:     "no headers",
			text:     "Just some regular text without any headers",
			expected: 0,
		},
		{
			name:     "headers with content",
			text:     "# Introduction\nThis is the intro.\n## Getting Started\nHere's how to start.",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.countHeaders(tt.text)
			if result != tt.expected {
				t.Errorf("countHeaders(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestContentQualityAnalyzer_IsNavigationPage(t *testing.T) {
	analyzer := NewContentQualityAnalyzer(QualityConfig{})

	tests := []struct {
		name     string
		content  ScrapedContent
		expected bool
	}{
		{
			name: "clear navigation page",
			content: ScrapedContent{
				Title:   "Table of Contents",
				Content: "This page contains links to all documentation sections.",
			},
			expected: true,
		},
		{
			name: "content page",
			content: ScrapedContent{
				Title:   "API Reference",
				Content: "This is a detailed guide about how to use our API with examples and explanations.",
			},
			expected: false,
		},
		{
			name: "index page",
			content: ScrapedContent{
				Title:   "Documentation Index",
				Content: "Welcome to our documentation index with navigation links.",
			},
			expected: true,
		},
		{
			name: "high link ratio page",
			content: ScrapedContent{
				Title:   "Quick Links",
				Content: "[Guide 1](link1) [Guide 2](link2) [Guide 3](link3) [Guide 4](link4) Text",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.IsNavigationPage(tt.content)
			if result != tt.expected {
				t.Errorf("IsNavigationPage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContentQualityAnalyzer_DetectIssues(t *testing.T) {
	config := QualityConfig{
		MinWordCount:      50,
		RequireTitle:      true,
		RequireContent:    true,
		BlacklistPatterns: []string{"error", "404"},
	}

	analyzer := NewContentQualityAnalyzer(config)

	tests := []struct {
		name           string
		content        ScrapedContent
		expectedIssues int
	}{
		{
			name: "perfect content",
			content: ScrapedContent{
				Title:   "Great Title",
				Content: strings.Repeat("This is great content with many words. ", 20),
			},
			expectedIssues: 0,
		},
		{
			name: "missing title",
			content: ScrapedContent{
				Title:   "",
				Content: strings.Repeat("Content without title. ", 20),
			},
			expectedIssues: 1,
		},
		{
			name: "short content",
			content: ScrapedContent{
				Title:   "Good Title",
				Content: "Short content",
			},
			expectedIssues: 1,
		},
		{
			name: "blacklisted content",
			content: ScrapedContent{
				Title:   "Error Page",
				Content: "Error 404 - page not found",
			},
			expectedIssues: 2, // Short content + blacklisted pattern
		},
		{
			name: "empty content",
			content: ScrapedContent{
				Title:   "Title",
				Content: "",
			},
			expectedIssues: 2, // Empty content + word count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := analyzer.DetectIssues(tt.content)
			if len(issues) != tt.expectedIssues {
				t.Errorf("DetectIssues() found %d issues, want %d", len(issues), tt.expectedIssues)
			}
		})
	}
}

func TestQualityScorer_CalculateScore(t *testing.T) {
	weights := QualityWeights{
		WordCount:     0.3,
		CodeBlocks:    0.2,
		Images:        0.1,
		Headers:       0.15,
		ContentRatio:  0.15,
		TitlePresence: 0.1,
	}

	scorer := QualityScorer{weights: weights}

	tests := []struct {
		name     string
		metrics  ContentMetrics
		minScore float64
		maxScore float64
	}{
		{
			name: "high quality content",
			metrics: ContentMetrics{
				WordCount:      500,
				CodeBlockCount: 3,
				HeaderCount:    3,
				ContentRatio:   0.8,
				HasTitle:       true,
			},
			minScore: 0.8,
			maxScore: 1.0,
		},
		{
			name: "medium quality content",
			metrics: ContentMetrics{
				WordCount:      200,
				CodeBlockCount: 1,
				HeaderCount:    2,
				ContentRatio:   0.6,
				HasTitle:       true,
			},
			minScore: 0.4,
			maxScore: 0.7,
		},
		{
			name: "low quality content",
			metrics: ContentMetrics{
				WordCount:      50,
				CodeBlockCount: 0,
				HeaderCount:    0,
				ContentRatio:   0.2,
				HasTitle:       false,
			},
			minScore: 0.0,
			maxScore: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.CalculateScore(tt.metrics)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("CalculateScore() = %f, want between %f and %f", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestContentQualityAnalyzer_ShouldSkip(t *testing.T) {
	config := QualityConfig{
		SkipNavigationPages: true,
	}

	analyzer := NewContentQualityAnalyzer(config)

	tests := []struct {
		name     string
		quality  ContentQuality
		expected bool
	}{
		{
			name: "high quality content",
			quality: ContentQuality{
				Score:            0.8,
				IsNavigationPage: false,
				Issues:           []QualityIssue{},
			},
			expected: false,
		},
		{
			name: "low quality content",
			quality: ContentQuality{
				Score:            0.2,
				IsNavigationPage: false,
				Issues:           []QualityIssue{},
			},
			expected: true,
		},
		{
			name: "navigation page",
			quality: ContentQuality{
				Score:            0.8,
				IsNavigationPage: true,
				Issues:           []QualityIssue{},
			},
			expected: true,
		},
		{
			name: "content with errors",
			quality: ContentQuality{
				Score:            0.8,
				IsNavigationPage: false,
				Issues: []QualityIssue{
					{Type: "missing_title", Severity: "error"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.ShouldSkip(tt.quality)
			if result != tt.expected {
				t.Errorf("ShouldSkip() = %v, want %v", result, tt.expected)
			}
		})
	}
}
