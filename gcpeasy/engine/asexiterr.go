package engine

import (
	"errors"
	"os/exec"
)

// asExitError is errors.As, named so the one call site reads as a question.
func asExitError(err error, target **exec.ExitError) bool { return errors.As(err, target) }
