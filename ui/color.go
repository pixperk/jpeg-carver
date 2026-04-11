// Package ui provides tiny ANSI color helpers for terminal output.
// Honors the NO_COLOR environment variable (https://no-color.org).
package ui

import "os"

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	blue   = "\x1b[34m"
	purple = "\x1b[35m"
	cyan   = "\x1b[36m"
	gray   = "\x1b[90m"
)

var enabled = os.Getenv("NO_COLOR") == ""

func wrap(code, s string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

func Bold(s string) string   { return wrap(bold, s) }
func Dim(s string) string    { return wrap(dim, s) }
func Red(s string) string    { return wrap(red, s) }
func Green(s string) string  { return wrap(green, s) }
func Yellow(s string) string { return wrap(yellow, s) }
func Blue(s string) string   { return wrap(blue, s) }
func Purple(s string) string { return wrap(purple, s) }
func Cyan(s string) string   { return wrap(cyan, s) }
func Gray(s string) string   { return wrap(gray, s) }

// Header returns a bold cyan heading.
func Header(s string) string { return wrap(bold+cyan, s) }

// Tag returns a bracketed, colored tag like "[sim]" or "[*]".
func Tag(code, label string) string { return wrap(code, "["+label+"]") }

// SimTag is the "[sim]" prefix used by the simulator.
func SimTag() string { return wrap(purple, "[sim]") }

// StarTag is the "[*]" prefix used for progress lines.
func StarTag() string { return wrap(blue, "[*]") }

// Step returns a "[n/total]" step counter in bold cyan.
func Step(n, total int) string {
	return wrap(bold+cyan, fmtStep(n, total))
}

func fmtStep(n, total int) string {
	// Small manual formatter to avoid importing fmt in the hot path.
	digits := func(x int) string {
		if x == 0 {
			return "0"
		}
		buf := [8]byte{}
		i := len(buf)
		for x > 0 {
			i--
			buf[i] = byte('0' + x%10)
			x /= 10
		}
		return string(buf[i:])
	}
	return "[" + digits(n) + "/" + digits(total) + "]"
}
