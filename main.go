package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"docscraper/config"
	"docscraper/devtools"
	"docscraper/output"
	"docscraper/scraper"
	"docscraper/utils"
)

func main() {
	// Command line flags
	var (
		rootURL       = flag.String("root-url", "", "Root URL to start scraping")
		outputFormat  = flag.String("output", "markdown", "Output format: markdown, text, json")
		outputType    = flag.String("format", "per-page", "Output type: single, per-page")
		outputDir     = flag.String("output-dir", "output", "Output directory")
		configFile    = flag.String("config", "", "Configuration file (YAML or JSON)")
		userAgentFile = flag.String("user-agent-list", "", "File containing user agent list")
		maxDepth      = flag.Int("max-depth", 3, "Maximum crawling depth")
		minDelay      = flag.Int("min-delay", 1, "Minimum delay between requests (seconds)")
		maxDelay      = flag.Int("max-delay", 3, "Maximum delay between requests (seconds)")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging")
		logFile       = flag.String("log-file", "scraper.log", "Log file path")
		
		// New feature flags
		hierarchical  = flag.Bool("hierarchical", false, "Use hierarchical output organization")
		debugMode     = flag.Bool("debug", false, "Enable debug mode")
		dryRun        = flag.Bool("dry-run", false, "Perform dry run without actual scraping")
		enableProfiling = flag.Bool("profile", false, "Enable performance profiling")
		skipDuplicates = flag.Bool("dedupe", false, "Enable duplicate URL detection")
		qualityFilter = flag.Bool("quality", false, "Enable content quality analysis")
	)
	flag.Parse()

	// Load configuration
	cfg := &config.Config{
		RootURL:       *rootURL,
		OutputFormat:  *outputFormat,
		OutputType:    *outputType,
		OutputDir:     *outputDir,
		MinDelay:      *minDelay,
		MaxDelay:      *maxDelay,
		MaxDepth:      *maxDepth,
		RespectRobots: true,
		LogFile:       *logFile,
		Verbose:       *verbose,
		UserAgents:    config.DefaultUserAgents,
	}

	// Load config file if provided
	if *configFile != "" {
		if err := config.LoadConfig(*configFile, cfg); err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}
	}

	// Apply command line overrides for new features
	if *hierarchical {
		hierarchicalVal := true
		cfg.UseHierarchicalOrdering = &hierarchicalVal
	}
	if *skipDuplicates {
		dedupeVal := true
		cfg.EnableDeduplication = &dedupeVal
	}
	if *qualityFilter {
		qualityVal := true
		cfg.EnableQualityAnalysis = &qualityVal
	}
	if *debugMode || *dryRun || *enableProfiling {
		devtoolsVal := true
		cfg.EnableDevTools = &devtoolsVal
		cfg.DevTools.EnableDebugMode = *debugMode
		cfg.DevTools.EnableDryRun = *dryRun
		cfg.DevTools.EnableProfiling = *enableProfiling
	}

	// Set defaults for configuration
	cfg.SetDefaults()

	// Override with command line arguments
	if *rootURL != "" {
		cfg.RootURL = *rootURL
	}

	if cfg.RootURL == "" {
		log.Fatal("Root URL is required. Use --root-url or provide it in config file.")
	}

	// Initialize DevTools if enabled
	var dt *devtools.DevTools
	if cfg.GetEnableDevTools() {
		dt = devtools.NewDevTools(cfg, cfg.DevTools.EnableDebugMode, cfg.DevTools.EnableDryRun)
		
		// Validate configuration
		if err := dt.ValidateConfiguration(); err != nil {
			log.Fatalf("Configuration validation failed: %v", err)
		}
		
		// Perform dry run if requested
		if cfg.DevTools.EnableDryRun {
			if err := dt.StartDryRun(); err != nil {
				log.Fatalf("Dry run failed: %v", err)
			}
			fmt.Println("Dry run completed successfully!")
			return
		}
		
		// Start profiling if enabled
		if cfg.DevTools.EnableProfiling {
			dt.StartProfiling()
		}
	}

	// Load user agent list if provided
	if *userAgentFile != "" {
		userAgents, err := utils.LoadFileLines(*userAgentFile)
		if err != nil {
			log.Printf("Warning: Could not load user agent file: %v", err)
		} else {
			cfg.UserAgents = userAgents
		}
	}

	// Create scraper with enhanced features
	s, err := scraper.NewWithFeatures(cfg)
	if err != nil {
		log.Fatalf("Error creating scraper: %v", err)
	}

	// Configure progress tracking if DevTools enabled
	if dt != nil && cfg.DevTools.EnableProgressBar {
		s.SetProgressCallback(func(current, total int, currentURL string) {
			dt.UpdateProgress(current, total, currentURL)
		})
	}

	// Start scraping
	fmt.Printf("Starting to scrape: %s\n", cfg.RootURL)
	pages, err := s.ScrapeWithFeatures()
	if err != nil {
		log.Fatalf("Error during scraping: %v", err)
	}

	// Generate output using appropriate generator
	// Convert scraper.PageData to output.PageData
	outputPages := make([]output.PageData, len(pages))
	for i, page := range pages {
		outputPages[i] = output.PageData{
			Title:     page.Title,
			URL:       page.URL,
			Content:   page.Content,
			Timestamp: page.Timestamp,
			Depth:     page.Depth,
		}
	}
	
	var generator interface{ Generate() error }
	
	if cfg.GetUseHierarchicalOrdering() {
		fmt.Println("Using hierarchical output organization...")
		generator = output.NewHierarchical(cfg, outputPages)
	} else {
		generator = output.New(cfg, outputPages)
	}

	if err := generator.Generate(); err != nil {
		log.Fatalf("Error generating output: %v", err)
	}

	// Stop profiling and save report if enabled
	if dt != nil && cfg.DevTools.EnableProfiling {
		report := dt.StopProfiling()
		if cfg.DevTools.SavePerformanceReport {
			reportFile := filepath.Join(cfg.OutputDir, "performance_report.txt")
			if err := report.SaveReport(reportFile); err != nil {
				log.Printf("Warning: Could not save performance report: %v", err)
			} else {
				fmt.Printf("Performance report saved to: %s\n", reportFile)
			}
		}
	}

	fmt.Printf("Scraping completed. Found %d pages. Output saved to: %s\n", 
		len(pages), cfg.OutputDir)

	// Print summary of features used
	if cfg.GetEnableDeduplication() {
		fmt.Println("✓ Duplicate detection enabled")
	}
	if cfg.GetEnableQualityAnalysis() {
		fmt.Println("✓ Content quality analysis enabled")
	}
	if cfg.GetUseHierarchicalOrdering() {
		fmt.Println("✓ Hierarchical organization enabled")
	}
	if cfg.GetEnableDevTools() {
		fmt.Println("✓ Development tools enabled")
	}
}