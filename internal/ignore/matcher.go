package ignore

import (
	"path/filepath"
	"regexp"
)

// Matcher evaluates paths against a set of ignore rules.
type Matcher struct {
	root  string
	rules []Rule
}

// Rule represents a single compiled ignore pattern.
type Rule struct {
	pattern *regexp.Regexp // Compiled regex
	negate  bool           // True if pattern started with '!'
	dirOnly bool           // True if pattern ended with '/'
}

// NewMatcher creates a Matcher anchored to a specific root directory.
func NewMatcher(root string) *Matcher {
	absRoot, _ := filepath.Abs(root)
	return &Matcher{
		root:  absRoot,
		rules: make([]Rule, 0),
	}
}

// AddPatterns processes compiled Rules and appends them.
func (m *Matcher) AddPatterns(patterns []Rule) {
	m.rules = append(m.rules, patterns...)
}

// Matches checks if the given absolute path should be ignored.
func (m *Matcher) Matches(absPath string, isDir bool) bool {
	// 1. Calculate relative path from root
	relPath, err := filepath.Rel(m.root, absPath)
	if err != nil {
		return false
	}

	// 2. Guard Clause: Never ignore the root directory itself.
	if relPath == "." {
		return false
	}

	// 3. Normalize to forward slashes for Regex matching (Windows compatibility)
	// Git patterns operate on forward slashes.
	relPath = filepath.ToSlash(relPath)

	ignored := false

	// 4. Iterate through rules in order (Last Match Wins)
	for _, r := range m.rules {
		if r.pattern == nil {
			continue
		}

		if r.dirOnly && !isDir {
			continue
		}

		if r.pattern.MatchString(relPath) {
			if r.negate {
				ignored = false
			} else {
				ignored = true
			}
		}
	}

	return ignored
}
