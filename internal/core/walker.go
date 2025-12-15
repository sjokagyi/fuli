package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

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

// ProgressCallback is a function type that handles events from the Walker.
type ProgressCallback func(msg interface{})

// Start begins the traversal in a separate goroutine.
// It accepts a callback function to report progress.
func (w *Walker) Start(onEvent ProgressCallback) {
	go w.run(onEvent)
}

func (w *Walker) run(onEvent ProgressCallback) {
	start := time.Now()
	var totalFiles int
	var totalBytes int64

	// Initialize Composer
	comp, err := NewComposer(w.cfg.OutputPath)
	if err != nil {
		onEvent(ErrorMsg{Err: err})
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
			logger.Error("Access error during walk", "path", path, "error", err)
			if w.cfg.IsVerbose {
				onEvent(ErrorMsg{Err: fmt.Errorf("access error at %s: %v", path, err)})
			}
			return nil // Continue walking other files
		}

		// Safety: Skip Symlinks explicitly
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
				onEvent(ErrorMsg{Err: fmt.Errorf("error checking binary %s: %v", path, err)})
			}
			return nil
		}
		if isBin {
			logger.Debug("Skipping file (binary detected)", "path", path)
			return nil
		}

		// 6. Composition
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
			onEvent(ErrorMsg{Err: err})
			return nil
		}

		totalFiles++
		totalBytes += bytesWritten

		// Send progress update
		onEvent(FileProcessedMsg{
			Path:  displayPath,
			Bytes: bytesWritten,
		})

		return nil
	})

	if err != nil {
		onEvent(ErrorMsg{Err: err})
	}

	// Send final completion stats
	onEvent(CompletionMsg{
		TotalFiles: totalFiles,
		TotalBytes: totalBytes,
		Elapsed:    time.Since(start),
	})
}
