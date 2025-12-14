package cmd

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sjokagyi/fuli/internal/config"
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
		Args: cobra.MaximumNArgs(1), // Optional source_directory argument
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
	rootCmd.Flags().StringArrayVarP(&ignoreFiles, "ignore", "i", []string{}, "Path to additional ignore files (e.g. .gitignore)")
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
	// Critical for debugging without messing up the TUI layout
	if err := logger.Init(cfg.IsVerbose); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize TUI Model
	// The model automatically determines if it needs to show the wizard (Form)
	// or jump straight to processing based on whether cfg is fully populated.
	model := tui.NewModel(cfg)

	// Initialize Program
	// We pass &model (pointer) so that we can bind the program reference to it
	// before the execution loop starts.
	p := tea.NewProgram(&model, tea.WithAltScreen())

	// Bind Program to Model
	// This allows the Model's internal Walker to send messages back to the Program loop.
	model.BindProgram(p)

	// Run
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}
