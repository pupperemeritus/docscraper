package scraper

import (
	"net/url"
	"sort"
	"strings"
)

// URLNormalizer defines options for URL normalization
type URLNormalizer struct {
	RemoveFragment    bool // Remove #section
	RemoveQuery       bool // Remove ?param=value
	RemoveTrailing    bool // Remove trailing slash
	LowerCase         bool // Convert to lowercase
	RemoveWWW         bool // Remove www. prefix
	SortQueryParams   bool // Sort query parameters
}

// LinkDeduplicator handles duplicate URL detection and filtering
type LinkDeduplicator struct {
	seenURLs        map[string]bool
	canonicalMap    map[string]string
	duplicateCount  int
	normalizeFunc   func(string) string
	normalizer      URLNormalizer
}

// NewLinkDeduplicator creates a new link deduplicator with the given configuration
func NewLinkDeduplicator(config URLNormalizer) *LinkDeduplicator {
	return &LinkDeduplicator{
		seenURLs:       make(map[string]bool),
		canonicalMap:   make(map[string]string),
		duplicateCount: 0,
		normalizer:     config,
	}
}

// NormalizeURL cleans and standardizes URLs according to the configuration
func (ld *LinkDeduplicator) NormalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Remove fragment if configured
	if ld.normalizer.RemoveFragment {
		parsedURL.Fragment = ""
	}

	// Remove query if configured
	if ld.normalizer.RemoveQuery {
		parsedURL.RawQuery = ""
	}

	// Sort query parameters if configured
	if ld.normalizer.SortQueryParams && !ld.normalizer.RemoveQuery {
		values := parsedURL.Query()
		sortedQuery := url.Values{}
		
		// Get sorted keys
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		// Rebuild query with sorted parameters
		for _, k := range keys {
			for _, v := range values[k] {
				sortedQuery.Add(k, v)
			}
		}
		parsedURL.RawQuery = sortedQuery.Encode()
	}

	// Remove www. prefix if configured
	if ld.normalizer.RemoveWWW {
		host := parsedURL.Host
		if strings.HasPrefix(strings.ToLower(host), "www.") {
			parsedURL.Host = host[4:]
		}
	}

	// Remove trailing slash if configured
	if ld.normalizer.RemoveTrailing && parsedURL.Path != "/" {
		parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
	}

	// Convert to lowercase if configured
	normalizedURL := parsedURL.String()
	if ld.normalizer.LowerCase {
		normalizedURL = strings.ToLower(normalizedURL)
	}

	return normalizedURL, nil
}

// IsDuplicate checks if the URL has already been processed
func (ld *LinkDeduplicator) IsDuplicate(rawURL string) bool {
	normalized, err := ld.NormalizeURL(rawURL)
	if err != nil {
		// If normalization fails, consider it as seen to avoid processing invalid URLs
		return true
	}

	return ld.seenURLs[normalized]
}

// AddURL adds a URL to the seen set, returns false if it was already seen (duplicate)
func (ld *LinkDeduplicator) AddURL(rawURL string) bool {
	normalized, err := ld.NormalizeURL(rawURL)
	if err != nil {
		// If normalization fails, don't add it and consider it a duplicate
		ld.duplicateCount++
		return false
	}

	if ld.seenURLs[normalized] {
		ld.duplicateCount++
		return false
	}

	ld.seenURLs[normalized] = true
	ld.canonicalMap[rawURL] = normalized
	return true
}

// GetDuplicateCount returns the number of duplicates found
func (ld *LinkDeduplicator) GetDuplicateCount() int {
	return ld.duplicateCount
}

// GetCanonicalURL returns the canonical version of a URL
func (ld *LinkDeduplicator) GetCanonicalURL(rawURL string) string {
	if canonical, exists := ld.canonicalMap[rawURL]; exists {
		return canonical
	}
	
	// If not in map, try to normalize it
	if normalized, err := ld.NormalizeURL(rawURL); err == nil {
		return normalized
	}
	
	// Return original if normalization fails
	return rawURL
}

// GetSeenURLsCount returns the number of unique URLs seen
func (ld *LinkDeduplicator) GetSeenURLsCount() int {
	return len(ld.seenURLs)
}

// Reset clears all stored URLs and counters
func (ld *LinkDeduplicator) Reset() {
	ld.seenURLs = make(map[string]bool)
	ld.canonicalMap = make(map[string]string)
	ld.duplicateCount = 0
}
