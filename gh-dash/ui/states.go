package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Sections as `comp.Tabs` with a count each, a table of pull requests, and a sidebar. CI's answer is a CHARACTER in the first column, so it survives the color being stripped.",
			Build: build()},
		{Name: "body", Note: "The sidebar shows the description in a `comp.Viewer` — the one component gh-dash's study said it was missing.",
			Build: build()},
		{Name: "checks", Note: "`c` swaps the body for the CI runs. `comp.StepList`: what will happen, what has, and what it cost.",
			Build: build("c")},
		{Name: "failing-checks", Note: "A section where CI is red. The glyphs are the tool's, not the component's.",
			Build: build("tab", "c")},
		{Name: "next-section", Note: "`tab`. Each section is its own query; gh-dash's are defined in YAML, which is the part tuikit cannot do.",
			Build: build("tab")},
		{Name: "sorted", Note: "`s` cycles the column. `comp.Sort` hands back indices; the comparison stays the tool's.",
			Build: build("s", "s", "s", "s", "s")},
		{Name: "marked", Note: "`space` on two. `comp.Marks`, keyed by repo and number.",
			Build: build("space", "j", "space")},
		{Name: "filtering", Note: "`/` captures the keyboard.",
			Build: build("/", "s", "p", "a", "r")},
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
