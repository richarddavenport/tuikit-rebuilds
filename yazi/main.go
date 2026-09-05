// Command yazi-rebuild draws yazi's interface on tuikit.
//
//	go run ./yazi          open it
//	go run ./yazi -shots   regenerate docs/screens.md
//
// It does not touch your filesystem. See yazi/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
	"github.com/richarddavenport/tuikit-rebuilds/yazi/ui"
)

func main() {
	shot.Main("yazi", func() app.Model { return ui.New() }, ui.States())
}
