// Command fx-rebuild draws fx's interface on tuikit.
//
//	go run ./fx          open it
//	go run ./fx -shots   regenerate docs/screens.md
//
// It does not parse JSON. See fx/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/fx/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	shot.Main("fx", func() app.Model { return ui.New() }, ui.States())
}
