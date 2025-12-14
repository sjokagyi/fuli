package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// NewDefaultSpinner creates a pre-configured spinner with the app's signature style.
func NewDefaultSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorHighlight)
	return s
}

// NewLogViewport creates a viewport optimized for scrolling text logs.
// Width and Height should be set by the main Update loop on WindowSizeMsg.
func NewLogViewport() viewport.Model {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true). // Left border only
		BorderForeground(colorSubtle).
		PaddingLeft(1)

	// Helper to scroll automatically
	return vp
}
