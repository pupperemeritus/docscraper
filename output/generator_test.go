package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"docscraper/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		OutputDir:    "test-output",
		OutputFormat: "markdown",
		OutputType:   "single",
	}

	pages := []PageData{
		{
			Title:     "Test Page 1",
			URL:       "https://example.com/page1",
			Content:   "Content of page 1",
			Timestamp: time.Now(),
			Depth:     1,
		},
		{
			Title:     "Test Page 2",
			URL:       "https://example.com/page2",
			Content:   "Content of page 2",
			Timestamp: time.Now(),
			Depth:     2,
		},
	}

	generator := New(cfg, pages)

	if generator == nil {
		t.Fatal("New() returned nil generator")
	}

	if generator.config != cfg {
		t.Error("Generator config not set correctly")
	}

	if len(generator.pages) != len(pages) {
		t.Errorf("Expected %d pages, got %d", len(pages), len(generator.pages))
	}

	for i, page := range generator.pages {
		if page.Title != pages[i].Title {
			t.Errorf("Page %d title mismatch: expected %s, got %s", i, pages[i].Title, page.Title)
		}
		if page.URL != pages[i].URL {
			t.Errorf("Page %d URL mismatch: expected %s, got %s", i, pages[i].URL, page.URL)
		}
		if page.Content != pages[i].Content {
			t.Errorf("Page %d content mismatch: expected %s, got %s", i, pages[i].Content, page.Content)
		}
	}
}

func TestGenerator_Generate_MarkdownSingle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputDir:    tmpDir,
		OutputFormat: "markdown",
		OutputType:   "single",
	}

	pages := []PageData{
		{
			Title:     "Test Page 1",
			URL:       "https://example.com/page1",
			Content:   "Content of page 1",
			Timestamp: time.Now(),
			Depth:     1,
		},
		{
			Title:     "Test Page 2",
			URL:       "https://example.com/page2",
			Content:   "Content of page 2",
			Timestamp: time.Now(),
			Depth:     2,
		},
	}

	generator := New(cfg, pages)
	err = generator.Generate()
	if err != nil {
		t.Errorf("Generate() error = %v", err)
	}

	// Check if file was created
	outputFile := filepath.Join(tmpDir, "documentation.md")
	if !fileExists(outputFile) {
		t.Error("Output file was not created")
	}

	// Check file contents
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Documentation Scrape Results") {
		t.Error("Output file should contain header")
	}
	if !strings.Contains(contentStr, "Test Page 1") {
		t.Error("Output file should contain page 1 title")
	}
	if !strings.Contains(contentStr, "Test Page 2") {
		t.Error("Output file should contain page 2 title")
	}
	if !strings.Contains(contentStr, "Content of page 1") {
		t.Error("Output file should contain page 1 content")
	}
}

func TestGenerator_Generate_MarkdownPerPage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputDir:    tmpDir,
		OutputFormat: "markdown",
		OutputType:   "per-page",
	}

	pages := []PageData{
		{
			Title:     "Test Page 1",
			URL:       "https://example.com/page1",
			Content:   "Content of page 1",
			Timestamp: time.Now(),
			Depth:     1,
		},
		{
			Title:     "Test Page 2",
			URL:       "https://example.com/page2",
			Content:   "Content of page 2",
			Timestamp: time.Now(),
			Depth:     2,
		},
	}

	generator := New(cfg, pages)
	err = generator.Generate()
	if err != nil {
		t.Errorf("Generate() error = %v", err)
	}

	// Check if index file was created
	indexFile := filepath.Join(tmpDir, "index.md")
	if !fileExists(indexFile) {
		t.Error("Index file was not created")
	}

	// Check if individual page files were created
	expectedFiles := []string{"page_001.md", "page_002.md"}
	for _, filename := range expectedFiles {
		fullPath := filepath.Join(tmpDir, filename)
		if !fileExists(fullPath) {
			t.Errorf("Expected file %s was not created", filename)
		}
	}
}

func TestGenerator_Generate_JSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputDir:    tmpDir,
		OutputFormat: "json",
		OutputType:   "single",
	}

	testTime := time.Date(2023, 12, 1, 10, 30, 0, 0, time.UTC)
	pages := []PageData{
		{
			Title:     "Test Page 1",
			URL:       "https://example.com/page1",
			Content:   "Content of page 1",
			Timestamp: testTime,
			Depth:     1,
		},
	}

	generator := New(cfg, pages)
	err = generator.Generate()
	if err != nil {
		t.Errorf("Generate() error = %v", err)
	}

	// Check if file was created
	outputFile := filepath.Join(tmpDir, "documentation.json")
	if !fileExists(outputFile) {
		t.Error("Output file was not created")
	}

	// Check JSON content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	if err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	// Check structure
	if result["root_url"] != "https://example.com" {
		t.Error("JSON should contain root_url")
	}
	if result["total_pages"] != float64(1) {
		t.Error("JSON should contain correct total_pages")
	}

	pages_data, ok := result["pages"].([]interface{})
	if !ok {
		t.Error("JSON should contain pages array")
	}
	if len(pages_data) != 1 {
		t.Error("JSON should contain one page")
	}
}

func TestGenerator_Generate_Text(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputDir:    tmpDir,
		OutputFormat: "text",
		OutputType:   "single",
	}

	pages := []PageData{
		{
			Title:     "Test Page 1",
			URL:       "https://example.com/page1",
			Content:   "Content of page 1",
			Timestamp: time.Now(),
			Depth:     1,
		},
	}

	generator := New(cfg, pages)
	err = generator.Generate()
	if err != nil {
		t.Errorf("Generate() error = %v", err)
	}

	// Check if file was created
	outputFile := filepath.Join(tmpDir, "documentation.txt")
	if !fileExists(outputFile) {
		t.Error("Output file was not created")
	}

	// Check content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "DOCUMENTATION SCRAPE RESULTS") {
		t.Error("Text output should contain header")
	}
	if !strings.Contains(contentStr, "Test Page 1") {
		t.Error("Text output should contain page title")
	}
	if !strings.Contains(contentStr, "Content of page 1") {
		t.Error("Text output should contain page content")
	}
}

func TestGenerator_Generate_UnsupportedFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		OutputDir:    tmpDir,
		OutputFormat: "unsupported",
		OutputType:   "single",
	}

	generator := New(cfg, []PageData{})
	err = generator.Generate()
	if err == nil {
		t.Error("Expected error for unsupported format")
	}

	expectedError := "unsupported output format: unsupported"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestGenerator_Generate_CreateDirError(t *testing.T) {
	cfg := &config.Config{
		OutputDir:    "/root/forbidden", // Directory that can't be created
		OutputFormat: "markdown",
		OutputType:   "single",
	}

	generator := New(cfg, []PageData{})
	err := generator.Generate()
	if err == nil {
		t.Error("Expected error when creating forbidden directory")
	}
}

// Helper function
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
