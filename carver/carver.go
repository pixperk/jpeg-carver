package carver

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"jpeg-carver/ui"
)

// Result holds metadata about a single carved file.
type Result struct {
	Index      int
	Offset     int64 // byte offset of SOI in the image
	Size       int64
	OutputPath string
	Valid      bool // set later by validator
	Truncated  bool // true if EOI was never found
}

// CarveJPEGs scans `imagePath` for JPEG SOI/EOI pairs and writes
// each recovered file into `outDir`. Returns a slice of results.
func CarveJPEGs(imagePath, outDir string) ([]Result, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	var results []Result
	count := 0
	n := len(data)

	for i := 0; i < n-2; i++ {
		// Look for SOI: FF D8 FF
		if data[i] == 0xFF && data[i+1] == 0xD8 && data[i+2] == 0xFF {
			soiOffset := i

			// Search forward for EOI: FF D9
			eoiFound := false
			j := i + 3
			for j < n-1 {
				if data[j] == 0xFF && data[j+1] == 0xD9 {
					eoiFound = true
					break
				}
				j++
			}

			var chunk []byte
			truncated := false
			if eoiFound {
				chunk = data[soiOffset : j+2] // include the FF D9
			} else {
				// No EOI found — extract to end of image (truncated file)
				chunk = data[soiOffset:]
				truncated = true
			}

			outName := fmt.Sprintf("recovered_%04d.jpg", count)
			outPath := filepath.Join(outDir, outName)
			if err := os.WriteFile(outPath, chunk, 0644); err != nil {
				return nil, fmt.Errorf("write %s: %w", outName, err)
			}

			results = append(results, Result{
				Index:      count,
				Offset:     int64(soiOffset),
				Size:       int64(len(chunk)),
				OutputPath: outPath,
				Truncated:  truncated,
			})

			count++

			// Advance past this JPEG to avoid re-finding embedded thumbnails
			// inside the same JPEG (EXIF thumbnails share the SOI marker).
			if eoiFound {
				i = j + 1
			} else {
				break
			}
		}
	}

	return results, nil
}

// Summary prints a human-readable summary to stdout.
func Summary(results []Result, elapsed time.Duration) {
	fmt.Printf("\n%s\n", ui.Header("=== Carving Summary ==="))

	valid, truncated := 0, 0
	for _, r := range results {
		if r.Valid {
			valid++
		}
		if r.Truncated {
			truncated++
		}
	}

	fmt.Printf("%s : %s\n", ui.Bold("Files recovered"), ui.Cyan(fmt.Sprintf("%d", len(results))))
	fmt.Printf("%s : %s\n", ui.Bold("Valid JPEGs    "), ui.Green(fmt.Sprintf("%d", valid)))
	fmt.Printf("%s : %s\n", ui.Bold("Truncated      "), ui.Yellow(fmt.Sprintf("%d", truncated)))
	fmt.Printf("%s : %s\n\n", ui.Bold("Time elapsed   "), ui.Dim(elapsed.String()))

	for _, r := range results {
		var tag string
		switch {
		case r.Truncated:
			tag = ui.Yellow("[TRUNCATED]")
		case !r.Valid:
			tag = ui.Red("[CORRUPT]  ")
		default:
			tag = ui.Green("[OK]       ")
		}
		fmt.Printf("  %s %s  %s  %s\n",
			tag,
			ui.Cyan(filepath.Base(r.OutputPath)),
			ui.Dim(fmt.Sprintf("offset=0x%08X", r.Offset)),
			ui.Dim(fmt.Sprintf("size=%d bytes", r.Size)),
		)
	}
}
