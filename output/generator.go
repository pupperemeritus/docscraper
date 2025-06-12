package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"docscraper/config"

	"gopkg.in/yaml.v2"
)

// PageData represents scraped page information (copied to avoid import cycle)
type PageData struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Depth     int       `json:"depth"`
}

// Generator handles output generation
type Generator struct {
	config *config.Config
	pages  []PageData
}

// New creates a new output generator
func New(cfg *config.Config, pages []PageData) *Generator {
	// Convert scraper.PageData to output.PageData
	outputPages := make([]PageData, len(pages))
	for i, page := range pages {
		outputPages[i] = PageData{
			Title:     page.Title,
			URL:       page.URL,
			Content:   page.Content,
			Timestamp: page.Timestamp,
			Depth:     page.Depth,
		}
	}

	return &Generator{
		config: cfg,
		pages:  outputPages,
	}
}

// Generate creates the output files based on configuration
func (g *Generator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	switch g.config.OutputFormat {
	case "markdown":
		return g.generateMarkdownOutput()
	case "text":
		return g.generateTextOutput()
	case "json":
		return g.generateJSONOutput()
	default:
		return fmt.Errorf("unsupported output format: %s", g.config.OutputFormat)
	}
}

// generateMarkdownOutput generates Markdown output
func (g *Generator) generateMarkdownOutput() error {
	if g.config.OutputType == "single" {
		return g.generateSingleMarkdown()
	}
	return g.generatePerPageMarkdown()
}

// generateSingleMarkdown creates a single Markdown file with all content
func (g *Generator) generateSingleMarkdown() error {
	filename := filepath.Join(g.config.OutputDir, "documentation.md")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "# Documentation Scrape Results\n\n")
	fmt.Fprintf(file, "**Scraped from:** %s  \n", g.config.RootURL)
	fmt.Fprintf(file, "**Generated:** %s  \n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "**Total Pages:** %d\n\n", len(g.pages))
	fmt.Fprintf(file, "---\n\n")

	// Write table of contents
	fmt.Fprintf(file, "## Table of Contents\n\n")
	for i, page := range g.pages {
		anchor := g.createAnchor(page.Title)
		fmt.Fprintf(file, "%d. [%s](#%s)\n", i+1, page.Title, anchor)
	}
	fmt.Fprintf(file, "\n---\n\n")

	// Write each page
	for i, page := range g.pages {
		anchor := g.createAnchor(page.Title)
		fmt.Fprintf(file, "## %s {#%s}\n\n", page.Title, anchor)
		fmt.Fprintf(file, "**URL:** %s  \n", page.URL)
		fmt.Fprintf(file, "**Scraped:** %s\n\n", page.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(file, "%s\n\n", page.Content)

		if i < len(g.pages)-1 {
			fmt.Fprintf(file, "---\n\n")
		}
	}

	return nil
}

// generatePerPageMarkdown creates separate Markdown files for each page
func (g *Generator) generatePerPageMarkdown() error {
	// Create individual page files
	for i, page := range g.pages {
		// Create numbered filename
		filename := fmt.Sprintf("page_%03d.md", i+1)
		filepath := filepath.Join(g.config.OutputDir, filename)

		file, err := os.Create(filepath)
		if err != nil {
			return err
		}

		// Write content
		fmt.Fprintf(file, "# %s\n\n", page.Title)
		fmt.Fprintf(file, "**URL:** %s  \n", page.URL)
		fmt.Fprintf(file, "**Scraped:** %s\n\n", page.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(file, "---\n\n")
		fmt.Fprintf(file, "%s\n", page.Content)

		file.Close()
	}

	// Create index file
	indexFile := filepath.Join(g.config.OutputDir, "index.md")
	file, err := os.Create(indexFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write index content
	fmt.Fprintf(file, "# Documentation Index\n\n")
	fmt.Fprintf(file, "**Scraped from:** %s  \n", g.config.RootURL)
	fmt.Fprintf(file, "**Generated:** %s  \n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "**Total Pages:** %d\n\n", len(g.pages))
	fmt.Fprintf(file, "## Pages\n\n")

	for i, page := range g.pages {
		pageFile := fmt.Sprintf("page_%03d.md", i+1)
		fmt.Fprintf(file, "%d. [%s](%s)\n", i+1, page.Title, pageFile)
	}

	return nil
}

// generateTextOutput generates plain text output
func (g *Generator) generateTextOutput() error {
	if g.config.OutputType == "single" {
		filename := filepath.Join(g.config.OutputDir, "documentation.txt")
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		// Write header
		fmt.Fprintf(file, "DOCUMENTATION SCRAPE RESULTS\n")
		fmt.Fprintf(file, "============================\n\n")
		fmt.Fprintf(file, "Scraped from: %s\n", g.config.RootURL)
		fmt.Fprintf(file, "Generated: %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(file, "Total Pages: %d\n\n", len(g.pages))
		fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 80))

		for i, page := range g.pages {
			fmt.Fprintf(file, "TITLE: %s\n", page.Title)
			fmt.Fprintf(file, "URL: %s\n", page.URL)
			fmt.Fprintf(file, "SCRAPED: %s\n", page.Timestamp.Format(time.RFC3339))
			fmt.Fprintf(file, "CONTENT:\n%s\n", page.Content)

			if i < len(g.pages)-1 {
				separator := "\n" + strings.Repeat("=", 80) + "\n\n"
				fmt.Fprint(file, separator)
			}
		}
	} else {
		for i, page := range g.pages {
			filename := g.createSafeFilename(page.Title, i, ".txt")
			filepath := filepath.Join(g.config.OutputDir, filename)

			file, err := os.Create(filepath)
			if err != nil {
				return err
			}

			fmt.Fprintf(file, "TITLE: %s\n", page.Title)
			fmt.Fprintf(file, "URL: %s\n", page.URL)
			fmt.Fprintf(file, "SCRAPED: %s\n\n", page.Timestamp.Format(time.RFC3339))
			fmt.Fprintf(file, "%s\n", page.Content)

			file.Close()
		}
		return g.generateMetadataFile()
	}

	return nil
}

// generateJSONOutput generates JSON output
func (g *Generator) generateJSONOutput() error {
	filename := filepath.Join(g.config.OutputDir, "documentation.json")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	output := map[string]interface{}{
		"root_url":    g.config.RootURL,
		"scraped_at":  time.Now().Format(time.RFC3339),
		"total_pages": len(g.pages),
		"pages":       g.pages,
	}

	return encoder.Encode(output)
}

// generateMetadataFile creates a metadata file for per-page outputs
func (g *Generator) generateMetadataFile() error {
	filename := filepath.Join(g.config.OutputDir, "metadata.yaml")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	metadata := map[string]interface{}{
		"scrape_info": map[string]interface{}{
			"root_url":    g.config.RootURL,
			"scraped_at":  time.Now().Format(time.RFC3339),
			"total_pages": len(g.pages),
		},
		"pages": make([]map[string]interface{}, len(g.pages)),
	}

	for i, page := range g.pages {
		metadata["pages"].([]map[string]interface{})[i] = map[string]interface{}{
			"title":     page.Title,
			"url":       page.URL,
			"timestamp": page.Timestamp.Format(time.RFC3339),
			"depth":     page.Depth,
		}
	}

	encoder := yaml.NewEncoder(file)
	return encoder.Encode(metadata)
}

// createSafeFilename creates a filesystem-safe filename
func (g *Generator) createSafeFilename(title string, index int, extension string) string {
	// Replace unsafe characters
	safe := regexp.MustCompile(`[^a-zA-Z0-9\-_\s]`).ReplaceAllString(title, "")
	safe = regexp.MustCompile(`\s+`).ReplaceAllString(safe, "_")
	safe = strings.Trim(safe, "_")

	// Limit length
	if len(safe) > 50 {
		safe = safe[:50]
	}

	// Ensure uniqueness
	if safe == "" {
		safe = fmt.Sprintf("page_%d", index)
	} else {
		safe = fmt.Sprintf("%s_%d", safe, index)
	}

	return safe + extension
}

// createAnchor creates a markdown anchor from a title
func (g *Generator) createAnchor(title string) string {
	// Convert to lowercase, replace spaces with dashes, remove special characters
	anchor := strings.ToLower(title)
	anchor = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(anchor, "")
	anchor = regexp.MustCompile(`\s+`).ReplaceAllString(anchor, "-")
	anchor = strings.Trim(anchor, "-")
	return anchor
}
