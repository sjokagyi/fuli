package core

import (
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
// It returns a tea.Cmd that listens to the channel, but for pure streaming
// in this architecture, we accept a send-only channel to push updates to the TUI.
//
// In Bubble Tea, the TUI usually triggers this via a Cmd that starts the goroutine.
// Start begins the traversal in a separate goroutine.
func (w *Walker) Start(prog *tea.Program) {
	go w.run(prog)
}

func (w *Walker) run(prog *tea.Program) {
	start := time.Now()
	var totalFiles int
	var totalBytes int64

	if w.cfg.IsVerbose {
		logger.Info("Starting Walker", "source", w.cfg.SourceDir, "output", w.cfg.OutputPath)
	}

	// Initialize Composer
	comp, err := NewComposer(w.cfg.OutputPath)
	if err != nil {
		prog.Send(ErrorMsg{Err: err})
		return
	}
	defer comp.Close()

	// Get absolute path of output file to prevent self-inclusion loops
	absOut, _ := filepath.Abs(w.cfg.OutputPath)

	// 1. Identify the Root Name
	rootName := filepath.Base(w.cfg.SourceDir)

	// Walk
	err = filepath.WalkDir(w.cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if w.cfg.IsVerbose {
				logger.Error("WalkDir error", "path", path, "error", err)
			}
			// Important: Return nil to continue walking if possible, or skip the dir
			return nil
		}

		// Calculate relative path for logic checks
		relFromSource, errRel := filepath.Rel(w.cfg.SourceDir, path)
		if errRel != nil {
			if w.cfg.IsVerbose {
				logger.Error("Failed to calculate relative path", "path", path, "error", errRel)
			}
			return nil
		}

		// --- 1. Root Directory Special Case ---
		// We must always enter the root directory to find its children.
		if relFromSource == "." {
			if w.cfg.IsVerbose {
				logger.Debug("Processing Root Directory", "path", path)
			}
			return nil
		}

		// --- 2. Prevent Self-Inclusion ---
		if absPath, _ := filepath.Abs(path); absPath == absOut {
			if w.cfg.IsVerbose {
				logger.Debug("Skipping Self (Output File)", "path", path)
			}
			return nil
		}

		// --- 3. Ignore Logic (ContextIgnore) ---
		// Passing the relative path to Matcher might be safer if Matcher expects relative,
		// but currently Matcher expects Absolute path.
		if w.matcher.Matches(path, d.IsDir()) {
			if w.cfg.IsVerbose {
				logger.Debug("Ignored by Matcher", "path", path, "isDir", d.IsDir())
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// --- 4. Binary Detection ---
		isBin, err := IsBinary(path)
		if err != nil {
			if w.cfg.IsVerbose {
				logger.Error("Binary check failed", "path", path, "error", err)
			}
			return nil
		}
		if isBin {
			if w.cfg.IsVerbose {
				logger.Debug("Skipping Binary File", "path", path)
			}
			return nil
		}

		// --- 5. Composition ---
		f, err := os.Open(path)
		if err != nil {
			if w.cfg.IsVerbose {
				logger.Error("Failed to open file", "path", path, "error", err)
			}
			return nil
		}
		defer f.Close()

		// Construct Display Path: "rootName/relativePath"
		displayPath := filepath.Join(rootName, relFromSource)
		relDir := filepath.Dir(displayPath)
		fileName := filepath.Base(path)

		if w.cfg.IsVerbose {
			logger.Info("Writing File", "path", path, "display", displayPath)
		}

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

	if w.cfg.IsVerbose {
		logger.Info("Walk Complete", "total_files", totalFiles, "total_bytes", totalBytes)
	}

	// Send final completion stats
	prog.Send(CompletionMsg{
		TotalFiles: totalFiles,
		TotalBytes: totalBytes,
		Elapsed:    time.Since(start),
	})
}
