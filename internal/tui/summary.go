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
	// Define the Markdown Template
	markdown := fmt.Sprintf(`
# Operation Complete

| Metric | Value |
| :--- | :--- |
| **Source** | %s |
| **Destination** | %s |
| **Files Processed** | %d |
| **Total Size** | %s |
| **Duration** | %s |

> *Content successfully aggregated.*
`,
		cfg.SourceDir,
		cfg.OutputPath,
		stats.TotalFiles,
		byteCountDecimal(stats.TotalBytes),
		stats.Elapsed.Round(time.Millisecond),
	)

	// Configure Glamour Renderer
	// We dynamically adjust the wrap width to the terminal size to prevent ugly wrapping.
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),        // Detects dark/light terminal
		glamour.WithWordWrap(width-10), // Padding for safety
	)
	if err != nil {
		return "Error generating summary."
	}

	// Render
	out, err := renderer.Render(markdown)
	if err != nil {
		return "Error rendering markdown."
	}

	// Wrap in a box using Lip Gloss (defined in styles.go)
	return SuccessBoxStyle.Render(strings.TrimSpace(out))
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
