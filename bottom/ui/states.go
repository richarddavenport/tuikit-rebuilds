package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Six widgets. `comp.Layout` nests — rows of columns of rows — which is bottom's grid without a constraint solver.",
			Build: build()},
		{Name: "charts", Note: "Every `comp.Sparkline` on screen. The newest sample is the LAST column, so the CPU-0 spike sits where now is.",
			Build: build()},
		{Name: "sorted-cpu", Note: "`s` to CPU%. `comp.Sort` hands back indices; the comparison stays the tool's.",
			Build: build("s", "s", "s", "s")},
		{Name: "sorted-desc", Note: "`S` reverses. Stable, so the two processes at 0% keep the order the fixture built them in.",
			Build: build("s", "s", "s", "s", "S")},
		{Name: "filtering", Note: "`/` captures the keyboard. `j` is the letter j while it does.",
			Build: build("/", "g", "o")},
		{Name: "disks-focused", Note: "`tab`. `comp.Focus` moves the keyboard by region NAME, so a widget added to the ring cannot silently steal it.",
			Build: build("tab")},
		{Name: "expanded", Note: "`e` gives the focused table the whole screen. The same draw function, a bigger rect.",
			Build: build("tab", "e")},
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
