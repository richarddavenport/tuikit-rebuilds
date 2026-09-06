// Package fake is a git repository that does not exist.
//
// The rebuild draws lazygit's interface, so what it needs is a plausible
// working tree and a diff — not git. A real backend would answer a different
// question and would be most of the code.
package fake

import "time"

// At is the moment this repository is frozen at, so ages hold still and a
// captured frame says the same thing tomorrow.
var At = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// Change is one path in the working tree, and what git thinks of it.
//
// Depth and Dir are what make it a tree row: lazygit flattens its file tree
// before drawing it, and comp.Tree takes the flattened form.
type Change struct {
	Path     string
	Depth    int
	Dir      bool
	Staged   rune // the first column of git status
	Unstaged rune
}

// Changes is the working tree, already flattened, parents before children.
func Changes() []Change {
	return []Change{
		{Path: "internal", Depth: 0, Dir: true, Staged: ' ', Unstaged: 'M'},
		{Path: "gui", Depth: 1, Dir: true, Staged: ' ', Unstaged: 'M'},
		{Path: "filetree.go", Depth: 2, Staged: 'M', Unstaged: ' '},
		{Path: "patch_exploring.go", Depth: 2, Staged: ' ', Unstaged: 'M'},
		{Path: "view.go", Depth: 2, Staged: ' ', Unstaged: 'M'},
		{Path: "commands", Depth: 1, Dir: true, Staged: 'A', Unstaged: ' '},
		{Path: "stash.go", Depth: 2, Staged: 'A', Unstaged: ' '},
		{Path: "pkg", Depth: 0, Dir: true, Staged: ' ', Unstaged: 'M'},
		{Path: "theme.go", Depth: 1, Staged: ' ', Unstaged: 'M'},
		{Path: "README.md", Depth: 0, Staged: ' ', Unstaged: 'M'},
		{Path: "go.sum", Depth: 0, Staged: ' ', Unstaged: 'D'},
	}
}

// Branch is one branch, with how far it has diverged.
type Branch struct {
	Name    string
	Current bool
	Ahead   int
	Behind  int
	When    time.Duration
}

// Branches is the local branch list.
func Branches() []Branch {
	return []Branch{
		{Name: "main", Current: true, Ahead: 2, When: 11 * time.Minute},
		{Name: "range-selection", Ahead: 7, Behind: 1, When: 3 * time.Hour},
		{Name: "tree-component", Ahead: 1, When: 26 * time.Hour},
		{Name: "viewer", Behind: 4, When: 70 * time.Hour},
	}
}

// Commit is one entry in the log.
type Commit struct {
	SHA     string
	Subject string
	Author  string
	When    time.Duration
}

// Commits is the recent history.
func Commits() []Commit {
	return []Commit{
		{SHA: "9192dc9", Subject: "Rebuild yazi, and correct the tree evidence", Author: "rd", When: 11 * time.Minute},
		{SHA: "6c97ce0", Subject: "comp.Viewer: the document seven tools built", Author: "rd", When: 4 * time.Hour},
		{SHA: "1c93993", Subject: "comp.Tree and List range selection", Author: "rd", When: 26 * time.Hour},
		{SHA: "cde4b75", Subject: "Widen the scope: state you act on", Author: "rd", When: 30 * time.Hour},
		{SHA: "f87da4b", Subject: "Rebuilds: could tuikit build the TUIs people use?", Author: "rd", When: 33 * time.Hour},
		{SHA: "56b4bdc", Subject: "Start a list of the other TUIs", Author: "rd", When: 51 * time.Hour},
	}
}

// Stash is one stashed change.
type Stash struct {
	Index   int
	Subject string
}

// Stashes is the stash list.
func Stashes() []Stash {
	return []Stash{
		{Index: 0, Subject: "WIP on main: half a scrollbar"},
		{Index: 1, Subject: "the layout experiment that did not work"},
	}
}

// DiffLine is one line of a unified diff, already classified.
//
// Classified here rather than by parsing in the interface, because deciding
// what a line MEANS is git's job and drawing it is the interface's. That split
// is the same one tuikit asks of an engine.
type DiffLine struct {
	Text string
	Kind DiffKind
	// Hunk is which hunk this line belongs to, so a range can snap to one.
	Hunk int
}

// DiffKind is what a diff line is.
type DiffKind int

// The kinds, in the order a diff presents them.
const (
	DiffContext DiffKind = iota
	DiffAdded
	DiffRemoved
	DiffHeader
	DiffHunk
)

// Diff is the change to the file the cursor is on.
func Diff() []DiffLine {
	raw := []struct {
		k DiffKind
		t string
	}{
		{DiffHeader, "diff --git a/internal/gui/patch_exploring.go b/internal/gui/patch_exploring.go"},
		{DiffHeader, "index 4b2c1a9..9f3e7d2 100644"},
		{DiffHeader, "--- a/internal/gui/patch_exploring.go"},
		{DiffHeader, "+++ b/internal/gui/patch_exploring.go"},
		{DiffHunk, "@@ -12,9 +12,7 @@ func (s *State) SelectedRange() (int, int) {"},
		{DiffContext, " func (s *State) SelectedRange() (int, int) {"},
		{DiffRemoved, "-\tif s.rangeStartIdx == nil {"},
		{DiffRemoved, "-\t\treturn s.selectedLineIdx, s.selectedLineIdx"},
		{DiffRemoved, "-\t}"},
		{DiffAdded, "+\tlo, hi, ok := s.list.Range()"},
		{DiffAdded, "+\tif !ok {"},
		{DiffAdded, "+\t\treturn s.list.Cursor(), s.list.Cursor()"},
		{DiffAdded, "+\t}"},
		{DiffContext, " \treturn lo, hi"},
		{DiffContext, " }"},
		{DiffHunk, "@@ -48,6 +46,4 @@ func (s *State) SetLineSelectMode() {"},
		{DiffContext, " func (s *State) SetLineSelectMode() {"},
		{DiffRemoved, "-\ts.rangeStartIdx = nil"},
		{DiffRemoved, "-\ts.selectMode = LINE"},
		{DiffAdded, "+\ts.list.ClearRange()"},
		{DiffContext, " }"},
	}
	out := make([]DiffLine, 0, len(raw))
	hunk := 0
	for _, r := range raw {
		if r.k == DiffHunk {
			hunk++
		}
		out = append(out, DiffLine{Text: r.t, Kind: r.k, Hunk: hunk})
	}
	return out
}
