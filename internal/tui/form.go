package tui

import (
	"fmt"

	"github.com/sjokagyi/fuli/internal/config"

	"github.com/charmbracelet/huh"
)

// CreateConfigForm builds a Huh form dynamically.
// It checks the provided config struct; if fields are already populated (via flags),
// it skips asking for them.
func CreateConfigForm(cfg *config.RunConfig) *huh.Form {
	var groups []*huh.Group

	// Source Directory Input
	if cfg.SourceDir == "" {
		input := huh.NewInput().
			Title("Source Directory").
			Description("Path to the folder you want to bundle.").
			Prompt("? ").
			Placeholder("./").
			Validate(func(s string) error {
				if len(s) == 0 {
					return fmt.Errorf("source directory cannot be empty")
				}
				return nil
			}).
			Value(&cfg.SourceDir)

		groups = append(groups, huh.NewGroup(input))
	}

	// Output File Input
	if cfg.OutputPath == "" {
		input := huh.NewInput().
			Title("Output Filename").
			Description("Where should the text content be saved?").
			Prompt("? ").
			Placeholder("fuli_output.txt").
			Validate(func(s string) error {
				if len(s) == 0 {
					return fmt.Errorf("output filename cannot be empty")
				}
				return nil
			}).
			Value(&cfg.OutputPath)

		groups = append(groups, huh.NewGroup(input))
	}

	// Confirm (Optional but good for UX)
	// Only show confirmation if we actually asked questions
	if len(groups) > 0 {
		confirm := huh.NewConfirm().
			Title("Ready to copy?").
			Affirmative("Yes, let's go!").
			Negative("Wait, cancel").
			Value(&cfg.DryRun) // Reuse DryRun temporarily or add a confirmed bool if needed,
			// but usually, if they say no, we just quit.

		groups = append(groups, huh.NewGroup(confirm))
	}

	// If all config was provided via flags, return nil to signal immediate execution
	if len(groups) == 0 {
		return nil
	}

	return huh.NewForm(groups...).
		WithTheme(huh.ThemeCharm()) // Use the standard Charm theme for consistency
}
