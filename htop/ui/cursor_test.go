package ui

import (
	"testing"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// The cursor against a row set that changes underneath it, which dive has
// reported five separate times. See ../dive/STUDY.md for the audit.

// dive#295, #259 and #308: hold ↓ past the end of a filtered list.
//
// dive#295 is a panic with a stack trace — `index out of range [15] with
// length 15` in `SetCursor`. It cannot happen here, because `List.resolve`
// takes the row count every frame and clamps to it rather than trusting a
// cursor it stored earlier.
func TestTheCursorCannotLeaveTheList(t *testing.T) {
	m := New()
	m.SetSize(shot.Width, shot.Height)
	r := app.New(m, app.WithSize(shot.Width, shot.Height))

	harness.Press(r, "f4", "p", "o", "s", "t", "enter")
	n := len(m.rows())
	if n == 0 || n >= 10 {
		t.Fatalf("the fixture changed: filter matched %d rows, wanted a few", n)
	}

	keys := make([]string, 60)
	for i := range keys {
		keys[i] = "j"
	}
	harness.Press(r, keys...)

	if got := m.procs.Cursor(); got >= n {
		t.Errorf("cursor at %d with %d rows — this is dive#295's panic", got, n)
	}
}

// dive#468 and #543, fixed rather than reproduced.
//
// This test used to assert the WRONG behaviour on purpose and fail loudly when
// tuikit#80 was fixed. It fired, so here is what it always wanted to say:
// filter to a few processes, move onto one, clear the filter, and the cursor is
// still on that process.
//
// The only change in the rebuild is `Key: itoa(rows[i].PID)` on the row.
// dive has this bug reported twice, five years apart, and still has it.
func TestTheCursorFollowsItsProcess(t *testing.T) {
	m := New()
	m.SetSize(shot.Width, shot.Height)
	r := app.New(m, app.WithSize(shot.Width, shot.Height))

	harness.Press(r, "f4", "p", "o", "s", "t", "enter", "j", "j")
	rows := m.rows()
	if len(rows) == 0 {
		t.Fatal("filter matched nothing")
	}
	before := rows[m.procs.Cursor()]

	harness.Press(r, "f4", "esc")
	after := m.rows()[m.procs.Cursor()]

	if after.PID != before.PID {
		t.Errorf("the cursor left its process when the filter cleared:\n  was %d %q\n  now %d %q",
			before.PID, before.Command, after.PID, after.Command)
	}
}
