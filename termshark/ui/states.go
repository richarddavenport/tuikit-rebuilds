package ui

import (
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing this rebuild can show.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Three panes. The selected field's bytes are highlighted in the dump — a `comp.Viewer` range, which is a background, so the hex digits keep their color.",
			Build: build()},
		{Name: "field-selected", Note: "`tab` to the tree, then down to the source address. The highlight in the dump follows. That link is termshark's whole trick.",
			Build: build("tab", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")},
		{Name: "sni", Note: "The TLS server name, 200 bytes in. `Viewer.Goto` scrolls the dump to it.",
			Build: build("tab", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")},
		{Name: "folded", Note: "`space` on a protocol. `comp.Tree` decides which rows exist.",
			Build: build("tab", " ")},
		{Name: "marked", Note: "`m` on two packets. `comp.Marks` is not a field on `List` — termshark has a copy mode over the table AND over the tree.",
			Build: build("m", "j", "j", "j", "m")},
		{Name: "filled", Note: "`M` fills from the nearest mark to the cursor.",
			Build: build("m", "j", "j", "j", "j", "M")},
		{Name: "filtering", Note: "`/` captures the keyboard.",
			Build: build("/", "t", "l", "s")},
		{Name: "hex-focused", Note: "`tab` twice more. `comp.Focus` moves the keyboard through three panes by region name.",
			Build: build("tab", "tab")},
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
