// Package ui is fx's interface, rebuilt on tuikit.
//
// One viewer and a search. fx is the study that concluded the ORIGINAL is
// better, and this rebuild is where that claim is checked rather than asserted.
//
// What it shows: comp.Viewer and comp.Tree do draw fx's screen. What it cannot
// show, because a fixture is 25 lines and fx's problem starts at a million, is
// the reason the study says no — comp.Tree.Visible walks every node on every
// frame where fx's linked list walks only what is visible.
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit-rebuilds/fx/fake"
	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
)

// Regions, named once.
const (
	regDoc    comp.Name = "doc"
	regHeader comp.Name = "header"
	regFooter comp.Name = "footer"
)

// Model is the whole interface.
type Model struct {
	sty styles

	// doc is the viewer. NoCursor is off because fx has a cursor: it is what
	// space folds and what the status line counts.
	doc  comp.Viewer
	tree comp.Tree

	query   string
	typing  bool
	matches []int

	keys comp.Keys
	help bool

	width, height int
	canvas        *comp.Canvas
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	m.doc = comp.Viewer{
		Name: regDoc, Numbers: true, Tab: 2,
		Selected:   &m.sty.selected,
		Ranged:     &m.sty.selected,
		Status:     &m.sty.muted,
		Number:     &m.sty.muted,
		EmptyStyle: &m.sty.muted,
		Empty:      "  nothing to show",
	}
	m.keys = comp.Keys{
		Overlay: true, Title: "fx, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "j/k", Label: "line"},
				{Key: "space", Label: "fold or unfold"},
				{Key: "e", Label: "unfold everything"},
				{Key: "E", Label: "fold everything"},
			}},
			{Name: "finding", Keys: []comp.Hint{
				{Key: "/", Label: "search"},
				{Key: "n", Label: "next match"},
			}},
			{Name: "everywhere", Keys: []comp.Hint{
				{Key: "?", Label: "this"}, {Key: "q", Label: "quit"},
			}},
		},
		TitleStyle: &m.sty.title, SectionStyle: &m.sty.key,
		KeyStyle: &m.sty.title, LabelStyle: &m.sty.muted, Border: &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the document is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// visible is the lines the folds admit, as indices into the document.
//
// # The one thing comp.Tree does not cover
//
// A container's closing bracket sits at the SAME depth as its opening one —
// that is what makes JSON readable — and comp.Tree's model is "rows deeper than
// me, until one that is not". So the close line is a sibling, not a child, and
// folding does not hide it.
//
// Found by drawing a frame: folding everything left a lone `}` on line 2.
//
// It is not a gap in the component. A tree of files or of packets has no
// closing row, and adding one to the model would be adding JSON's punctuation
// to a type that six other tools share. The fix belongs here: a close line
// whose opener is shut is dropped, and the folded line's `{ … }` preview shows
// the bracket instead.
func (m *Model) visible() []int {
	doc := fake.Doc()
	nodes := make([]comp.Node, len(doc))
	for i, l := range doc {
		key := ""
		if l.Open != "" {
			key = l.Path
		}
		nodes[i] = comp.Node{Depth: l.Depth, Key: key}
	}
	rows := m.tree.Visible(nodes)

	out := rows[:0]
	for _, i := range rows {
		if l := doc[i]; l.Close != "" && m.tree.IsCollapsed(strings.TrimSuffix(l.Path, "/end")) {
			continue
		}
		out = append(out, i)
	}
	return out
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		return m, app.Keys{Capture: m.capture(), Screen: m.screenKey, Global: m.globalKey}.Route(msg)
	}
	return m, nil
}

func (m *Model) capture() app.Handled {
	switch {
	case m.help:
		return func(tea.KeyMsg) (tea.Cmd, bool) { m.help = false; return nil, true }
	case m.typing:
		return m.searchKey
	}
	return nil
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "j", "down":
		m.doc.Move(1)
	case "k", "up":
		m.doc.Move(-1)
	case " ":
		m.fold()
	case "e":
		m.tree.Expand()
	case "E":
		m.foldAll()
	case "/":
		m.typing, m.query = true, ""
	case "n":
		m.nextMatch()
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) globalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "?":
		m.help = true
		return nil, true
	case "esc":
		m.query, m.matches = "", nil
		return nil, true
	}
	return nil, false
}

func (m *Model) searchKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		m.typing = false
		m.search()
	case "esc":
		m.typing, m.query = false, ""
	case "backspace":
		if m.query != "" {
			m.query = m.query[:len(m.query)-1]
		}
	default:
		if len(msg.Runes) == 1 {
			m.query += string(msg.Runes)
		}
	}
	return nil, true
}

// fold opens or shuts the container under the cursor.
func (m *Model) fold() {
	doc, rows := fake.Doc(), m.visible()
	if m.doc.Cursor() >= len(rows) {
		return
	}
	if l := doc[rows[m.doc.Cursor()]]; l.Open != "" {
		m.tree.Toggle(l.Path)
	}
}

func (m *Model) foldAll() {
	doc := fake.Doc()
	nodes := make([]comp.Node, len(doc))
	for i, l := range doc {
		key := ""
		if l.Open != "" {
			key = l.Path
		}
		nodes[i] = comp.Node{Depth: l.Depth, Key: key}
	}
	m.tree.Collapse(nodes)
	m.doc.Goto(0)
}

// search finds every visible line containing the query, and goes to the first.
func (m *Model) search() {
	m.matches = nil
	if m.query == "" {
		return
	}
	doc, rows := fake.Doc(), m.visible()
	for i, idx := range rows {
		if strings.Contains(strings.ToLower(text(doc[idx])), strings.ToLower(m.query)) {
			m.matches = append(m.matches, i)
		}
	}
	if len(m.matches) > 0 {
		m.doc.Goto(m.matches[0])
	}
}

func (m *Model) nextMatch() {
	for _, at := range m.matches {
		if at > m.doc.Cursor() {
			m.doc.Goto(at)
			return
		}
	}
	if len(m.matches) > 0 {
		m.doc.Goto(m.matches[0])
	}
}

// text is a line without its styling, for searching.
func text(l fake.Line) string {
	var b strings.Builder
	b.WriteString(strings.Repeat("  ", l.Depth))
	if l.Key != "" {
		b.WriteString(`"` + l.Key + `": `)
	}
	b.WriteString(l.Open + l.Close + l.Value)
	return b.String()
}
