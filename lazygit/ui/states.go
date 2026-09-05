package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
//
// ONE list, walked by the goldens, the narrow-terminal run, the colour check
// and the capture. A screen added without a frame is a screen added without any
// of them, which is the only arrangement in which they cannot drift apart.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Four panels and a diff. The file panel is a `comp.List` whose rows a `comp.Tree` decides.",
			Build: build()},
		{Name: "folded", Note: "`space` on a directory. `comp.Tree` holds the collapse state, keyed by path rather than by index, and draws nothing.",
			Build: build("space")},
		{Name: "line-select", Note: "`v` enters line select. The diff is a `comp.Viewer`, which opens at the top and does not tail.",
			Build: build("v")},
		{Name: "range", Note: "`J` extends. The range is a background only, so every line keeps its own added or removed colour.",
			Build: build("v", "J", "J", "J")},
		{Name: "whole-hunk", Note: "`a` snaps the range to the hunk. lazygit carries 13 kB of state for this.",
			Build: build("v", "j", "j", "j", "j", "a")},
		{Name: "commits", Note: "`4` moves the keyboard. Only the focused panel's cursor is drawn bright.",
			Build: build("4", "j")},
		{Name: "help", Note: "`?` is `comp.Keys`, built from the same list the footer is.",
			Build: build("?")},
	}
}

// build returns a model with the keys already pressed, redrawing between each
// the way a running program does.
func build(keys ...string) func() app.Model {
	return func() app.Model {
		m := New()
		m.SetSize(shot.Width, shot.Height)
		harness.Press(app.New(m, app.WithSize(shot.Width, shot.Height)), keys...)
		return m
	}
}
