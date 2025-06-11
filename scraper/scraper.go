package scraper

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"docscraper/config"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

// PageData represents scraped page information
type PageData struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Depth     int       `json:"depth"`
}

// ProgressCallback defines the signature for progress tracking callbacks
type ProgressCallback func(current, total int, currentURL string)

// EnhancedScraper extends Scraper with advanced features
type EnhancedScraper struct {
	*Scraper
	deduplicator      *LinkDeduplicator
	qualityAnalyzer   *ContentQualityAnalyzer
	progressCallback  ProgressCallback
	currentProgress   int
	totalEstimated    int
}

// Scraper handles the web scraping functionality
type Scraper struct {
	config    *config.Config
	collector *colly.Collector
	pages     []PageData
	logger    *log.Logger
	extractor *ContentExtractor
}

// New creates a new scraper instance
func New(cfg *config.Config) (*Scraper, error) {
	// Validate configuration first
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Setup logging
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "", log.LstdFlags)

	// Create collector with optional SSL ignore setting
	c := colly.NewCollector(
		colly.Async(true),
	)

	// Set limits including concurrent requests
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: cfg.GetConcurrentRequests(),
		Delay:       time.Duration(cfg.MinDelay) * time.Second,
	})

	// Set allowed domains to prevent following external links
	rootURL, err := url.Parse(cfg.RootURL)
	if err != nil {
		return nil, fmt.Errorf("invalid root URL")
	}
	c.AllowedDomains = []string{rootURL.Host}

	// Configure proxy if available (optional feature)
	if cfg.HasProxies() {
		rp, err := proxy.RoundRobinProxySwitcher(cfg.Proxies...)
		if err != nil {
			return nil, fmt.Errorf("failed to setup proxy switcher: %v", err)
		}
		c.SetProxyFunc(rp)
		logger.Printf("Configured %d proxies for rotation", len(cfg.Proxies))
	}

	// Set depth limit
	// This will be combined with other OnRequest logic in setupCallbacks

	// Enable debug mode if verbose
	if cfg.Verbose {
		// Note: Colly v2 debugger is not publicly accessible
		// Verbose logging is handled through our custom logger
	}

	scraper := &Scraper{
		config:    cfg,
		collector: c,
		pages:     make([]PageData, 0),
		logger:    logger,
		extractor: NewContentExtractor(),
	}

	// Setup collector callbacks
	scraper.setupCallbacks()

	return scraper, nil
}

// setupCallbacks configures the colly collector with callbacks
func (s *Scraper) setupCallbacks() {
	// Rotate user agents, add delays, and check depth
	s.collector.OnRequest(func(r *colly.Request) {
		// Check depth limit
		if r.Depth > s.config.MaxDepth {
			s.logger.Printf("Skipping URL at depth %d (max: %d): %s", r.Depth, s.config.MaxDepth, r.URL.String())
			r.Abort()
			return
		}

		if len(s.config.UserAgents) > 0 {
			userAgent := s.config.UserAgents[rand.Intn(len(s.config.UserAgents))]
			r.Headers.Set("User-Agent", userAgent)
		}

		// Add random delay
		if s.config.MaxDelay > s.config.MinDelay {
			delay := rand.Intn(s.config.MaxDelay-s.config.MinDelay) + s.config.MinDelay
			time.Sleep(time.Duration(delay) * time.Second)
		}

		s.logger.Printf("Visiting: %s (depth: %d)", r.URL.String(), r.Depth)
	})

	// Handle HTML responses
	s.collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Count total links on the page
		linkCount := e.DOM.Find("a[href]").Length()
		s.logger.Printf("Page %s contains %d links", e.Request.URL.String(), linkCount)
		
		s.extractPageContent(e)
	})

	// Find and follow links
	linkCounter := 0
	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		linkCounter++
		link := e.Attr("href")
		s.logger.Printf("Processing link #%d: %s (current depth: %d)", linkCounter, link, e.Request.Depth)
		
		if s.shouldFollowLink(link, e.Request.URL) {
			s.logger.Printf("Following link #%d: %s", linkCounter, link)
			e.Request.Visit(link)
		} else {
			s.logger.Printf("Rejected link #%d: %s", linkCounter, link)
		}
	})

	// Handle errors
	s.collector.OnError(func(r *colly.Response, err error) {
		s.logger.Printf("Error visiting %s: %v", r.Request.URL, err)
	})

	// Log responses
	s.collector.OnResponse(func(r *colly.Response) {
		s.logger.Printf("Response from %s: %d", r.Request.URL, r.StatusCode)
	})
}

// extractPageContent extracts and processes content from a page
func (s *Scraper) extractPageContent(e *colly.HTMLElement) {
	doc := e.DOM

	// Extract title
	title := s.extractor.ExtractTitle(doc)
	
	// Extract main content
	content := s.extractor.ExtractContent(doc)
	
	if strings.TrimSpace(content) == "" {
		s.logger.Printf("No content found for: %s", e.Request.URL.String())
		return
	}

	// Create page data
	pageData := PageData{
		Title:     strings.TrimSpace(title),
		URL:       e.Request.URL.String(),
		Content:   content,
		Timestamp: time.Now(),
		Depth:     e.Request.Depth,
	}

	s.pages = append(s.pages, pageData)
	s.logger.Printf("Extracted content from: %s (Title: %s)", pageData.URL, pageData.Title)
}

// shouldFollowLink determines if a link should be followed
func (s *Scraper) shouldFollowLink(link string, baseURL *url.URL) bool {
	s.logger.Printf("Evaluating link: %s from base: %s", link, baseURL.String())
	
	// Parse the link
	linkURL, err := url.Parse(link)
	if err != nil {
		s.logger.Printf("Failed to parse link '%s': %v", link, err)
		return false
	}

	// Check for invalid formats that look like malformed absolute URLs or contain invalid characters
	if !linkURL.IsAbs() && (strings.Contains(link, "://") || 
		(len(link) > 0 && !strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "#") && 
		 !strings.HasPrefix(link, "?") && !strings.Contains(link, ".") && 
		 !strings.Contains(link, "/") && len(link) > 10)) {
		s.logger.Printf("Invalid URL format: %s", link)
		return false
	}

	// Resolve relative URLs
	resolvedURL := baseURL.ResolveReference(linkURL)
	s.logger.Printf("Resolved URL: %s (Host: %s, Path: %s)", resolvedURL.String(), resolvedURL.Host, resolvedURL.Path)

	// Check if the resolved URL is valid (has a scheme and host for absolute URLs)
	if linkURL.IsAbs() && (linkURL.Scheme == "" || linkURL.Host == "") {
		s.logger.Printf("Invalid absolute URL: %s", link)
		return false
	}

	// Only follow links from the same domain
	if resolvedURL.Host != baseURL.Host {
		s.logger.Printf("Skipping external domain: %s vs %s", resolvedURL.Host, baseURL.Host)
		return false
	}

	// Skip certain file types
	skipExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".zip", ".tar", ".gz", ".mp4", ".avi", ".mov"}
	for _, ext := range skipExtensions {
		if strings.HasSuffix(strings.ToLower(resolvedURL.Path), ext) {
			s.logger.Printf("Skipping file extension %s for: %s", ext, resolvedURL.String())
			return false
		}
	}

	// Skip fragments and query-only links
	if resolvedURL.Path == baseURL.Path && resolvedURL.Fragment != "" {
		s.logger.Printf("Skipping fragment-only link: %s", resolvedURL.String())
		return false
	}

	// Skip common non-content paths
	skipPaths := []string{"/login", "/register", "/api/", "/admin/", "/search"}
	for _, path := range skipPaths {
		if strings.Contains(resolvedURL.Path, path) {
			s.logger.Printf("Skipping non-content path '%s' for: %s", path, resolvedURL.String())
			return false
		}
	}

	s.logger.Printf("Link approved for following: %s", resolvedURL.String())
	return true
}

// Scrape starts the scraping process
func (s *Scraper) Scrape() error {
	// Check robots.txt if enabled
	if s.config.RespectRobots {
		if allowed, err := s.checkRobotsTxt(s.config.RootURL); err != nil {
			s.logger.Printf("Warning: Could not check robots.txt: %v", err)
		} else if !allowed {
			return fmt.Errorf("robots.txt disallows scraping this site")
		}
	}

	s.logger.Printf("Starting scrape of: %s with max depth: %d", s.config.RootURL, s.config.MaxDepth)
	
	// Start scraping
	s.collector.Visit(s.config.RootURL)
	s.collector.Wait()
	
	s.logger.Printf("Scraping completed. Total pages found: %d", len(s.pages))
	for i, page := range s.pages {
		s.logger.Printf("Page %d: %s (depth: %d)", i+1, page.URL, page.Depth)
	}

	return nil
}

// checkRobotsTxt checks if scraping is allowed by robots.txt
func (s *Scraper) checkRobotsTxt(rootURL string) (bool, error) {
	parsedURL, err := url.Parse(rootURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false, fmt.Errorf("invalid URL")
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	
	resp, err := http.Get(robotsURL)
	if err != nil {
		return true, nil // If robots.txt doesn't exist, assume allowed
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return true, nil // No robots.txt found, assume allowed
	}

	// Simple robots.txt parsing (basic implementation)
	scanner := bufio.NewScanner(resp.Body)
	userAgentMatch := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToLower(line), "user-agent:") {
			agent := strings.TrimSpace(strings.TrimPrefix(line, "user-agent:"))
			if agent == "*" {
				userAgentMatch = true
			}
		} else if userAgentMatch && strings.HasPrefix(strings.ToLower(line), "disallow:") {
			disallow := strings.TrimSpace(strings.TrimPrefix(line, "disallow:"))
			if disallow == "/" {
				return false, nil
			}
		}
	}

	return true, nil
}

// GetPageCount returns the number of scraped pages
func (s *Scraper) GetPageCount() int {
	return len(s.pages)
}

// GetPages returns the scraped pages
func (s *Scraper) GetPages() []PageData {
	return s.pages
}

// NewWithFeatures creates a new enhanced scraper with advanced features
func NewWithFeatures(cfg *config.Config) (*EnhancedScraper, error) {
	// Create base scraper
	baseScraper, err := New(cfg)
	if err != nil {
		return nil, err
	}

	enhanced := &EnhancedScraper{
		Scraper: baseScraper,
	}

	// Initialize deduplicator if enabled
	if cfg.GetEnableDeduplication() {
		enhanced.deduplicator = NewLinkDeduplicator(URLNormalizer{
			RemoveFragment:  cfg.Deduplication.RemoveFragments,
			RemoveQuery:     cfg.Deduplication.RemoveQueryParams,
			LowerCase:       cfg.Deduplication.IgnoreCase,
			RemoveWWW:       cfg.Deduplication.IgnoreWWW,
			RemoveTrailing:  cfg.Deduplication.IgnoreTrailingSlash,
			SortQueryParams: true,
		})
	}

	// Initialize quality analyzer if enabled
	if cfg.GetEnableQualityAnalysis() {
		qualityConfig := QualityConfig{
			MinWordCount:        cfg.QualityAnalysis.MinWordCount,
			RequireTitle:        cfg.QualityAnalysis.RequireTitle,
			RequireContent:      cfg.QualityAnalysis.RequireContent,
			SkipNavigationPages: cfg.QualityAnalysis.SkipNavigation,
			BlacklistPatterns:   cfg.QualityAnalysis.BlacklistedPatterns,
		}
		enhanced.qualityAnalyzer = NewContentQualityAnalyzer(qualityConfig)
	}

	return enhanced, nil
}

// SetProgressCallback sets the progress tracking callback
func (es *EnhancedScraper) SetProgressCallback(callback ProgressCallback) {
	es.progressCallback = callback
}

// ScrapeWithFeatures performs scraping with all advanced features enabled
func (es *EnhancedScraper) ScrapeWithFeatures() ([]PageData, error) {
	// Override the link following logic to include deduplication
	if es.deduplicator != nil {
		es.setupDeduplicationCallbacks()
	}

	// Override the content extraction to include quality analysis
	if es.qualityAnalyzer != nil {
		es.setupQualityAnalysisCallbacks()
	}

	// Start scraping
	if err := es.Scraper.Scrape(); err != nil {
		return nil, err
	}

	return es.GetPages(), nil
}

// setupDeduplicationCallbacks modifies the scraper to use deduplication
func (es *EnhancedScraper) setupDeduplicationCallbacks() {
	// Replace the original link handling with deduplication-aware version
	es.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		
		// Check if this URL should be followed (use existing logic)
		if !es.shouldFollowLink(link, e.Request.URL) {
			return
		}

		// Resolve absolute URL
		absoluteURL := e.Request.AbsoluteURL(link)

		// Check for duplicates
		if es.deduplicator.IsDuplicate(absoluteURL) {
			es.logger.Printf("Skipping duplicate URL: %s", absoluteURL)
			return
		}

		// Add to deduplicator and visit
		es.deduplicator.AddURL(absoluteURL)
		e.Request.Visit(link)
	})
}

// setupQualityAnalysisCallbacks modifies the scraper to use quality analysis
func (es *EnhancedScraper) setupQualityAnalysisCallbacks() {
	// Replace the original HTML handling with quality-aware version
	es.collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Extract content using DOM
		title := es.extractor.ExtractTitle(e.DOM)
		content := es.extractor.ExtractContent(e.DOM)
		
		// Create ScrapedContent struct for quality analysis
		scrapedContent := ScrapedContent{
			URL:     e.Request.URL.String(),
			Title:   title,
			Content: content,
			Metadata: NodeMetadata{
				WordCount:    len(strings.Fields(content)),
				LastModified: time.Now(),
				ContentType:  "text/html",
			},
		}
		
		// Analyze quality
		quality := es.qualityAnalyzer.AnalyzeContent(scrapedContent)
		
		// Check if content meets quality standards
		if quality.Score < es.config.QualityAnalysis.MinScore {
			es.logger.Printf("Skipping low quality page (score: %.2f): %s", quality.Score, e.Request.URL.String())
			return
		}

		if len(strings.Fields(content)) < es.config.QualityAnalysis.MinWordCount {
			es.logger.Printf("Skipping page with insufficient content (%d words): %s", 
				len(strings.Fields(content)), e.Request.URL.String())
			return
		}

		// Check for blacklisted patterns
		for _, pattern := range es.config.QualityAnalysis.BlacklistedPatterns {
			if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
				es.logger.Printf("Skipping page containing blacklisted pattern '%s': %s", pattern, e.Request.URL.String())
				return
			}
		}

		// If quality analysis passed, save the page
		page := PageData{
			Title:     title,
			URL:       e.Request.URL.String(),
			Content:   content,
			Timestamp: time.Now(),
			Depth:     e.Request.Depth,
		}

		es.pages = append(es.pages, page)
		
		// Update progress
		if es.progressCallback != nil {
			es.currentProgress++
			es.progressCallback(es.currentProgress, es.totalEstimated, e.Request.URL.String())
		}

		es.logger.Printf("Scraped page: %s (quality score: %.2f)", page.URL, quality.Score)
	})
}