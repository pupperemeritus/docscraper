package scraper

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

// ContentQuality represents quality metrics for content
type ContentQuality struct {
	Score            float64        `json:"score"`
	WordCount        int            `json:"word_count"`
	CodeBlockCount   int            `json:"code_block_count"`
	ImageCount       int            `json:"image_count"`
	LinkCount        int            `json:"link_count"`
	EmptyLineRatio   float64        `json:"empty_line_ratio"`
	ContentRatio     float64        `json:"content_ratio"`
	HasTitle         bool           `json:"has_title"`
	HasHeaders       bool           `json:"has_headers"`
	IsNavigationPage bool           `json:"is_navigation_page"`
	Language         string         `json:"language"`
	Issues           []QualityIssue `json:"issues"`
	Tags             []string       `json:"tags"`
}

// QualityIssue represents a content quality issue
type QualityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// NodeMetadata contains metadata about a document node
type NodeMetadata struct {
	WordCount     int            `json:"word_count"`
	LastModified  time.Time      `json:"last_modified"`
	ContentType   string         `json:"content_type"`
	HasCodeBlocks bool           `json:"has_code_blocks"`
	HasImages     bool           `json:"has_images"`
	Quality       ContentQuality `json:"quality"`
	Tags          []string       `json:"tags"`
}

// DocumentNode represents a node in the documentation tree
type DocumentNode struct {
	URL      string          `json:"url"`
	Path     string          `json:"path"`
	Title    string          `json:"title"`
	Content  string          `json:"content,omitempty"`
	Depth    int             `json:"depth"`
	Level    int             `json:"level"`
	Parent   *DocumentNode   `json:"-"`
	Children []*DocumentNode `json:"children"`
	Metadata NodeMetadata    `json:"metadata"`
	Index    int             `json:"index"`
}

// DocumentTree represents the complete documentation tree structure
type DocumentTree struct {
	Root       *DocumentNode            `json:"root"`
	NodeMap    map[string]*DocumentNode `json:"-"`
	MaxDepth   int                      `json:"max_depth"`
	TotalNodes int                      `json:"total_nodes"`
	BuildTime  time.Time                `json:"build_time"`
}

// SortCriteria defines how to sort tree nodes
type SortCriteria string

const (
	SortByIndex SortCriteria = "index"
	SortByTitle SortCriteria = "title"
	SortByURL   SortCriteria = "url"
	SortByDate  SortCriteria = "date"
)

// TreeConfig defines configuration for tree building
type TreeConfig struct {
	BuildingStrategy      string       `yaml:"building_strategy"`
	UseURLHierarchy       bool         `yaml:"use_url_hierarchy"`
	UseBreadcrumbs        bool         `yaml:"use_breadcrumbs"`
	UseNavigation         bool         `yaml:"use_navigation"`
	FallbackToRoot        bool         `yaml:"fallback_to_root"`
	SortChildren          bool         `yaml:"sort_children"`
	SortBy                SortCriteria `yaml:"sort_by"`
	SortOrder             string       `yaml:"sort_order"`
	AutoIndex             bool         `yaml:"auto_index"`
	PreserveOriginalOrder bool         `yaml:"preserve_original_order"`
}

// ScrapedContent represents content scraped from a page
type ScrapedContent struct {
	URL      string
	Title    string
	Content  string
	Metadata NodeMetadata
}

// TreeBuilder builds documentation trees from scraped content
type TreeBuilder struct {
	config TreeConfig
}

// NewTreeBuilder creates a new tree builder with the given configuration
func NewTreeBuilder(config TreeConfig) *TreeBuilder {
	return &TreeBuilder{
		config: config,
	}
}

// BuildTree builds a documentation tree from URLs and content
func (tb *TreeBuilder) BuildTree(urls []string, contents map[string]ScrapedContent) *DocumentTree {
	tree := &DocumentTree{
		NodeMap:   make(map[string]*DocumentNode),
		BuildTime: time.Now(),
	}

	// Create root node
	if len(urls) > 0 {
		rootURL := urls[0]
		if content, exists := contents[rootURL]; exists {
			tree.Root = &DocumentNode{
				URL:      rootURL,
				Path:     tb.extractPath(rootURL),
				Title:    content.Title,
				Content:  content.Content,
				Depth:    0,
				Level:    0,
				Children: make([]*DocumentNode, 0),
				Metadata: content.Metadata,
				Index:    0,
			}
			tree.NodeMap[rootURL] = tree.Root
		}
	}

	// Add all other nodes
	for i, url := range urls {
		if i == 0 && tree.Root != nil {
			continue // Skip root URL as it's already added
		}
		if content, exists := contents[url]; exists {
			tb.AddNode(tree, url, content)
		}
	}

	// Calculate depth and levels
	tb.CalculateDepthAndLevel(tree)

	// Sort children if configured
	if tb.config.SortChildren {
		tb.sortTreeChildren(tree.Root)
	}

	// Auto-index if configured
	if tb.config.AutoIndex {
		tb.autoIndexNodes(tree.Root, 0)
	}

	// Update tree statistics
	tree.TotalNodes = len(tree.NodeMap)
	tree.MaxDepth = tb.calculateMaxDepth(tree.Root)

	return tree
}

// AddNode adds a new node to the tree
func (tb *TreeBuilder) AddNode(tree *DocumentTree, url string, content ScrapedContent) error {
	if tree.NodeMap[url] != nil {
		return fmt.Errorf("node with URL %s already exists", url)
	}

	node := &DocumentNode{
		URL:      url,
		Path:     tb.extractPath(url),
		Title:    content.Title,
		Content:  content.Content,
		Children: make([]*DocumentNode, 0),
		Metadata: content.Metadata,
	}

	// Determine parent
	parent := tb.DetermineParent(node, tree)
	if parent != nil {
		node.Parent = parent
		parent.Children = append(parent.Children, node)
	} else if tb.config.FallbackToRoot && tree.Root != nil {
		node.Parent = tree.Root
		tree.Root.Children = append(tree.Root.Children, node)
	}

	tree.NodeMap[url] = node
	return nil
}

// DetermineParent determines the parent node for a given node
func (tb *TreeBuilder) DetermineParent(node *DocumentNode, tree *DocumentTree) *DocumentNode {
	if tb.config.UseURLHierarchy {
		return tb.findParentByURLHierarchy(node, tree)
	}
	// Other parent detection strategies can be added here
	return nil
}

// findParentByURLHierarchy finds parent based on URL path hierarchy
func (tb *TreeBuilder) findParentByURLHierarchy(node *DocumentNode, tree *DocumentTree) *DocumentNode {
	parsedURL, err := url.Parse(node.URL)
	if err != nil {
		return nil
	}

	urlPath := parsedURL.Path
	if urlPath == "/" || urlPath == "" {
		return tree.Root
	}

	// Try to find parent by removing the last path segment
	parentPath := path.Dir(urlPath)
	if parentPath == "." {
		parentPath = "/"
	}

	// Look for existing nodes that could be parents
	for existingURL, existingNode := range tree.NodeMap {
		existingParsed, err := url.Parse(existingURL)
		if err != nil {
			continue
		}

		if existingParsed.Host == parsedURL.Host {
			if existingParsed.Path == parentPath ||
				(parentPath != "/" && strings.HasPrefix(urlPath, existingParsed.Path+"/")) {
				return existingNode
			}
		}
	}

	return nil
}

// extractPath extracts the path component from a URL
func (tb *TreeBuilder) extractPath(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsedURL.Path
}

// CalculateDepthAndLevel calculates depth and level for all nodes
func (tb *TreeBuilder) CalculateDepthAndLevel(tree *DocumentTree) {
	if tree.Root != nil {
		tb.calculateNodeDepthAndLevel(tree.Root, 0, 0)
	}
}

// calculateNodeDepthAndLevel recursively calculates depth and level
func (tb *TreeBuilder) calculateNodeDepthAndLevel(node *DocumentNode, depth, level int) {
	node.Depth = depth
	node.Level = level

	for _, child := range node.Children {
		tb.calculateNodeDepthAndLevel(child, depth+1, level+1)
	}
}

// SortChildren sorts the children of a node according to the configured criteria
func (tb *TreeBuilder) SortChildren(node *DocumentNode, sortBy SortCriteria) {
	if len(node.Children) <= 1 {
		return
	}

	switch sortBy {
	case SortByIndex:
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Index < node.Children[j].Index
		})
	case SortByTitle:
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Title < node.Children[j].Title
		})
	case SortByURL:
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].URL < node.Children[j].URL
		})
	case SortByDate:
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Metadata.LastModified.Before(node.Children[j].Metadata.LastModified)
		})
	}

	// Recursively sort children's children
	for _, child := range node.Children {
		tb.SortChildren(child, sortBy)
	}
}

// sortTreeChildren sorts all children in the tree
func (tb *TreeBuilder) sortTreeChildren(root *DocumentNode) {
	tb.SortChildren(root, tb.config.SortBy)
}

// autoIndexNodes automatically assigns indices to nodes
func (tb *TreeBuilder) autoIndexNodes(node *DocumentNode, startIndex int) int {
	node.Index = startIndex
	currentIndex := startIndex + 1

	for _, child := range node.Children {
		currentIndex = tb.autoIndexNodes(child, currentIndex)
	}

	return currentIndex
}

// calculateMaxDepth calculates the maximum depth in the tree
func (tb *TreeBuilder) calculateMaxDepth(node *DocumentNode) int {
	if node == nil {
		return 0
	}

	maxChildDepth := 0
	for _, child := range node.Children {
		childDepth := tb.calculateMaxDepth(child)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return node.Depth + maxChildDepth
}

// Tree traversal methods

// TraverseDepthFirst performs depth-first traversal of the tree
func (dt *DocumentTree) TraverseDepthFirst(visitor func(*DocumentNode) error) error {
	if dt.Root == nil {
		return nil
	}
	return dt.traverseDepthFirstNode(dt.Root, visitor)
}

// traverseDepthFirstNode recursively traverses nodes depth-first
func (dt *DocumentTree) traverseDepthFirstNode(node *DocumentNode, visitor func(*DocumentNode) error) error {
	if err := visitor(node); err != nil {
		return err
	}

	for _, child := range node.Children {
		if err := dt.traverseDepthFirstNode(child, visitor); err != nil {
			return err
		}
	}

	return nil
}

// TraverseBreadthFirst performs breadth-first traversal of the tree
func (dt *DocumentTree) TraverseBreadthFirst(visitor func(*DocumentNode) error) error {
	if dt.Root == nil {
		return nil
	}

	queue := []*DocumentNode{dt.Root}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if err := visitor(node); err != nil {
			return err
		}

		queue = append(queue, node.Children...)
	}

	return nil
}

// FindNode finds a node by URL
func (dt *DocumentTree) FindNode(url string) *DocumentNode {
	return dt.NodeMap[url]
}

// GetAllNodes returns all nodes in the tree
func (dt *DocumentTree) GetAllNodes() []*DocumentNode {
	nodes := make([]*DocumentNode, 0, len(dt.NodeMap))
	for _, node := range dt.NodeMap {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodesByLevel returns all nodes at a specific level
func (dt *DocumentTree) GetNodesByLevel(level int) []*DocumentNode {
	var nodes []*DocumentNode

	dt.TraverseDepthFirst(func(node *DocumentNode) error {
		if node.Level == level {
			nodes = append(nodes, node)
		}
		return nil
	})

	return nodes
}
