// Package ui is lazygit's interface, rebuilt on tuikit.
//
// Four panels down the left, a main panel on the right, and a diff you can
// select ranges of. What it is NOT is a git client: see fake.
//
// # What this rebuild is testing
//
// The lazygit study found three holes, and all three are now components:
// comp.Tree for the file tree, List.Extend/Range for staging a range of a diff,
// and comp.Viewer for the diff itself. This is those three used in anger, and
// the point is whether they hold up together rather than one at a time.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/lazygit/fake"
)

// panel is which of the five left-hand panels has the keyboard.
//
// Written by hand because tuikit has no focus manager — Focused is a field
// every component can be TOLD, and nothing holds the answer. Five tools in the
// survey wrote this, which is tuikit issue 59.
type panel int

// The five, in the order lazygit numbers them with 1-5.
const (
	panelStatus panel = iota
	panelFiles
	panelBranches
	panelCommits
	panelStash
	panelCount
)

func (p panel) title() string {
	return [...]string{"Status", "Files", "Branches", "Commits", "Stash"}[p]
}

func (p panel) region() comp.Name {
	return [...]comp.Name{"status", "files", "branches", "commits", "stash"}[p]
}

func (p panel) row() comp.Name {
	return [...]comp.Name{"status.row", "files.row", "branches.row", "commits.row", "stash.row"}[p]
}

// Model is the whole interface.
type Model struct {
	sty styles

	lists [panelCount]comp.List
	focus panel

	// tree is collapse state over the flattened working tree. It draws
	// nothing and owns no node type: the rows stay fake.Change, comp.List
	// draws them, and Row.Depth indents them.
	tree comp.Tree

	// diff is the main panel. A Viewer rather than a LogPane because a diff
	// opens at the TOP and does not follow anything.
	diff comp.Viewer
	// staging is line-select mode: j and k move the cursor inside the diff
	// instead of inside the file list, and shift-arrow extends a range.
	staging bool

	split comp.Split
	mouse app.Mouse

	keys comp.Keys
	help bool

	width, height int
	canvas        *comp.Canvas
}

// New builds the interface with its world already in it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	for p := panelStatus; p < panelCount; p++ {
		m.lists[p] = comp.List{
			Name:       p.row(),
			Empty:      "  (none)",
			Selected:   &m.sty.selected,
			Unfocused:  &m.sty.plain,
			Status:     &m.sty.muted,
			EmptyStyle: &m.sty.muted,
			NoStatus:   true,
		}
	}
	m.focus = panelFiles
	m.lists[panelFiles].Focused = true

	m.diff = comp.Viewer{
		Name: "diff", Tab: 4,
		Selected:   &m.sty.selected,
		Ranged:     &m.sty.ranged,
		Status:     &m.sty.muted,
		EmptyStyle: &m.sty.muted,
		Empty:      "  no changes",
		NoCursor:   true,
	}
	m.split = comp.Split{Name: "split", Ratio: [2]int{1, 3}, Min: 24}
	m.keys = comp.Keys{
		Overlay: true, Title: "lazygit, rebuilt",
		Sections:     helpSections(),
		TitleStyle:   &m.sty.title,
		SectionStyle: &m.sty.accent,
		KeyStyle:     &m.sty.title,
		LabelStyle:   &m.sty.muted,
		Border:       &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame, so a script can address a region by name.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the world is a fixture and it is already here.
func (m *Model) Init() tea.Cmd { return nil }

// visible is the file rows the tree admits this frame.
//
// comp.Tree.Visible takes depths and gives back indices into the caller's own
// slice, so the rows stay fake.Change and nothing is copied to be drawn.
func (m *Model) visible() []int {
	changes := fake.Changes()
	nodes := make([]comp.Node, len(changes))
	for i, c := range changes {
		key := ""
		if c.Dir {
			// Keyed by the path, not the index. A row inserted above a
			// collapsed directory would otherwise collapse a different one.
			key = pathOf(changes, i)
		}
		nodes[i] = comp.Node{Depth: c.Depth, Key: key}
	}
	return m.tree.Visible(nodes)
}

// pathOf is a row's full path, built from the depths above it. It is the tree's
// key, so it has to be stable across a filter or a refresh.
func pathOf(changes []fake.Change, i int) string {
	parts := []string{changes[i].Path}
	depth := changes[i].Depth
	for j := i - 1; j >= 0 && depth > 0; j-- {
		if changes[j].Depth == depth-1 {
			parts = append([]string{changes[j].Path}, parts...)
			depth--
		}
	}
	out := ""
	for k, p := range parts {
		if k > 0 {
			out += "/"
		}
		out += p
	}
	return out
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.onMouse(msg)
	case tea.KeyMsg:
		return m, m.key(msg)
	}
	return m, nil
}
