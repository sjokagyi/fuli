package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var (
	// Default global logger
	l *log.Logger
)

// Init sets up the global logger.
// If verbose is true, it logs to a file (fuli.debug.log).
// If verbose is false, it discards logs to keep the TUI clean.
func Init(verbose bool) error {
	var output io.Writer

	if verbose {
		// Log to a file in the current directory for easy debugging
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		logFile := filepath.Join(cwd, "fuli.debug.log")
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = f
	} else {
		// Discard output if not verbose
		output = io.Discard
	}

	l = log.New(output)
	l.SetReportTimestamp(true)
	l.SetLevel(log.DebugLevel) // Capture everything in the file

	if verbose {
		l.Info("Logger initialized", "verbose", true)
	}

	return nil
}

// Info logs useful information.
func Info(msg string, keyvals ...interface{}) {
	if l != nil {
		l.Info(msg, keyvals...)
	}
}

// Debug logs detailed developer info.
func Debug(msg string, keyvals ...interface{}) {
	if l != nil {
		l.Debug(msg, keyvals...)
	}
}

// Error logs error conditions.
func Error(msg string, keyvals ...interface{}) {
	if l != nil {
		l.Error(msg, keyvals...)
	}
}

// Close ensures any open file handles are synced/closed (best effort).
// Note: Standard os.File doesn't strictly need a close for logger if the app is exiting,
// but it's good practice if we expand this later.
func Close() {
	// Implementation depends on if we store the file handle struct,
	// simpler to leave empty for this scope or rely on OS cleanup.
}
