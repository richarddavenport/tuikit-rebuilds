package cli

import (
	"testing"

	"github.com/richarddavenport/tuikit/guard"
)

// Every action reachable by mouse has a keyboard path.
//
// It lives here rather than beside the interface's other guards because it
// reads the command tree, and the tree is this package's — internal/tui imports
// nothing from internal/cli, which is what keeps the interface from knowing how
// it was launched.
func TestEveryActionHasAKeyboardPath(t *testing.T) {
	guard.Reachable(t, Commands())
}

// And nothing has taken a key that means something else in every other tuikit
// tool. ctrl+c, q, esc and ? mean one thing each, everywhere.
//
// The one rule tuikit imposes rather than offers, and the reason it is a rule
// is that all of its value is in being the same: a reader who has used another
// of these tools presses q expecting to leave, and a tool where q means "queue"
// has set a trap using the other tools' credibility. See spec.Reserved.
//
// A tool may still put a question in front of q — what LEAVING costs is the
// tool's business. It may not make q mean something that is not leaving.
func TestNothingTakesAReservedKey(t *testing.T) {
	guard.Reserved(t, Commands())
}
