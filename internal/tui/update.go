package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/sjokagyi/fuli/internal/core"
)

// Update is the main event loop for the Bubble Tea model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// --- System Events ---
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Resize viewport to fit available space
		headerHeight := 4
		footerHeight := 2
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - headerHeight - footerHeight

	case tea.KeyMsg:
		// Global Quit Keys
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Allow 'q' to quit only when not typing in a form
		if m.state != StateConfig && msg.String() == "q" {
			return m, tea.Quit
		}

	// --- Core Engine Events ---
	case core.FileProcessedMsg:
		// Add formatted line to viewport
		line := fmt.Sprintf("%s %s", StyleKeyword("COPIED"), StylePath(msg.Path))
		m.processedFiles = append(m.processedFiles, line)

		// Update viewport content
		m.viewport.SetContent(strings.Join(m.processedFiles, "\n"))
		m.viewport.GotoBottom() // Auto-scroll
		return m, nil

	case core.ErrorMsg:
		m.lastError = msg.Err
		m.state = StateError
		return m, nil

	case core.CompletionMsg:
		m.state = StateDone
		// Generate the fancy summary using the helper
		m.summary = GenerateSummary(msg, m.cfg, m.width)
		return m, nil // Stay in StateDone to show summary
	}

	// --- State Machine ---
	switch m.state {

	case StateConfig:
		// Pass messages to Huh form
		if m.form != nil {
			var formModel tea.Model
			formModel, cmd = m.form.Update(msg)
			if f, ok := formModel.(*huh.Form); ok {
				m.form = f
			}
			cmds = append(cmds, cmd)

			// Check if form is finished
			if m.form.State == huh.StateCompleted {
				// Start the real work
				m.state = StateRunning
				cmds = append(cmds, m.spinner.Tick, m.startWalkerCmd)
			}
		}

	case StateRunning:
		// Animate spinner
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Handle viewport scrolling
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case StateDone:
		// Allow scrolling through summary or logs if needed
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
