package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileLines(t *testing.T) {
	// Create a temporary file with test content
	content := `# This is a comment
line1
line2

# Another comment
line3
    line4    
`
	tmpfile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Test loading the file
	lines, err := LoadFileLines(tmpfile.Name())
	if err != nil {
		t.Errorf("LoadFileLines() error = %v", err)
	}

	expected := []string{"line1", "line2", "line3", "line4"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Line %d: expected %s, got %s", i, expected[i], line)
		}
	}

	// Test non-existent file
	_, err = LoadFileLines("non-existent-file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestWriteLinesToFile(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "test-dir-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	filename := filepath.Join(tmpdir, "test-output.txt")
	lines := []string{"line1", "line2", "line3"}

	// Test writing lines
	err = WriteLinesToFile(filename, lines)
	if err != nil {
		t.Errorf("WriteLinesToFile() error = %v", err)
	}

	// Verify the file was created and has correct content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}

	expected := "line1\nline2\nline3\n"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}

	// Test writing to invalid path
	err = WriteLinesToFile("/invalid/path/file.txt", lines)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Test existing file
	if !FileExists(tmpfile.Name()) {
		t.Error("FileExists() should return true for existing file")
	}

	// Test non-existent file
	if FileExists("non-existent-file.txt") {
		t.Error("FileExists() should return false for non-existent file")
	}

	// Test directory
	tmpdir, err := os.MkdirTemp("", "test-dir-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	if !FileExists(tmpdir) {
		t.Error("FileExists() should return true for existing directory")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "test-parent-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Test creating nested directories
	testDir := filepath.Join(tmpdir, "level1", "level2", "level3")
	err = EnsureDir(testDir)
	if err != nil {
		t.Errorf("EnsureDir() error = %v", err)
	}

	// Verify directory was created
	if !FileExists(testDir) {
		t.Error("EnsureDir() should create the directory")
	}

	// Test with existing directory (should not error)
	err = EnsureDir(testDir)
	if err != nil {
		t.Errorf("EnsureDir() should not error for existing directory: %v", err)
	}

	// Test with empty string (should create current directory, which should succeed)
	err = EnsureDir("")
	if err != nil {
		t.Errorf("EnsureDir() should handle empty string: %v", err)
	}
}
