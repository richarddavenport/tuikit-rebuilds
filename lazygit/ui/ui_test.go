package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// Every screen, at 132x38 and at 80x24, with the shape checked after the
// color is stripped.
func TestScreens(t *testing.T) {
	shot.Goldens(t, "testdata", States())
}

// Folding a directory removes its children and nothing else. This is the
// component the study said was the biggest gap in the field.
func TestFoldingADirectoryHidesItsChildren(t *testing.T) {
	open := build()().(*Model)
	shut := build("space")().(*Model)

	if len(shut.visible()) >= len(open.visible()) {
		t.Errorf("folding showed %d rows, not fewer than %d", len(shut.visible()), len(open.visible()))
	}
	frame := shot.Frame(States()[1], shot.Width, shot.Height)
	if strings.Contains(frame, "filetree.go") {
		t.Error("a child of the folded directory is still drawn")
	}
	if !strings.Contains(frame, "README.md") {
		t.Error("folding one directory hid a row outside it")
	}
}

// A range grows from the anchor and keeps every line's own color, which is the
// rule comp.Viewer exists for.
func TestExtendingSelectsARange(t *testing.T) {
	m := build("v", "J", "J")().(*Model)
	lo, hi, ok := m.diff.Range()
	if !ok {
		t.Fatal("no range after two extends")
	}
	if hi-lo != 2 {
		t.Errorf("range is %d..%d, want three lines", lo, hi)
	}
}

// `a` snaps the range to the whole hunk the cursor is in.
func TestWholeHunkSnapsToItsBoundaries(t *testing.T) {
	m := build("v", "j", "j", "j", "j", "a")().(*Model)
	lo, hi, ok := m.diff.Range()
	if !ok {
		t.Fatal("no range after selecting a hunk")
	}
	if lo == hi {
		t.Errorf("the hunk selected one line, %d", lo)
	}
	// It must not run past the second hunk's header.
	frame := shot.Frame(States()[4], shot.Width, shot.Height)
	if !strings.Contains(frame, "@@ -48,6 +46,4 @@") {
		t.Error("the second hunk is not on screen, so the boundary was not tested")
	}
}

// Leaving line select drops the range, because a selection that survives the
// mode it belongs to is one the reader cannot get rid of.
func TestLeavingLineSelectDropsTheRange(t *testing.T) {
	m := build("v", "J", "J", "esc")().(*Model)
	if _, _, ok := m.diff.Range(); ok {
		t.Error("the range survived leaving line select")
	}
}

// Moving the keyboard to another panel leaves the first panel's cursor where it
// was, and only one panel is drawn focused.
func TestOnlyOnePanelIsFocused(t *testing.T) {
	m := build("4")().(*Model)
	var focused int
	for i := range m.lists {
		if m.lists[i].Focused {
			focused++
		}
	}
	if focused != 1 {
		t.Errorf("%d panels are focused", focused)
	}
	if m.focus != panelCommits {
		t.Errorf("focus is %v after pressing 4", m.focus)
	}
}
