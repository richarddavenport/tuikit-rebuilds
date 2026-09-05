// Command gh-dash-rebuild draws gh-dash's interface on tuikit.
//
//	go run ./gh-dash          open it
//	go run ./gh-dash -shots   regenerate docs/screens.md
//
// It does not talk to GitHub. See gh-dash/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/gh-dash/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	shot.Main("gh-dash", func() app.Model { return ui.New() }, ui.States())
}
