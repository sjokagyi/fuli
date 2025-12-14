package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sjokagyi/fuli/internal/config"
	"github.com/sjokagyi/fuli/internal/ignore"
	"github.com/sjokagyi/fuli/internal/logger"
)

// Messages for the Bubble Tea loop
type FileProcessedMsg struct {
	Path  string
	Bytes int64
}

type ErrorMsg struct {
	Err error
}

type CompletionMsg struct {
	TotalFiles int
	TotalBytes int64
	Elapsed    time.Duration
}

// Walker orchestrates the directory traversal.
type Walker struct {
	cfg     config.RunConfig
	matcher *ignore.Matcher
}

func NewWalker(cfg config.RunConfig, m *ignore.Matcher) *Walker {
	return &Walker{
		cfg:     cfg,
		matcher: m,
	}
}

// Start begins the traversal in a separate goroutine.
func (w *Walker) Start(prog *tea.Program) {
	go w.run(prog)
}

func (w *Walker) run(prog *tea.Program) {
	start := time.Now()
	var totalFiles int
	var totalBytes int64

	// Initialize Composer
	comp, err := NewComposer(w.cfg.OutputPath)
	if err != nil {
		prog.Send(ErrorMsg{Err: err})
		return
	}
	defer comp.Close()

	// Get absolute path of output file to prevent self-inclusion loops
	absOut, _ := filepath.Abs(w.cfg.OutputPath)

	// Identify the Root Name
	rootName := filepath.Base(w.cfg.SourceDir)

	// Walk
	err = filepath.WalkDir(w.cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log I/O errors (permission denied, etc.)
			logger.Error("Access error during walk", "path", path, "error", err)
			if w.cfg.IsVerbose {
				prog.Send(ErrorMsg{Err: fmt.Errorf("access error at %s: %v", path, err)})
			}
			return nil // Continue walking other files
		}

		// Safety: Skip Symlinks explicitly
		// This prevents infinite loops and referencing files outside the project scope unexpectedly.
		if d.Type()&fs.ModeSymlink != 0 {
			logger.Debug("Skipping symlink", "path", path)
			return nil
		}

		// Prevent infinite recursion (Self-Inclusion)
		if absPath, _ := filepath.Abs(path); absPath == absOut {
			logger.Debug("Skipping output file (self-inclusion)", "path", path)
			return nil
		}

		// Ignore Logic (ContextIgnore)
		if w.matcher.Matches(path, d.IsDir()) {
			if d.IsDir() {
				logger.Debug("Skipping directory (ignored)", "path", path)
				return filepath.SkipDir
			}
			logger.Debug("Skipping file (ignored)", "path", path)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Binary Detection
		isBin, err := IsBinary(path)
		if err != nil {
			logger.Error("Error detecting file type", "path", path, "error", err)
			if w.cfg.IsVerbose {
				prog.Send(ErrorMsg{Err: fmt.Errorf("error checking binary %s: %v", path, err)})
			}
			return nil
		}
		if isBin {
			logger.Debug("Skipping file (binary detected)", "path", path)
			return nil
		}

		// Composition
		f, err := os.Open(path)
		if err != nil {
			logger.Error("Failed to open file", "path", path, "error", err)
			return nil // Skip unreadable files
		}
		defer f.Close()

		// --- ROOT-ANCHORED PATH CALCULATION ---
		relFromSource, err := filepath.Rel(w.cfg.SourceDir, path)
		if err != nil {
			relFromSource = filepath.Base(path)
		}

		displayPath := filepath.Join(rootName, relFromSource)
		relDir := filepath.Dir(displayPath)
		fileName := filepath.Base(path)

		bytesWritten, err := comp.Append(fileName, relDir, f)
		if err != nil {
			prog.Send(ErrorMsg{Err: err})
			return nil
		}

		totalFiles++
		totalBytes += bytesWritten

		// Send progress update to UI
		prog.Send(FileProcessedMsg{
			Path:  displayPath,
			Bytes: bytesWritten,
		})

		return nil
	})

	if err != nil {
		prog.Send(ErrorMsg{Err: err})
	}

	// Send final completion stats
	prog.Send(CompletionMsg{
		TotalFiles: totalFiles,
		TotalBytes: totalBytes,
		Elapsed:    time.Since(start),
	})
}
