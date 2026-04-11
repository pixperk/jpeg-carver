package validator

import (
	"os"
)

// ValidateJPEG checks basic structural integrity of a JPEG file.
// Returns true if the file has valid SOI and EOI markers and is
// at least 1000 bytes (tiny files are usually junk).
func ValidateJPEG(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil || len(data) < 4 {
		return false
	}

	// Check SOI
	if data[0] != 0xFF || data[1] != 0xD8 || data[2] != 0xFF {
		return false
	}

	// Check EOI
	n := len(data)
	if data[n-2] != 0xFF || data[n-1] != 0xD9 {
		return false
	}

	// Reject suspiciously small files (likely false positives)
	if n < 1000 {
		return false
	}

	return true
}
