# jpeg-carver

A small JPEG file carver written in Go. It scans a raw disk image byte-by-byte
for JPEG start-of-image (`FF D8 FF`) and end-of-image (`FF D9`) markers, extracts
every matching pair as a standalone `.jpg`, validates the structure, and writes a
plain-text forensic report.

Also includes a **simulation mode** that builds a fake raw disk image by burying
real JPEGs in random junk — so you can demo recovery without needing an actual
forensic image.

## Build

```bash
go build -o jpeg-carver .
```

## Usage

```bash
# Self-contained demo: generates test JPEGs, builds a fake disk,
# carves them back, validates, and writes a report.
./jpeg-carver demo

# Build a simulated disk image from real JPEGs
./jpeg-carver simulate photo1.jpg photo2.jpg photo3.jpg
# -> writes simulated_disk.dd

# Carve JPEGs from any raw image (real or simulated)
./jpeg-carver carve simulated_disk.dd recovered/
# -> writes recovered/recovered_NNNN.jpg + recovered/forensic_report.txt
```

## Example: end-to-end with real photos

```bash
./jpeg-carver simulate sample_photos/photo_1.jpg sample_photos/photo_2.jpg \
                       sample_photos/photo_3.jpg sample_photos/photo_4.jpg
./jpeg-carver carve simulated_disk.dd recovered_photos
md5sum sample_photos/*.jpg recovered_photos/*.jpg   # originals == recovered
```

## Project layout

```
carver/      Core scanning + extraction engine
validator/   Structural JPEG validation (SOI/EOI + size sanity)
simulator/   Fake raw disk image generator
report/      Plain-text forensic report writer
main.go      CLI entry point (simulate / carve / demo)
```
