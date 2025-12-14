package core_test

import (
	"os"
	"testing"

	"github.com/sjokagyi/fuli/internal/core"
)

func TestIsBinary(t *testing.T) {
	// Create temporary test files
	tmp := t.TempDir()

	// Text File
	txtFile := tmp + "/test.txt"
	os.WriteFile(txtFile, []byte("Hello World\nfunc main() {}"), 0644)

	// Binary File (simulate with NUL byte)
	binFile := tmp + "/test.bin"
	os.WriteFile(binFile, []byte{0xFF, 0xD8, 0xFF, 0x00, 0x10}, 0644)

	// PNG File (by extension logic)
	imgFile := tmp + "/image.png"
	os.WriteFile(imgFile, []byte("fake content"), 0644)

	tests := []struct {
		path   string
		expect bool
	}{
		{txtFile, false},
		{binFile, true},
		{imgFile, true},
	}

	for _, tt := range tests {
		isBin, err := core.IsBinary(tt.path)
		if err != nil {
			t.Fatalf("Unexpected error for %s: %v", tt.path, err)
		}
		if isBin != tt.expect {
			t.Errorf("File %s: expected binary=%v, got %v", tt.path, tt.expect, isBin)
		}
	}
}
