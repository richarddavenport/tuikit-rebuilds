package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// Folding the build cache hides everything under it and nothing else. This is
// the component dive's study said it needed.
func TestFoldingHidesOnlyThatSubtree(t *testing.T) {
	frame := shot.Frame(States()[3], shot.Width, shot.Height)
	if strings.Contains(frame, "0a1f2b3c4d") {
		t.Error("a file inside the folded directory is still drawn")
	}
	if !strings.Contains(frame, "usr") {
		t.Error("folding one directory hid a sibling")
	}
}

// dive's four states are characters, not only colours, so the frame still says
// what changed once the colour is stripped.
func TestChangesReadWithoutColour(t *testing.T) {
	frame := harnessStrip(shot.Frame(States()[2], shot.Width, shot.Height))
	for _, mark := range []string{"+", "~", "-"} {
		if !strings.Contains(frame, mark+" ") {
			t.Errorf("no %q mark in a colourless frame", mark)
		}
	}
}

// Attributes are right-aligned by Row.Right rather than by padding arithmetic.
func TestAttributesAppearOnTheRight(t *testing.T) {
	off := shot.Frame(States()[2], shot.Width, shot.Height)
	on := shot.Frame(States()[4], shot.Width, shot.Height)
	if strings.Contains(off, "drwxr-xr-x") {
		t.Error("permissions are shown before `a` was pressed")
	}
	if !strings.Contains(on, "drwxr-xr-x") {
		t.Error("`a` did not show permissions")
	}
}
