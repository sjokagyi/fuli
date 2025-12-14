package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Composer handles the buffered writing of files into the master output file.
type Composer struct {
	file   *os.File
	writer *bufio.Writer
}

// NewComposer creates a new Composer.
// It opens the file with O_TRUNC to overwrite existing files, ensuring a fresh start.
func NewComposer(outputPath string) (*Composer, error) {
	// 0644 provides read/write for the user and read-only for others.
	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	// Use a 4KB buffer (default) to reduce syscalls when writing many small lines.
	return &Composer{
		file:   f,
		writer: bufio.NewWriter(f),
	}, nil
}

// Append writes the header and content of a source file to the output.
// It uses io.Copy for memory efficiency, streaming from source to destination
// without loading the entire source file into RAM.
func (c *Composer) Append(fileName, relativeDir string, source io.Reader) (int64, error) {
	// Construct Header according to specification:
	// This is the content of <file_name> in the <directory> directory:
	header := fmt.Sprintf("This is the content of %s in the %s directory:\n\"\n", fileName, relativeDir)

	// Write Header
	if _, err := c.writer.WriteString(header); err != nil {
		return 0, fmt.Errorf("failed to write header: %w", err)
	}

	// Stream Content
	written, err := io.Copy(c.writer, source)
	if err != nil {
		return written, fmt.Errorf("failed to copy content: %w", err)
	}

	// Write Footer (Closing quote and newline)
	// Ensure there's a newline before the closing quote if the file didn't end with one?
	// The prompt format is strict: content -> " -> newline.
	footer := "\"\n\n"
	if _, err := c.writer.WriteString(footer); err != nil {
		return written, fmt.Errorf("failed to write footer: %w", err)
	}

	return written, nil
}

// Close flushes the buffer and closes the underlying file handle.
func (c *Composer) Close() error {
	if err := c.writer.Flush(); err != nil {
		c.file.Close() // Attempt to close even if flush fails
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	return c.file.Close()
}
