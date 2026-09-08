// Package fake is a git repository that does not exist.
//
// gitui is lazygit's control: same subject, different team, different
// substrate. This fixture is deliberately DIFFERENT from lazygit's, so the two
// rebuilds do not look like the same program twice.
package fake

import "time"

// At is the moment the repository is frozen at.
var At = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// Tab is one of gitui's top-level tabs.
type Tab int

// The tabs, in the order gitui numbers them.
const (
	Status Tab = iota
	Log
	Files
	Stashing
	TabCount
)

// Name is what the tab strip shows.
func (t Tab) Name() string {
	return [...]string{"Status", "Log", "Files", "Stashing"}[t]
}

// Change is one path in the working tree.
type Change struct {
	Path   string
	Status rune
	Staged bool
}

// Changes is the working tree.
func Changes() []Change {
	return []Change{
		{Path: "src/components/diff.rs", Status: 'M', Staged: true},
		{Path: "src/components/syntax_text.rs", Status: 'M', Staged: true},
		{Path: "src/popups/goto_line.rs", Status: 'A', Staged: true},
		{Path: "src/app.rs", Status: 'M'},
		{Path: "src/keys/key_config.rs", Status: 'M'},
		{Path: "asyncgit/src/sync/diff.rs", Status: 'M'},
		{Path: "CHANGELOG.md", Status: 'M'},
		{Path: "src/tabs/revlog.rs", Status: 'D'},
	}
}

// Commit is one entry in the log.
type Commit struct {
	SHA     string
	Subject string
	Author  string
	When    time.Duration
	// Refs are branch and tag names pointing here.
	Refs []string
}

// Commits is the history.
func Commits() []Commit {
	return []Commit{
		{SHA: "a4f19c2", Subject: "popups: a modal stack, not a screen stack", Author: "extrawurst", When: 40 * time.Minute, Refs: []string{"HEAD -> master"}},
		{SHA: "77b0e51", Subject: "diff: keep syntax color under the selection", Author: "extrawurst", When: 5 * time.Hour},
		{SHA: "1d9a3f8", Subject: "status_tree: collapse by path", Author: "cruessler", When: 27 * time.Hour},
		{SHA: "9e2c7b4", Subject: "asyncgit: cancel a diff that nobody is waiting for", Author: "extrawurst", When: 2 * 24 * time.Hour, Refs: []string{"v0.27.0"}},
		{SHA: "5c81d0a", Subject: "textinput: undo stack", Author: "cruessler", When: 4 * 24 * time.Hour},
		{SHA: "b2704ef", Subject: "keys: make the config file the source of truth", Author: "extrawurst", When: 6 * 24 * time.Hour},
	}
}

// Stash is one stashed change.
type Stash struct {
	SHA     string
	Subject string
	When    time.Duration
}

// Stashes is the stash list.
func Stashes() []Stash {
	return []Stash{
		{SHA: "e91b3c7", Subject: "On master: half a popup stack", When: 3 * time.Hour},
		{SHA: "44a0d18", Subject: "On master: the layout experiment", When: 30 * time.Hour},
	}
}

// DiffKind is what a diff line is.
type DiffKind int

// The kinds.
const (
	Context DiffKind = iota
	Added
	Removed
	Hunk
	Header
)

// DiffLine is one line, already classified.
type DiffLine struct {
	Text string
	Kind DiffKind
	Hunk int
}

// Diff is the change to the file the cursor is on.
func Diff() []DiffLine {
	raw := []struct {
		k DiffKind
		t string
	}{
		{Header, "diff --git a/src/components/diff.rs b/src/components/diff.rs"},
		{Header, "index 7c1a4f2..2b90e13 100644"},
		{Hunk, "@@ -204,11 +204,9 @@ impl DiffComponent {"},
		{Context, "     fn get_selection(&self) -> Selection {"},
		{Removed, "-        if let Some(start) = self.selection_start {"},
		{Removed, "-            Selection::Multiple(start, self.selected_line)"},
		{Removed, "-        } else {"},
		{Removed, "-            Selection::Single(self.selected_line)"},
		{Removed, "-        }"},
		{Added, "+        match self.list.range() {"},
		{Added, "+            Some((lo, hi)) => Selection::Multiple(lo, hi),"},
		{Added, "+            None => Selection::Single(self.list.cursor()),"},
		{Added, "+        }"},
		{Context, "     }"},
		{Hunk, "@@ -311,6 +309,4 @@ impl DiffComponent {"},
		{Context, "     fn reset_selection(&mut self) {"},
		{Removed, "-        self.selection_start = None;"},
		{Added, "+        self.list.clear_range();"},
		{Context, "     }"},
	}
	out := make([]DiffLine, 0, len(raw))
	hunk := 0
	for _, r := range raw {
		if r.k == Hunk {
			hunk++
		}
		out = append(out, DiffLine{Text: r.t, Kind: r.k, Hunk: hunk})
	}
	return out
}

// Popup is one of gitui's 32 popup files, as a menu entry.
//
// The study's point about gitui is that 32 popups with a consistent open, close
// and stack discipline is a SYSTEM, and app.Stack is about screens rather than
// about things layered over one.
type Popup struct {
	Key   string
	Label string
}

// Popups is what the command palette offers.
func Popups() []Popup {
	return []Popup{
		{Key: "b", Label: "branches"},
		{Key: "c", Label: "commit"},
		{Key: "p", Label: "push"},
		{Key: "f", Label: "fetch"},
		{Key: "P", Label: "pull"},
		{Key: "t", Label: "tag"},
		{Key: "B", Label: "blame this file"},
		{Key: "L", Label: "log search"},
		{Key: "g", Label: "go to line"},
		{Key: "o", Label: "options"},
	}
}
