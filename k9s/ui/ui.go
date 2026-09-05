// Package ui is k9s's interface, rebuilt on tuikit.
//
// One table, a `:` command bar, marks, a sort, and a viewer for `describe`.
// This is the rebuild that exercises the three components its own study argued
// for: comp.Marks, comp.Sort and comp.Focus.
package ui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/k9s/fake"
)

// Regions, named once.
const (
	regTable  comp.Name = "table"
	regRow    comp.Name = "table.row"
	regHeader comp.Name = "header"
	regPrompt comp.Name = "prompt"
	regView   comp.Name = "view"
	regFooter comp.Name = "footer"
)

// Model is the whole interface.
type Model struct {
	sty styles

	list  comp.List
	table comp.Table
	sortB comp.Sort
	marks comp.Marks

	// view is the describe overlay. Empty means the table is showing.
	view      comp.Viewer
	viewing   bool
	viewTitle string
	viewLines []comp.Line

	prompt  string
	typing  bool
	filter  string
	mouse   app.Mouse
	keys    comp.Keys
	help    bool
	pending string

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected, header lipgloss.Style
	ready, warn, danger, marked            lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		header:   lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		ready:    lipgloss.NewStyle().Foreground(p.Success),
		warn:     lipgloss.NewStyle().Foreground(p.Pending),
		danger:   lipgloss.NewStyle().Foreground(p.Danger),
		marked:   lipgloss.NewStyle().Foreground(p.Pending).Bold(true),
	}
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	m.list = comp.List{
		Name: regRow, Focused: true, Marker: "› ", Blank: "  ",
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted,
		Empty: "  no resources match",
	}
	// A table lays out; a List selects and scrolls. Two components composed,
	// which is why comp.Table has no cursor of its own.
	m.table = comp.Table{Gap: 2, Columns: []comp.Column{
		{}, {Fill: true}, {}, {}, {Right: true}, {Right: true}, {Right: true}, {Right: true},
	}}
	m.view = comp.Viewer{
		Name: regView, Numbers: true, NoCursor: true,
		Status: &m.sty.muted, Number: &m.sty.muted, EmptyStyle: &m.sty.muted,
	}
	m.keys = comp.Keys{
		Overlay: true, Title: "k9s, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "j/k", Label: "move"}, {Key: ":", Label: "switch resource"},
				{Key: "/", Label: "filter"},
			}},
			{Name: "acting", Keys: []comp.Hint{
				{Key: "space", Label: "mark this one"},
				{Key: "ctrl+space", Label: "fill from the nearest mark"},
				{Key: "d", Label: "describe"},
				{Key: "ctrl+d", Label: "delete the marked, or this one"},
			}},
			{Name: "ordering", Keys: []comp.Hint{
				{Key: "s", Label: "sort by the next column"},
				{Key: "S", Label: "reverse"},
			}},
			{Name: "everywhere", Keys: []comp.Hint{
				{Key: "?", Label: "this"}, {Key: "esc", Label: "back"}, {Key: "q", Label: "quit"},
			}},
		},
		TitleStyle: &m.sty.title, SectionStyle: &m.sty.header,
		KeyStyle: &m.sty.title, LabelStyle: &m.sty.muted, Border: &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the cluster is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// rows is the pods on screen: filtered, then ordered.
//
// comp.Sort hands back indices rather than rows, so the fixture is never
// copied and the caller keeps its own types. The comparison stays here, which
// is the half a framework must not guess at — ordering a Kubernetes age against
// a byte count is a domain question.
func (m *Model) rows() []fake.Pod {
	all := fake.Pods()
	var kept []fake.Pod
	for _, p := range all {
		if m.filter == "" || strings.Contains(strings.ToLower(p.Key()), strings.ToLower(m.filter)) {
			kept = append(kept, p)
		}
	}
	order := m.sortB.Apply(len(kept), func(a, b, col int) bool { return less(kept[a], kept[b], col) })
	out := make([]fake.Pod, len(order))
	for i, at := range order {
		out[i] = kept[at]
	}
	return out
}

// less is the comparison, per column. The tool's, always.
func less(a, b fake.Pod, col int) bool {
	switch col {
	case 0:
		return a.Namespace < b.Namespace
	case 2:
		return a.Ready < b.Ready
	case 3:
		return a.Status < b.Status
	case 4:
		return a.Restarts < b.Restarts
	case 5:
		return a.CPU < b.CPU
	case 6:
		return a.Mem < b.Mem
	case 7:
		return a.Age < b.Age
	}
	return a.Name < b.Name
}

// cursor is the pod under the cursor, and whether there is one.
func (m *Model) cursor() (fake.Pod, bool) {
	rows := m.rows()
	if len(rows) == 0 {
		return fake.Pod{}, false
	}
	return rows[min(m.list.Cursor(), len(rows)-1)], true
}

// keys is every visible row's identity, in order, for comp.Marks.Fill.
func (m *Model) keysOf(rows []fake.Pod) []string {
	out := make([]string, len(rows))
	for i, p := range rows {
		out[i] = p.Key()
	}
	return out
}

// acting is what ctrl+d would delete: the marks, or the cursor.
//
// comp.Marks.Acting is the method that makes marks worth having, and this is
// the call site it exists for. Without it every action writes the same `if`.
func (m *Model) acting() []string {
	key := ""
	if p, ok := m.cursor(); ok {
		key = p.Key()
	}
	return m.marks.Acting(key)
}

func sortedJoin(s []string) string {
	c := append([]string(nil), s...)
	sort.Strings(c)
	return strings.Join(c, ", ")
}
