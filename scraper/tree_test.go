package scraper

import (
	"testing"
	"time"
)

func TestNewTreeBuilder(t *testing.T) {
	config := TreeConfig{
		BuildingStrategy: "url_path",
		UseURLHierarchy:  true,
		SortChildren:     true,
		SortBy:           SortByTitle,
		SortOrder:        "asc",
		AutoIndex:        true,
		FallbackToRoot:   true,
	}

	builder := NewTreeBuilder(config)
	if builder == nil {
		t.Fatal("NewTreeBuilder() returned nil")
	}

	if builder.config.BuildingStrategy != "url_path" {
		t.Errorf("Expected building strategy 'url_path', got %s", builder.config.BuildingStrategy)
	}
}

func TestTreeBuilder_BuildTree(t *testing.T) {
	config := TreeConfig{
		UseURLHierarchy: true,
		SortChildren:    true,
		SortBy:          SortByTitle,
		AutoIndex:       true,
		FallbackToRoot:  true,
	}

	builder := NewTreeBuilder(config)

	// Create test content
	urls := []string{
		"https://example.com/",
		"https://example.com/docs/",
		"https://example.com/docs/guide/",
		"https://example.com/docs/api/",
		"https://example.com/docs/guide/getting-started",
		"https://example.com/docs/api/reference",
	}

	contents := map[string]ScrapedContent{
		"https://example.com/": {
			URL:     "https://example.com/",
			Title:   "Home",
			Content: "Welcome to the documentation",
			Metadata: NodeMetadata{
				WordCount:     5,
				LastModified:  time.Now(),
				ContentType:   "text/html",
				HasCodeBlocks: false,
			},
		},
		"https://example.com/docs/": {
			URL:     "https://example.com/docs/",
			Title:   "Documentation",
			Content: "Documentation overview",
			Metadata: NodeMetadata{
				WordCount:     3,
				LastModified:  time.Now(),
				ContentType:   "text/html",
				HasCodeBlocks: false,
			},
		},
		"https://example.com/docs/guide/": {
			URL:     "https://example.com/docs/guide/",
			Title:   "Guide",
			Content: "User guide",
			Metadata: NodeMetadata{
				WordCount:     2,
				LastModified:  time.Now(),
				ContentType:   "text/html",
				HasCodeBlocks: false,
			},
		},
		"https://example.com/docs/api/": {
			URL:     "https://example.com/docs/api/",
			Title:   "API",
			Content: "API documentation",
			Metadata: NodeMetadata{
				WordCount:     2,
				LastModified:  time.Now(),
				ContentType:   "text/html",
				HasCodeBlocks: true,
			},
		},
		"https://example.com/docs/guide/getting-started": {
			URL:     "https://example.com/docs/guide/getting-started",
			Title:   "Getting Started",
			Content: "How to get started",
			Metadata: NodeMetadata{
				WordCount:     4,
				LastModified:  time.Now(),
				ContentType:   "text/html",
				HasCodeBlocks: false,
			},
		},
		"https://example.com/docs/api/reference": {
			URL:     "https://example.com/docs/api/reference",
			Title:   "API Reference",
			Content: "Complete API reference",
			Metadata: NodeMetadata{
				WordCount:     3,
				LastModified:  time.Now(),
				ContentType:   "text/html",
				HasCodeBlocks: true,
			},
		},
	}

	tree := builder.BuildTree(urls, contents)

	// Test basic tree structure
	if tree == nil {
		t.Fatal("BuildTree() returned nil")
	}

	if tree.Root == nil {
		t.Fatal("Tree root is nil")
	}

	if tree.Root.Title != "Home" {
		t.Errorf("Expected root title 'Home', got %s", tree.Root.Title)
	}

	if tree.TotalNodes != 6 {
		t.Errorf("Expected 6 total nodes, got %d", tree.TotalNodes)
	}

	// Test node mapping
	if tree.NodeMap["https://example.com/"] == nil {
		t.Error("Root node not found in NodeMap")
	}

	if tree.NodeMap["https://example.com/docs/"] == nil {
		t.Error("Docs node not found in NodeMap")
	}

	// Test hierarchy
	docsNode := tree.NodeMap["https://example.com/docs/"]
	if docsNode.Parent != tree.Root {
		t.Error("Docs node should be child of root")
	}

	if len(tree.Root.Children) == 0 {
		t.Error("Root should have children")
	}

	// Test depth calculation
	if tree.Root.Depth != 0 {
		t.Errorf("Expected root depth 0, got %d", tree.Root.Depth)
	}

	if docsNode.Depth != 1 {
		t.Errorf("Expected docs depth 1, got %d", docsNode.Depth)
	}
}

func TestTreeBuilder_AddNode(t *testing.T) {
	config := TreeConfig{
		UseURLHierarchy: true,
		FallbackToRoot:  true,
	}

	builder := NewTreeBuilder(config)

	// Create a simple tree with root
	tree := &DocumentTree{
		NodeMap:   make(map[string]*DocumentNode),
		BuildTime: time.Now(),
	}

	// Add root
	rootContent := ScrapedContent{
		URL:     "https://example.com/",
		Title:   "Home",
		Content: "Home page",
		Metadata: NodeMetadata{
			WordCount: 2,
		},
	}

	tree.Root = &DocumentNode{
		URL:      "https://example.com/",
		Title:    "Home",
		Content:  "Home page",
		Children: make([]*DocumentNode, 0),
		Metadata: rootContent.Metadata,
	}
	tree.NodeMap["https://example.com/"] = tree.Root

	// Add child node
	childContent := ScrapedContent{
		URL:     "https://example.com/docs",
		Title:   "Documentation",
		Content: "Docs",
		Metadata: NodeMetadata{
			WordCount: 1,
		},
	}

	err := builder.AddNode(tree, "https://example.com/docs", childContent)
	if err != nil {
		t.Fatalf("AddNode() failed: %v", err)
	}

	// Verify node was added
	if tree.NodeMap["https://example.com/docs"] == nil {
		t.Error("Child node not added to NodeMap")
	}

	// Test duplicate addition
	err = builder.AddNode(tree, "https://example.com/docs", childContent)
	if err == nil {
		t.Error("Expected error when adding duplicate node")
	}
}

func TestDocumentTree_TraverseDepthFirst(t *testing.T) {
	// Create a simple tree structure
	root := &DocumentNode{
		URL:      "https://example.com/",
		Title:    "Root",
		Children: make([]*DocumentNode, 0),
	}

	child1 := &DocumentNode{
		URL:    "https://example.com/page1",
		Title:  "Page 1",
		Parent: root,
	}

	child2 := &DocumentNode{
		URL:    "https://example.com/page2",
		Title:  "Page 2",
		Parent: root,
	}

	grandchild := &DocumentNode{
		URL:    "https://example.com/page1/sub",
		Title:  "Subpage",
		Parent: child1,
	}

	child1.Children = []*DocumentNode{grandchild}
	root.Children = []*DocumentNode{child1, child2}

	tree := &DocumentTree{
		Root: root,
		NodeMap: map[string]*DocumentNode{
			"https://example.com/":          root,
			"https://example.com/page1":     child1,
			"https://example.com/page2":     child2,
			"https://example.com/page1/sub": grandchild,
		},
	}

	// Test traversal order
	var visitedTitles []string
	err := tree.TraverseDepthFirst(func(node *DocumentNode) error {
		visitedTitles = append(visitedTitles, node.Title)
		return nil
	})

	if err != nil {
		t.Fatalf("TraverseDepthFirst() failed: %v", err)
	}

	expectedOrder := []string{"Root", "Page 1", "Subpage", "Page 2"}
	if len(visitedTitles) != len(expectedOrder) {
		t.Errorf("Expected %d nodes, visited %d", len(expectedOrder), len(visitedTitles))
	}

	for i, expected := range expectedOrder {
		if i >= len(visitedTitles) || visitedTitles[i] != expected {
			t.Errorf("Expected node %d to be %s, got %s", i, expected, visitedTitles[i])
		}
	}
}

func TestDocumentTree_TraverseBreadthFirst(t *testing.T) {
	// Create a simple tree structure
	root := &DocumentNode{
		URL:      "https://example.com/",
		Title:    "Root",
		Children: make([]*DocumentNode, 0),
	}

	child1 := &DocumentNode{
		URL:    "https://example.com/page1",
		Title:  "Page 1",
		Parent: root,
	}

	child2 := &DocumentNode{
		URL:    "https://example.com/page2",
		Title:  "Page 2",
		Parent: root,
	}

	grandchild1 := &DocumentNode{
		URL:    "https://example.com/page1/sub1",
		Title:  "Subpage 1",
		Parent: child1,
	}

	grandchild2 := &DocumentNode{
		URL:    "https://example.com/page1/sub2",
		Title:  "Subpage 2",
		Parent: child1,
	}

	child1.Children = []*DocumentNode{grandchild1, grandchild2}
	root.Children = []*DocumentNode{child1, child2}

	tree := &DocumentTree{
		Root: root,
		NodeMap: map[string]*DocumentNode{
			"https://example.com/":           root,
			"https://example.com/page1":      child1,
			"https://example.com/page2":      child2,
			"https://example.com/page1/sub1": grandchild1,
			"https://example.com/page1/sub2": grandchild2,
		},
	}

	// Test traversal order
	var visitedTitles []string
	err := tree.TraverseBreadthFirst(func(node *DocumentNode) error {
		visitedTitles = append(visitedTitles, node.Title)
		return nil
	})

	if err != nil {
		t.Fatalf("TraverseBreadthFirst() failed: %v", err)
	}

	expectedOrder := []string{"Root", "Page 1", "Page 2", "Subpage 1", "Subpage 2"}
	if len(visitedTitles) != len(expectedOrder) {
		t.Errorf("Expected %d nodes, visited %d", len(expectedOrder), len(visitedTitles))
	}

	for i, expected := range expectedOrder {
		if i >= len(visitedTitles) || visitedTitles[i] != expected {
			t.Errorf("Expected node %d to be %s, got %s", i, expected, visitedTitles[i])
		}
	}
}

func TestDocumentTree_FindNode(t *testing.T) {
	tree := &DocumentTree{
		NodeMap: map[string]*DocumentNode{
			"https://example.com/": {
				URL:   "https://example.com/",
				Title: "Home",
			},
			"https://example.com/docs": {
				URL:   "https://example.com/docs",
				Title: "Documentation",
			},
		},
	}

	// Test finding existing node
	node := tree.FindNode("https://example.com/")
	if node == nil {
		t.Error("FindNode() should return existing node")
	}
	if node.Title != "Home" {
		t.Errorf("Expected title 'Home', got %s", node.Title)
	}

	// Test finding non-existing node
	node = tree.FindNode("https://example.com/nonexistent")
	if node != nil {
		t.Error("FindNode() should return nil for non-existing node")
	}
}

func TestDocumentTree_GetNodesByLevel(t *testing.T) {
	// Create nodes with different levels
	root := &DocumentNode{
		URL:   "https://example.com/",
		Title: "Root",
		Level: 0,
	}

	level1Node1 := &DocumentNode{
		URL:   "https://example.com/page1",
		Title: "Page 1",
		Level: 1,
	}

	level1Node2 := &DocumentNode{
		URL:   "https://example.com/page2",
		Title: "Page 2",
		Level: 1,
	}

	level2Node := &DocumentNode{
		URL:   "https://example.com/page1/sub",
		Title: "Subpage",
		Level: 2,
	}

	// Set up parent-child relationships for traversal
	level1Node1.Children = []*DocumentNode{level2Node}
	root.Children = []*DocumentNode{level1Node1, level1Node2}

	tree := &DocumentTree{
		Root: root,
		NodeMap: map[string]*DocumentNode{
			"https://example.com/":          root,
			"https://example.com/page1":     level1Node1,
			"https://example.com/page2":     level1Node2,
			"https://example.com/page1/sub": level2Node,
		},
	}

	// Test getting nodes by level
	level0Nodes := tree.GetNodesByLevel(0)
	if len(level0Nodes) != 1 {
		t.Errorf("Expected 1 node at level 0, got %d", len(level0Nodes))
	}
	if level0Nodes[0].Title != "Root" {
		t.Errorf("Expected root node at level 0, got %s", level0Nodes[0].Title)
	}

	level1Nodes := tree.GetNodesByLevel(1)
	if len(level1Nodes) != 2 {
		t.Errorf("Expected 2 nodes at level 1, got %d", len(level1Nodes))
	}

	level2Nodes := tree.GetNodesByLevel(2)
	if len(level2Nodes) != 1 {
		t.Errorf("Expected 1 node at level 2, got %d", len(level2Nodes))
	}

	// Test non-existing level
	level5Nodes := tree.GetNodesByLevel(5)
	if len(level5Nodes) != 0 {
		t.Errorf("Expected 0 nodes at level 5, got %d", len(level5Nodes))
	}
}

func TestTreeBuilder_SortChildren(t *testing.T) {
	config := TreeConfig{
		SortChildren: true,
		SortBy:       SortByTitle,
	}

	builder := NewTreeBuilder(config)

	// Create parent node with unsorted children
	parent := &DocumentNode{
		URL:   "https://example.com/",
		Title: "Parent",
		Children: []*DocumentNode{
			{URL: "https://example.com/zebra", Title: "Zebra", Index: 3},
			{URL: "https://example.com/apple", Title: "Apple", Index: 1},
			{URL: "https://example.com/bear", Title: "Bear", Index: 2},
		},
	}

	// Test sorting by title
	builder.SortChildren(parent, SortByTitle)

	expectedTitles := []string{"Apple", "Bear", "Zebra"}
	for i, expected := range expectedTitles {
		if parent.Children[i].Title != expected {
			t.Errorf("Expected child %d to be %s, got %s", i, expected, parent.Children[i].Title)
		}
	}

	// Test sorting by index
	builder.SortChildren(parent, SortByIndex)

	expectedIndices := []int{1, 2, 3}
	for i, expected := range expectedIndices {
		if parent.Children[i].Index != expected {
			t.Errorf("Expected child %d to have index %d, got %d", i, expected, parent.Children[i].Index)
		}
	}

	// Test sorting by URL
	builder.SortChildren(parent, SortByURL)

	expectedURLs := []string{
		"https://example.com/apple",
		"https://example.com/bear",
		"https://example.com/zebra",
	}
	for i, expected := range expectedURLs {
		if parent.Children[i].URL != expected {
			t.Errorf("Expected child %d to have URL %s, got %s", i, expected, parent.Children[i].URL)
		}
	}
}
