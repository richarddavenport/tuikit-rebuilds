// Command dive-rebuild draws dive's interface on tuikit.
//
//	go run ./dive          open it
//	go run ./dive -shots   regenerate docs/screens.md
//
// It does not read docker images. See dive/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/dive/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	shot.Main("dive", func() app.Model { return ui.New() }, ui.States())
}
