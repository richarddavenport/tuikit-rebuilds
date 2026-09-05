// Package ui is bottom's interface, rebuilt on tuikit.
//
// A grid of widgets, all live: CPU, memory, network, processes, disks and
// temperatures. bottom is the study that came back "no — the charts are the
// tool", so this rebuild exists to check that comp.Sparkline is the component
// it was missing.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/bottom/fake"
)

// Regions, named once. The ring is the order tab moves through.
const (
	regCPU   comp.Name = "cpu"
	regMem   comp.Name = "memory"
	regNet   comp.Name = "network"
	regProc  comp.Name = "processes"
	regProcR comp.Name = "processes.row"
	regDisk  comp.Name = "disks"
	regDiskR comp.Name = "disks.row"
	regTemp  comp.Name = "temperatures"
	regTempR comp.Name = "temperatures.row"
)

// Model is the whole interface.
type Model struct {
	sty styles

	focus comp.Focus
	procs comp.List
	disks comp.List
	temps comp.List
	table comp.Table
	sortB comp.Sort

	// expanded is bottom's `e`: one widget takes the whole screen.
	expanded bool

	filter string
	typing bool
	mouse  app.Mouse
	keys   comp.Keys
	help   bool

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected, header lipgloss.Style
	cool, warm, hot                        lipgloss.Style
	rx, tx                                 lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		header:   lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		cool:     lipgloss.NewStyle().Foreground(p.Success),
		warm:     lipgloss.NewStyle().Foreground(p.Pending),
		hot:      lipgloss.NewStyle().Foreground(p.Danger),
		rx:       lipgloss.NewStyle().Foreground(p.Success),
		tx:       lipgloss.NewStyle().Foreground(p.Accent),
	}
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	base := comp.List{
		Marker: "› ", Blank: "  ",
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, NoStatus: true,
		Empty: "  nothing matches",
	}
	m.procs, m.disks, m.temps = base, base, base
	m.procs.Name, m.disks.Name, m.temps.Name = regProcR, regDiskR, regTempR
	m.focus.Ring = []comp.Name{regProc, regDisk, regTemp}
	m.table = comp.Table{Gap: 2, Columns: []comp.Column{
		{Right: true}, {Fill: true}, {}, {Right: true}, {Right: true}, {},
	}}
	m.keys = comp.Keys{
		Overlay: true, Title: "bottom, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "tab", Label: "next table"}, {Key: "j/k", Label: "move"},
				{Key: "e", Label: "expand this widget"},
			}},
			{Name: "processes", Keys: []comp.Hint{
				{Key: "s", Label: "sort by the next column"}, {Key: "S", Label: "reverse"},
				{Key: "/", Label: "filter"},
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

// Init has nothing to do: the machine is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// processes is the table's rows: filtered, then ordered.
func (m *Model) processes() []fake.Process {
	all := fake.Processes()
	var kept []fake.Process
	for _, p := range all {
		if m.filter == "" || strings.Contains(strings.ToLower(p.Name), strings.ToLower(m.filter)) {
			kept = append(kept, p)
		}
	}
	order := m.sortB.Apply(len(kept), func(a, b, col int) bool { return less(kept[a], kept[b], col) })
	out := make([]fake.Process, len(order))
	for i, at := range order {
		out[i] = kept[at]
	}
	return out
}

// less is the comparison, and it stays the tool's.
func less(a, b fake.Process, col int) bool {
	switch col {
	case 0:
		return a.PID < b.PID
	case 2:
		return a.User < b.User
	case 3:
		return a.CPU < b.CPU
	case 4:
		return a.Mem < b.Mem
	case 5:
		return a.State < b.State
	}
	return a.Name < b.Name
}

// list is whichever table has the keyboard.
func (m *Model) list() *comp.List {
	switch {
	case m.focus.Is(regDisk):
		return &m.disks
	case m.focus.Is(regTemp):
		return &m.temps
	}
	return &m.procs
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
			Wheel: func(_ comp.ID, by int) tea.Cmd { m.list().Scroll(by); return nil },
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
		m.list().Move(1)
	case "k", "up":
		m.list().Move(-1)
	case "e":
		m.expanded = !m.expanded
	case "s":
		m.sortB.Toggle((m.sortCol() + 1) % len(fake.Columns()))
		m.procs.Reset()
	case "S":
		m.sortB.Toggle(m.sortCol())
		m.procs.Reset()
	case "/":
		m.typing, m.filter = true, ""
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) sortCol() int {
	if col, _, ok := m.sortB.By(); ok {
		return col
	}
	return len(fake.Columns()) - 1
}

func (m *Model) globalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "?":
		m.help = true
		return nil, true
	case "esc":
		m.expanded, m.filter = false, ""
		m.procs.Reset()
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
		m.procs.Reset()
	case "backspace":
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.procs.Reset()
		}
	default:
		if len(msg.Runes) == 1 {
			m.filter += string(msg.Runes)
			m.procs.Reset()
		}
	}
	return nil, true
}

// pct is a percentage as bottom writes it.
func pct(v float64) string { return fmt.Sprintf("%.1f%%", v) }
