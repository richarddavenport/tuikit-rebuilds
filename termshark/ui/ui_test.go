package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// termshark's whole trick: selecting a field highlights its bytes. The range is
// over LINES, and turning a byte span into a line span is the tool's
// arithmetic, because only the tool knows how many bytes a row holds.
func TestSelectingAFieldHighlightsItsBytes(t *testing.T) {
	m := build("tab", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")().(*Model)
	f, ok := m.field()
	if !ok || f.Len == 0 {
		t.Fatalf("no field with bytes under the cursor: %+v", f)
	}
	lo, hi, has := m.hex.Range()
	if !has {
		t.Fatal("no range in the hex dump")
	}
	wantLo, wantHi := f.Off/bytesPerRow, (f.Off+f.Len-1)/bytesPerRow
	if lo != wantLo || hi != wantHi {
		t.Errorf("range is %d..%d, want %d..%d for bytes %d+%d", lo, hi, wantLo, wantHi, f.Off, f.Len)
	}
}

// A field far into the packet scrolls the dump to it.
func TestAFieldDeepInThePacketScrollsTheDump(t *testing.T) {
	m := build("tab", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")().(*Model)
	if m.hex.Offset() == 0 {
		t.Error("the dump is still at the top for a field 200 bytes in")
	}
	frame := shot.Frame(States()[2], shot.Width, shot.Height)
	if !strings.Contains(frame, "api.acme.internal") {
		t.Error("the SNI bytes are not on screen")
	}
}

// Three panes, one focused, moved by region name.
func TestFocusMovesThroughThreePanes(t *testing.T) {
	for i, want := range []string{"fields", "hex", "packets"} {
		keys := make([]string, i+1)
		for j := range keys {
			keys[j] = "tab"
		}
		m := build(keys...)().(*Model)
		if got := string(m.focus.Current()); got != want {
			t.Errorf("%d tabs: focus is %q, want %q", i+1, got, want)
		}
	}
}

// Marks are kept by packet number, so a display filter cannot move one.
func TestMarksSurviveAFilter(t *testing.T) {
	m := build("m", "j", "j", "j", "m", "/", "t", "l", "s", "enter")().(*Model)
	if m.marks.Len() != 2 {
		t.Errorf("%d marks survived the filter, want 2", m.marks.Len())
	}
}
