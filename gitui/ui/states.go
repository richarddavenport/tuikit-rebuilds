package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "status", Note: "Unstaged and staged on the left, the diff on the right. `comp.Tabs` carries a count on the tab it belongs to.",
			Build: build()},
		{Name: "staged-focused", Note: "`shift+tab`. `comp.Focus` moves between the two lists by region name.",
			Build: build("shift+tab")},
		{Name: "line-select", Note: "`v`. The diff is a `comp.Viewer` — it opens at the top and follows nothing.",
			Build: build("v")},
		{Name: "range", Note: "`J` three times. The range is a background only, so every line keeps its added or removed colour.",
			Build: build("v", "J", "J", "J")},
		{Name: "whole-hunk", Note: "`a` snaps to the hunk boundaries.",
			Build: build("v", "j", "j", "a")},
		{Name: "palette", Note: "`space`. gitui has 32 popup files; this is ten of them in a `comp.Palette`.",
			Build: build(" ")},
		{Name: "stacked", Note: "**The gap this rebuild found.** A popup opened FROM a popup. The palette is still underneath and `esc` comes back to it — `app.Stack` is about screens and has nothing to say about this.",
			Build: build(" ", "enter")},
		{Name: "help", Note: "`?` from the screen. A popup CAPTURES every key — `app.Keys` says a capture takes the key whether or not it does anything with it — so `?` inside a confirm closes the confirm rather than stacking help on it. That is the routing working, not a bug.",
			Build: build("?")},
		{Name: "log", Note: "`2`. `Row.Right` puts the author and age against the edge without anyone doing padding arithmetic.",
			Build: build("2")},
		{Name: "stashing", Note: "`4`.", Build: build("4")},
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
