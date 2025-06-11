# Go Documentation Scraper

A powerful and flexible web scraper designed specifically for scraping documentation websites with advanced features including hierarchical organization, duplicate detection, content quality analysis, and comprehensive development tools.

## ✨ Features

### 🚀 Core Features

- **Multi-format Output**: Markdown, Text, and JSON formats
- **Depth Control**: Configurable crawling depth
- **Rate Limiting**: Respectful scraping with configurable delays
- **Content Extraction**: Intelligent content extraction from documentation pages
- **Robots.txt Support**: Optional respect for robots.txt files
- **Logging**: Comprehensive logging for monitoring and debugging

### 🎯 Advanced Features (NEW!)

- **🌳 Hierarchical Organization**: Automatically organize scraped content based on URL structure
- **🔍 Duplicate Link Detection**: Prevent redundant crawling with intelligent URL normalization
- **⭐ Content Quality Analysis**: Filter and score content based on quality metrics
- **🛠️ DevTools**: Development and debugging tools with profiling and validation

### 🔧 Optional Features

- **🔄 Proxy Support**: HTTP/HTTPS and SOCKS5 proxy rotation
- **⚡ Performance Tuning**: Configurable concurrent requests and timeouts
- **👤 User Agent Rotation**: Multiple user agents for better compatibility

## 📦 Installation

```bash
go build -o docscraper
```

## 🚀 Quick Start

### Basic Usage

1. Create a simple `config.yaml`:

  ```yaml
  root_url: "https://docs.example.com"
  output_format: "markdown"
  output_type: "single"
  output_dir: "output"
  min_delay: 2
  max_delay: 4
  max_depth: 2
  respect_robots: false
  log_file: "scraper.log"
  verbose: true
  ```

2. Run the scraper:

  ```bash
  ./docscraper
  ```

### Advanced Usage with All Features

1. Use the advanced configuration template:

  ```bash
  cp config-advanced.yaml config.yaml
  # Edit config.yaml as needed
  ./docscraper
  ```

2. Or use command-line flags for quick testing:

  ```bash
  # Enable all advanced features
  ./docscraper --root-url="https://pkg.go.dev/net/http" \
              --hierarchical \
              --dedupe \
              --quality \
              --debug \
              --profile
  ```

## 🎯 Advanced Features Guide

### 🌳 Hierarchical Organization

Automatically organize scraped content based on URL structure for better navigation:

```yaml
# Enable hierarchical organization
use_hierarchical_ordering: true

# Output will be organized in a tree structure:
# docs/
#   ├── api/
#   │   ├── index.md
#   │   ├── users/
#   │   │   └── index.md
#   │   └── auth/
#   │       └── index.md
#   └── guides/
#       └── index.md
```

**Benefits:**

- Natural navigation structure
- Hierarchical table of contents
- Nested directory organization
- Cross-references between related pages

### 🔍 Duplicate Link Detection

Prevent redundant crawling and content duplication:

```yaml
# Enable deduplication
enable_deduplication: true

# Configure URL normalization
deduplication:
  remove_fragments: true      # Treat page.html#section same as page.html
  remove_query_params: false  # Keep ?param=value differences
  ignore_case: true          # Treat Page.html same as page.html
  ignore_www: true           # Treat www.site.com same as site.com
  ignore_trailing_slash: true # Treat /path/ same as /path
```

**Benefits:**

- Faster scraping (fewer requests)
- Reduced storage requirements
- Cleaner output without duplicates
- Better resource utilization

### ⭐ Content Quality Analysis

Filter and analyze content based on quality metrics:

```yaml
# Enable quality analysis
enable_quality_analysis: true

# Configure quality standards
quality_analysis:
  min_score: 0.6              # Minimum quality score (0.0-1.0)
  min_word_count: 100         # Minimum word count
  require_title: true         # Skip pages without titles
  require_content: true       # Skip pages with minimal content
  skip_navigation: true       # Skip navigation/index pages
  filter_by_language: "en"    # Filter by detected language
  blacklisted_patterns:       # Skip pages containing these patterns
    - "404"
    - "not found"
    - "error"
    - "maintenance"
```

**Quality Metrics:**

- Word count and content ratio
- Presence of code blocks and images
- Header structure and title quality
- Navigation page detection
- Language detection
- Custom pattern filtering

### 🛠️ DevTools

Comprehensive development and debugging tools:

```yaml
# Enable development tools
enable_devtools: true

# Configure devtools features
devtools:
  enable_debug_mode: true         # Detailed debug logging
  enable_dry_run: false          # Simulate without scraping
  enable_profiling: true         # Performance profiling
  enable_progress_bar: true      # Show progress during scraping
  validation_level: "strict"     # Config validation: strict, normal, relaxed
  save_performance_report: true  # Save performance report to file
```

**DevTools Features:**

- **Configuration Validation**: Comprehensive config checking
- **Dry Run Mode**: Test configuration without actual scraping
- **Performance Profiling**: Track timing, memory, and errors
- **Progress Tracking**: Real-time progress with ETA
- **Debug Mode**: Detailed logging for troubleshooting

## 📋 Configuration Reference

### Required Settings

Option           | Type   | Description
---------------- | ------ | -----------------------------------------
`root_url`       | string | Starting URL for scraping
`output_format`  | string | Output format: "markdown", "text", "json"
`output_type`    | string | Output type: "single", "per-page"
`output_dir`     | string | Directory for output files
`min_delay`      | int    | Minimum delay between requests (seconds)
`max_delay`      | int    | Maximum delay between requests (seconds)
`max_depth`      | int    | Maximum crawling depth
`respect_robots` | bool   | Whether to respect robots.txt
`log_file`       | string | Path to log file
`verbose`        | bool   | Enable verbose logging

### Advanced Feature Settings

Option                      | Type | Default | Description
--------------------------- | ---- | ------- | ---------------------------------------
`use_hierarchical_ordering` | bool | false   | Enable hierarchical output organization
`enable_deduplication`      | bool | true    | Enable duplicate URL detection
`enable_quality_analysis`   | bool | false   | Enable content quality analysis
`enable_devtools`           | bool | false   | Enable development tools

### Command Line Flags

```bash
# Basic flags
--root-url string     Root URL to start scraping
--output string       Output format (default "markdown")
--format string       Output type (default "per-page")
--output-dir string   Output directory (default "output")
--max-depth int       Maximum crawling depth (default 3)
--verbose             Enable verbose logging

# Advanced feature flags
--hierarchical        Enable hierarchical output organization
--dedupe             Enable duplicate URL detection
--quality            Enable content quality analysis
--debug              Enable debug mode
--dry-run            Perform dry run without actual scraping
--profile            Enable performance profiling
```

./docscraper

````
### Advanced Configuration

For advanced features, see `config-advanced.yaml` for a complete example with all optional features.

## Configuration Reference

### Required Settings

Option           | Type   | Description
---------------- | ------ | --------------------------------------------
`root_url`       | string | Starting URL for scraping
`output_format`  | string | Output format: "markdown", "text", or "json"
`output_type`    | string | Output type: "single" or "per-page"
`output_dir`     | string | Directory for output files
`min_delay`      | int    | Minimum delay between requests (seconds)
`max_delay`      | int    | Maximum delay between requests (seconds)
`max_depth`      | int    | Maximum crawling depth
`respect_robots` | bool   | Whether to respect robots.txt
`log_file`       | string | Path to log file
`verbose`        | bool   | Enable verbose logging

### Optional Settings

#### Proxy Configuration

```yaml
# Optional proxy support - remove section if not needed
proxies:
    - "http://proxy1.example.com:8080"
    - "http://user:pass@proxy2.example.com:3128"
    - "socks5://127.0.0.1:1080"
````

#### Performance Tuning

```yaml
# Optional performance settings - remove if defaults are fine
concurrent_requests: 3        # Default: 2
request_timeout: 45          # Default: 30 seconds
retry_attempts: 2            # Default: 0 (no retries)
ignore_ssl_errors: false     # Default: false
```

#### User Agent Rotation

```yaml
# Optional custom user agents - remove to use defaults
user_agents:
    - "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    - "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
```

## Usage Examples

### Basic Documentation Scraping

```bash
# Scrape a documentation site with default settings
./docscraper
```

### With Proxy Support

```yaml
# config.yaml
root_url: "https://docs.example.com"
# ... other basic settings ...
proxies:
    - "http://proxy.example.com:8080"
```

### High-Performance Scraping

```yaml
# config.yaml
root_url: "https://docs.example.com"
# ... other basic settings ...
concurrent_requests: 5
request_timeout: 60
retry_attempts: 3
```

## Output Formats

### Markdown (default)

Generates clean Markdown files suitable for documentation systems.

### Text

Plain text output for simple processing.

### JSON

Structured JSON output for programmatic processing:

```json
{
  "title": "Page Title",
  "url": "https://example.com/page",
  "content": "Page content...",
  "timestamp": "2023-12-01T10:30:00Z",
  "depth": 1
}
```

## Best Practices

### Respectful Scraping

- Always set appropriate delays (`min_delay`, `max_delay`)
- Respect robots.txt when possible
- Use reasonable concurrent request limits
- Monitor the target site's load

### Proxy Usage

- Use proxies responsibly and legally
- Rotate proxies to avoid rate limiting
- Test proxy connectivity before scraping
- Consider geographic restrictions

### Performance Optimization

- Start with conservative settings
- Increase concurrency gradually
- Monitor memory usage for large sites
- Use appropriate timeouts

## Troubleshooting

### Common Issues

1. **SSL Certificate Errors**

  ```yaml
  ignore_ssl_errors: true  # Use with caution
  ```

2. **Rate Limiting**

  ```yaml
  min_delay: 5
  max_delay: 10
  concurrent_requests: 1
  ```

3. **Proxy Connection Issues**

  ```yaml
  request_timeout: 60
  retry_attempts: 3
  ```

### Debug Mode

Enable verbose logging to see detailed information:

```yaml
verbose: true
```

## Project Structure

```
godocscraper/
├── main.go                 # Main application entry point
├── config/
│   └── config.go          # Configuration management
├── scraper/
│   ├── scraper.go         # Core scraping logic
│   └── extractor.go       # Content extraction
├── output/
│   └── generator.go       # Output generation
├── utils/
│   └── file.go           # File utilities
├── config.yaml           # Basic configuration
└── config-advanced.yaml  # Advanced configuration example
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Changelog

### Version 1.1.0

- ✅ Added proxy support (HTTP/HTTPS/SOCKS5)
- ✅ Added performance tuning options
- ✅ Added retry logic for failed requests
- ✅ Added SSL certificate options
- ✅ Improved configuration validation
- ✅ All new features are optional and backward compatible

### Version 1.0.0

- Initial release with basic scraping functionality
