package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Layers left, the filesystem they produced right. The `!` marks a layer whose bytes a later one throws away.",
			Build: build()},
		{Name: "wasted-layer", Note: "`go mod download` adds 64 MB and `rm -rf` deletes it. `comp.Detail` says so in facts.",
			Build: build("j", "j", "j", "j")},
		{Name: "tree-focused", Note: "`tab`. `comp.Focus` holds which pane has the keyboard; the cursor is a different CHARACTER, not just a different colour.",
			Build: build("tab")},
		{Name: "folded", Note: "`space` on a directory. `comp.Tree` decides which rows exist; the ▸ is the tool's.",
			Build: build("tab", "j", "j", " ")},
		{Name: "attributes", Note: "`a` adds permissions and sizes on the right. `Row.Right` right-aligns them without anyone doing padding arithmetic.",
			Build: build("tab", "a")},
		{Name: "deep", Note: "Inside the build cache, which is what dive exists to find.",
			Build: build("tab", "j", "j", "j", "j", "j", "a")},
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
