package ignore

import (
	"path/filepath"
	"strings"
)

// Matcher evaluates paths against a set of ignore rules.
// It manages the complexity of relative paths and rule precedence.
type Matcher struct {
	root  string
	rules []rule
}

// rule represents a single parsed line from an ignore file.
type rule struct {
	pattern  string
	negate   bool // True if the pattern starts with '!'
	dirOnly  bool // True if the pattern ends with '/'
	rootOnly bool // True if the pattern contains a separator (e.g., "subdir/file")
}

// NewMatcher creates a Matcher anchored to a specific root directory.
// All file checks will be calculated relative to this root.
func NewMatcher(root string) *Matcher {
	absRoot, _ := filepath.Abs(root)
	return &Matcher{
		root:  absRoot,
		rules: make([]rule, 0),
	}
}

// AddPatterns processes raw strings into rules and appends them to the matcher.
// Patterns added later have higher precedence (last match wins).
func (m *Matcher) AddPatterns(patterns []string) {
	for _, p := range patterns {
		m.rules = append(m.rules, parseRule(p))
	}
}

// Matches checks if the given absolute path should be ignored.
func (m *Matcher) Matches(absPath string, isDir bool) bool {
	// Calculate the path relative to the project root.
	// This is crucial because ignore patterns usually apply to the relative path.
	relPath, err := filepath.Rel(m.root, absPath)
	if err != nil {
		return false
	}

	// Never ignore the root directory itself.
	// If we matched the root, filepath.Rel returns "."
	if relPath == "." {
		return false
	}

	// Normalize to forward slashes for consistent matching, as gitignore uses '/'
	relPath = filepath.ToSlash(relPath)
	fileName := filepath.Base(absPath)

	ignored := false

	for _, r := range m.rules {
		// Optimization: If the rule only applies to directories, skip if this is a file.
		if r.dirOnly && !isDir {
			continue
		}

		var matched bool

		if r.rootOnly {
			// If rule contains a slash (e.g. "logs/debug.log"), match against the full relative path.
			matched, _ = filepath.Match(r.pattern, relPath)
		} else {
			// If rule has no slash (e.g. "*.o"), match against the base filename.
			// This allows "*.o" to match "main.o" and "src/utils.o".
			matched, _ = filepath.Match(r.pattern, fileName)
		}

		if matched {
			if r.negate {
				ignored = false // Re-include the file
			} else {
				ignored = true
			}
		}
	}

	return ignored
}

// parseRule converts a raw string into an efficient rule struct.
func parseRule(raw string) rule {
	r := rule{}

	// Handle Negation
	if strings.HasPrefix(raw, "!") {
		r.negate = true
		raw = raw[1:]
	}

	// Handle Directory Specificity
	if strings.HasSuffix(raw, "/") {
		r.dirOnly = true
		raw = raw[:len(raw)-1]
	}

	// Handle Root Anchoring
	// If there is a separator in the pattern (start or middle), it matches relative to root.
	// Example: "foo" matches everywhere. "/foo" or "bar/foo" matches specifically.
	if strings.Contains(raw, "/") {
		r.rootOnly = true
		// Strip leading slash for filepath.Match compatibility, as Rel() returns paths without leading slashes.
		raw = strings.TrimPrefix(raw, "/")
	}

	r.pattern = raw
	return r
}
