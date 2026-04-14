# JPEG Forensic Carver

A forensic file carver written in Go that recovers JPEG images from raw disk images.

---

## Quick Demo

```bash
# build
go build -o jpeg-carver .

# create a fake disk image with 4 sample photos buried in random junk
./jpeg-carver simulate sample_photos/photo_1.jpg sample_photos/photo_2.jpg \
                       sample_photos/photo_3.jpg sample_photos/photo_4.jpg

# carve them back out
./jpeg-carver carve simulated_disk.dd recovered/
```

That's it — two commands. `recovered/` will contain the extracted JPEGs and a
`forensic_report.txt`. You can verify the recovery was lossless:

```bash
md5sum sample_photos/*.jpg recovered/*.jpg
```

There is also a self-contained `demo` command that does everything in one shot
using auto-generated test JPEGs (no input files needed):

```bash
./jpeg-carver demo
```

---

## What This Tool Does

When a file is deleted, the operating system only removes the directory entry
that points to it — the actual bytes remain on disk until something else
overwrites them. This is why "deleted" photos can often be recovered.

This tool performs **file carving**: it scans the raw bytes of a disk image,
completely ignoring the file system, and looks for JPEG file signatures to
extract every image it can find.

### How the carving works

1. The entire disk image is read into memory.
2. Every byte position is scanned for the JPEG start-of-image marker (`FF D8 FF`).
3. From each SOI, it scans forward for the end-of-image marker (`FF D9`).
4. The byte range `[SOI ... EOI+2]` is extracted and written as a new `.jpg` file.
5. If no EOI is found before the end of the image, the file is extracted to the
   end and flagged as **truncated** — this means it was likely partially
   overwritten by newer data.
6. The scan advances past the EOI and repeats.

### Validation

Each recovered file is checked for structural integrity:

- Must start with `FF D8 FF` (valid JPEG SOI)
- Must end with `FF D9` (valid JPEG EOI)
- Must be at least 1000 bytes (smaller matches are almost always false
  positives — embedded EXIF thumbnails, metadata fragments, etc.)

Files that fail are flagged as **CORRUPT** or **TRUNCATED** in the output.

### SHA-256 Hashing

Every recovered file gets a SHA-256 hash computed and printed. This serves as
an evidence integrity fingerprint — if anyone re-hashes the file later and gets
the same value, it proves the file has not been modified since recovery. This is
standard practice in digital forensics to maintain chain of custody.

### Hex Dump Preview

The first 32 bytes of each recovered file are displayed as a hex dump. This lets
you visually confirm the file signature at a glance — you'll see `FF D8 FF E1`
(SOI + EXIF marker), `45 78 69 66` ("Exif" in ASCII), and the TIFF header bytes,
confirming these are real photographs with metadata intact.

### Forensic Report

A plain-text report (`forensic_report.txt`) is generated with every carve run,
documenting:

- Examiner name and timestamp
- Source image path and total size
- Number of files recovered, how many are valid vs truncated
- Per-file details: filename, status, byte offset (hex + decimal), size,
  SHA-256 hash, and hex dump

This is the kind of documentation you would include in a real forensic
investigation to preserve the chain of evidence.

### Disk Image Simulator

Since we don't have an actual forensic disk image to work with, the tool includes
a simulator that builds a realistic fake one. It takes real JPEG files and:

1. Inserts random junk data (4–64 KB) between them, simulating unallocated
   sectors on a formatted drive.
2. Scrubs the junk to remove accidental `FF D8` sequences that would cause
   false positives.
3. Optionally truncates the last JPEG at the midpoint (removes its EOI marker),
   simulating a file that was partially overwritten.

The result is a `.dd` file that behaves exactly like a raw disk image from `dd`.

---

## Project Structure

```
jpeg-carver/
├── main.go                CLI entry point (simulate / carve / demo)
├── carver/
│   └── carver.go          Core scanning engine + SHA-256 hashing + hex dump
├── validator/
│   └── validator.go       Structural JPEG validation
├── simulator/
│   └── simulator.go       Fake disk image generator
├── report/
│   └── report.go          Plain-text forensic report writer
├── ui/
│   └── color.go           Colored terminal output (respects NO_COLOR)
├── sample_photos/         4 sample JPEGs for the simulate command
│   ├── photo_1.jpg
│   ├── photo_2.jpg
│   ├── photo_3.jpg
│   └── photo_4.jpg
└── go.mod
```

---

## Why This Matters

File carving is one of the foundational techniques in digital forensics. It's
how investigators recover evidence from formatted USB drives, wiped phones, and
corrupted storage. Tools like Scalpel, Foremost, and PhotoRec all work on the
same principle — scanning raw bytes for known signatures.

This project demonstrates that principle from scratch: no libraries, no
frameworks, just raw byte scanning in Go. It handles the key real-world
scenarios — successful recovery, truncated files from partial overwrites,
false positive filtering, and proper evidence documentation with hashing.
