package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestNewContentExtractor(t *testing.T) {
	extractor := NewContentExtractor()

	if extractor == nil {
		t.Fatal("NewContentExtractor() returned nil")
	}

	if len(extractor.contentSelectors) == 0 {
		t.Error("ContentExtractor should have content selectors")
	}

	if len(extractor.removeSelectors) == 0 {
		t.Error("ContentExtractor should have remove selectors")
	}

	// Check that some expected selectors are present
	expectedContentSelectors := []string{"main", ".content", "article"}
	for _, expected := range expectedContentSelectors {
		found := false
		for _, selector := range extractor.contentSelectors {
			if selector == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected content selector %s not found", expected)
		}
	}

	expectedRemoveSelectors := []string{"script", "style", "nav"}
	for _, expected := range expectedRemoveSelectors {
		found := false
		for _, selector := range extractor.removeSelectors {
			if selector == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected remove selector %s not found", expected)
		}
	}
}

func TestContentExtractor_ExtractTitle(t *testing.T) {
	extractor := NewContentExtractor()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "title tag",
			html:     `<html><head><title>Test Title</title></head><body></body></html>`,
			expected: "Test Title",
		},
		{
			name:     "h1 tag",
			html:     `<html><body><h1>Main Heading</h1><p>Content</p></body></html>`,
			expected: "Main Heading",
		},
		{
			name:     "page-title class",
			html:     `<html><body><div class="page-title">Page Title</div></body></html>`,
			expected: "Page Title",
		},
		{
			name:     "data-title attribute",
			html:     `<html><body><div data-title="Data Title">Content</div></body></html>`,
			expected: "Data Title",
		},
		{
			name:     "no title found",
			html:     `<html><body><p>Just content</p></body></html>`,
			expected: "Untitled",
		},
		{
			name:     "title with whitespace",
			html:     `<html><head><title>   Spaced Title   </title></head></html>`,
			expected: "Spaced Title",
		},
		{
			name:     "multiple h1 tags - first one wins",
			html:     `<html><body><h1>First H1</h1><h1>Second H1</h1></body></html>`,
			expected: "First H1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := extractor.ExtractTitle(doc.Selection)
			if result != tt.expected {
				t.Errorf("ExtractTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContentExtractor_ExtractContent(t *testing.T) {
	extractor := NewContentExtractor()

	tests := []struct {
		name     string
		html     string
		expected string
		contains []string
		notContains []string
	}{
		{
			name: "main content area",
			html: `<html><body>
				<nav>Navigation</nav>
				<main>Main content here</main>
				<footer>Footer content</footer>
			</body></html>`,
			expected: "Main content here",
		},
		{
			name: "content class",
			html: `<html><body>
				<div class="sidebar">Sidebar</div>
				<div class="content">Article content</div>
			</body></html>`,
			expected: "Article content",
		},
		{
			name: "article tag",
			html: `<html><body>
				<header>Header</header>
				<article>Article text content</article>
			</body></html>`,
			expected: "Article text content",
		},
		{
			name: "removes scripts and styles",
			html: `<html><body>
				<script>alert('test');</script>
				<style>body { color: red; }</style>
				<main>Clean content</main>
			</body></html>`,
			expected: "Clean content",
		},
		{
			name: "fallback to body",
			html: `<html><body>
				<p>Just body content</p>
			</body></html>`,
			expected: "Just body content",
		},
		{
			name: "removes navigation elements",
			html: `<html><body>
				<nav class="navigation">Nav content</nav>
				<div class="sidebar">Sidebar</div>
				<main>Main content only</main>
			</body></html>`,
			expected: "Main content only",
		},
		{
			name: "substantial content check",
			html: `<html><body>
				<div class="content">Short</div>
				<article>This is a much longer piece of content that should be substantial enough to be considered the main content of the page</article>
			</body></html>`,
			contains: []string{"much longer piece"},
			notContains: []string{"Short"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := extractor.ExtractContent(doc.Selection)
			result = strings.TrimSpace(result)

			if tt.expected != "" && result != tt.expected {
				t.Errorf("ExtractContent() = %q, want %q", result, tt.expected)
			}

			for _, contains := range tt.contains {
				if !strings.Contains(result, contains) {
					t.Errorf("ExtractContent() result should contain %q, got %q", contains, result)
				}
			}

			for _, notContains := range tt.notContains {
				if strings.Contains(result, notContains) {
					t.Errorf("ExtractContent() result should not contain %q, got %q", notContains, result)
				}
			}
		})
	}
}

func TestContentExtractor_cleanText(t *testing.T) {
	extractor := NewContentExtractor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove excessive whitespace",
			input:    "Text   with    lots\n\n\nof   whitespace",
			expected: "Text with lots of whitespace",
		},
		{
			name:     "trim leading and trailing whitespace",
			input:    "   Text with spaces   ",
			expected: "Text with spaces",
		},
		{
			name:     "remove skip to content",
			input:    "Skip to main content This is the actual content",
			expected: "This is the actual content",
		},
		{
			name:     "remove click here patterns",
			input:    "Content here Click here to subscribe More content",
			expected: "Content here More content",
		},
		{
			name:     "remove subscribe patterns",
			input:    "Main content Subscribe to our newsletter End content",
			expected: "Main content End content",
		},
		{
			name:     "remove follow us patterns",
			input:    "Article text Follow us on Twitter Final text",
			expected: "Article text Final text",
		},
		{
			name:     "case insensitive pattern matching",
			input:    "Text SKIP TO MAIN CONTENT more text",
			expected: "more text",
		},
		{
			name:     "no changes needed",
			input:    "Clean text content",
			expected: "Clean text content",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText() = %q, want %q", result, tt.expected)
			}
		})
	}
}
