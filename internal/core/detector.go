package core

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// BinaryExtensions is a set of common file extensions that should definitely be skipped.
// Using a map for O(1) lookups.
var BinaryExtensions = map[string]struct{}{
	".exe": {}, ".bin": {}, ".dll": {}, ".so": {}, ".dylib": {},
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".ico": {},
	".zip": {}, ".tar": {}, ".gz": {}, ".7z": {}, ".rar": {},
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {},
	".pyc": {}, ".class": {}, ".jar": {}, ".o": {}, ".a": {},
	".DS_Store": {},
}

// IsBinary determines if a file is binary or text.
// It uses a tiered approach:
// 1. Check file extension against a known blacklist.
// 2. Read the first 512 bytes (sniffing buffer).
// 3. Check for NUL bytes (common in binaries).
// 4. Use http.DetectContentType as a final fallback.
func IsBinary(path string) (bool, error) {
	// Fast fail on extension
	ext := strings.ToLower(filepath.Ext(path))
	if _, ok := BinaryExtensions[ext]; ok {
		return true, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read first 512 bytes
	// 512 bytes is the standard sniffing size used by http.DetectContentType
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	// If empty file, it's effectively text (safe to copy)
	if n == 0 {
		return false, nil
	}

	slice := buf[:n]

	// Check for NUL byte (The Git heuristic)
	// A NUL byte in the first 512 bytes is a very strong indicator of a binary file.
	if bytes.IndexByte(slice, 0) != -1 {
		return true, nil
	}

	// MIME sniffing fallback
	contentType := http.DetectContentType(slice)
	return strings.HasPrefix(contentType, "application/octet-stream"), nil
}
