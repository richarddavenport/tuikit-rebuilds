package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Miller columns: parent, current, preview. `comp.Layout.Cols` and nothing more. **No tree** — yazi has none, which is what reading its source corrected.",
			Build: build()},
		{Name: "file-preview", Note: "The third column previews a file. A `comp.Viewer`, so it scrolls and does not tail.",
			Build: build("j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")},
		{Name: "into", Note: "`l` descends. The cursor is remembered per directory, so `h` and `l` feel like moving rather than teleporting.",
			Build: build("j", "l")},
		{Name: "back-up", Note: "`h` comes back and lands on the directory you came out of — not at the top.",
			Build: build("j", "l", "h")},
		{Name: "marked", Note: "`space` on three. `comp.Marks` keys by path, so the selection survives leaving the directory and coming back.",
			Build: build("space", "j", "space", "j", "j", "space")},
		{Name: "visual", Note: "`v` starts a dragged range. This is yazi's gesture: the range is primary and commits into the set. k9s does the opposite, and `comp.Marks` offers both.",
			Build: build("v", "j", "j", "j")},
		{Name: "visual-committed", Note: "`v` again commits it. The range is gone and four paths are marked.",
			Build: build("v", "j", "j", "j", "v")},
		{Name: "marks-survive", Note: "Marked, then down a directory and back. Kept by identity, so nothing moved.",
			Build: build("space", "j", "space", "j", "l", "h")},
		{Name: "searching", Note: "`/` captures the keyboard.",
			Build: build("/", "g", "o")},
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
