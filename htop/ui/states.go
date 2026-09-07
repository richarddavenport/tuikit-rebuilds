package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "The default arrangement: meters in two columns, then the process table. The header's height is computed from the arrangement, not written down.",
			Build: build()},
		{Name: "tree", Note: "`F5`. The branch lines are the tool's — `comp.Tree` says which rows exist and `Row.Depth` indents, but nothing in `comp` draws ├ and │. See tuikit#78.",
			Build: build("f5")},
		{Name: "folded", Note: "`space` on a subtree. `comp.Tree` keys collapse state by PID, so a filter that removes a row above cannot fold a different process.",
			Build: build("f5", "j", "j", " ")},
		{Name: "setup", Note: "`F2`, and the reason htop is here. The left pane is the header being edited; the right is every meter that exists. The list is closed, which is what keeps the guards working.",
			Build: build("f2")},
		{Name: "setup-modes", Note: "`space` cycles a meter's mode. htop's four — Bar, Text, Graph, LED — are a per-meter choice a reader saves, not a decision the drawing makes.",
			Build: build("f2", " ", "j", " ", " ")},
		{Name: "meter-modes", Note: "The same three numbers drawn four ways at once. Graph is `comp.Sparkline`; the LED digits are htop's own and nothing should supply them.",
			Build: buildModes()},
		{Name: "four-columns", Note: "`L` in Setup. Thirteen named splits in htop's `HeaderLayout.h` are thirteen sets of `comp.Percent` constraints.",
			Build: buildLayout()},
		{Name: "search", Note: "`F3`. The cursor jumps to a match and every row stays on screen.",
			Build: build("f3", "f", "i", "r", "e")},
		{Name: "filter", Note: "`F4`, the same editor and a different meaning: rows that do not match are gone. htop shares one `IncSet` between the two.",
			Build: build("f4", "p", "o", "s", "t")},
		{Name: "sorted", Note: "`F6` then enter. `comp.Sort` holds the column and the direction and marks the header.",
			Build: build("f6", "j", "j", "j", "enter")},
		{Name: "signal", Note: "`F9`. The same picker as F6 with different contents, which is what htop's `Panel` is for.",
			Build: build("f9")},
		{Name: "help", Note: "`F1`.", Build: build("f1")},
	}
}

func build(keys ...string) func() app.Model {
	return func() app.Model {
		m := New()
		m.SetSize(shot.Width, shot.Height)
		harness.Press(app.New(m, app.WithSize(shot.Width, shot.Height)), keys...)
		return m
	}
}

// buildModes puts one meter in each mode, so the four are comparable in a
// single frame rather than across four screenshots.
func buildModes() func() app.Model {
	return func() app.Model {
		m := New()
		m.SetSize(shot.Width, shot.Height)
		m.columns = [][]Slot{
			{{Name: "CPU", Mode: ModeBar}, {Name: "Memory", Mode: ModeText}, {Name: "Swap", Mode: ModeGraph}},
			{{Name: "Load average", Mode: ModeLED}, {Name: "Tasks", Mode: ModeText}},
		}
		return m
	}
}

// buildLayout shows the four-column split, which is the arrangement that makes
// the header's computed height obvious.
func buildLayout() func() app.Model {
	return func() app.Model {
		m := New()
		m.SetSize(shot.Width, shot.Height)
		m.columns = [][]Slot{
			{{Name: "CPU", Mode: ModeBar}},
			{{Name: "Memory", Mode: ModeBar}, {Name: "Swap", Mode: ModeBar}},
			{{Name: "Tasks", Mode: ModeText}, {Name: "Load average", Mode: ModeText}},
			{{Name: "Uptime", Mode: ModeText}, {Name: "Disk IO", Mode: ModeText}},
		}
		m.layout = 5
		return m
	}
}
