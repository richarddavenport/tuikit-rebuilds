package ui

import "github.com/richarddavenport/tuikit/harness"

// harnessStrip is harness.Strip, aliased so a test reads as what it is asking.
func harnessStrip(frame string) string { return harness.Strip(frame) }
