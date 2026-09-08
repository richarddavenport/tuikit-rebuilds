package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// The gap this rebuild is about: a popup can open a popup, and esc closes ONE.
func TestPopupsStack(t *testing.T) {
	m := build(" ", "enter")().(*Model)
	if len(m.stack) != 2 {
		t.Fatalf("stack is %v, want the palette and a confirm", m.stack)
	}
	if m.top() != popupConfirm {
		t.Errorf("the top of the stack is %v", m.top())
	}
}

// Esc closes one, not all. That distinction is the reason a stack is a stack.
func TestEscClosesOnePopup(t *testing.T) {
	m := build(" ", "enter", "esc")().(*Model)
	if len(m.stack) != 1 {
		t.Fatalf("esc left %d popups, want 1", len(m.stack))
	}
	if m.top() != popupPalette {
		t.Errorf("esc came back to %v, want the palette", m.top())
	}
}

// The innermost popup gets the keyboard, and it takes the key whether or not it
// does anything with it.
//
// So `?` inside a confirm closes the confirm rather than opening help on top of
// it. That is app.Keys' contract working — "an unrecognized key inside a filter
// box is a character, not a chance for the screen underneath to act" — and it
// is worth a test because the first version of this rebuild expected the other
// thing.
func TestAPopupTakesEveryKey(t *testing.T) {
	m := build(" ", "enter", "?")().(*Model)
	if m.top() != popupPalette {
		t.Errorf("top is %v — ? should have closed the confirm, not stacked on it", m.top())
	}
	if len(m.stack) != 1 {
		t.Errorf("stack is %d deep, want 1", len(m.stack))
	}
}

// The footer says how deep the stack is, which is what makes esc predictable.
func TestTheFooterCountsTheStack(t *testing.T) {
	frame := shot.Frame(States()[6], shot.Width, shot.Height)
	if !strings.Contains(frame, "close one of 2") {
		t.Errorf("the footer does not count the stack:\n%s", strings.Split(frame, "\n")[37])
	}
}

// A range grows from the anchor and keeps each line's own color.
func TestExtendingSelectsARange(t *testing.T) {
	m := build("v", "J", "J")().(*Model)
	lo, hi, ok := m.diff.Range()
	if !ok || hi-lo != 2 {
		t.Errorf("range is %d..%d (%v)", lo, hi, ok)
	}
}

// The hunk snap stops at the hunk boundary rather than running into the next.
func TestWholeHunkStopsAtTheBoundary(t *testing.T) {
	m := build("v", "j", "j", "a")().(*Model)
	lo, hi, ok := m.diff.Range()
	if !ok {
		t.Fatal("no range after a")
	}
	if lo == hi {
		t.Error("the hunk selected one line")
	}
	frame := shot.Frame(States()[4], shot.Width, shot.Height)
	if !strings.Contains(frame, "@@ -311,6 +309,4 @@") {
		t.Error("the second hunk is off screen, so the boundary was not tested")
	}
}
