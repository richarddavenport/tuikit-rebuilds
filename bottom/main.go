// Command bottom-rebuild draws bottom's interface on tuikit.
//
//	go run ./bottom          open it
//	go run ./bottom -shots   regenerate docs/screens.md
//
// It does not read this machine. See bottom/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/bottom/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	shot.Main("bottom", func() app.Model { return ui.New() }, ui.States())
}
