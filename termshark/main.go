// Command termshark-rebuild draws termshark's interface on tuikit.
//
//	go run ./termshark          open it
//	go run ./termshark -shots   regenerate docs/screens.md
//
// It does not run tshark. See termshark/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
	"github.com/richarddavenport/tuikit-rebuilds/termshark/ui"
)

func main() {
	shot.Main("termshark", func() app.Model { return ui.New() }, ui.States())
}
