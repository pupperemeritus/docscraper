package output

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"docscraper/config"
)

// DocumentNode represents a node in the documentation tree (local copy to avoid import cycle)
type DocumentNode struct {
	URL      string            `json:"url"`
	Path     string            `json:"path"`
	Title    string            `json:"title"`
	Content  string            `json:"content,omitempty"`
	Depth    int               `json:"depth"`
	Level    int               `json:"level"`
	Parent   *DocumentNode     `json:"-"`
	Children []*DocumentNode   `json:"children"`
	Index    int               `json:"index"`
	Timestamp time.Time        `json:"timestamp"`
}

// DocumentTree represents the complete documentation tree structure (local copy)
type DocumentTree struct {
	Root       *DocumentNode            `json:"root"`
	NodeMap    map[string]*DocumentNode `json:"-"`
	MaxDepth   int                      `json:"max_depth"`
	TotalNodes int                      `json:"total_nodes"`
	BuildTime  time.Time                `json:"build_time"`
}

// HierarchicalGenerator handles hierarchical output generation
type HierarchicalGenerator struct {
	config *config.Config
	tree   *DocumentTree
}

// NewHierarchical creates a new hierarchical output generator
func NewHierarchical(cfg *config.Config, pages []PageData) *HierarchicalGenerator {
	// Convert PageData to DocumentNode and build tree
	tree := buildTreeFromPages(pages)

	return &HierarchicalGenerator{
		config: cfg,
		tree:   tree,
	}
}

// buildTreeFromPages creates a document tree from page data
func buildTreeFromPages(pages []PageData) *DocumentTree {
	// Create root node
	root := &DocumentNode{
		Title:    "Root",
		Children: make([]*DocumentNode, 0),
	}

	// Create nodes and a map for quick lookup
	nodeMap := make(map[string]*DocumentNode)
	nodes := make([]*DocumentNode, len(pages))

	for i, page := range pages {
		node := &DocumentNode{
			URL:       page.URL,
			Path:      extractPathFromURL(page.URL),
			Title:     page.Title,
			Content:   page.Content,
			Depth:     page.Depth,
			Level:     0,
			Children:  make([]*DocumentNode, 0),
			Index:     i,
			Timestamp: page.Timestamp,
		}
		nodes[i] = node
		nodeMap[page.URL] = node
	}

	// Build hierarchy based on URL paths
	for _, node := range nodes {
		parent := findParentNode(node, nodes)
		if parent != nil {
			parent.Children = append(parent.Children, node)
			node.Parent = parent
			node.Level = parent.Level + 1
		} else {
			root.Children = append(root.Children, node)
			node.Parent = root
			node.Level = 1
		}
	}

	return &DocumentTree{
		Root:       root,
		NodeMap:    nodeMap,
		MaxDepth:   calculateMaxDepth(root),
		TotalNodes: len(pages),
		BuildTime:  time.Now(),
	}
}

// extractPathFromURL extracts the path component from a URL
func extractPathFromURL(urlStr string) string {
	if u, err := url.Parse(urlStr); err == nil {
		return u.Path
	}
	return ""
}

// findParentNode finds the most appropriate parent for a node based on URL hierarchy
func findParentNode(node *DocumentNode, allNodes []*DocumentNode) *DocumentNode {
	if node.Path == "" || node.Path == "/" {
		return nil
	}

	// Find the node with the longest matching path prefix
	var bestParent *DocumentNode
	maxMatchLength := 0

	for _, candidate := range allNodes {
		if candidate == node || candidate.Path == "" {
			continue
		}

		if strings.HasPrefix(node.Path, candidate.Path) && len(candidate.Path) > maxMatchLength {
			// Additional check: ensure it's actually a parent path, not just a prefix
			if node.Path == candidate.Path+"/" || 
			   (len(node.Path) > len(candidate.Path) && node.Path[len(candidate.Path)] == '/') {
				bestParent = candidate
				maxMatchLength = len(candidate.Path)
			}
		}
	}

	return bestParent
}

// calculateMaxDepth calculates the maximum depth of the tree
func calculateMaxDepth(root *DocumentNode) int {
	if root == nil || len(root.Children) == 0 {
		return 0
	}

	maxDepth := 0
	for _, child := range root.Children {
		depth := 1 + calculateMaxDepth(child)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// GetAllNodes returns all nodes in the tree
func (t *DocumentTree) GetAllNodes() []*DocumentNode {
	var nodes []*DocumentNode
	t.collectNodes(t.Root, &nodes)
	return nodes
}

// collectNodes recursively collects all nodes
func (t *DocumentTree) collectNodes(node *DocumentNode, nodes *[]*DocumentNode) {
	if node == nil {
		return
	}

	if node.Title != "Root" { // Skip root node
		*nodes = append(*nodes, node)
	}

	for _, child := range node.Children {
		t.collectNodes(child, nodes)
	}
}

// Generate creates hierarchically organized output
func (h *HierarchicalGenerator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(h.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	switch h.config.OutputFormat {
	case "markdown":
		return h.generateHierarchicalMarkdown()
	case "text":
		return h.generateHierarchicalText()
	case "json":
		return h.generateHierarchicalJSON()
	default:
		return fmt.Errorf("unsupported output format: %s", h.config.OutputFormat)
	}
}

// generateHierarchicalMarkdown generates hierarchically ordered markdown
func (h *HierarchicalGenerator) generateHierarchicalMarkdown() error {
	if h.config.OutputType == "single" {
		return h.generateSingleHierarchicalMarkdown()
	}
	return h.generatePerPageHierarchicalMarkdown()
}

// generateSingleHierarchicalMarkdown creates a single markdown file with hierarchical organization
func (h *HierarchicalGenerator) generateSingleHierarchicalMarkdown() error {
	filename := filepath.Join(h.config.OutputDir, "documentation_hierarchical.md")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "# Documentation Scrape Results (Hierarchical)\n\n")
	fmt.Fprintf(file, "**Scraped from:** %s  \n", h.config.RootURL)
	fmt.Fprintf(file, "**Generated:** %s  \n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "**Total Pages:** %d\n\n", len(h.tree.GetAllNodes()))
	fmt.Fprintf(file, "---\n\n")

	// Generate hierarchical table of contents
	fmt.Fprintf(file, "## Table of Contents\n\n")
	h.writeHierarchicalTOC(file, h.tree.Root, 0)
	fmt.Fprintf(file, "\n---\n\n")

	// Write content in hierarchical order
	h.writeHierarchicalContent(file, h.tree.Root, 0)

	return nil
}

// generatePerPageHierarchicalMarkdown creates separate files organized hierarchically
func (h *HierarchicalGenerator) generatePerPageHierarchicalMarkdown() error {
	// Create directory structure that mirrors the document hierarchy
	err := h.createHierarchicalDirectories(h.tree.Root, h.config.OutputDir)
	if err != nil {
		return err
	}

	// Generate files for each node
	err = h.writeHierarchicalFiles(h.tree.Root, h.config.OutputDir)
	if err != nil {
		return err
	}

	// Create main index file
	return h.generateHierarchicalIndex()
}

// writeHierarchicalTOC writes a hierarchical table of contents
func (h *HierarchicalGenerator) writeHierarchicalTOC(file *os.File, node *DocumentNode, level int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", level)
	if node.Title != "" && node.Title != "Root" { // Skip root node
		anchor := h.createAnchor(node.Title)
		fmt.Fprintf(file, "%s- [%s](#%s)\n", indent, node.Title, anchor)
	}

	// Sort children for consistent ordering
	children := make([]*DocumentNode, len(node.Children))
	copy(children, node.Children)
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	for _, child := range children {
		h.writeHierarchicalTOC(file, child, level+1)
	}
}

// writeHierarchicalContent writes content in hierarchical order
func (h *HierarchicalGenerator) writeHierarchicalContent(file *os.File, node *DocumentNode, level int) {
	if node == nil {
		return
	}

	if node.Title != "" && node.Title != "Root" { // Skip root node
		// Write section header based on hierarchy level
		headerLevel := level + 1
		if headerLevel > 6 {
			headerLevel = 6 // Markdown only supports up to h6
		}
		headerPrefix := strings.Repeat("#", headerLevel)
		anchor := h.createAnchor(node.Title)
		
		fmt.Fprintf(file, "%s %s {#%s}\n\n", headerPrefix, node.Title, anchor)
		fmt.Fprintf(file, "**URL:** %s  \n", node.URL)
		fmt.Fprintf(file, "**Scraped:** %s\n\n", node.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(file, "%s\n\n", node.Content)
	}

	// Sort children for consistent ordering
	children := make([]*DocumentNode, len(node.Children))
	copy(children, node.Children)
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	for _, child := range children {
		h.writeHierarchicalContent(file, child, level+1)
	}
}

// createHierarchicalDirectories creates directory structure mirroring document hierarchy
func (h *HierarchicalGenerator) createHierarchicalDirectories(node *DocumentNode, basePath string) error {
	if node == nil {
		return nil
	}

	var currentPath string
	if node.Title != "" && node.Title != "Root" { // Skip root node
		safeName := h.createSafeDirectoryName(node.Title)
		currentPath = filepath.Join(basePath, safeName)
		if err := os.MkdirAll(currentPath, 0755); err != nil {
			return err
		}
	} else {
		currentPath = basePath
	}

	for _, child := range node.Children {
		if err := h.createHierarchicalDirectories(child, currentPath); err != nil {
			return err
		}
	}

	return nil
}

// writeHierarchicalFiles writes individual files in hierarchical structure
func (h *HierarchicalGenerator) writeHierarchicalFiles(node *DocumentNode, basePath string) error {
	if node == nil {
		return nil
	}

	var currentPath string
	if node.Title != "" && node.Title != "Root" { // Skip root node
		safeName := h.createSafeDirectoryName(node.Title)
		currentPath = filepath.Join(basePath, safeName)
		
		// Write content file
		filename := filepath.Join(currentPath, "index.md")
		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		fmt.Fprintf(file, "# %s\n\n", node.Title)
		fmt.Fprintf(file, "**URL:** %s  \n", node.URL)
		fmt.Fprintf(file, "**Scraped:** %s\n\n", node.Timestamp.Format(time.RFC3339))
		
		// Add navigation to children if any
		if len(node.Children) > 0 {
			fmt.Fprintf(file, "## Sub-sections\n\n")
			for _, child := range node.Children {
				childSafeName := h.createSafeDirectoryName(child.Title)
				fmt.Fprintf(file, "- [%s](%s/index.md)\n", child.Title, childSafeName)
			}
			fmt.Fprintf(file, "\n")
		}
		
		fmt.Fprintf(file, "---\n\n")
		fmt.Fprintf(file, "%s\n", node.Content)
		file.Close()
	} else {
		currentPath = basePath
	}

	for _, child := range node.Children {
		if err := h.writeHierarchicalFiles(child, currentPath); err != nil {
			return err
		}
	}

	return nil
}

// generateHierarchicalIndex creates a main index file for hierarchical structure
func (h *HierarchicalGenerator) generateHierarchicalIndex() error {
	filename := filepath.Join(h.config.OutputDir, "index.md")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# Documentation Index (Hierarchical)\n\n")
	fmt.Fprintf(file, "**Scraped from:** %s  \n", h.config.RootURL)
	fmt.Fprintf(file, "**Generated:** %s  \n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "**Total Pages:** %d\n\n", len(h.tree.GetAllNodes()))
	fmt.Fprintf(file, "## Structure\n\n")

	h.writeHierarchicalIndex(file, h.tree.Root, 0)

	return nil
}

// writeHierarchicalIndex writes hierarchical index links
func (h *HierarchicalGenerator) writeHierarchicalIndex(file *os.File, node *DocumentNode, level int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", level)
	if node.Title != "" && node.Title != "Root" { // Skip root node
		safeName := h.createSafeDirectoryName(node.Title)
		fmt.Fprintf(file, "%s- [%s](%s/index.md)\n", indent, node.Title, safeName)
	}

	// Sort children for consistent ordering
	children := make([]*DocumentNode, len(node.Children))
	copy(children, node.Children)
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	for _, child := range children {
		h.writeHierarchicalIndex(file, child, level+1)
	}
}

// generateHierarchicalText generates hierarchical text output
func (h *HierarchicalGenerator) generateHierarchicalText() error {
	filename := filepath.Join(h.config.OutputDir, "documentation_hierarchical.txt")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "DOCUMENTATION SCRAPE RESULTS (HIERARCHICAL)\n")
	fmt.Fprintf(file, "===========================================\n\n")
	fmt.Fprintf(file, "Scraped from: %s\n", h.config.RootURL)
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "Total Pages: %d\n\n", len(h.tree.GetAllNodes()))

	h.writeHierarchicalTextContent(file, h.tree.Root, 0)

	return nil
}

// writeHierarchicalTextContent writes text content in hierarchical order
func (h *HierarchicalGenerator) writeHierarchicalTextContent(file *os.File, node *DocumentNode, level int) {
	if node == nil {
		return
	}

	if node.Title != "" && node.Title != "Root" { // Skip root node
		indent := strings.Repeat("  ", level)
		separator := strings.Repeat("=", 80-len(indent))
		
		fmt.Fprintf(file, "%s%s\n", indent, separator)
		fmt.Fprintf(file, "%sTITLE: %s\n", indent, node.Title)
		fmt.Fprintf(file, "%sURL: %s\n", indent, node.URL)
		fmt.Fprintf(file, "%sSCRAPED: %s\n", indent, node.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(file, "%s%s\n\n", indent, separator)
		
		// Indent content
		contentLines := strings.Split(node.Content, "\n")
		for _, line := range contentLines {
			fmt.Fprintf(file, "%s%s\n", indent, line)
		}
		fmt.Fprintf(file, "\n")
	}

	// Sort children for consistent ordering
	children := make([]*DocumentNode, len(node.Children))
	copy(children, node.Children)
	sort.Slice(children, func(i, j int) bool {
		return children[i].Title < children[j].Title
	})

	for _, child := range children {
		h.writeHierarchicalTextContent(file, child, level+1)
	}
}

// generateHierarchicalJSON generates hierarchical JSON output
func (h *HierarchicalGenerator) generateHierarchicalJSON() error {
	filename := filepath.Join(h.config.OutputDir, "documentation_hierarchical.json")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	output := map[string]interface{}{
		"root_url":    h.config.RootURL,
		"scraped_at":  time.Now().Format(time.RFC3339),
		"total_pages": len(h.tree.GetAllNodes()),
		"hierarchy":   h.nodeToJSON(h.tree.Root),
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// nodeToJSON converts a document node to JSON representation
func (h *HierarchicalGenerator) nodeToJSON(node *DocumentNode) map[string]interface{} {
	if node == nil {
		return nil
	}

	result := map[string]interface{}{
		"url":       node.URL,
		"path":      node.Path,
		"title":     node.Title,
		"content":   node.Content,
		"timestamp": node.Timestamp.Format(time.RFC3339),
		"depth":     node.Depth,
		"level":     node.Level,
		"index":     node.Index,
		"children":  make([]map[string]interface{}, 0),
	}

	for _, child := range node.Children {
		childJSON := h.nodeToJSON(child)
		if childJSON != nil {
			result["children"] = append(result["children"].([]map[string]interface{}), childJSON)
		}
	}

	return result
}

// createSafeDirectoryName creates a filesystem-safe directory name
func (h *HierarchicalGenerator) createSafeDirectoryName(title string) string {
	// Replace unsafe characters
	safe := regexp.MustCompile(`[^a-zA-Z0-9\-_\s]`).ReplaceAllString(title, "")
	safe = regexp.MustCompile(`\s+`).ReplaceAllString(safe, "_")
	safe = strings.Trim(safe, "_")
	
	// Limit length
	if len(safe) > 30 {
		safe = safe[:30]
	}
	
	if safe == "" {
		safe = "untitled"
	}
	
	return strings.ToLower(safe)
}

// createAnchor creates a markdown anchor from a title
func (h *HierarchicalGenerator) createAnchor(title string) string {
	// Convert to lowercase, replace spaces with dashes, remove special characters
	anchor := strings.ToLower(title)
	anchor = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(anchor, "")
	anchor = regexp.MustCompile(`\s+`).ReplaceAllString(anchor, "-")
	anchor = strings.Trim(anchor, "-")
	return anchor
}
