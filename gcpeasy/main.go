// Command gcpeasy-rebuild draws gcpeasy's interface on tuikit.
//
//	go run ./gcpeasy            open it, against real GCP
//	go run ./gcpeasy -fixture   open it, against canned data
//	go run ./gcpeasy -shots     regenerate docs/screens.md
//
// The one rebuild here that talks to a real backend. Everything else in this
// repository draws against a fixture, which is the right shape for asking
// "could tuikit draw this". gcpeasy asks the other question — what a tool built
// on tuikit from the start actually looks like — and that needs gcloud.
package main

import (
	"os"

	"github.com/richarddavenport/tuikit/app"

	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/engine"
	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/ui"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-fixture" {
		defer engine.Fixture()()
	}
	shot.Main("gcpeasy", func() app.Model { return ui.New() }, ui.States())
}
