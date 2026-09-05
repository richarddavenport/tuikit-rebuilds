// Command k9s-rebuild draws k9s's interface on tuikit.
//
//	go run ./k9s          open it
//	go run ./k9s -shots   regenerate docs/screens.md
//
// It does not talk to Kubernetes. See k9s/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
	"github.com/richarddavenport/tuikit-rebuilds/k9s/ui"
)

func main() {
	shot.Main("k9s", func() app.Model { return ui.New() }, ui.States())
}
