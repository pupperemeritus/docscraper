package utils

import (
	"bufio"
	"os"
	"strings"
)

// LoadFileLines reads a file and returns its lines as a slice of strings
func LoadFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") { // Skip empty lines and comments
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

// WriteLinesToFile writes a slice of strings to a file, one per line
func WriteLinesToFile(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dir string) error {
	if dir == "" {
		return nil // Empty directory means current directory, which always exists
	}
	return os.MkdirAll(dir, 0755)
}
