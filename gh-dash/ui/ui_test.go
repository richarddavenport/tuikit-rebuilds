package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// The measurement this rebuild exists to make. gh-dash's Update is 741 lines,
// 39% of a 1,917-line file. This one is not, and the difference is that
// app.Keys owns the routing order and comp.List owns its own viewport.
func TestUpdateIsSmall(t *testing.T) {
	src, err := readFile("ui.go")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(src, "func (m *Model) Update(")
	if start < 0 {
		t.Fatal("no Update method")
	}
	rest := src[start:]
	end := strings.Index(rest, "\nfunc ")
	if end < 0 {
		end = len(rest)
	}
	lines := strings.Count(rest[:end], "\n")
	if lines > 60 {
		t.Errorf("Update is %d lines — the point of this rebuild is that it is not 741", lines)
	}
	t.Logf("Update is %d lines", lines)
}

// CI's answer is a character, so it survives the colour being stripped.
func TestCheckStateReadsWithoutColour(t *testing.T) {
	frame := harnessStrip(shot.Frame(States()[3], shot.Width, shot.Height))
	if !strings.Contains(frame, "✗") {
		t.Error("a failing check is not visible without colour")
	}
}

// Marks are kept by repo and number, so a filter cannot move one.
func TestMarksSurviveAFilter(t *testing.T) {
	m := build("space", "j", "space", "/", "s", "p", "a", "r", "enter")().(*Model)
	if m.marks.Len() != 2 {
		t.Errorf("%d marks survived, want 2: %v", m.marks.Len(), m.marks.Keys())
	}
}

// Each section is its own query and its own list, and tab moves between them.
func TestTabMovesBetweenSections(t *testing.T) {
	m := build("tab")().(*Model)
	if m.section != 1 {
		t.Errorf("section is %d after tab", m.section)
	}
	if m.list.Cursor() != 0 {
		t.Error("the cursor did not reset when the rows changed under it")
	}
}
