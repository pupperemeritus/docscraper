package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"docscraper/config"
	"docscraper/devtools"
	"docscraper/output"
	"docscraper/scraper"
)

// TestAdvancedFeaturesIntegration tests all advanced features working together
func TestAdvancedFeaturesIntegration(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := os.MkdirTemp("", "advanced_integration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create advanced configuration with all features enabled
	cfg := &config.Config{
		RootURL:      "https://httpbin.org", // Using httpbin for reliable testing
		OutputFormat: "markdown",
		OutputType:   "per-page",
		OutputDir:    tempDir,
		MinDelay:     0, // No delay for tests
		MaxDelay:     0,
		MaxDepth:     2,
		RespectRobots: false, // Ignore robots.txt for testing
		Verbose:      true,
		LogFile:      filepath.Join(tempDir, "test.log"),
		UserAgents:   []string{"TestAgent/1.0"},
	}

	// Enable all advanced features
	hierarchicalVal := true
	dedupeVal := true
	qualityVal := true
	devtoolsVal := true

	cfg.UseHierarchicalOrdering = &hierarchicalVal
	cfg.EnableDeduplication = &dedupeVal
	cfg.EnableQualityAnalysis = &qualityVal
	cfg.EnableDevTools = &devtoolsVal

	// Set defaults
	cfg.SetDefaults()

	// Override quality settings for testing
	cfg.QualityAnalysis.MinScore = 0.1          // Lower threshold for testing
	cfg.QualityAnalysis.MinWordCount = 5        // Lower word count for testing
	cfg.QualityAnalysis.RequireTitle = false    // Don't require titles for testing
	cfg.QualityAnalysis.RequireContent = false  // Don't require content for testing

	// Override devtools settings
	cfg.DevTools.EnableDebugMode = true
	cfg.DevTools.EnableDryRun = false
	cfg.DevTools.EnableProfiling = true
	cfg.DevTools.SavePerformanceReport = true

	t.Run("DevTools Configuration Validation", func(t *testing.T) {
		dt := devtools.NewDevTools(cfg, true, false)
		err := dt.ValidateConfiguration()
		if err != nil {
			t.Errorf("Configuration validation failed: %v", err)
		}
	})

	t.Run("Dry Run Mode", func(t *testing.T) {
		dryRunCfg := *cfg
		dryRunCfg.DevTools.EnableDryRun = true
		
		dt := devtools.NewDevTools(&dryRunCfg, false, true)
		err := dt.StartDryRun()
		if err != nil {
			t.Errorf("Dry run failed: %v", err)
		}
	})

	t.Run("Enhanced Scraper with All Features", func(t *testing.T) {
		// Create enhanced scraper
		enhancedScraper, err := scraper.NewWithFeatures(cfg)
		if err != nil {
			t.Fatalf("Failed to create enhanced scraper: %v", err)
		}

		// Set up progress tracking
		progressUpdates := 0
		enhancedScraper.SetProgressCallback(func(current, total int, currentURL string) {
			progressUpdates++
			t.Logf("Progress: %d/%d - %s", current, total, currentURL)
		})

		// Start profiling
		dt := devtools.NewDevTools(cfg, true, false)
		dt.StartProfiling()

		// Perform limited scraping (just a few pages to avoid overwhelming the test)
		limitedCfg := *cfg
		limitedCfg.MaxDepth = 1 // Limit depth for faster testing

		// Note: For integration test, we'll simulate the scraping with mock data
		// instead of actual HTTP requests to avoid external dependencies
		mockPages := []scraper.PageData{
			{
				Title:     "Test Page 1",
				URL:       "https://httpbin.org/",
				Content:   "This is test content for page 1 with sufficient words for quality analysis.",
				Timestamp: time.Now(),
				Depth:     0,
			},
			{
				Title:     "Test Page 2", 
				URL:       "https://httpbin.org/get",
				Content:   "This is test content for page 2 with different content structure.",
				Timestamp: time.Now(),
				Depth:     1,
			},
			{
				Title:     "Test Page 3",
				URL:       "https://httpbin.org/headers",
				Content:   "This is test content for page 3 in a different section.",
				Timestamp: time.Now(),
				Depth:     1,
			},
		}

		// Stop profiling and get report
		report := dt.StopProfiling()
		if report == nil {
			t.Error("Expected performance report, got nil")
		}

		// Test deduplication
		t.Run("Deduplication", func(t *testing.T) {
			dedup := scraper.NewLinkDeduplicator(scraper.URLNormalizer{
				RemoveFragment: true,
				RemoveQuery:    false,
				LowerCase:      true,
				RemoveWWW:      true,
				RemoveTrailing: true,
			})

			// Test duplicate detection
			url1 := "https://httpbin.org/get"
			url2 := "https://httpbin.org/get/"
			url3 := "https://httpbin.org/get#section"

			dedup.AddURL(url1)
			if !dedup.IsDuplicate(url2) {
				t.Error("Expected url2 to be detected as duplicate of url1")
			}
			if !dedup.IsDuplicate(url3) {
				t.Error("Expected url3 to be detected as duplicate of url1 (fragment should be removed)")
			}
		})

		// Test quality analysis
		t.Run("Quality Analysis", func(t *testing.T) {
			qualityConfig := scraper.QualityConfig{
				MinWordCount:        10,
				RequireTitle:        true,
				RequireContent:      true,
				SkipNavigationPages: true,
			}
			analyzer := scraper.NewContentQualityAnalyzer(qualityConfig)
			
			scrapedContent := scraper.ScrapedContent{
				URL:     "https://httpbin.org/test",
				Title:   "Test Page",
				Content: "This is a test page with sufficient content for quality analysis testing.",
			}

			quality := analyzer.AnalyzeContent(scrapedContent)
			if quality.Score <= 0 {
				t.Error("Expected positive quality score")
			}
			if quality.WordCount <= 0 {
				t.Error("Expected positive word count")
			}
		})

		// Test output generation
		t.Run("Standard Output Generation", func(t *testing.T) {
			// Convert to output format
			outputPages := make([]output.PageData, len(mockPages))
			for i, page := range mockPages {
				outputPages[i] = output.PageData{
					Title:     page.Title,
					URL:       page.URL,
					Content:   page.Content,
					Timestamp: page.Timestamp,
					Depth:     page.Depth,
				}
			}

			generator := output.New(cfg, outputPages)
			err := generator.Generate()
			if err != nil {
				t.Errorf("Standard output generation failed: %v", err)
			}

			// Check if output files were created
			indexFile := filepath.Join(cfg.OutputDir, "index.md")
			if _, err := os.Stat(indexFile); os.IsNotExist(err) {
				t.Error("Expected index.md file to be created")
			}
		})

		t.Run("Hierarchical Output Generation", func(t *testing.T) {
			// Create hierarchical output directory
			hierarchicalDir := filepath.Join(tempDir, "hierarchical")
			hierarchicalCfg := *cfg
			hierarchicalCfg.OutputDir = hierarchicalDir

			// Convert to output format
			outputPages := make([]output.PageData, len(mockPages))
			for i, page := range mockPages {
				outputPages[i] = output.PageData{
					Title:     page.Title,
					URL:       page.URL,
					Content:   page.Content,
					Timestamp: page.Timestamp,
					Depth:     page.Depth,
				}
			}

			hierarchicalGenerator := output.NewHierarchical(&hierarchicalCfg, outputPages)
			err := hierarchicalGenerator.Generate()
			if err != nil {
				t.Errorf("Hierarchical output generation failed: %v", err)
			}

			// Check if hierarchical output was created
			hierarchicalIndex := filepath.Join(hierarchicalDir, "index.md")
			if _, err := os.Stat(hierarchicalIndex); os.IsNotExist(err) {
				t.Error("Expected hierarchical index.md file to be created")
			}
		})

		t.Run("Performance Report Generation", func(t *testing.T) {
			if cfg.DevTools.SavePerformanceReport {
				reportFile := filepath.Join(cfg.OutputDir, "performance_report.txt")
				err := report.SaveReport(reportFile)
				if err != nil {
					t.Errorf("Failed to save performance report: %v", err)
				}

				// Check if report file exists
				if _, err := os.Stat(reportFile); os.IsNotExist(err) {
					t.Error("Expected performance report file to be created")
				}
			}
		})
	})
}

// TestConfigurationFeatures tests the new configuration options
func TestConfigurationFeatures(t *testing.T) {
	t.Run("Advanced Configuration Loading", func(t *testing.T) {
		cfg := &config.Config{}
		
		// Test loading advanced config
		err := config.LoadConfig("config-advanced.yaml", cfg)
		if err != nil {
			t.Fatalf("Failed to load advanced config: %v", err)
		}

		// Verify advanced features are loaded
		if !cfg.GetUseHierarchicalOrdering() {
			t.Error("Expected hierarchical ordering to be enabled")
		}
		if !cfg.GetEnableDeduplication() {
			t.Error("Expected deduplication to be enabled")
		}
		if !cfg.GetEnableQualityAnalysis() {
			t.Error("Expected quality analysis to be enabled")
		}
		if !cfg.GetEnableDevTools() {
			t.Error("Expected devtools to be enabled")
		}

		// Test configuration validation
		err = cfg.Validate()
		if err != nil {
			t.Errorf("Advanced configuration validation failed: %v", err)
		}
	})

	t.Run("Default Values", func(t *testing.T) {
		cfg := &config.Config{
			RootURL:      "https://example.com",
			OutputFormat: "markdown",
			OutputType:   "single",
			OutputDir:    "/tmp/test",
		}

		cfg.SetDefaults()

		// Check that defaults were set
		if len(cfg.UserAgents) == 0 {
			t.Error("Expected default user agents to be set")
		}
		if cfg.MinDelay == 0 && cfg.MaxDelay == 0 {
			t.Error("Expected default delays to be set")
		}
		if cfg.MaxDepth == 0 {
			t.Error("Expected default max depth to be set")
		}
	})
}

// TestFeatureToggling tests enabling/disabling features
func TestFeatureToggling(t *testing.T) {
	baseConfig := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/test",
	}

	t.Run("All Features Disabled", func(t *testing.T) {
		cfg := *baseConfig
		falseVal := false
		cfg.UseHierarchicalOrdering = &falseVal
		cfg.EnableDeduplication = &falseVal
		cfg.EnableQualityAnalysis = &falseVal
		cfg.EnableDevTools = &falseVal

		if cfg.GetUseHierarchicalOrdering() {
			t.Error("Expected hierarchical ordering to be disabled")
		}
		if cfg.GetEnableDeduplication() {
			t.Error("Expected deduplication to be disabled")
		}
		if cfg.GetEnableQualityAnalysis() {
			t.Error("Expected quality analysis to be disabled")
		}
		if cfg.GetEnableDevTools() {
			t.Error("Expected devtools to be disabled")
		}
	})

	t.Run("Selective Feature Enabling", func(t *testing.T) {
		cfg := *baseConfig
		trueVal := true
		falseVal := false
		
		cfg.UseHierarchicalOrdering = &trueVal
		cfg.EnableDeduplication = &falseVal
		cfg.EnableQualityAnalysis = &trueVal
		cfg.EnableDevTools = &falseVal

		if !cfg.GetUseHierarchicalOrdering() {
			t.Error("Expected hierarchical ordering to be enabled")
		}
		if cfg.GetEnableDeduplication() {
			t.Error("Expected deduplication to be disabled")
		}
		if !cfg.GetEnableQualityAnalysis() {
			t.Error("Expected quality analysis to be enabled")
		}
		if cfg.GetEnableDevTools() {
			t.Error("Expected devtools to be disabled")
		}
	})
}

// TestEdgeCases tests edge cases and error handling
func TestEdgeCases(t *testing.T) {
	t.Run("Empty Page List", func(t *testing.T) {
		cfg := &config.Config{
			RootURL:      "https://example.com",
			OutputFormat: "markdown",
			OutputType:   "single",
			OutputDir:    "/tmp/test",
		}

		emptyPages := []output.PageData{}
		
		// Test standard generator with empty pages
		generator := output.New(cfg, emptyPages)
		err := generator.Generate()
		if err != nil {
			t.Errorf("Standard generator should handle empty pages: %v", err)
		}

		// Test hierarchical generator with empty pages
		hierarchicalGenerator := output.NewHierarchical(cfg, emptyPages)
		err = hierarchicalGenerator.Generate()
		if err != nil {
			t.Errorf("Hierarchical generator should handle empty pages: %v", err)
		}
	})

	t.Run("Invalid Configuration", func(t *testing.T) {
		cfg := &config.Config{
			RootURL:      "", // Invalid empty URL
			OutputFormat: "invalid",
			OutputType:   "invalid",
			OutputDir:    "",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected validation to fail for invalid configuration")
		}
	})

	t.Run("Quality Analysis Edge Cases", func(t *testing.T) {
		qualityConfig := scraper.QualityConfig{
			MinWordCount:   100,
			RequireTitle:   true,
			RequireContent: true,
		}
		
		analyzer := scraper.NewContentQualityAnalyzer(qualityConfig)
		
		// Test with empty content
		emptyContent := scraper.ScrapedContent{
			URL:     "https://example.com",
			Title:   "",
			Content: "",
		}
		
		quality := analyzer.AnalyzeContent(emptyContent)
		if quality.Score > 0.5 { // Should be low quality
			t.Error("Expected low quality score for empty content")
		}
		
		// Test with minimal content
		minimalContent := scraper.ScrapedContent{
			URL:     "https://example.com",
			Title:   "Test",
			Content: "Short content.",
		}
		
		quality = analyzer.AnalyzeContent(minimalContent)
		if quality.WordCount != 2 {
			t.Errorf("Expected word count 2, got %d", quality.WordCount)
		}
	})
}

// BenchmarkAdvancedFeatures benchmarks the performance of advanced features
func BenchmarkAdvancedFeatures(b *testing.B) {
	// Create test data
	pages := make([]output.PageData, 100)
	for i := 0; i < 100; i++ {
		pages[i] = output.PageData{
			Title:     "Test Page",
			URL:       "https://example.com/page" + string(rune(i)),
			Content:   strings.Repeat("This is test content. ", 50),
			Timestamp: time.Now(),
			Depth:     i % 5,
		}
	}

	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/benchmark",
	}

	b.Run("Standard Generation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			generator := output.New(cfg, pages)
			_ = generator.Generate()
		}
	})

	b.Run("Hierarchical Generation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			generator := output.NewHierarchical(cfg, pages)
			_ = generator.Generate()
		}
	})

	b.Run("Deduplication", func(b *testing.B) {
		dedup := scraper.NewLinkDeduplicator(scraper.URLNormalizer{
			LowerCase: true,
		})
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				url := "https://example.com/page" + string(rune(j))
				if !dedup.IsDuplicate(url) {
					dedup.AddURL(url)
				}
			}
		}
	})

	b.Run("Quality Analysis", func(b *testing.B) {
		analyzer := scraper.NewContentQualityAnalyzer(scraper.QualityConfig{})
		content := scraper.ScrapedContent{
			Title:   "Test Page",
			Content: strings.Repeat("This is test content for quality analysis. ", 20),
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = analyzer.AnalyzeContent(content)
		}
	})
}
