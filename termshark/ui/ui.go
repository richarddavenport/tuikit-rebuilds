// Package ui is termshark's interface, rebuilt on tuikit.
//
// Three panes stacked: the packet list, the dissection tree, and the hex dump.
// termshark is the study that came back "no", because gowid gives it a real
// focus system and tuikit had only a bool. comp.Focus is the answer to that,
// and this rebuild is where it is checked.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/termshark/fake"
)

// Regions, named once. The first three are the focus ring.
const (
	regList     comp.Name = "packets"
	regListRow  comp.Name = "packets.row"
	regFields   comp.Name = "fields"
	regFieldRow comp.Name = "fields.row"
	regHex      comp.Name = "hex"
	regFilter   comp.Name = "filter"
	regFooter   comp.Name = "footer"
)

// Model is the whole interface.
type Model struct {
	sty styles

	packets comp.List
	fields  comp.List
	tree    comp.Tree
	hex     comp.Viewer
	focus   comp.Focus

	filter string
	typing bool

	// marks is termshark's copy mode, which it has over the table AND over the
	// tree. comp.Marks is not a field on either, for that reason.
	marks comp.Marks

	mouse app.Mouse
	keys  comp.Keys
	help  bool

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected, header lipgloss.Style
	bad, proto, field, value, hit, marked  lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		header:   lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		bad:      lipgloss.NewStyle().Foreground(p.Danger),
		proto:    lipgloss.NewStyle().Foreground(p.Success),
		field:    lipgloss.NewStyle().Foreground(p.Accent),
		value:    lipgloss.NewStyle().Foreground(p.Muted),
		// The bytes a selected field covers. A background, so the hex digits
		// keep their own colour underneath — which is comp.Viewer's rule.
		hit:    lipgloss.NewStyle().Background(p.SelectionBG),
		marked: lipgloss.NewStyle().Foreground(p.Pending).Bold(true),
	}
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	base := comp.List{
		Marker: "› ", Blank: "  ",
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, NoStatus: true,
		Empty: "  no packets match",
	}
	m.packets, m.fields = base, base
	m.packets.Name, m.fields.Name = regListRow, regFieldRow
	m.focus.Ring = []comp.Name{regList, regFields, regHex}

	// NoCursor is deliberately OFF, and it is worth saying why.
	//
	// The highlight in the dump is not a selection the reader made — it is
	// derived from whichever field is selected in the tree. But
	// Viewer.NoCursor disables the cursor AND the range together, so a derived
	// highlight cannot exist without a cursor to anchor it.
	//
	// Leaving the cursor on is harmless here: Viewer only paints it when
	// Focused, and the hex pane is focused only when the reader tabs to it. One
	// tool so far, so this is written down rather than filed.
	m.hex = comp.Viewer{
		Name:   regHex,
		Ranged: &m.sty.hit, Selected: &m.sty.hit,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted,
	}
	m.keys = comp.Keys{
		Overlay: true, Title: "termshark, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "tab", Label: "next pane"}, {Key: "j/k", Label: "move"},
				{Key: "space", Label: "fold a field"},
			}},
			{Name: "selecting", Keys: []comp.Hint{
				{Key: "m", Label: "mark this packet"},
				{Key: "M", Label: "fill from the nearest mark"},
			}},
			{Name: "finding", Keys: []comp.Hint{{Key: "/", Label: "display filter"}}},
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

// Init has nothing to do: the capture is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// shown is the packets the display filter admits.
func (m *Model) shown() []fake.Packet {
	if m.filter == "" {
		return fake.Packets()
	}
	var out []fake.Packet
	for _, p := range fake.Packets() {
		hay := strings.ToLower(p.Proto + " " + p.Src + " " + p.Dst + " " + p.Info)
		if strings.Contains(hay, strings.ToLower(m.filter)) {
			out = append(out, p)
		}
	}
	return out
}

// visible is the dissection rows the folds admit.
func (m *Model) visible() []int {
	fs := fake.Dissection()
	nodes := make([]comp.Node, len(fs))
	for i, f := range fs {
		key := ""
		if f.Dir {
			key = f.Path
		}
		nodes[i] = comp.Node{Depth: f.Depth, Key: key}
	}
	return m.tree.Visible(nodes)
}

// field is the dissection row under the cursor, which decides which bytes are
// highlighted. That link is termshark's whole trick.
func (m *Model) field() (fake.Field, bool) {
	rows := m.visible()
	if len(rows) == 0 {
		return fake.Field{}, false
	}
	fs := fake.Dissection()
	return fs[rows[min(m.fields.Cursor(), len(rows)-1)]], true
}

// list is whichever list has the keyboard.
func (m *Model) list() *comp.List {
	if m.focus.Is(regFields) {
		return &m.fields
	}
	return &m.packets
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.mouse.Route(msg, m.canvas, app.Handler{
			Blocked: func() bool { return m.help || m.typing },
			Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
				m.focus.On(id)
				if id.Index != comp.NoIndex {
					m.list().Select(id.Index)
				}
				return nil
			},
			Wheel: func(id comp.ID, by int) tea.Cmd {
				if id.Name == regHex {
					m.hex.Scroll(by)
				} else {
					m.list().Scroll(by)
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
		return m.filterKey
	}
	return nil
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		m.focus.Next()
	case "shift+tab":
		m.focus.Prev()
	case "j", "down":
		if m.focus.Is(regHex) {
			m.hex.Scroll(1)
		} else {
			m.list().Move(1)
		}
	case "k", "up":
		if m.focus.Is(regHex) {
			m.hex.Scroll(-1)
		} else {
			m.list().Move(-1)
		}
	case " ":
		m.fold()
	case "m":
		if p, ok := m.packet(); ok {
			m.marks.Toggle(fmt.Sprintf("%d", p.No))
		}
	case "M":
		rows := m.shown()
		keys := make([]string, len(rows))
		for i, p := range rows {
			keys[i] = fmt.Sprintf("%d", p.No)
		}
		m.marks.Fill(keys, m.packets.Cursor())
	case "/":
		m.typing, m.filter = true, ""
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) packet() (fake.Packet, bool) {
	rows := m.shown()
	if len(rows) == 0 {
		return fake.Packet{}, false
	}
	return rows[min(m.packets.Cursor(), len(rows)-1)], true
}

func (m *Model) globalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "?":
		m.help = true
		return nil, true
	case "esc":
		m.filter = ""
		m.marks.Clear()
		m.packets.Reset()
		return nil, true
	}
	return nil, false
}

func (m *Model) filterKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		m.typing = false
	case "esc":
		m.typing, m.filter = false, ""
		m.packets.Reset()
	case "backspace":
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.packets.Reset()
		}
	default:
		if len(msg.Runes) == 1 {
			m.filter += string(msg.Runes)
			m.packets.Reset()
		}
	}
	return nil, true
}

func (m *Model) fold() {
	if !m.focus.Is(regFields) {
		return
	}
	if f, ok := m.field(); ok && f.Dir {
		m.tree.Toggle(f.Path)
	}
}
