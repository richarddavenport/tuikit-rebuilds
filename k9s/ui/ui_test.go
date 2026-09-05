package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// The rule that makes marks worth having: one key means "the marked ones" and
// "this one", through one code path.
func TestActingFallsBackToTheCursor(t *testing.T) {
	none := build("j", "j", "ctrl+d")().(*Model)
	if got := none.acting(); len(got) != 1 {
		t.Errorf("with nothing marked, acting on %v", got)
	}
	some := build("space", "j", "space", "ctrl+d")().(*Model)
	if got := some.acting(); len(got) != 2 {
		t.Errorf("with two marked, acting on %v", got)
	}
}

// A mark is kept by identity, so a filter that removes rows above it does not
// move it to a different pod.
func TestMarksSurviveAFilter(t *testing.T) {
	m := build("space", "j", "space", "/", "j", "o", "b", "s", "enter")().(*Model)
	if m.marks.Len() != 2 {
		t.Errorf("%d marks survived the filter, want 2", m.marks.Len())
	}
	for _, k := range m.marks.Keys() {
		if !strings.HasPrefix(k, "acme/") {
			t.Errorf("a mark moved to %q — the filter shows acme-jobs", k)
		}
	}
}

// Fill reaches back to the nearest mark, which is k9s's gesture and has no
// anchor at all.
func TestFillMarksTheSpan(t *testing.T) {
	m := build("space", "j", "j", "j", "j", "ctrl+@")().(*Model)
	if m.marks.Len() != 5 {
		t.Errorf("fill marked %d rows, want 5: %v", m.marks.Len(), m.marks.Keys())
	}
}

// Sorting reorders and the arrow says which column.
func TestSortingShowsWhichColumn(t *testing.T) {
	frame := shot.Frame(States()[3], shot.Width, shot.Height)
	if !strings.Contains(frame, "↑") {
		t.Errorf("no sort arrow in the header:\n%s", strings.Split(frame, "\n")[2])
	}
	desc := shot.Frame(States()[4], shot.Width, shot.Height)
	if !strings.Contains(desc, "↓") {
		t.Error("reversing did not change the arrow")
	}
}

// Describe opens at the TOP. That is the difference from a LogPane and the
// reason comp.Viewer exists.
func TestDescribeOpensAtTheTop(t *testing.T) {
	m := build("d")().(*Model)
	if m.view.Offset() != 0 {
		t.Errorf("describe opened at line %d", m.view.Offset())
	}
	frame := shot.Frame(States()[7], shot.Width, shot.Height)
	if !strings.Contains(frame, "Name:") {
		t.Error("the first line of describe is not on screen")
	}
}
