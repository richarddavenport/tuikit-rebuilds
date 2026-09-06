package engine

import (
	"testing"

	"github.com/richarddavenport/tuikit/guard"
)

// The engine has never heard of a terminal.
//
// No colour, no width, no keys, no framework. The split is worth as much as it
// is enforced, it is mechanically checkable, so it is checked — from the first
// commit, before there is anything to be tempted by.
func TestTheEngineHasNeverHeardOfATerminal(t *testing.T) {
	guard.Engine(t, ".")
}
