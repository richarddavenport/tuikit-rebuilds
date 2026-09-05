package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "One table. `comp.Table` lays the columns out and `comp.List` scrolls and selects them — two components composed, which is why `Table` has no cursor.",
			Build: build()},
		{Name: "marked", Note: "`space` on three rows. `comp.Marks` keys by namespace/name, so a filter cannot move a mark to a different pod.",
			Build: build("space", "j", "space", "j", "j", "space")},
		{Name: "filled", Note: "`ctrl+space` fills from the nearest mark to the cursor. No anchor: the set is primary and the range is derived. k9s's gesture.",
			Build: build("space", "j", "j", "j", "j", "ctrl+@")},
		{Name: "sorted", Note: "`s` cycles the sort column; the arrow comes from `theme.Chrome`. The comparison stays the tool's — ordering an age against a byte count is a domain question.",
			Build: build("s", "s", "s", "s", "s")},
		{Name: "sorted-desc", Note: "`S` reverses. `comp.Sort` is stable, so rows that tie keep the order the fixture built them in.",
			Build: build("s", "s", "s", "s", "s", "S")},
		{Name: "filtering", Note: "`/` captures the keyboard. While it does, `j` is the letter j.",
			Build: build("/", "j", "o", "b", "s")},
		{Name: "filtered", Note: "Enter. Marks made before the filter survive it, because they are kept by identity.",
			Build: build("space", "j", "space", "/", "j", "o", "b", "s", "enter")},
		{Name: "describe", Note: "`d` opens `kubectl describe` in a `comp.Viewer`. Not a `LogPane`: describe output opens at the top and follows nothing.",
			Build: build("d")},
		{Name: "delete", Note: "`ctrl+d`. `Marks.Acting` returns the marks, or the cursor when there are none — one code path for both.",
			Build: build("space", "j", "space", "ctrl+d")},
		{Name: "delete-one", Note: "The same key with nothing marked. This is the `if` every call site would otherwise write.",
			Build: build("j", "j", "ctrl+d")},
		{Name: "help", Note: "`?`.", Build: build("?")},
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
