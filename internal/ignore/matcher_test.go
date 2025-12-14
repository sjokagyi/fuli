package ignore_test

import (
	"path/filepath"
	"testing"

	"github.com/sjokagyi/fuli/internal/ignore"
)

func TestMatcher(t *testing.T) {
	root := "/project"
	matcher := ignore.NewMatcher(root)

	// Simulate a .contextignore file
	patterns := []string{
		"*.log",      // Ignore all logs
		"!keep.log",  // Exception: Keep this specific log
		"/build",     // Ignore root build folder only
		"temp/",      // Ignore temp directories anywhere
		"secret.txt", // Ignore specific file anywhere
	}
	matcher.AddPatterns(patterns)

	tests := []struct {
		path   string
		isDir  bool
		expect bool // true = ignored, false = included
	}{
		{"/project/app.log", false, true},           // Matches *.log
		{"/project/keep.log", false, false},         // Matches !keep.log
		{"/project/src/app.log", false, true},       // Matches *.log (global)
		{"/project/build", true, true},              // Matches /build
		{"/project/src/build", true, false},         // Should NOT match /build (anchored)
		{"/project/src/temp", true, true},           // Matches temp/
		{"/project/src/temp/file.go", false, false}, // Inside ignored dir, but file itself not matched
		{"/project/secret.txt", false, true},        // Matches secret.txt
		{"/project/src/secret.txt", false, true},    // Matches secret.txt
		{"/project/main.go", false, false},          // No match
	}

	for _, tt := range tests {
		// Fix path separators for Windows compatibility in tests
		path := filepath.FromSlash(tt.path)
		result := matcher.Matches(path, tt.isDir)
		if result != tt.expect {
			t.Errorf("Path %s (Dir: %v): expected %v, got %v", path, tt.isDir, tt.expect, result)
		}
	}
}
