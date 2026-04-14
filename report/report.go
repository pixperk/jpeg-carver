package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"jpeg-carver/carver"
)

// Generate writes a forensic report to `outPath`.
func Generate(results []carver.Result, imagePath, outPath string, elapsed time.Duration) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "============================================\n")
	fmt.Fprintf(f, "       JPEG FORENSIC CARVING REPORT\n")
	fmt.Fprintf(f, "============================================\n\n")
	fmt.Fprintf(f, "Examiner    : Yashaswi Mishra (235805294)\n")
	fmt.Fprintf(f, "Date        : %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "Source Image: %s\n", imagePath)

	stat, _ := os.Stat(imagePath)
	fmt.Fprintf(f, "Image Size  : %d bytes\n", stat.Size())
	fmt.Fprintf(f, "Elapsed     : %s\n\n", elapsed)

	valid, truncated := 0, 0
	for _, r := range results {
		if r.Valid {
			valid++
		}
		if r.Truncated {
			truncated++
		}
	}

	fmt.Fprintf(f, "--- Summary ---\n")
	fmt.Fprintf(f, "Total files recovered : %d\n", len(results))
	fmt.Fprintf(f, "Structurally valid    : %d\n", valid)
	fmt.Fprintf(f, "Truncated (no EOI)    : %d\n\n", truncated)

	fmt.Fprintf(f, "--- Recovered Files ---\n\n")
	for _, r := range results {
		status := "VALID"
		if r.Truncated {
			status = "TRUNCATED"
		} else if !r.Valid {
			status = "CORRUPT"
		}
		fmt.Fprintf(f, "  File   : %s\n", filepath.Base(r.OutputPath))
		fmt.Fprintf(f, "  Status : %s\n", status)
		fmt.Fprintf(f, "  Offset : 0x%08X (%d)\n", r.Offset, r.Offset)
		fmt.Fprintf(f, "  Size   : %d bytes\n", r.Size)
		if r.SHA256 != "" {
			fmt.Fprintf(f, "  SHA-256: %s\n", r.SHA256)
		}
		if r.HexDump != "" {
			fmt.Fprintf(f, "  Hex    : %s\n", r.HexDump)
		}
		fmt.Fprintln(f)
	}

	fmt.Fprintf(f, "============================================\n")
	fmt.Fprintf(f, "                END OF REPORT\n")
	fmt.Fprintf(f, "============================================\n")

	return nil
}
