// Command lazygit-rebuild draws lazygit's interface on tuikit.
//
//	go run ./lazygit          open it
//	go run ./lazygit -shots   regenerate docs/screens.md
//
// It does not talk to git. See lazygit/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
	"github.com/richarddavenport/tuikit-rebuilds/lazygit/ui"
)

func main() {
	shot.Main("lazygit", func() app.Model { return ui.New() }, ui.States())
}
