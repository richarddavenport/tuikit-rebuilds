// Command htop-rebuild draws htop's interface on tuikit.
//
//	go run ./htop          open it
//	go run ./htop -shots   regenerate docs/screens.md
//
// It does not read this machine. See htop/fake.
package main

import (
	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/htop/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	shot.Main("htop", func() app.Model { return ui.New() }, ui.States())
}
