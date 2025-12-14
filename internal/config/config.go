package config

// RunConfig defines the runtime configuration for Fuli.
// It aggregates inputs from CLI flags and TUI forms.
type RunConfig struct {
	// SourceDir is the absolute path to the directory being processed.
	SourceDir string

	// OutputPath is the absolute path to the destination text file.
	OutputPath string

	// IgnoreFiles contains a list of specific ignore file paths (e.g., .gitignore, .contextignore)
	// that should be parsed in addition to default behavior.
	IgnoreFiles []string

	// IsVerbose toggles detailed logging for debugging file access and ignore logic.
	IsVerbose bool

	// DryRun simulates the traversal without performing actual disk writes.
	DryRun bool
}
