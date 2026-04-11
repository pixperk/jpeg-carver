package simulator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"jpeg-carver/ui"
)

// Config controls how the simulated image is generated.
type Config struct {
	// Paths to real JPEG files to embed in the image.
	SourceJPEGs []string

	// How many bytes of random junk to insert before, between,
	// and after the embedded JPEGs (simulates unallocated space).
	MinJunkBytes int
	MaxJunkBytes int

	// If true, the last JPEG will be truncated (no EOI),
	// simulating a partially overwritten file.
	SimulateTruncation bool
}

// GenerateImage creates a simulated raw disk image at `outPath`.
func GenerateImage(cfg Config, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for i, jpegPath := range cfg.SourceJPEGs {
		// Write random junk (simulated deleted/unallocated data)
		junkSize := randRange(cfg.MinJunkBytes, cfg.MaxJunkBytes)
		junk := make([]byte, junkSize)
		rand.Read(junk)

		// Make sure no accidental FF D8 FF in the junk
		for k := 0; k < len(junk)-2; k++ {
			if junk[k] == 0xFF && junk[k+1] == 0xD8 {
				junk[k+1] = 0x00
			}
		}

		if _, err := f.Write(junk); err != nil {
			return err
		}

		// Read the real JPEG
		jpegData, err := os.ReadFile(jpegPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", filepath.Base(jpegPath), err)
		}

		// Optionally truncate the last one
		if cfg.SimulateTruncation && i == len(cfg.SourceJPEGs)-1 {
			cutPoint := len(jpegData) / 2
			jpegData = jpegData[:cutPoint]
			fmt.Printf("%s %s %s %s %s\n",
				ui.SimTag(),
				ui.Yellow("Truncating"),
				ui.Cyan(filepath.Base(jpegPath)),
				ui.Dim(fmt.Sprintf("at %d bytes", cutPoint)),
				ui.Dim("(simulating partial overwrite)"),
			)
		}

		if _, err := f.Write(jpegData); err != nil {
			return err
		}

		fmt.Printf("%s %s %s %s\n",
			ui.SimTag(),
			ui.Green("Embedded"),
			ui.Cyan(filepath.Base(jpegPath)),
			ui.Dim(fmt.Sprintf("(%d bytes) after %d bytes of junk", len(jpegData), junkSize)),
		)
	}

	// Trailing junk
	trailing := make([]byte, randRange(cfg.MinJunkBytes, cfg.MaxJunkBytes))
	rand.Read(trailing)
	for k := 0; k < len(trailing)-2; k++ {
		if trailing[k] == 0xFF && trailing[k+1] == 0xD8 {
			trailing[k+1] = 0x00
		}
	}
	f.Write(trailing)

	stat, _ := f.Stat()
	fmt.Printf("%s %s %s %s\n",
		ui.SimTag(),
		ui.Bold("Disk image created:"),
		ui.Cyan(outPath),
		ui.Dim(fmt.Sprintf("(%d bytes)", stat.Size())),
	)
	return nil
}

func randRange(min, max int) int {
	diff := max - min
	if diff <= 0 {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	return min + int(n.Int64())
}
