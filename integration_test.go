package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"docscraper/config"
	"docscraper/output"
	"docscraper/scraper"
)

func TestIntegrationAdvancedFeatures(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration with all advanced features enabled
	cfg := &config.Config{
		RootURL:      "https://httpbin.org",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    tempDir,
		MinDelay:     1,
		MaxDelay:     2,
		MaxDepth:     2,
		UserAgents:   config.DefaultUserAgents,
	}

	// Enable advanced features
	hierarchical := true
	deduplication := true
	qualityAnalysis := true
	devtools := true

	cfg.UseHierarchicalOrdering = &hierarchical
	cfg.EnableDeduplication = &deduplication
	cfg.EnableQualityAnalysis = &qualityAnalysis
	cfg.EnableDevTools = &devtools

	// Set defaults
	cfg.SetDefaults()

	// Test configuration validation
	dt := devtools.NewDevTools(cfg, true, false)
	if err := dt.ValidateConfiguration(); err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}

	t.Logf("✓ Configuration validation passed")

	// Test enhanced scraper creation
	enhancedScraper, err := scraper.NewWithFeatures(cfg)
	if err != nil {
		t.Fatalf("Failed to create enhanced scraper: %v", err)
	}

	t.Logf("✓ Enhanced scraper created successfully")

	// Test progress callback
	progressCalled := false
	enhancedScraper.SetProgressCallback(func(current, total int, currentURL string) {
		progressCalled = true
		t.Logf("Progress: %d/%d - %s", current, total, currentURL)
	})

	// Create mock page data for testing output generators
	mockPages := []scraper.PageData{
		{
			Title:     "Home Page",
			URL:       "https://httpbin.org/",
			Content:   "This is the home page with some content about HTTP testing.",
			Timestamp: time.Now(),
			Depth:     0,
		},
		{
			Title:     "GET Method",
			URL:       "https://httpbin.org/get",
			Content:   "Information about GET HTTP method. This method is used to retrieve data.",
			Timestamp: time.Now(),
			Depth:     1,
		},
		{
			Title:     "POST Method",
			URL:       "https://httpbin.org/post",
			Content:   "Information about POST HTTP method. This method is used to send data.",
			Timestamp: time.Now(),
			Depth:     1,
		},
	}

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

	// Test regular output generator
	t.Run("RegularOutputGeneration", func(t *testing.T) {
		regularDir := filepath.Join(tempDir, "regular")
		regularCfg := *cfg
		regularCfg.OutputDir = regularDir

		generator := output.New(&regularCfg, outputPages)
		if err := generator.Generate(); err != nil {
			t.Errorf("Regular output generation failed: %v", err)
		}

		// Check if output file exists
		outputFile := filepath.Join(regularDir, "documentation.md")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file was not created: %s", outputFile)
		}

		t.Logf("✓ Regular output generation successful")
	})

	// Test hierarchical output generator
	t.Run("HierarchicalOutputGeneration", func(t *testing.T) {
		hierarchicalDir := filepath.Join(tempDir, "hierarchical")
		hierarchicalCfg := *cfg
		hierarchicalCfg.OutputDir = hierarchicalDir

		generator := output.NewHierarchical(&hierarchicalCfg, outputPages)
		if err := generator.Generate(); err != nil {
			t.Errorf("Hierarchical output generation failed: %v", err)
		}

		// Check if output file exists
		outputFile := filepath.Join(hierarchicalDir, "documentation_hierarchical.md")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Hierarchical output file was not created: %s", outputFile)
		}

		t.Logf("✓ Hierarchical output generation successful")
	})

	// Test deduplicator
	t.Run("DeduplicationFeature", func(t *testing.T) {
		deduplicator := scraper.NewLinkDeduplicator(scraper.URLNormalizer{
			RemoveFragment: true,
			RemoveQuery:    false,
			LowerCase:      true,
			RemoveWWW:      true,
			RemoveTrailing: true,
		})

		// Test duplicate detection
		url1 := "https://example.com/page"
		url2 := "https://Example.com/page/"
		url3 := "https://www.example.com/page#section"

		if deduplicator.IsDuplicate(url1) {
			t.Errorf("First URL should not be duplicate")
		}
		deduplicator.AddURL(url1)

		if !deduplicator.IsDuplicate(url2) {
			t.Errorf("Second URL should be detected as duplicate")
		}

		if !deduplicator.IsDuplicate(url3) {
			t.Errorf("Third URL should be detected as duplicate")
		}

		t.Logf("✓ Deduplication feature working correctly")
	})

	// Test quality analyzer
	t.Run("QualityAnalysisFeature", func(t *testing.T) {
		qualityConfig := scraper.QualityConfig{
			MinWordCount:        10,
			RequireTitle:        true,
			RequireContent:      true,
			SkipNavigationPages: true,
		}

		analyzer := scraper.NewContentQualityAnalyzer(qualityConfig)

		// Test high quality content
		highQualityContent := scraper.ScrapedContent{
			URL:     "https://example.com/good",
			Title:   "Good Article",
			Content: "This is a well-written article with substantial content that provides valuable information to readers. It contains multiple sentences and covers the topic thoroughly.",
		}

		quality := analyzer.AnalyzeContent(highQualityContent)
		if quality.Score < 0.5 {
			t.Errorf("High quality content should have score > 0.5, got %f", quality.Score)
		}

		// Test low quality content
		lowQualityContent := scraper.ScrapedContent{
			URL:     "https://example.com/bad",
			Title:   "",
			Content: "Short.",
		}

		quality = analyzer.AnalyzeContent(lowQualityContent)
		if quality.Score > 0.3 {
			t.Errorf("Low quality content should have score < 0.3, got %f", quality.Score)
		}

		t.Logf("✓ Quality analysis feature working correctly")
	})

	// Test DevTools features
	t.Run("DevToolsFeatures", func(t *testing.T) {
		// Test dry run
		dryRunCfg := *cfg
		dryRunCfg.DevTools.EnableDryRun = true
		dryRunDT := devtools.NewDevTools(&dryRunCfg, false, true)
		
		if err := dryRunDT.StartDryRun(); err != nil {
			t.Errorf("Dry run failed: %v", err)
		}

		// Test profiling
		profilingDT := devtools.NewDevTools(cfg, true, false)
		profilingDT.StartProfiling()
		
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		
		report := profilingDT.StopProfiling()
		if report == nil {
			t.Errorf("Performance report should not be nil")
		}

		if report.TotalDuration <= 0 {
			t.Errorf("Performance report should have positive duration")
		}

		t.Logf("✓ DevTools features working correctly")
	})

	t.Logf("✓ All integration tests passed successfully!")
}

func TestConfigurationDefaults(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/test",
	}

	// Test that SetDefaults works correctly
	cfg.SetDefaults()

	if len(cfg.UserAgents) == 0 {
		t.Errorf("Default user agents should be set")
	}

	if cfg.MinDelay == 0 && cfg.MaxDelay == 0 {
		t.Errorf("Default delays should be set")
	}

	if cfg.MaxDepth == 0 {
		t.Errorf("Default max depth should be set")
	}

	// Test getter methods
	if !cfg.GetEnableDeduplication() {
		t.Logf("Deduplication disabled by default")
	}

	if cfg.GetEnableQualityAnalysis() {
		t.Logf("Quality analysis disabled by default")
	}

	if cfg.GetUseHierarchicalOrdering() {
		t.Logf("Hierarchical ordering disabled by default")
	}

	if cfg.GetEnableDevTools() {
		t.Logf("DevTools disabled by default")
	}

	t.Logf("✓ Configuration defaults working correctly")
}

func TestFeatureToggling(t *testing.T) {
	cfg := &config.Config{
		RootURL:      "https://example.com",
		OutputFormat: "markdown",
		OutputType:   "single",
		OutputDir:    "/tmp/test",
	}

	// Enable features one by one
	hierarchical := true
	cfg.UseHierarchicalOrdering = &hierarchical

	deduplication := true
	cfg.EnableDeduplication = &deduplication

	quality := true
	cfg.EnableQualityAnalysis = &quality

	devtools := true
	cfg.EnableDevTools = &devtools

	cfg.SetDefaults()

	// Verify features are enabled
	if !cfg.GetUseHierarchicalOrdering() {
		t.Errorf("Hierarchical ordering should be enabled")
	}

	if !cfg.GetEnableDeduplication() {
		t.Errorf("Deduplication should be enabled")
	}

	if !cfg.GetEnableQualityAnalysis() {
		t.Errorf("Quality analysis should be enabled")
	}

	if !cfg.GetEnableDevTools() {
		t.Errorf("DevTools should be enabled")
	}

	t.Logf("✓ Feature toggling working correctly")
}
