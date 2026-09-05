// Command gitui-rebuild draws gitui's interface on tuikit.
//
//	go run ./gitui          open it
//	go run ./gitui -shots   regenerate docs/screens.md
//
// It does not talk to git. See gitui/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/gitui/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	shot.Main("gitui", func() app.Model { return ui.New() }, ui.States())
}
