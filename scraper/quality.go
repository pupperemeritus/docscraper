package scraper

import (
	"regexp"
	"strings"
	"time"
)

// QualityConfig defines configuration for content quality analysis
type QualityConfig struct {
	MinWordCount        int      `yaml:"min_word_count"`
	MinCodeBlockCount   int      `yaml:"min_code_block_count"`
	MaxEmptyLineRatio   float64  `yaml:"max_empty_line_ratio"`
	RequireTitle        bool     `yaml:"require_title"`
	RequireContent      bool     `yaml:"require_content"`
	SkipNavigationPages bool     `yaml:"skip_navigation_pages"`
	MinContentRatio     float64  `yaml:"min_content_ratio"`
	BlacklistPatterns   []string `yaml:"blacklist_patterns"`
	WhitelistPatterns   []string `yaml:"whitelist_patterns"`
}

// QualityWeights defines weights for different quality metrics
type QualityWeights struct {
	WordCount     float64 `yaml:"word_count"`
	CodeBlocks    float64 `yaml:"code_blocks"`
	Images        float64 `yaml:"images"`
	Headers       float64 `yaml:"headers"`
	ContentRatio  float64 `yaml:"content_ratio"`
	TitlePresence float64 `yaml:"title_presence"`
}

// ContentMetrics represents various content metrics
type ContentMetrics struct {
	WordCount      int
	CodeBlockCount int
	ImageCount     int
	LinkCount      int
	HeaderCount    int
	EmptyLineRatio float64
	ContentRatio   float64
	HasTitle       bool
	HasHeaders     bool
	TotalLines     int
	EmptyLines     int
}

// ElementCounts represents counts of different HTML elements
type ElementCounts struct {
	Images  int
	Links   int
	Headers int
	Code    int
}

// CodeBlock represents a code block found in content
type CodeBlock struct {
	Language  string
	Content   string
	LineCount int
}

// QualityScorer handles scoring of content quality
type QualityScorer struct {
	weights QualityWeights
}

// QualityStats tracks overall quality statistics
type QualityStats struct {
	TotalPages       int     `json:"total_pages"`
	PassedPages      int     `json:"passed_pages"`
	FailedPages      int     `json:"failed_pages"`
	AverageScore     float64 `json:"average_score"`
	AverageWordCount int     `json:"average_word_count"`
}

// QualityReport provides a summary of quality analysis
type QualityReport struct {
	Stats        QualityStats  `json:"stats"`
	TopPages     []PageQuality `json:"top_pages"`
	WorstPages   []PageQuality `json:"worst_pages"`
	CommonIssues []IssueCount  `json:"common_issues"`
	GeneratedAt  time.Time     `json:"generated_at"`
}

// PageQuality represents quality data for a specific page
type PageQuality struct {
	URL     string         `json:"url"`
	Title   string         `json:"title"`
	Quality ContentQuality `json:"quality"`
}

// IssueCount represents the frequency of specific issues
type IssueCount struct {
	IssueType string `json:"issue_type"`
	Count     int    `json:"count"`
}

// ContentQualityAnalyzer analyzes and rates content quality
type ContentQualityAnalyzer struct {
	config QualityConfig
	scorer QualityScorer
	stats  QualityStats
}

// NewContentQualityAnalyzer creates a new content quality analyzer
func NewContentQualityAnalyzer(config QualityConfig) *ContentQualityAnalyzer {
	// Default weights if not provided
	weights := QualityWeights{
		WordCount:     0.3,
		CodeBlocks:    0.2,
		Images:        0.1,
		Headers:       0.15,
		ContentRatio:  0.15,
		TitlePresence: 0.1,
	}

	return &ContentQualityAnalyzer{
		config: config,
		scorer: QualityScorer{weights: weights},
		stats:  QualityStats{},
	}
}

// AnalyzeContent analyzes the quality of scraped content
func (cqa *ContentQualityAnalyzer) AnalyzeContent(content ScrapedContent) ContentQuality {
	metrics := cqa.extractMetrics(content)
	score := cqa.scorer.CalculateScore(metrics)
	language := cqa.DetectLanguage(content.Content)
	issues := cqa.DetectIssues(content)
	tags := cqa.generateTags(content, metrics)

	quality := ContentQuality{
		Score:            score,
		WordCount:        metrics.WordCount,
		CodeBlockCount:   metrics.CodeBlockCount,
		ImageCount:       metrics.ImageCount,
		LinkCount:        metrics.LinkCount,
		EmptyLineRatio:   metrics.EmptyLineRatio,
		ContentRatio:     metrics.ContentRatio,
		HasTitle:         metrics.HasTitle,
		HasHeaders:       metrics.HasHeaders,
		IsNavigationPage: cqa.IsNavigationPage(content),
		Language:         language,
		Issues:           issues,
		Tags:             tags,
	}

	// Update statistics
	cqa.updateStats(quality)

	return quality
}

// extractMetrics extracts various metrics from content
func (cqa *ContentQualityAnalyzer) extractMetrics(content ScrapedContent) ContentMetrics {
	text := content.Content
	lines := strings.Split(text, "\n")

	metrics := ContentMetrics{
		WordCount:      cqa.countWords(text),
		CodeBlockCount: cqa.countCodeBlocks(text),
		HasTitle:       content.Title != "" && content.Title != "Untitled",
		TotalLines:     len(lines),
	}

	// Count empty lines
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			metrics.EmptyLines++
		}
	}

	if metrics.TotalLines > 0 {
		metrics.EmptyLineRatio = float64(metrics.EmptyLines) / float64(metrics.TotalLines)
	}

	// Count headers (lines starting with # in markdown or h1-h6 patterns)
	metrics.HeaderCount = cqa.countHeaders(text)
	metrics.HasHeaders = metrics.HeaderCount > 0

	// Estimate content ratio (actual content vs. boilerplate)
	metrics.ContentRatio = cqa.estimateContentRatio(text)

	return metrics
}

// countWords counts words in text
func (cqa *ContentQualityAnalyzer) countWords(text string) int {
	// Simple word counting - split by whitespace and filter empty strings
	words := strings.Fields(text)
	return len(words)
}

// countCodeBlocks counts code blocks in text
func (cqa *ContentQualityAnalyzer) countCodeBlocks(text string) int {
	// Count markdown code blocks (```...```)
	codeBlockPattern := regexp.MustCompile("```[\\s\\S]*?```")
	matches := codeBlockPattern.FindAllString(text, -1)
	count := len(matches)

	// Remove code blocks from text to avoid counting their content as inline code
	textWithoutBlocks := codeBlockPattern.ReplaceAllString(text, "")

	// Also count inline code (`...`) in the remaining text
	inlineCodePattern := regexp.MustCompile("`[^`]+`")
	inlineMatches := inlineCodePattern.FindAllString(textWithoutBlocks, -1)

	// Add inline code but with less weight (3 inline = 1 block)
	count += len(inlineMatches) / 3

	return count
}

// countHeaders counts header elements in text
func (cqa *ContentQualityAnalyzer) countHeaders(text string) int {
	// Count markdown headers (# ## ### etc.)
	headerPattern := regexp.MustCompile(`(?m)^#{1,6}\s+.+`)
	matches := headerPattern.FindAllString(text, -1)
	return len(matches)
}

// estimateContentRatio estimates the ratio of actual content vs. boilerplate
func (cqa *ContentQualityAnalyzer) estimateContentRatio(text string) float64 {
	if len(text) == 0 {
		return 0.0
	}

	// Common boilerplate patterns
	boilerplatePatterns := []string{
		`(?i)copyright`,
		`(?i)all rights reserved`,
		`(?i)privacy policy`,
		`(?i)terms of service`,
		`(?i)cookie policy`,
		`(?i)newsletter`,
		`(?i)subscribe`,
		`(?i)follow us`,
		`(?i)social media`,
		`(?i)navigation`,
		`(?i)menu`,
		`(?i)footer`,
		`(?i)header`,
	}

	totalLength := len(text)
	boilerplateLength := 0

	for _, pattern := range boilerplatePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			boilerplateLength += len(match)
		}
	}

	contentLength := totalLength - boilerplateLength
	if contentLength < 0 {
		contentLength = 0
	}

	return float64(contentLength) / float64(totalLength)
}

// CalculateScore calculates a quality score based on metrics
func (qs *QualityScorer) CalculateScore(metrics ContentMetrics) float64 {
	score := 0.0

	// Word count score (0-1)
	wordScore := 0.0
	if metrics.WordCount >= 500 {
		wordScore = 1.0
	} else if metrics.WordCount >= 100 {
		wordScore = float64(metrics.WordCount) / 500.0
	}
	score += wordScore * qs.weights.WordCount

	// Code block score (0-1)
	codeScore := 0.0
	if metrics.CodeBlockCount >= 3 {
		codeScore = 1.0
	} else if metrics.CodeBlockCount > 0 {
		codeScore = float64(metrics.CodeBlockCount) / 3.0
	}
	score += codeScore * qs.weights.CodeBlocks

	// Header score (0-1)
	headerScore := 0.0
	if metrics.HeaderCount >= 3 {
		headerScore = 1.0
	} else if metrics.HeaderCount > 0 {
		headerScore = float64(metrics.HeaderCount) / 3.0
	}
	score += headerScore * qs.weights.Headers

	// Content ratio score
	contentRatioScore := metrics.ContentRatio
	if contentRatioScore > 1.0 {
		contentRatioScore = 1.0
	}
	score += contentRatioScore * qs.weights.ContentRatio

	// Title presence score
	titleScore := 0.0
	if metrics.HasTitle {
		titleScore = 1.0
	}
	score += titleScore * qs.weights.TitlePresence

	return score
}

// DetectLanguage attempts to detect the primary language of content
func (cqa *ContentQualityAnalyzer) DetectLanguage(content string) string {
	// Simple language detection based on common patterns
	// In a real implementation, you might use a proper language detection library

	// Count common English words
	englishWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
	englishCount := 0

	words := strings.Fields(strings.ToLower(content))
	wordMap := make(map[string]int)
	for _, word := range words {
		wordMap[word]++
	}

	for _, englishWord := range englishWords {
		englishCount += wordMap[englishWord]
	}

	if englishCount > len(words)/20 { // If more than 5% are common English words
		return "en"
	}

	return "unknown"
}

// ExtractCodeBlocks extracts code blocks from content
func (cqa *ContentQualityAnalyzer) ExtractCodeBlocks(content string) []CodeBlock {
	var blocks []CodeBlock

	// Extract markdown code blocks
	codeBlockPattern := regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")
	matches := codeBlockPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			language := match[1]
			if language == "" {
				language = "text"
			}
			code := match[2]
			lineCount := len(strings.Split(code, "\n"))

			blocks = append(blocks, CodeBlock{
				Language:  language,
				Content:   code,
				LineCount: lineCount,
			})
		}
	}

	return blocks
}

// IsNavigationPage determines if a page is primarily navigation
func (cqa *ContentQualityAnalyzer) IsNavigationPage(content ScrapedContent) bool {
	text := strings.ToLower(content.Content)

	// Navigation indicators
	navIndicators := []string{
		"table of contents",
		"navigation",
		"site map",
		"index",
		"directory",
		"menu",
		"links",
	}

	indicatorCount := 0
	for _, indicator := range navIndicators {
		if strings.Contains(text, indicator) {
			indicatorCount++
		}
	}

	// If title suggests navigation
	title := strings.ToLower(content.Title)
	for _, indicator := range navIndicators {
		if strings.Contains(title, indicator) {
			indicatorCount += 2 // Weight title indicators more heavily
		}
	}

	// If high ratio of links to content
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	linkCount := len(linkPattern.FindAllString(content.Content, -1))
	wordCount := len(strings.Fields(content.Content))

	if wordCount > 0 && float64(linkCount)/float64(wordCount) > 0.3 {
		indicatorCount++
	}

	return indicatorCount >= 2
}

// DetectIssues detects quality issues in content
func (cqa *ContentQualityAnalyzer) DetectIssues(content ScrapedContent) []QualityIssue {
	var issues []QualityIssue

	wordCount := len(strings.Fields(content.Content))

	// Check minimum word count
	if wordCount < cqa.config.MinWordCount {
		issues = append(issues, QualityIssue{
			Type:        "word_count",
			Severity:    "warning",
			Description: "Content has fewer words than recommended minimum",
		})
	}

	// Check for missing title
	if cqa.config.RequireTitle && (content.Title == "" || content.Title == "Untitled") {
		issues = append(issues, QualityIssue{
			Type:        "missing_title",
			Severity:    "error",
			Description: "Page is missing a title",
		})
	}

	// Check for empty content
	if cqa.config.RequireContent && len(strings.TrimSpace(content.Content)) == 0 {
		issues = append(issues, QualityIssue{
			Type:        "empty_content",
			Severity:    "error",
			Description: "Page has no content",
		})
	}

	// Check blacklist patterns
	text := strings.ToLower(content.Content)
	blacklistFound := false
	var foundPatterns []string
	for _, pattern := range cqa.config.BlacklistPatterns {
		if strings.Contains(text, strings.ToLower(pattern)) {
			blacklistFound = true
			foundPatterns = append(foundPatterns, pattern)
		}
	}

	if blacklistFound {
		description := "Content contains blacklisted pattern"
		if len(foundPatterns) > 1 {
			description += "s: " + strings.Join(foundPatterns, ", ")
		} else {
			description += ": " + foundPatterns[0]
		}
		issues = append(issues, QualityIssue{
			Type:        "blacklisted_content",
			Severity:    "warning",
			Description: description,
		})
	}

	return issues
}

// ShouldSkip determines if content should be skipped based on quality
func (cqa *ContentQualityAnalyzer) ShouldSkip(quality ContentQuality) bool {
	// Skip if below minimum score threshold
	if quality.Score < 0.3 { // Default minimum acceptable score
		return true
	}

	// Skip navigation pages if configured
	if cqa.config.SkipNavigationPages && quality.IsNavigationPage {
		return true
	}

	// Skip if has critical issues
	for _, issue := range quality.Issues {
		if issue.Severity == "error" {
			return true
		}
	}

	return false
}

// generateTags generates tags based on content analysis
func (cqa *ContentQualityAnalyzer) generateTags(content ScrapedContent, metrics ContentMetrics) []string {
	var tags []string

	// Content type tags
	if metrics.CodeBlockCount > 0 {
		tags = append(tags, "technical")
	}

	if metrics.WordCount > 1000 {
		tags = append(tags, "long-form")
	} else if metrics.WordCount < 200 {
		tags = append(tags, "short-form")
	}

	if metrics.HasHeaders {
		tags = append(tags, "structured")
	}

	// Quality tags
	if metrics.WordCount >= 500 && metrics.CodeBlockCount >= 2 {
		tags = append(tags, "comprehensive")
	}

	if metrics.ContentRatio < 0.3 {
		tags = append(tags, "low-content")
	}

	return tags
}

// updateStats updates internal statistics
func (cqa *ContentQualityAnalyzer) updateStats(quality ContentQuality) {
	cqa.stats.TotalPages++

	if quality.Score >= 0.4 { // Default passing score
		cqa.stats.PassedPages++
	} else {
		cqa.stats.FailedPages++
	}

	// Update average score
	totalScore := cqa.stats.AverageScore * float64(cqa.stats.TotalPages-1)
	cqa.stats.AverageScore = (totalScore + quality.Score) / float64(cqa.stats.TotalPages)

	// Update average word count
	totalWords := cqa.stats.AverageWordCount * (cqa.stats.TotalPages - 1)
	cqa.stats.AverageWordCount = (totalWords + quality.WordCount) / cqa.stats.TotalPages
}

// GenerateReport generates a quality analysis report
func (cqa *ContentQualityAnalyzer) GenerateReport() QualityReport {
	return QualityReport{
		Stats:        cqa.stats,
		TopPages:     []PageQuality{}, // Would be populated with actual data
		WorstPages:   []PageQuality{}, // Would be populated with actual data
		CommonIssues: []IssueCount{},  // Would be populated with actual data
		GeneratedAt:  time.Now(),
	}
}
