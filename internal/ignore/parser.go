package ignore

import (
	"bufio"
	"os"
	"strings"
)

// ParseFile reads an ignore file (like .contextignore) and returns a slice of active patterns.
// It implements standard gitignore parsing rules regarding comments and whitespace.
func ParseFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Skip comments (lines starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Handle escaped comments
		// If a line starts with "\#", it is treated as a pattern starting with "#".
		if strings.HasPrefix(line, "\\#") {
			line = strings.TrimPrefix(line, "\\")
		}

		// Trim trailing spaces
		// Spaces at the end of the line are ignored unless escaped (not handled here for simplicity).
		line = strings.TrimRight(line, " ")

		patterns = append(patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}
