package ui

import "testing"

// Every digit is three rows of four columns, or the meter goes ragged.
func TestEveryDigitIsFourColumnsWide(t *testing.T) {
	for d := 0; d <= 9; d++ {
		for i, row := range ledRows(d) {
			if got := len([]rune(row)); got != 4 {
				t.Errorf("digit %d row %d is %d columns (%q), want 4", d, i, got, row)
			}
		}
	}
}

// Every digit is distinguishable from every other one, which is the only thing
// a seven-segment display has to get right — and the risk in deriving the
// glyphs rather than copying a table someone drew by eye.
func TestEveryDigitLooksDifferent(t *testing.T) {
	seen := map[string]int{}
	for d := 0; d <= 9; d++ {
		r := ledRows(d)
		key := r[0] + "|" + r[1] + "|" + r[2]
		if prev, dup := seen[key]; dup {
			t.Errorf("digit %d draws the same as %d:\n%s\n%s\n%s", d, prev, r[0], r[1], r[2])
		}
		seen[key] = d
	}
}
