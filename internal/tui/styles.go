package tui

import "github.com/charmbracelet/lipgloss"

// Color Palette
var (
	colorSubtle    = lipgloss.Color("241")
	colorHighlight = lipgloss.Color("205")     // Pink
	colorSuccess   = lipgloss.Color("42")      // Green
	colorError     = lipgloss.Color("196")     // Red
	colorText      = lipgloss.Color("252")     // White-ish
	colorHeader    = lipgloss.Color("#FFF7DB") // Warm white
)

// Common Styles
var (
	// Base application style
	AppStyle = lipgloss.NewStyle().
			Margin(1, 2)

	// TitleStyle for the main app header
	TitleStyle = lipgloss.NewStyle().
			Foreground(colorHeader).
			Background(colorHighlight).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	// SubtleStyle for hints and help text
	SubtleStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// LogLineStyle for the scrolling file list
	LogLineStyle = lipgloss.NewStyle().
			Foreground(colorText).
			PaddingLeft(1)

	// SuccessBoxStyle for the final summary report
	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSuccess).
			Padding(1, 2).
			MarginTop(1)

	// ErrorBoxStyle for displaying fatal errors
	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Foreground(colorError).
			Padding(1, 2).
			MarginTop(1)
)

// Helper to style specific words in the log
func StyleKeyword(s string) string {
	return lipgloss.NewStyle().Foreground(colorHighlight).Render(s)
}

func StylePath(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Render(s) // Blue
}
