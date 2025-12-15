package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sjokagyi/fuli/internal/config"
	"github.com/sjokagyi/fuli/internal/core"
	"github.com/sjokagyi/fuli/internal/ignore"
	"github.com/sjokagyi/fuli/internal/logger"
	"github.com/sjokagyi/fuli/internal/tui"
)

var (
	// CLI Flags
	outputPath  string
	ignoreFiles []string
	verbose     bool
	dryRun      bool

	// Root Command Definition
	rootCmd = &cobra.Command{
		Use:   "fuli [source_directory]",
		Short: "Fuli: The File Aggregator TUI",
		Long: `Fuli recursively traverses a directory, respects ignore files (like .contextignore),
and aggregates all file contents into a single formatted text file.

It features a TUI wizard for interactive use, but can also be fully controlled 
via flags for scripts and CI/CD pipelines.`,
		Args: cobra.MaximumNArgs(1),
		RunE: run,
	}
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Bind Flags
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to the destination text file")
	rootCmd.Flags().StringArrayVarP(&ignoreFiles, "ignore", "i", []string{}, "Path to additional ignore files")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate traversal without writing to disk")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging to fuli.debug.log")
}

func run(cmd *cobra.Command, args []string) error {
	// Populate Configuration
	cfg := &config.RunConfig{
		OutputPath:  outputPath,
		IgnoreFiles: ignoreFiles,
		IsVerbose:   verbose,
		DryRun:      dryRun,
	}

	// Handle optional positional argument
	if len(args) > 0 {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("invalid source directory: %w", err)
		}
		cfg.SourceDir = absPath
	}

	// Initialize Logger
	if err := logger.Init(cfg.IsVerbose); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// DECISION: Headless vs Interactive
	// If SourceDir AND OutputPath are set via flags/args, run in Headless mode (better for scripts).
	if cfg.SourceDir != "" && cfg.OutputPath != "" {
		return runHeadless(cfg)
	}

	// Interactive TUI Mode
	model := tui.NewModel(cfg)
	p := tea.NewProgram(&model, tea.WithAltScreen())
	model.BindProgram(p)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}
	return nil
}

// runHeadless executes the logic without the TUI loop, ensuring proper exit codes and stdout logging.
func runHeadless(cfg *config.RunConfig) error {
	// Validate (Fail Fast)
	if info, err := os.Stat(cfg.SourceDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source directory does not exist: %s", cfg.SourceDir)
		}
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", cfg.SourceDir)
	}

	// Normalize Output Path
	absOutput, err := filepath.Abs(cfg.OutputPath)
	if err != nil {
		return err
	}
	cfg.OutputPath = absOutput

	// Setup Matcher
	matcher := ignore.NewMatcher(cfg.SourceDir)
	defaultIgnorePath := filepath.Join(cfg.SourceDir, ".contextignore")
	if patterns, err := ignore.ParseFile(defaultIgnorePath); err == nil {
		matcher.AddPatterns(patterns)
		manualRules := ignore.CompileIgnoreLines([]string{".contextignore"})
		matcher.AddPatterns(manualRules)
	}
	for _, file := range cfg.IgnoreFiles {
		if patterns, err := ignore.ParseFile(file); err == nil {
			matcher.AddPatterns(patterns)
		}
	}

	// Setup Walker & Sync Channel
	walker := core.NewWalker(*cfg, matcher)
	done := make(chan error) // Used to block until Walker finishes

	// Start Walker with Callback
	fmt.Printf("Indexing %s...\n", cfg.SourceDir)
	walker.Start(func(msg interface{}) {
		switch m := msg.(type) {
		case core.FileProcessedMsg:
			fmt.Printf("✓ COPIED %s (%d bytes)\n", m.Path, m.Bytes)
		case core.ErrorMsg:
			done <- m.Err
		case core.CompletionMsg:
			fmt.Println("\n--- Operation Complete ---")
			fmt.Printf("Files: %d | Size: %s | Time: %s\n",
				m.TotalFiles,
				byteCountDecimal(m.TotalBytes),
				m.Elapsed.Round(time.Millisecond))

			if m.TotalFiles == 0 {
				fmt.Println("WARNING: No files were processed.")
			} else {
				fmt.Printf("Output saved to: %s\n", cfg.OutputPath)
			}
			done <- nil // Success
		}
	})

	// Block until done
	return <-done
}

// Helper for headless printing
func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
