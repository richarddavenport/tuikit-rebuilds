// Package ui is yazi's interface, rebuilt on tuikit.
//
// Miller columns — parent, current, preview — with a header, a status line and
// overlays. No tree anywhere: yazi has none, which is what checking its source
// corrected in its own study.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/yazi/fake"
)

// Regions, named once.
const (
	regParent  comp.Name = "parent"
	regCurrent comp.Name = "current"
	regCurRow  comp.Name = "current.row"
	regPreview comp.Name = "preview"
	regHeader  comp.Name = "header"
	regStatus  comp.Name = "status"
)

// Root is where the fixture starts.
const Root = "/Users/richardd/Developer/richarddavenport/tuikit"

// Model is the whole interface.
type Model struct {
	sty styles

	cwd string
	// cursors remembers where you were in each directory, keyed by path. Going
	// up and coming back has to land where you left, which is the thing every
	// file manager gets right and every naive one does not.
	cursors map[string]int

	current comp.List
	parent  comp.List
	preview comp.Viewer

	// marks is yazi's selection: a SET, keyed by path, that survives leaving a
	// directory and coming back. visual is the range being dragged, which
	// commits into it.
	marks  comp.Marks
	visual bool
	// visLow is the anchor. The OTHER end is not stored: it is read from the
	// cursor whenever the range is needed.
	//
	// It used to be stored, updated in the key handler, and it lagged by one
	// row. comp.List.Move is deferred — "Cursor() between a Move and a Draw is
	// the old value" — so a range accumulated in the handler is always a row
	// behind what the reader can see.
	//
	// Deriving it costs nothing and cannot go stale.
	visLow int

	search string
	typing bool

	mouse app.Mouse
	keys  comp.Keys
	help  bool

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected  lipgloss.Style
	dir, exec, image, marked, hover lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		dir:      lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		exec:     lipgloss.NewStyle().Foreground(p.Success),
		image:    lipgloss.NewStyle().Foreground(p.Pending),
		marked:   lipgloss.NewStyle().Foreground(p.Pending).Bold(true),
		hover:    lipgloss.NewStyle().Background(p.SelectionBG),
	}
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles(), cwd: Root, cursors: map[string]int{}}
	base := comp.List{
		Marker: "› ", Blank: "  ",
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, NoStatus: true,
		Empty: "  (empty)",
	}
	m.current, m.parent = base, base
	m.current.Name, m.parent.Name = regCurRow, "parent.row"
	m.current.Focused = true
	m.preview = comp.Viewer{
		Name: regPreview, NoCursor: true, Tab: 4,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted,
		Empty: "  no preview",
	}
	m.keys = comp.Keys{
		Overlay: true, Title: "yazi, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "j/k", Label: "move"}, {Key: "h", Label: "up a directory"},
				{Key: "l", Label: "into the directory"},
			}},
			{Name: "selecting", Keys: []comp.Hint{
				{Key: "space", Label: "mark this one"},
				{Key: "v", Label: "visual mode: drag a range"},
				{Key: "esc", Label: "clear the marks"},
			}},
			{Name: "finding", Keys: []comp.Hint{{Key: "/", Label: "search"}}},
			{Name: "everywhere", Keys: []comp.Hint{
				{Key: "?", Label: "this"}, {Key: "q", Label: "quit"},
			}},
		},
		TitleStyle: &m.sty.title, SectionStyle: &m.sty.dir,
		KeyStyle: &m.sty.title, LabelStyle: &m.sty.muted, Border: &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the filesystem is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// entries is the current directory, after the search.
func (m *Model) entries() []fake.Entry {
	all := fake.Dir(m.cwd)
	if m.search == "" {
		return all
	}
	var out []fake.Entry
	for _, e := range all {
		if strings.Contains(strings.ToLower(e.Name), strings.ToLower(m.search)) {
			out = append(out, e)
		}
	}
	return out
}

// hovered is the entry under the cursor.
func (m *Model) hovered() (fake.Entry, bool) {
	rows := m.entries()
	if len(rows) == 0 {
		return fake.Entry{}, false
	}
	return rows[min(m.current.Cursor(), len(rows)-1)], true
}

// path is an entry's full path, and its identity for comp.Marks.
func (m *Model) path(e fake.Entry) string { return m.cwd + "/" + e.Name }

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.mouse.Route(msg, m.canvas, app.Handler{
			Blocked: func() bool { return m.help || m.typing },
			Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
				if id.Name == regCurRow && id.Index != comp.NoIndex {
					m.current.Select(id.Index)
				}
				return nil
			},
			Wheel: func(id comp.ID, by int) tea.Cmd {
				if id.Name == regPreview {
					m.preview.Scroll(by)
				} else {
					m.current.Scroll(by)
				}
				return nil
			},
		})
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
		m.current.Move(1)
	case "k", "up":
		m.current.Move(-1)
	case "h", "left":
		m.up()
	case "l", "right", "enter":
		m.into()
	case " ":
		if e, ok := m.hovered(); ok {
			m.marks.Toggle(m.path(e))
		}
	case "v":
		m.toggleVisual()
	case "/":
		m.typing, m.search = true, ""
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
		m.marks.Clear()
		m.visual, m.search = false, ""
		m.current.Reset()
		return nil, true
	}
	return nil, false
}

func (m *Model) searchKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		m.typing = false
	case "esc":
		m.typing, m.search = false, ""
		m.current.Reset()
	case "backspace":
		if m.search != "" {
			m.search = m.search[:len(m.search)-1]
			m.current.Reset()
		}
	default:
		if len(msg.Runes) == 1 {
			m.search += string(msg.Runes)
			m.current.Reset()
		}
	}
	return nil, true
}

// toggleVisual starts or commits a dragged range.
//
// yazi's gesture, and the other half of comp.Marks: the range is primary and
// it commits into the set. k9s does the opposite, and Marks offers both.
func (m *Model) toggleVisual() {
	if m.visual {
		m.commitVisual()
		m.visual = false
		return
	}
	m.visual = true
	m.visLow = m.current.Cursor()
}

// visualRange is the anchor and the cursor, ordered.
func (m *Model) visualRange() (lo, hi int) {
	lo, hi = m.visLow, m.current.Cursor()
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi
}

// commitVisual hands the range's keys to the set.
func (m *Model) commitVisual() {
	rows := m.entries()
	lo, hi := m.visualRange()
	var keys []string
	for i := lo; i <= hi && i < len(rows); i++ {
		keys = append(keys, m.path(rows[i]))
	}
	m.marks.Span(keys)
}

// inVisual reports whether a row is inside the range being dragged.
func (m *Model) inVisual(i int) bool {
	if !m.visual {
		return false
	}
	lo, hi := m.visualRange()
	return i >= lo && i <= hi
}

// up and into move between directories, remembering where the cursor was.
func (m *Model) up() {
	parent := fake.Parent(m.cwd)
	if parent == "" {
		return
	}
	m.cursors[m.cwd] = m.current.Cursor()
	was := m.cwd
	m.cwd = parent
	m.current.Reset()
	// Land on the directory we came out of, which is what makes h and l feel
	// like moving rather than like teleporting.
	for i, e := range m.entries() {
		if m.path(e) == was {
			m.current.Select(i)
		}
	}
	m.visual = false
}

func (m *Model) into() {
	e, ok := m.hovered()
	if !ok || !e.Dir {
		return
	}
	next := m.path(e)
	if fake.Dir(next) == nil {
		return
	}
	m.cursors[m.cwd] = m.current.Cursor()
	m.cwd = next
	m.current.Reset()
	m.current.Select(m.cursors[next])
	m.visual = false
}

// size is bytes as a file manager says them.
func size(e fake.Entry) string {
	if e.Dir {
		return ""
	}
	switch {
	case e.Size < 1000:
		return fmt.Sprintf("%d B", e.Size)
	case e.Size < 1000_000:
		return fmt.Sprintf("%.1f kB", float64(e.Size)/1000)
	}
	return fmt.Sprintf("%.1f MB", float64(e.Size)/1000_000)
}
