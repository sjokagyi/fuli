package ignore

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ParseFile reads an ignore file and returns a slice of compiled rules.
func ParseFile(path string) ([]Rule, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return CompileIgnoreLines(lines), nil
}

// CompileIgnoreLines converts a slice of raw strings into compiled Rules.
// This helper allows other packages (like tui or tests) to generate rules from strings.
func CompileIgnoreLines(lines []string) []Rule {
	var rules []Rule
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle escaped comments "\#" -> "#"
		if strings.HasPrefix(line, "\\#") {
			line = strings.TrimPrefix(line, "\\")
		}

		// Handle trailing spaces (unless escaped with \)
		// Git behavior: "foo " -> "foo", "foo\ " -> "foo "
		if !strings.HasSuffix(line, "\\ ") {
			line = strings.TrimRight(line, " ")
		} else {
			line = strings.TrimSuffix(line, "\\ ") + " "
		}

		rules = append(rules, compileRule(line))
	}
	return rules
}

// compileRule parses a raw line into a Rule struct with a compiled Regex.
func compileRule(raw string) Rule {
	r := Rule{}

	// 1. Handle Negation ('!')
	if strings.HasPrefix(raw, "!") {
		r.negate = true
		raw = raw[1:]
	}

	// 2. Handle Directory Specificity ('/')
	if strings.HasSuffix(raw, "/") {
		r.dirOnly = true
		raw = raw[:len(raw)-1]
	}

	// 3. Translate Glob to Regex
	regexStr := translateGitIgnoreToRegex(raw)

	// 4. Compile Regex
	re, err := regexp.Compile(regexStr)
	if err == nil {
		r.pattern = re
	}

	return r
}

// translateGitIgnoreToRegex converts a gitignore glob pattern to a Go regex string.
func translateGitIgnoreToRegex(pattern string) string {
	// Root anchoring check: patterns with '/' (not at end) are anchored to root.
	// e.g. "foo/bar" is anchored. "foo" is not.
	isRootAnchored := strings.Contains(strings.TrimSuffix(pattern, "/"), "/")
	hasLeadingSlash := strings.HasPrefix(pattern, "/")

	if hasLeadingSlash {
		pattern = pattern[1:]
	}

	var sb strings.Builder
	sb.WriteString("^") // Start of line

	// If not anchored to root, allow prefix
	if !isRootAnchored && !hasLeadingSlash {
		sb.WriteString("((.*/)|)")
	}

	n := len(pattern)
	for i := 0; i < n; i++ {
		c := pattern[i]
		switch c {
		case '/':
			// Lookahead for "/**" (Trailing)
			if i+2 < n && pattern[i+1] == '*' && pattern[i+2] == '*' && i+3 == n {
				sb.WriteString("(/.*)?") // Matches "" (if dir matched elsewhere) or "/..."
				i += 2
				continue
			}
			// Lookahead for "/**/" (Middle)
			if i+3 < n && pattern[i+1] == '*' && pattern[i+2] == '*' && pattern[i+3] == '/' {
				sb.WriteString("(?:/|/.+/)") // Matches "/" (zero dirs) or "/.../" (nested)
				i += 3
				continue
			}
			// Normal Slash
			sb.WriteString("\\/")

		case '*':
			// Lookahead for "**/..." (Leading)
			if i == 0 && i+2 < n && pattern[i+1] == '*' && pattern[i+2] == '/' {
				sb.WriteString("(.*/)?") // Matches "" or "foo/..."
				i += 2
				continue
			}
			// Double star without slashes (e.g. "foo**bar") -> Treat as normal greedy wildcard
			if i+1 < n && pattern[i+1] == '*' {
				sb.WriteString(".*")
				i++
				continue
			}
			// Single star
			sb.WriteString("[^/]*")

		case '?':
			// '?' matches one character, but not a slash
			sb.WriteString("[^/]")

		case '.', '+', '(', ')', '|', '^', '$', '{', '}':
			// Escape regex specials
			sb.WriteString("\\" + string(c))

		case '[':
			// Simple character class pass-through.
			sb.WriteString("\\[")

		case ']':
			sb.WriteString("\\]")

		case '\\':
			// Handle escapes in the glob pattern
			if i+1 < n {
				sb.WriteString(regexp.QuoteMeta(string(pattern[i+1])))
				i++
			}

		default:
			// For all other characters, escape them if they are regex specials, otherwise write literally
			sb.WriteString(regexp.QuoteMeta(string(c)))
		}
	}

	sb.WriteString("$") // End of line
	return sb.String()
}
