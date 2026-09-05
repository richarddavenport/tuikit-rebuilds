package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// Folding a container hides everything inside it AND its closing bracket, which
// comp.Tree gets right for free because the close line sits one depth inside
// the subtree the open line owns.
func TestFoldingHidesTheClosingBracketToo(t *testing.T) {
	frame := shot.Frame(States()[1], shot.Width, shot.Height)
	if strings.Contains(frame, "1.3.10") {
		t.Error("a folded container still shows its contents")
	}
	if !strings.Contains(frame, "▸") {
		t.Error("no fold marker")
	}
	if !strings.Contains(frame, "keywords") {
		t.Error("folding one container hid a sibling")
	}
}

// A folded container previews its brackets, which is what makes folding a lock
// file readable rather than just shorter.
func TestAFoldedContainerSaysWhatItIs(t *testing.T) {
	frame := shot.Frame(States()[2], shot.Width, shot.Height)
	if !strings.Contains(frame, "{ … }") && !strings.Contains(frame, "[ … ]") {
		t.Errorf("no preview on a folded container:\n%s", frame)
	}
}

// Search moves the cursor to the first hit and counts them.
func TestSearchCountsAndJumps(t *testing.T) {
	m := build("/", "v", "e", "r", "enter")().(*Model)
	if len(m.matches) == 0 {
		t.Fatal("no matches for ver")
	}
	if m.doc.Cursor() != m.matches[0] {
		t.Errorf("cursor is at %d, first match at %d", m.doc.Cursor(), m.matches[0])
	}
}

// While the search box has the keyboard, j is the letter j.
func TestTypingCapturesTheKeyboard(t *testing.T) {
	m := build("/", "j")().(*Model)
	if m.query != "j" {
		t.Errorf("query is %q — j moved the cursor instead of being typed", m.query)
	}
	if m.doc.Cursor() != 0 {
		t.Errorf("the cursor moved to %d while typing", m.doc.Cursor())
	}
}
