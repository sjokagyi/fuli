package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/sjokagyi/fuli/internal/config"
	"github.com/sjokagyi/fuli/internal/core"
	"github.com/sjokagyi/fuli/internal/ignore"
)

// Application States
type AppState int

const (
	StateConfig AppState = iota
	StateRunning
	StateDone
	StateError
)

type Model struct {
	// Infrastructure
	state   AppState
	cfg     *config.RunConfig
	walker  *core.Walker
	program *tea.Program

	// Dimensions
	width  int
	height int

	// Components
	form     *huh.Form
	spinner  spinner.Model
	viewport viewport.Model

	// Data
	processedFiles []string
	summary        string
	lastError      error
}

// NewModel initializes the TUI state.
func NewModel(cfg *config.RunConfig) Model {
	m := Model{
		state:          StateConfig,
		cfg:            cfg,
		spinner:        NewDefaultSpinner(), // Uses factory from components.go
		processedFiles: make([]string, 0),
		viewport:       NewLogViewport(), // Uses factory from components.go
	}

	// Initialize Configuration Form using the factory in form.go
	m.form = CreateConfigForm(cfg)

	// If no form was generated (all flags provided), skip to running
	if m.form == nil {
		m.state = StateRunning
	}

	return m
}

// BindProgram connects the tea.Program to the model so the Walker can send messages.
func (m *Model) BindProgram(p *tea.Program) {
	m.program = p
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	if m.state == StateConfig && m.form != nil {
		cmds = append(cmds, m.form.Init())
	} else if m.state == StateRunning {
		// If we started in Running state (CLI mode), start immediately
		cmds = append(cmds, m.spinner.Tick, m.startWalkerCmd)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.state {
	case StateConfig:
		if m.form != nil {
			return lipgloss.JoinVertical(lipgloss.Left,
				TitleStyle.Render("Configuration"),
				m.form.View(),
			)
		}

	case StateRunning:
		header := fmt.Sprintf("%s Indexing %s...", m.spinner.View(), m.cfg.SourceDir)
		return lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("Processing"),
			header,
			m.viewport.View(),
			SubtleStyle.Render("\nPress Ctrl+C to cancel"),
		)

	case StateDone:
		return lipgloss.JoinVertical(lipgloss.Left,
			m.summary,
			SubtleStyle.Render("\nPress q to quit"),
		)

	case StateError:
		content := fmt.Sprintf("Error: %v", m.lastError)
		return lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("Error"),
			ErrorBoxStyle.Render(content),
			SubtleStyle.Render("\nPress q to quit"),
		)
	}

	return ""
}

// --- Helpers ---

// startWalkerCmd triggers the Core Walker.
func (m *Model) startWalkerCmd() tea.Msg {
	if m.program == nil {
		return core.ErrorMsg{Err: fmt.Errorf("internal error: program reference not bound")}
	}

	// Normalize paths to Absolute Paths
	absSource, err := filepath.Abs(m.cfg.SourceDir)
	if err != nil {
		return core.ErrorMsg{Err: fmt.Errorf("invalid source path: %w", err)}
	}
	m.cfg.SourceDir = absSource

	absOutput, err := filepath.Abs(m.cfg.OutputPath)
	if err != nil {
		return core.ErrorMsg{Err: fmt.Errorf("invalid output path: %w", err)}
	}
	m.cfg.OutputPath = absOutput

	// FAIL FAST: Check if Source Directory actually exists
	if info, err := os.Stat(m.cfg.SourceDir); err != nil {
		if os.IsNotExist(err) {
			return core.ErrorMsg{Err: fmt.Errorf("source directory does not exist: %s", m.cfg.SourceDir)}
		}
		return core.ErrorMsg{Err: fmt.Errorf("unable to access source directory: %w", err)}
	} else if !info.IsDir() {
		return core.ErrorMsg{Err: fmt.Errorf("source path is not a directory: %s", m.cfg.SourceDir)}
	}

	// Initialize Ignore Logic
	matcher := ignore.NewMatcher(m.cfg.SourceDir)

	defaultIgnorePath := filepath.Join(m.cfg.SourceDir, ".contextignore")
	if patterns, err := ignore.ParseFile(defaultIgnorePath); err == nil {
		matcher.AddPatterns(patterns)
		// Explicitly ignore the config file itself
		manualRules := ignore.CompileIgnoreLines([]string{".contextignore"})
		matcher.AddPatterns(manualRules)
	}

	for _, file := range m.cfg.IgnoreFiles {
		if patterns, err := ignore.ParseFile(file); err == nil {
			matcher.AddPatterns(patterns)
		}
	}

	// Initialize and Start Walker
	m.walker = core.NewWalker(*m.cfg, matcher)
	m.walker.Start(func(msg interface{}) {
		m.program.Send(msg)
	})

	return nil
}
