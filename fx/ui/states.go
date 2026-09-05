package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "One `comp.Viewer`. Each line is spans, so the cursor's background goes under the syntax colours rather than over them.",
			Build: build()},
		{Name: "folded", Note: "`space` on a container. `comp.Tree` holds the state; the ▸ and the `{ … }` preview are the tool's.",
			Build: build("j", "j", "j", "j", "j", "j", "j", " ")},
		{Name: "all-folded", Note: "`E`. Every container shut, which is how you read a lock file.",
			Build: build("E")},
		{Name: "searching", Note: "`/` captures the keyboard. While it does, `j` is the letter j — `app.Keys` makes that the shape of the routing.",
			Build: build("/", "v", "e", "r")},
		{Name: "found", Note: "Enter. The cursor is on the first match and the status line counts them.",
			Build: build("/", "v", "e", "r", "enter")},
		{Name: "deep", Note: "Scrolled into the dependencies. `comp.Viewer` numbers lines from the document, so the gutter does not change width as you go.",
			Build: build("j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")},
		{Name: "help", Note: "`?`.",
			Build: build("?")},
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
