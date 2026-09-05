// Package fake is a filesystem that is not yours.
//
// yazi is Miller columns: parent, current, preview. It has no tree and no
// collapse state anywhere — the first pass of its study said it did, and
// checking the source is what corrected that.
package fake

import "time"

// At is the moment the filesystem is frozen at.
var At = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// Entry is one file or directory.
type Entry struct {
	Name  string
	Dir   bool
	Size  int64
	Mode  string
	Mod   time.Duration
	Link  bool
	Exec  bool
	Image bool
}

// Dir is a directory's contents, by path. Directories first, then names,
// because that is what a file manager shows and sorting it here keeps the
// interface free of an opinion about it.
func Dir(path string) []Entry {
	switch path {
	case "/Users/richardd/Developer":
		return []Entry{
			{Name: "richarddavenport", Dir: true, Mode: "drwxr-xr-x", Mod: 40 * time.Minute},
			{Name: "teaching", Dir: true, Mode: "drwxr-xr-x", Mod: 20 * time.Hour},
			{Name: "archive", Dir: true, Mode: "drwxr-xr-x", Mod: 30 * 24 * time.Hour},
			{Name: "BOARD.md", Size: 2_140, Mode: "-rw-r--r--", Mod: 2 * time.Hour},
			{Name: "NOW.md", Size: 18_902, Mode: "-rw-r--r--", Mod: 3 * time.Hour},
			{Name: "SESSIONS.md", Size: 6_104, Mode: "-rw-r--r--", Mod: 40 * time.Minute},
		}
	case "/Users/richardd/Developer/richarddavenport":
		return []Entry{
			{Name: "tuikit", Dir: true, Mode: "drwxr-xr-x", Mod: 35 * time.Minute},
			{Name: "tuikit-rebuilds", Dir: true, Mode: "drwxr-xr-x", Mod: 12 * time.Minute},
			{Name: "azctl", Dir: true, Mode: "drwxr-xr-x", Mod: 6 * 24 * time.Hour},
			{Name: "pgctl", Dir: true, Mode: "drwxr-xr-x", Mod: 9 * 24 * time.Hour},
			{Name: "swarmctl", Dir: true, Mode: "drwxr-xr-x", Mod: 11 * 24 * time.Hour},
			{Name: "docket", Dir: true, Mode: "drwxr-xr-x", Mod: 4 * 24 * time.Hour},
		}
	case "/Users/richardd/Developer/richarddavenport/tuikit":
		return []Entry{
			{Name: "app", Dir: true, Mode: "drwxr-xr-x", Mod: 3 * time.Hour},
			{Name: "comp", Dir: true, Mode: "drwxr-xr-x", Mod: 35 * time.Minute},
			{Name: "design", Dir: true, Mode: "drwxr-xr-x", Mod: time.Hour},
			{Name: "docs", Dir: true, Mode: "drwxr-xr-x", Mod: 5 * time.Hour},
			{Name: "gallery", Dir: true, Mode: "drwxr-xr-x", Mod: 40 * time.Minute},
			{Name: "guard", Dir: true, Mode: "drwxr-xr-x", Mod: 2 * time.Hour},
			{Name: "harness", Dir: true, Mode: "drwxr-xr-x", Mod: 20 * time.Hour},
			{Name: "spec", Dir: true, Mode: "drwxr-xr-x", Mod: 26 * time.Hour},
			{Name: "theme", Dir: true, Mode: "drwxr-xr-x", Mod: 50 * time.Minute},
			{Name: "go.mod", Size: 812, Mode: "-rw-r--r--", Mod: 30 * 24 * time.Hour},
			{Name: "LICENSE", Size: 1_063, Mode: "-rw-r--r--", Mod: 30 * 24 * time.Hour},
			{Name: "Makefile", Size: 1_744, Mode: "-rw-r--r--", Mod: 8 * 24 * time.Hour},
			{Name: "README.md", Size: 7_218, Mode: "-rw-r--r--", Mod: 4 * time.Hour},
			{Name: "logo.png", Size: 44_910, Mode: "-rw-r--r--", Mod: 20 * 24 * time.Hour, Image: true},
			{Name: "tuikit", Size: 14_882_000, Mode: "-rwxr-xr-x", Mod: 35 * time.Minute, Exec: true},
		}
	case "/Users/richardd/Developer/richarddavenport/tuikit/comp":
		return []Entry{
			{Name: "canvas.go", Size: 14_211, Mode: "-rw-r--r--", Mod: 26 * time.Hour},
			{Name: "focus.go", Size: 3_902, Mode: "-rw-r--r--", Mod: 40 * time.Minute},
			{Name: "list.go", Size: 21_884, Mode: "-rw-r--r--", Mod: time.Hour},
			{Name: "marks.go", Size: 5_140, Mode: "-rw-r--r--", Mod: 38 * time.Minute},
			{Name: "sparkline.go", Size: 4_402, Mode: "-rw-r--r--", Mod: 25 * time.Minute},
			{Name: "tree.go", Size: 5_009, Mode: "-rw-r--r--", Mod: 20 * time.Hour},
			{Name: "viewer.go", Size: 12_770, Mode: "-rw-r--r--", Mod: 6 * time.Hour},
		}
	}
	return nil
}

// Parent is the path above one, or "" at the root of the fixture.
func Parent(path string) string {
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' {
			if Dir(path[:i]) == nil {
				return ""
			}
			return path[:i]
		}
	}
	return ""
}

// Preview is what the third column shows for a file.
func Preview(name string) []string {
	switch name {
	case "README.md":
		return []string{
			"# tuikit",
			"",
			"### Build the tool, not the terminal.",
			"",
			"A Go framework for terminal applications that **show you state**",
			"**and let you act on it** — a git client, a file manager, a",
			"cluster browser, a system monitor, an API client, a deploy tool.",
			"",
			"Three things here that a widget library does not give you:",
			"",
			"**Nothing is invented.** Every one of the 26 components was",
			"pulled out of tools that had already written it.",
		}
	case "go.mod":
		return []string{
			"module github.com/richarddavenport/tuikit",
			"",
			"go 1.25.0",
			"",
			"require (",
			"\tgithub.com/charmbracelet/bubbletea v1.3.10",
			"\tgithub.com/charmbracelet/lipgloss v1.1.0",
			"\tgithub.com/charmbracelet/x/ansi v0.11.7",
			"\tgolang.org/x/sys v0.31.0",
			")",
		}
	case "Makefile":
		return []string{
			"## check: everything CI runs, before you push",
			"check: test lint",
			"\t@echo \"all checks passed\"",
		}
	case "LICENSE":
		return []string{"MIT License", "", "Copyright (c) 2026 Richard Davenport"}
	}
	return nil
}
