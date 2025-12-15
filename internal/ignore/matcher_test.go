package ignore_test

import (
	"path/filepath"
	"testing"

	"github.com/sjokagyi/fuli/internal/ignore"
)

func TestMatcher(t *testing.T) {
	root := "/project"
	matcher := ignore.NewMatcher(root)

	patterns := []string{
		// 1. Basic Wildcards
		"*.log", // Recursive match of .log files
		"temp/", // Recursive match of temp directories

		// 2. Anchoring
		"/build",   // Root-only match
		"doc/*.md", // Matches doc/foo.md anywhere? No, "doc" anchor implies relative path matching.
		// Git behavior: "doc/*.md" matches "doc/foo.md" at root, OR "src/doc/foo.md"?
		// Correction: "doc/*.md" contains a slash, so it is treated as relative to root (anchored).
		// To match "doc" anywhere, one would need "**/doc/*.md" in standard git behavior if it wasn't a simple name.

		// 3. Double Star (The new features)
		"**/nested/file", // Matches nested/file at any depth
		"src/**/test",    // Matches test inside src at any depth
		"assets/**",      // Matches everything inside assets

		// 4. Negation
		"!important.log", // Override *.log for this specific name
	}

	// Compile strings to Rules
	rules := ignore.CompileIgnoreLines(patterns)
	matcher.AddPatterns(rules)

	tests := []struct {
		path   string
		isDir  bool
		expect bool
		desc   string
	}{
		// --- Basic Wildcards ---
		{"/project/app.log", false, true, "Standard wildcard *.log"},
		{"/project/src/error.log", false, true, "Recursive wildcard *.log"},
		{"/project/important.log", false, false, "Negated pattern !important.log"},

		// --- Directory Matching ---
		{"/project/temp", true, true, "Directory match temp/"},
		{"/project/src/temp", true, true, "Recursive directory match temp/"},
		{"/project/temp/file.txt", false, false, "File inside ignored dir (Walker handles skipping, matcher shouldn't ignore file explicitly unless rule says so)"},
		// Note: The Matcher returns false here because the rule "temp/" matches the directory entity.
		// It does NOT match "temp/file.txt" string. The walker skips the parent, so this file is never visited.
		// This behavior remains consistent.

		// --- Anchoring ---
		{"/project/build", true, true, "Root anchor /build"},
		{"/project/src/build", true, false, "Root anchor /build should NOT match nested build"},
		{"/project/doc/readme.md", false, true, "Anchored pattern doc/*.md"},
		{"/project/src/doc/readme.md", false, false, "Anchored pattern doc/*.md should NOT match nested doc"},

		// --- Double Star (New Functionality) ---
		{"/project/a/b/nested/file", false, true, "Leading double star **/nested/file"},
		{"/project/nested/file", false, true, "Leading double star matches at root too"},

		{"/project/src/a/b/test", true, true, "Middle double star src/**/test"},
		{"/project/src/test", true, true, "Middle double star zero-width match"},

		{"/project/assets/image.png", false, true, "Trailing double star assets/**"},
		{"/project/assets/css/style.css", false, true, "Trailing double star recursive"},
	}

	for _, tt := range tests {
		// Simulate cross-platform path separators for input
		path := filepath.FromSlash(tt.path)
		result := matcher.Matches(path, tt.isDir)

		if result != tt.expect {
			t.Errorf("FAIL: %s\nPath: %s (Dir: %v)\nExpected: %v, Got: %v",
				tt.desc, path, tt.isDir, tt.expect, result)
		}
	}
}

// TestRegexTranslation validates the internal translation logic directly
// This ensures our regex generation assumes correct Git behavior.
func TestRegexTranslation(t *testing.T) {
	// Accessing internal functionality usually requires being in the same package
	// or exporting it. Since this is white-box testing for design verification:

	// Note: You might need to temporarily export translateGitIgnoreToRegex or copy logic here.
	// For this design, we assume we are testing the behavior via public API (CompileIgnoreLines -> internal logic).

	patterns := []string{
		"foo/**/bar",
	}

	// Create a matcher with specific patterns to debug specific regex edge cases
	matcher := ignore.NewMatcher("/root")
	matcher.AddPatterns(ignore.CompileIgnoreLines(patterns))

	if !matcher.Matches("/root/foo/a/b/c/bar", false) {
		t.Error("Failed to match middle double star")
	}
	if !matcher.Matches("/root/foo/bar", false) {
		t.Error("Failed to match middle double star with zero depth")
	}
}

func TestMatcher_Negation(t *testing.T) {
	root := "/project"
	matcher := ignore.NewMatcher(root)

	patterns := []string{
		// 1. General ignore
		"*.secret",

		// 2. Specific Exception (Re-include)
		"!public.secret",

		// 3. Directory Ignore
		"vendor/",

		// 4. Attempt to re-include file in ignored dir
		// (Note: In Git/Fuli, this technically "matches" as False in the matcher,
		// but the Walker will never reach this file because 'vendor/' returned True,
		// causing filepath.SkipDir. This test verifies the MATCH logic, not the Walk logic.)
		"!vendor/README.md",

		// 5. Bad Ordering Regression Test
		// If we negate first, then ignore later, the file SHOULD be ignored.
		"!bad_order.log",
		"*.log",
	}

	// Compile and add
	matcher.AddPatterns(ignore.CompileIgnoreLines(patterns))

	tests := []struct {
		path   string
		isDir  bool
		expect bool // true = ignored, false = included
		desc   string
	}{
		// Standard Negation
		{"/project/api.secret", false, true, "Standard *.secret should be ignored"},
		{"/project/public.secret", false, false, "Negated !public.secret should be INCLUDED"},

		// Directory Logic
		{"/project/vendor", true, true, "Vendor directory is ignored"},

		// Order of Operations
		{"/project/bad_order.log", false, true, "File should be ignored because *.log came AFTER !bad_order.log"},
		{"/project/ok.log", false, true, "Standard *.log matches"},
	}

	for _, tt := range tests {
		path := filepath.FromSlash(tt.path)
		result := matcher.Matches(path, tt.isDir)

		if result != tt.expect {
			t.Errorf("FAIL: %s\nPath: %s\nExpected Ignored: %v, Got: %v",
				tt.desc, path, tt.expect, result)
		}
	}
}
