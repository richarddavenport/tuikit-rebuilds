package engine

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// run is how the engine reaches gcloud and kubectl.
//
// A package variable so a test can answer without a cloud account, and so the
// fixture that the screens are rendered from is the same code path as the real
// thing rather than a second one that can drift.
//
// Not an interface with one method, and not a struct threaded through every
// call: both make every signature carry a dependency that exactly one line of
// each function uses.
var run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	// Nothing here is allowed to ask the person a question, and a gcloud that
	// decides to prompt would hang a TUI with no way to answer it.
	cmd.Env = append(cmd.Environ(), "CLOUDSDK_CORE_DISABLE_PROMPTS=1")
	out, err := cmd.Output()
	if err != nil {
		return nil, describeExitError(name, args, err)
	}
	return out, nil
}

// describeExitError turns an *exec.ExitError into something a person can act
// on.
//
// The default is "exit status 1", which says nothing. gcloud and kubectl both
// write a real explanation to stderr and then exit non-zero, so the useful
// message is already there and merely thrown away.
func describeExitError(name string, args []string, err error) error {
	var ee *exec.ExitError
	if !asExitError(err, &ee) {
		if _, lookErr := exec.LookPath(name); lookErr != nil {
			return fmt.Errorf("%s is not installed or not on PATH", name)
		}
		return fmt.Errorf("running %s: %w", name, err)
	}
	msg := strings.TrimSpace(string(ee.Stderr))
	if msg == "" {
		return fmt.Errorf("%s %s exited %d", name, strings.Join(args, " "), ee.ExitCode())
	}
	// The last line is the one that says what went wrong; the ones above it are
	// usually a stack of API noise.
	lines := strings.Split(msg, "\n")
	return fmt.Errorf("%s: %s", name, strings.TrimSpace(lines[len(lines)-1]))
}

// clock is time.Now, replaceable so a fixture's ages do not change between
// runs. A pod's age is drawn on the screen, so a moving clock is a golden that
// fails every minute.
var clock = time.Now

// SetRunner replaces how the engine reaches the outside world, and returns a
// function that puts the old one back.
//
// Exported for tests and for the fixture. It is the only seam, which is why it
// is worth having exactly one.
func SetRunner(r func(ctx context.Context, name string, args ...string) ([]byte, error)) func() {
	prev := run
	run = r
	return func() { run = prev }
}

// SetClock does the same for time, so a fixture's ages hold still.
func SetClock(now func() time.Time) func() {
	prev := clock
	clock = now
	return func() { clock = prev }
}

// Now is the engine's clock, so a snapshot and the ages inside it agree.
//
// Exported because the interface stamps a snapshot with it, and a second call
// to time.Now there would put the fetch a few microseconds after the pod ages
// it is describing.
func Now() time.Time { return clock() }
