package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"jpeg-carver/carver"
	"jpeg-carver/report"
	"jpeg-carver/simulator"
	"jpeg-carver/ui"
	"jpeg-carver/validator"
)

func errf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", ui.Red("[error]"), fmt.Sprintf(format, a...))
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "simulate":
		runSimulate()
	case "carve":
		runCarve()
	case "demo":
		runDemo()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(ui.Header("JPEG Forensic Carver"))
	fmt.Println()
	fmt.Println(ui.Bold("Usage:"))
	fmt.Printf("  %s %s\n", ui.Green("jpeg-carver simulate"), ui.Dim("<jpg1> <jpg2> ..."))
	fmt.Println("      Creates a simulated raw disk image with embedded JPEGs.")
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.Green("jpeg-carver carve"), ui.Dim("<image.dd> [output_dir]"))
	fmt.Println("      Carves JPEG files from a raw disk image.")
	fmt.Println()
	fmt.Printf("  %s\n", ui.Green("jpeg-carver demo"))
	fmt.Println("      Full demo: generates test images, builds a simulated disk,")
	fmt.Println("      carves them back, validates, and produces a report.")
}

func runSimulate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jpeg-carver simulate <jpg1> <jpg2> ...")
		os.Exit(1)
	}

	cfg := simulator.Config{
		SourceJPEGs:        os.Args[2:],
		MinJunkBytes:       4096,
		MaxJunkBytes:       65536,
		SimulateTruncation: true,
	}

	err := simulator.GenerateImage(cfg, "simulated_disk.dd")
	if err != nil {
		errf("%v", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s %s\n", ui.Green("OK"), ui.Bold("Disk image ready: simulated_disk.dd"))
	fmt.Printf("%s %s\n", ui.Dim("Now run:"), ui.Cyan("jpeg-carver carve simulated_disk.dd"))
}

func runCarve() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jpeg-carver carve <image.dd> [output_dir]")
		os.Exit(1)
	}

	imagePath := os.Args[2]
	outDir := "recovered"
	if len(os.Args) >= 4 {
		outDir = os.Args[3]
	}

	fmt.Printf("%s Carving JPEGs from %s ...\n", ui.StarTag(), ui.Cyan(imagePath))
	start := time.Now()

	results, err := carver.CarveJPEGs(imagePath, outDir)
	if err != nil {
		errf("%v", err)
		os.Exit(1)
	}

	// Validate each recovered file
	for i := range results {
		results[i].Valid = validator.ValidateJPEG(results[i].OutputPath)
	}

	elapsed := time.Since(start)
	carver.Summary(results, elapsed)

	// Generate report
	reportPath := filepath.Join(outDir, "forensic_report.txt")
	if err := report.Generate(results, imagePath, reportPath, elapsed); err != nil {
		fmt.Fprintf(os.Stderr, "%s report generation failed: %v\n", ui.Yellow("[warn]"), err)
	} else {
		fmt.Printf("\n%s Forensic report saved to: %s\n", ui.Green("[ok]"), ui.Cyan(reportPath))
	}
}

func runDemo() {
	fmt.Println(ui.Header("=== JPEG Forensic Carver — Full Demo ==="))
	fmt.Println()

	// Step 1: Create some tiny test JPEGs
	fmt.Printf("%s %s\n", ui.Step(1, 4), ui.Bold("Generating test JPEG files ..."))
	os.MkdirAll("demo_input", 0755)

	for i := 0; i < 4; i++ {
		createMinimalJPEG(fmt.Sprintf("demo_input/test_%d.jpg", i), 5000+i*3000)
	}

	// Step 2: Build simulated disk image
	fmt.Printf("%s %s\n", ui.Step(2, 4), ui.Bold("Building simulated disk image ..."))
	cfg := simulator.Config{
		SourceJPEGs: []string{
			"demo_input/test_0.jpg",
			"demo_input/test_1.jpg",
			"demo_input/test_2.jpg",
			"demo_input/test_3.jpg",
		},
		MinJunkBytes:       8192,
		MaxJunkBytes:       32768,
		SimulateTruncation: true,
	}
	if err := simulator.GenerateImage(cfg, "demo_disk.dd"); err != nil {
		errf("%v", err)
		os.Exit(1)
	}

	// Step 3: Carve
	fmt.Printf("\n%s %s\n", ui.Step(3, 4), ui.Bold("Carving JPEGs from simulated disk ..."))
	start := time.Now()
	results, err := carver.CarveJPEGs("demo_disk.dd", "demo_recovered")
	if err != nil {
		errf("%v", err)
		os.Exit(1)
	}
	for i := range results {
		results[i].Valid = validator.ValidateJPEG(results[i].OutputPath)
	}
	elapsed := time.Since(start)

	// Step 4: Report
	fmt.Printf("%s %s\n", ui.Step(4, 4), ui.Bold("Generating forensic report ..."))
	report.Generate(results, "demo_disk.dd", "demo_recovered/forensic_report.txt", elapsed)

	carver.Summary(results, elapsed)
	fmt.Printf("\n%s\n", ui.Header("=== Demo complete! ==="))
	fmt.Printf("  %s : %s\n", ui.Bold("Simulated disk "), ui.Cyan("demo_disk.dd"))
	fmt.Printf("  %s : %s\n", ui.Bold("Recovered files"), ui.Cyan("demo_recovered/"))
	fmt.Printf("  %s : %s\n", ui.Bold("Report         "), ui.Cyan("demo_recovered/forensic_report.txt"))
}

// createMinimalJPEG writes a minimal but valid JPEG file of approximately
// `size` bytes. It uses SOI + APP0 (JFIF) + padding via COM markers + EOI.
func createMinimalJPEG(path string, size int) {
	header := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, // SOI + APP0
		0x00, 0x10, // APP0 length (16 bytes)
		0x4A, 0x46, 0x49, 0x46, 0x00, // "JFIF\0"
		0x01, 0x01, // version 1.1
		0x00,       // aspect ratio units
		0x00, 0x01, // X density
		0x00, 0x01, // Y density
		0x00, 0x00, // no thumbnail
	}

	paddingSize := size - len(header) - 2 // -2 for EOI
	if paddingSize < 4 {
		paddingSize = 4
	}
	if paddingSize > 65533 {
		var body []byte
		remaining := paddingSize
		for remaining > 0 {
			blockSize := remaining
			if blockSize > 65533 {
				blockSize = 65533
			}
			commentLen := blockSize + 2
			block := []byte{0xFF, 0xFE, byte(commentLen >> 8), byte(commentLen & 0xFF)}
			pad := make([]byte, blockSize)
			for k := range pad {
				pad[k] = byte(k % 251) // avoid FF D8 and FF D9 patterns
			}
			block = append(block, pad...)
			body = append(body, block...)
			remaining -= blockSize
		}
		header = append(header, body...)
	} else {
		commentLen := paddingSize + 2
		comment := []byte{0xFF, 0xFE, byte(commentLen >> 8), byte(commentLen & 0xFF)}
		pad := make([]byte, paddingSize)
		for k := range pad {
			pad[k] = byte(k % 251)
		}
		comment = append(comment, pad...)
		header = append(header, comment...)
	}

	// EOI
	header = append(header, 0xFF, 0xD9)

	os.WriteFile(path, header, 0644)
}
