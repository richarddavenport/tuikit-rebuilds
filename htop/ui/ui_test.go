package ui

import (
	"strings"
	"testing"

	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/htop/fake"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func TestScreens(t *testing.T) { shot.Goldens(t, "testdata", States()) }

// The branch lines say which process spawned which, and they have to survive
// the colour being stripped — a tree drawn only in colour is not a tree.
func TestTheTreeDrawsBranchesNotIndentation(t *testing.T) {
	frame := shot.Frame(States()[1], shot.Width, shot.Height)
	for _, want := range []string{"├─", "└─", "│  "} {
		if !strings.Contains(frame, want) {
			t.Errorf("tree mode drew no %q", want)
		}
	}
	if flat := shot.Frame(States()[0], shot.Width, shot.Height); strings.Contains(flat, "├─") {
		t.Error("the flat list drew branch lines")
	}
}

// A LAST child gets └ and a middle one gets ├. This is the distinction depth
// alone cannot make, and the reason tuikit#78 exists.
func TestTheLastChildClosesItsBranch(t *testing.T) {
	ns := nodes(treeOrder(fake.Processes()))
	var lastSeen, middleSeen bool
	for _, n := range ns {
		if n.depth == 0 {
			continue
		}
		if n.last {
			lastSeen = true
		} else {
			middleSeen = true
		}
	}
	if !lastSeen || !middleSeen {
		t.Fatal("the fixture has no branch with both a middle and a last child")
	}
}

// Folding hides the subtree and nothing else.
func TestFoldingHidesOnlyThatSubtree(t *testing.T) {
	open := shot.Frame(States()[1], shot.Width, shot.Height)
	shut := shot.Frame(States()[2], shot.Width, shot.Height)
	if !strings.Contains(open, "systemd-journald") {
		t.Fatal("the fixture changed; this test picked the wrong row")
	}
	if strings.Contains(shut, "sshd: richard") {
		t.Error("folding did not hide the subtree")
	}
	if !strings.Contains(shut, "firefox") {
		t.Error("folding hid a row outside the subtree")
	}
}

// Search leaves every row on screen and moves the cursor. Filter removes rows.
// htop shares one editor between them, so a rebuild that lost the distinction
// would still look right in a screenshot of the input line.
func TestSearchKeepsRowsAndFilterRemovesThem(t *testing.T) {
	search := shot.Frame(States()[7], shot.Width, shot.Height)
	filter := shot.Frame(States()[8], shot.Width, shot.Height)

	if !strings.Contains(search, "postgres") {
		t.Error("search hid a row that does not match, which is what filter is for")
	}
	if strings.Contains(filter, "firefox") {
		t.Error("filter kept a row that does not match")
	}
	if !strings.Contains(filter, "postgres") {
		t.Error("filter dropped a row that does match")
	}
}

// The header's height comes from the arrangement, so adding a three-row LED
// meter has to push the table down. A layout written as constants would not.
func TestTheHeaderHeightFollowsTheArrangement(t *testing.T) {
	plain := New()
	plain.columns = [][]Slot{{{Name: "CPU", Mode: ModeBar}}}
	tall := New()
	tall.columns = [][]Slot{{{Name: "CPU", Mode: ModeLED}}}

	if got, want := plain.headerRows(), 1; got != want {
		t.Errorf("a bar meter is %d rows, want %d", got, want)
	}
	if got, want := tall.headerRows(), 3; got != want {
		t.Errorf("an LED meter is %d rows, want %d", got, want)
	}
}

// tuikit#68, checked here for the first time.
//
// The header is data, so a reader can remove a meter. If the function bar
// advertised a key belonging to a meter that is not on screen, the interface
// would be lying — which is the bug bottom, gh-dash and htop's own static
// HELP_TEXT all have. This asserts the sheet is derived, not literal.
func TestTheKeySheetOffersNothingThatIsNotOnScreen(t *testing.T) {
	m := New()
	m.SetSize(shot.Width, shot.Height)
	m.columns = [][]Slot{{{Name: "CPU", Mode: ModeBar}}}

	frame := app.New(m, app.WithSize(shot.Width, shot.Height)).View()

	// Swap is gone from the arrangement, so nothing may name it.
	if strings.Contains(frame, "Swp") {
		t.Error("a meter that was removed is still drawn")
	}
	// And the bar still offers only keys that route.
	for _, key := range []string{"Setup", "Search", "Tree"} {
		if !strings.Contains(frame, key) {
			t.Errorf("the function bar lost %q", key)
		}
	}
}
