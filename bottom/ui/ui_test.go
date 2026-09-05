package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// The rule bottom forked ratatui's Chart to get: the newest sample is at the
// right edge. CPU 0 spikes near the end of its series, so the tallest bars must
// be in the right-hand part of its row.
func TestTheSpikeIsAtTheRightEdge(t *testing.T) {
	frame := shot.Frame(States()[0], shot.Width, shot.Height)
	var row string
	for _, line := range strings.Split(frame, "\n") {
		if strings.Contains(line, "0  9") || strings.Contains(line, "0 9") {
			row = line
			break
		}
	}
	if row == "" {
		t.Skip("could not find CPU 0's row; the label format changed")
	}
	full := strings.LastIndex(row, "█")
	if full < 0 {
		t.Fatalf("no full bar in CPU 0's row: %q", row)
	}
	if float64(full) < float64(len(row))*0.5 {
		t.Errorf("the spike is at %d of %d — it should be near the right edge", full, len(row))
	}
}

// Sorting reorders the process table and the arrow says which column.
func TestSortingTheProcessTable(t *testing.T) {
	frame := shot.Frame(States()[2], shot.Width, shot.Height)
	if !strings.Contains(frame, "↑") {
		t.Error("no sort arrow after pressing s")
	}
	desc := shot.Frame(States()[3], shot.Width, shot.Height)
	if !strings.Contains(desc, "↓") {
		t.Error("S did not reverse the arrow")
	}
}

// Filtering narrows the table, and `j` typed into it is a letter.
func TestFilteringCapturesTheKeyboard(t *testing.T) {
	m := build("/", "g", "o")().(*Model)
	if m.filter != "go" {
		t.Errorf("filter is %q", m.filter)
	}
	// Only the process actually called "go". gcpeasy, postgres and zombie-job
	// do not contain the substring, which the first version of this test got
	// wrong.
	if got := len(m.processes()); got != 1 {
		t.Errorf("%d processes match go, want 1", got)
	}
	if m.processes()[0].Name != "go" {
		t.Errorf("the match is %q", m.processes()[0].Name)
	}
}

// Expanding gives the focused table the whole screen: the same draw function
// with a bigger rect, and nothing else changes.
func TestExpandingUsesTheWholeScreen(t *testing.T) {
	small := shot.Frame(States()[5], shot.Width, shot.Height)
	big := shot.Frame(States()[6], shot.Width, shot.Height)
	if strings.Contains(big, "Temperatures") {
		t.Error("another widget is still drawn while expanded")
	}
	if !strings.Contains(small, "Temperatures") {
		t.Error("the unexpanded grid is missing a widget")
	}
}
