package scraper

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ContentExtractor handles content extraction from HTML
type ContentExtractor struct {
	// Content selectors in order of preference
	contentSelectors []string
	// Selectors to remove from content
	removeSelectors []string
}

// NewContentExtractor creates a new content extractor
func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{
		contentSelectors: []string{
			"main", ".main", "#main",
			".content", "#content", ".main-content",
			"article", ".article",
			".documentation", ".docs", ".doc-content",
			".post-content", ".entry-content",
			".page-content", ".body-content",
			".markdown-body", ".wiki-content",
		},
		removeSelectors: []string{
			"script", "style", "nav", "footer", "header",
			".navigation", ".sidebar", ".menu", ".breadcrumb",
			".toc", ".table-of-contents",
			".related", ".tags", ".metadata",
			".comments", ".social-share",
			".advertisement", ".ads",
		},
	}
}

// ExtractTitle extracts the page title
func (e *ContentExtractor) ExtractTitle(doc *goquery.Selection) string {
	// Try different title sources
	titleSelectors := []string{
		"title",
		"h1",
		".page-title", ".main-title", ".doc-title",
		"[data-title]",
	}

	for _, selector := range titleSelectors {
		if selector == "[data-title]" {
			if element := doc.Find(selector).First(); element.Length() > 0 {
				if title, exists := element.Attr("data-title"); exists && title != "" {
					return strings.TrimSpace(title)
				}
			}
		} else {
			if title := doc.Find(selector).First().Text(); title != "" {
				return strings.TrimSpace(title)
			}
		}
	}

	return "Untitled"
}

// ExtractContent extracts clean text content from HTML
func (e *ContentExtractor) ExtractContent(doc *goquery.Selection) string {
	// Remove unwanted elements
	for _, selector := range e.removeSelectors {
		doc.Find(selector).Remove()
	}

	// Try to find main content areas
	var content string
	for _, selector := range e.contentSelectors {
		if contentEl := doc.Find(selector); contentEl.Length() > 0 {
			content = contentEl.Text()
			if len(strings.TrimSpace(content)) > 100 { // Ensure substantial content
				break
			}
		}
	}

	// Fallback to body if no main content found
	if content == "" {
		content = doc.Find("body").Text()
	}

	// Clean up the content
	return e.cleanText(content)
}

// cleanText cleans and normalizes extracted text
func (e *ContentExtractor) cleanText(text string) string {
	// Remove common noise patterns first (before whitespace normalization)
	noisePatterns := []string{
		`(?i).*?Skip to .*?content\s*`,
		`(?i)Click here to \w+\s*`,
		`(?i)Subscribe to \w+ \w+\s*`,
		`(?i)Follow us on \w+\s*`,
	}

	for _, pattern := range noisePatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	// Remove excessive whitespace
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Remove leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}
