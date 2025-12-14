package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/sjokagyi/fuli/internal/config"
	"github.com/sjokagyi/fuli/internal/core"
)

// GenerateSummary creates a rendered Markdown string containing execution statistics.
func GenerateSummary(stats core.CompletionMsg, cfg *config.RunConfig, width int) string {
	// Determine Title and Footer based on success
	title := "# Operation Complete"
	footer := "> *Content successfully aggregated.*"
	boxStyle := SuccessBoxStyle

	if stats.TotalFiles == 0 {
		title = "# Operation Finished (Empty)"
		footer = "> **Warning**: No files were copied.\n> Check your source path, .contextignore rules, or run with `-v` to debug."
		boxStyle = WarningBoxStyle
	}

	// Define the Markdown Template
	markdown := fmt.Sprintf(`
%s

| Metric | Value |
| :--- | :--- |
| **Source** | %s |
| **Destination** | %s |
| **Files Processed** | %d |
| **Total Size** | %s |
| **Duration** | %s |

%s
`,
		title,
		cfg.SourceDir,
		cfg.OutputPath,
		stats.TotalFiles,
		byteCountDecimal(stats.TotalBytes),
		stats.Elapsed.Round(time.Millisecond),
		footer,
	)

	// Configure Glamour Renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-10),
	)
	if err != nil {
		return "Error generating summary."
	}

	// Render
	out, err := renderer.Render(markdown)
	if err != nil {
		return "Error rendering markdown."
	}

	// Return styled box
	return boxStyle.Render(strings.TrimSpace(out))
}

// byteCountDecimal formats bytes into human-readable strings (kB, MB, GB).
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
