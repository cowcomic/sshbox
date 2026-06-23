package cmd

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// displayWidth returns the display width of a string, counting CJK characters as 2.
func displayWidth(s string) int {
	w := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if isCJK(r) {
			w += 2
		} else {
			w += 1
		}
		i += size
	}
	return w
}

// padRight pads s with spaces to reach the target display width.
func padRight(s string, width int) string {
	dw := displayWidth(s)
	if dw >= width {
		return s
	}
	return s + spaces(width-dw)
}

func spaces(n int) string {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = ' '
	}
	return string(buf)
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r)
}

// printTable writes a table with properly aligned columns.
// headers is the list of column headers, rows is a list of rows (each row is a list of cell values).
func printTable(w io.Writer, headers []string, rows [][]string) {
	// Calculate max display width per column
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if dw := displayWidth(cell); dw > widths[i] {
					widths[i] = dw
				}
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Fprint(w, padRight(h, widths[i]))
		if i < len(headers)-1 {
			fmt.Fprint(w, "  ")
		}
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprint(w, padRight(cell, widths[i]))
			}
			if i < len(row)-1 {
				fmt.Fprint(w, "  ")
			}
		}
		fmt.Fprintln(w)
	}
}
