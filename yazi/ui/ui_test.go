package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// yazi has no tree. Its study first said it did, and reading the source
// corrected that — this rebuild is Miller columns and nothing else.
func TestThereIsNoTree(t *testing.T) {
	frame := shot.Frame(States()[0], shot.Width, shot.Height)
	for _, marker := range []string{"▸", "▾"} {
		if strings.Contains(frame, marker) {
			t.Errorf("a fold marker %q appeared — this is Miller columns", marker)
		}
	}
}

// Going down and coming back lands where you left, which is what makes h and l
// feel like moving.
func TestComingBackLandsWhereYouLeft(t *testing.T) {
	m := build("j", "l", "h")().(*Model)
	if m.cwd != Root {
		t.Fatalf("cwd is %q after h", m.cwd)
	}
	e, ok := m.hovered()
	if !ok || e.Name != "comp" {
		t.Errorf("landed on %q, want the directory we came out of", e.Name)
	}
}

// The marks are a SET kept by path, so leaving a directory and coming back
// keeps them.
func TestMarksSurviveLeavingTheDirectory(t *testing.T) {
	m := build("space", "j", "space", "j", "l", "h")().(*Model)
	if m.marks.Len() != 2 {
		t.Errorf("%d marks survived, want 2: %v", m.marks.Len(), m.marks.Keys())
	}
	for _, k := range m.marks.Keys() {
		if !strings.HasPrefix(k, Root+"/") {
			t.Errorf("a mark is %q — it should still be the original path", k)
		}
	}
}

// yazi's gesture: drag a range, then commit it into the set.
func TestVisualCommitsIntoTheSet(t *testing.T) {
	dragging := build("v", "j", "j", "j")().(*Model)
	if dragging.marks.Len() != 0 {
		t.Errorf("dragging marked %d before committing", dragging.marks.Len())
	}
	done := build("v", "j", "j", "j", "v")().(*Model)
	if done.marks.Len() != 4 {
		t.Errorf("committing marked %d, want 4: %v", done.marks.Len(), done.marks.Keys())
	}
	if done.visual {
		t.Error("still in visual mode after committing")
	}
}
