// Package ui is htop's interface, rebuilt on tuikit.
//
// Meters across the top in columns, a process table below, a function bar at
// the bottom. What makes htop worth a study is none of those — it is F2, the
// Setup screen, which lets a user rearrange the meters and save the result.
// See STUDY.md.
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/htop/fake"
)

// Regions, named once.
const (
	regHeader  comp.Name = "header"
	regMeter   comp.Name = "header.meter"
	regProcs   comp.Name = "procs"
	regProcRow comp.Name = "procs.row"
	regSetup   comp.Name = "setup"
	regSetupL  comp.Name = "setup.left"
	regSetupR  comp.Name = "setup.right"
	regFnBar   comp.Name = "fnbar"
	regIncSet  comp.Name = "incset"
	regSort    comp.Name = "sort"
	regSignal  comp.Name = "signal"
	regHelp    comp.Name = "help"
)

// MeterMode is one of htop's four ways to draw the same number.
//
// A user picks per meter, in Setup, and htop persists it. That is the reason
// this is a field on a slot rather than a decision the drawing makes: the same
// CPU meter is a bar for one reader and a graph for another.
type MeterMode int

// The four, in the order Setup cycles them — htop's Meter_nextSupportedMode.
const (
	ModeBar MeterMode = iota
	ModeText
	ModeGraph
	ModeLED
)

func (m MeterMode) String() string {
	return [...]string{"Bar", "Text", "Graph", "LED"}[m]
}

// Rows is the height one meter costs in this mode.
func (m MeterMode) Rows() int {
	switch m {
	case ModeGraph:
		return 2
	case ModeLED:
		return 3
	default:
		return 1
	}
}

// Slot is one meter in the header: what it shows and how it is drawn.
//
// The arrangement is data. That is the whole finding — see STUDY.md — and it
// is why this is a slice a user edits rather than a call in Draw.
type Slot struct {
	Name string
	Mode MeterMode
}

// Layout is a header arrangement: how many columns, and their shares.
//
// htop offers thirteen of these as a closed enum (HeaderLayout.h) rather than
// letting a user type percentages. That is what keeps it checkable: a reader
// arranges the meters, and cannot invent a fourteenth split.
type Layout struct {
	Name   string
	Widths []int
}

// Layouts is the closed set, matching htop's HeaderLayout_layouts.
var Layouts = []Layout{
	{"one_100", []int{100}},
	{"two_50_50", []int{50, 50}},
	{"two_33_67", []int{33, 67}},
	{"three_33_34_33", []int{33, 34, 33}},
	{"three_40_20_40", []int{40, 20, 40}},
	{"four_25_25_25_25", []int{25, 25, 25, 25}},
}

// screen is which whole-screen view is up. Setup replaces the interface rather
// than sitting over it, which is why it is here and not in overlay.
type screen int

const (
	screenMain screen = iota
	screenSetup
)

// overlay is what is drawn ON TOP of the current screen.
//
// htop has one at a time and so does this. gitui has thirty-two and a real
// stack for them, which is tuikit#69 — the gap this rebuild did not hit only
// because htop does not need it.
type overlay int

const (
	noOverlay overlay = iota
	overlaySort
	overlaySignal
	overlayHelp
)

// incMode is htop's IncSet: one control, two meanings.
//
// Search moves the cursor to a match and leaves every row on screen. Filter
// removes the rows that do not match. htop binds them to F3 and F4 and shares
// the editor between them, which is a distinction comp.Input does not make.
type incMode int

const (
	incOff incMode = iota
	incSearch
	incFilter
)

// Model is the whole interface.
type Model struct {
	sty styles

	// header is the arrangement, and it is state rather than code.
	columns [][]Slot
	layout  int

	procs comp.List
	table comp.Table
	sortB comp.Sort

	// tree is collapse state over the process hierarchy. comp.Tree answers
	// which rows are visible; the branch lines are drawn in tree.go, because
	// nothing in comp draws them — see STUDY.md and tuikit#78.
	tree     comp.Tree
	treeView bool

	inc      incMode
	incQuery string
	// filter is the COMMITTED query, which outlives the input being closed.
	// htop's F4 applies while you navigate; dive cannot do that, which is its
	// issue 627.
	filter string

	// setupCol is which of Setup's two panes has the keyboard, and setupSel
	// the row inside each. A tiny two-pane focus, which comp.Focus holds.
	focus    comp.Focus
	setupSel [2]int

	screen  screen
	overlay overlay

	sortSel, signalSel int

	keys  comp.Keys
	mouse app.Mouse

	width, height int
	canvas        *comp.Canvas
}

// New builds it, with htop's own default arrangement: CPU and memory left,
// tasks, load and uptime right.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles(Palette), layout: 1}
	m.columns = [][]Slot{
		{{Name: "CPU", Mode: ModeBar}, {Name: "Memory", Mode: ModeBar}, {Name: "Swap", Mode: ModeBar}},
		{{Name: "Tasks", Mode: ModeText}, {Name: "Load average", Mode: ModeText}, {Name: "Uptime", Mode: ModeText}},
	}
	m.procs = comp.List{
		Name: regProcRow, Focused: true,
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, NoStatus: true,
		Empty: "no process matches",
	}
	m.table = comp.Table{Gap: 1, Columns: []comp.Column{
		{Width: 6, Right: true}, {Width: 9}, {Width: 4, Right: true}, {Width: 4, Right: true},
		{Width: 7, Right: true}, {Width: 8, Right: true}, {Width: 1},
		{Width: 5, Right: true}, {Width: 5, Right: true}, {Width: 9, Right: true}, {Fill: true},
	}}
	m.focus.Ring = []comp.Name{regSetupL, regSetupR}
	m.keys = comp.Keys{
		Overlay: true, Title: "htop, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "j/k", Label: "move"},
				{Key: "t", Label: "tree view"},
				{Key: "space", Label: "fold a subtree"},
			}},
			{Name: "finding", Keys: []comp.Hint{
				{Key: "F3", Label: "search — jump to a match"},
				{Key: "F4", Label: "filter — hide the rest"},
				{Key: "F6", Label: "sort by a column"},
			}},
			{Name: "doing", Keys: []comp.Hint{
				{Key: "F2", Label: "setup — arrange the meters"},
				{Key: "F9", Label: "send a signal"},
			}},
			{Name: "everywhere", Keys: []comp.Hint{
				{Key: "F1/?", Label: "this"}, {Key: "q", Label: "quit"},
			}},
		},
		TitleStyle: &m.sty.title, SectionStyle: &m.sty.header,
		KeyStyle: &m.sty.key, LabelStyle: &m.sty.label, Border: &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the machine is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// signals is the list F9 offers, which is htop's SignalsPanel.
var signals = []string{
	"SIGTERM 15", "SIGKILL 9", "SIGHUP 1", "SIGINT 2", "SIGQUIT 3",
	"SIGSTOP 19", "SIGCONT 18", "SIGUSR1 10", "SIGUSR2 12",
}

// rows is the processes to draw, after the tree, the filter and the sort.
//
// Order matters and is htop's: build the hierarchy first, then hide what a
// filter excludes, then sort. Sorting before the tree would scatter children
// away from their parents, which is the bug the tree exists to prevent.
func (m *Model) rows() []fake.Process {
	all := fake.Processes()

	if m.treeView {
		all = treeOrder(all)
		if q := m.filterQuery(); q != "" {
			all = keepMatching(all, q)
		}
		return m.visibleTree(all)
	}

	if q := m.filterQuery(); q != "" {
		all = keepMatching(all, q)
	}
	if col, desc, ok := m.sortB.By(); ok {
		idx := m.sortB.Apply(len(all), func(a, b, _ int) bool {
			return lessBy(all[a], all[b], col)
		})
		out := make([]fake.Process, len(idx))
		for i, j := range idx {
			out[i] = all[j]
		}
		if desc {
			for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
				out[i], out[j] = out[j], out[i]
			}
		}
		all = out
	}
	return all
}

// filterQuery is the filter in force, whether or not its input is open.
func (m *Model) filterQuery() string {
	if m.inc == incFilter {
		return m.incQuery
	}
	return m.filter
}

func keepMatching(in []fake.Process, q string) []fake.Process {
	q = strings.ToLower(q)
	out := in[:0:0]
	for _, p := range in {
		if strings.Contains(strings.ToLower(p.Command), q) || strings.Contains(strings.ToLower(p.User), q) {
			out = append(out, p)
		}
	}
	return out
}

func lessBy(a, b fake.Process, col int) bool {
	switch col {
	case 0:
		return a.PID < b.PID
	case 1:
		return a.User < b.User
	case 5:
		return a.ResMiB < b.ResMiB
	case 7:
		return a.CPU < b.CPU
	case 8:
		return a.Mem < b.Mem
	default:
		return a.Command < b.Command
	}
}

// sortColumns is what F6 offers, and the index each maps to.
var sortColumns = []struct {
	Label string
	Col   int
}{
	{"PID", 0}, {"USER", 1}, {"RES", 5}, {"CPU%", 7}, {"MEM%", 8}, {"Command", 10},
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.mouse.Route(msg, m.canvas, app.Handler{
			Blocked: func() bool { return m.overlay != noOverlay },
			Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
				if m.screen == screenSetup {
					m.focus.On(id)
					return nil
				}
				if id.Index != comp.NoIndex {
					m.procs.Select(id.Index)
				}
				return nil
			},
			Wheel: func(_ comp.ID, by int) tea.Cmd { m.procs.Scroll(by); return nil },
		})
	case tea.KeyMsg:
		return m, app.Keys{Capture: m.capture(), Screen: m.screenKey, Global: m.globalKey}.Route(msg)
	}
	return m, nil
}

// capture is what takes every key while something modal is up.
//
// The contract app.Keys enforces is that a capture takes the key WHETHER OR NOT
// it does anything with it, so a stray `?` inside a signal list closes the list
// rather than stacking help on top of it.
func (m *Model) capture() app.Handled {
	switch {
	case m.overlay == overlayHelp:
		return func(tea.KeyMsg) (tea.Cmd, bool) { m.overlay = noOverlay; return nil, true }
	case m.overlay == overlaySort:
		return m.sortKey
	case m.overlay == overlaySignal:
		return m.signalKey
	case m.inc != incOff:
		return m.incKey
	}
	return nil
}

func (m *Model) sortKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "j", "down":
		m.sortSel = min(m.sortSel+1, len(sortColumns)-1)
	case "k", "up":
		m.sortSel = max(m.sortSel-1, 0)
	case "enter":
		m.sortB.Toggle(sortColumns[m.sortSel].Col)
		m.overlay = noOverlay
	default:
		m.overlay = noOverlay
	}
	return nil, true
}

func (m *Model) signalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "j", "down":
		m.signalSel = min(m.signalSel+1, len(signals)-1)
	case "k", "up":
		m.signalSel = max(m.signalSel-1, 0)
	default:
		m.overlay = noOverlay
	}
	return nil, true
}

// incKey edits the search or filter query.
//
// One handler for both, because htop shares one LineEditor between them and the
// only difference is what the result is used for. That is the argument for them
// being one control rather than two.
func (m *Model) incKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		// Esc abandons: the input closes and the filter goes with it.
		m.inc, m.incQuery, m.filter = incOff, "", ""
	case "enter":
		// Enter commits. For a filter that means the rows stay narrowed while
		// the keyboard goes back to the table, which is the whole point of
		// having a filter rather than a search.
		if m.inc == incFilter {
			m.filter = m.incQuery
		}
		m.inc, m.incQuery = incOff, ""
	case "backspace":
		if m.incQuery != "" {
			m.incQuery = m.incQuery[:len(m.incQuery)-1]
		}
	default:
		if s := msg.String(); len(s) == 1 {
			m.incQuery += s
		}
	}
	if m.inc == incSearch {
		m.jumpToMatch()
	}
	return nil, true
}

// jumpToMatch is what makes search different from filter: the cursor moves and
// nothing is hidden.
func (m *Model) jumpToMatch() {
	if m.incQuery == "" {
		return
	}
	q := strings.ToLower(m.incQuery)
	for i, p := range m.rows() {
		if strings.Contains(strings.ToLower(p.Command), q) {
			m.procs.Select(i)
			return
		}
	}
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if m.screen == screenSetup {
		return m.setupKey(msg)
	}
	switch msg.String() {
	case "j", "down":
		m.procs.Move(1)
	case "k", "up":
		m.procs.Move(-1)
	case "t", "f5":
		m.treeView = !m.treeView
		m.procs.Reset()
	case " ":
		m.foldHere()
	case "f2", "S":
		m.screen = screenSetup
		m.focus.Set(regSetupL)
	case "f3", "/":
		m.inc, m.incQuery = incSearch, ""
	case "f4", "\\":
		m.inc, m.incQuery = incFilter, ""
	case "f6", ">":
		m.overlay = overlaySort
	case "f9", "K":
		m.overlay = overlaySignal
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) globalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "f1", "?":
		m.overlay = overlayHelp
	case "esc":
		if m.screen == screenSetup {
			m.screen = screenMain
			return nil, true
		}
		return nil, false
	default:
		return nil, false
	}
	return nil, true
}

// setupKey drives the Setup screen: pick a meter on the right, add it to a
// column on the left, cycle its mode, remove it.
//
// This is the whole finding. The arrangement is a value being edited, and the
// available meters are a closed list the reader chooses from — so nothing here
// can name a meter the program cannot draw.
func (m *Model) setupKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	left := m.focus.Is(regSetupL)
	i := 0
	if !left {
		i = 1
	}
	switch msg.String() {
	case "tab", "l", "right", "h":
		m.focus.Next()
	case "j", "down":
		m.setupSel[i] = min(m.setupSel[i]+1, m.setupLen(left)-1)
	case "k", "up":
		m.setupSel[i] = max(m.setupSel[i]-1, 0)
	case " ", "enter":
		if left {
			m.cycleMode()
		} else {
			m.addMeter()
		}
	case "delete", "backspace", "d":
		if left {
			m.removeMeter()
		}
	case "L":
		m.layout = (m.layout + 1) % len(Layouts)
		m.reshape()
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) setupLen(left bool) int {
	if left {
		return len(m.columns[0])
	}
	return len(fake.MeterNames())
}

func (m *Model) cycleMode() {
	if len(m.columns[0]) == 0 {
		return
	}
	i := min(m.setupSel[0], len(m.columns[0])-1)
	m.columns[0][i].Mode = (m.columns[0][i].Mode + 1) % 4
}

func (m *Model) addMeter() {
	names := fake.MeterNames()
	name := names[min(m.setupSel[1], len(names)-1)]
	m.columns[0] = append(m.columns[0], Slot{Name: name, Mode: ModeBar})
}

func (m *Model) removeMeter() {
	if len(m.columns[0]) == 0 {
		return
	}
	i := min(m.setupSel[0], len(m.columns[0])-1)
	m.columns[0] = append(m.columns[0][:i], m.columns[0][i+1:]...)
	m.setupSel[0] = max(0, min(m.setupSel[0], len(m.columns[0])-1))
}

// reshape redistributes the meters when the column count changes, so a reader
// switching from two columns to four does not lose the ones that had nowhere
// to go.
func (m *Model) reshape() {
	var all []Slot
	for _, col := range m.columns {
		all = append(all, col...)
	}
	n := len(Layouts[m.layout].Widths)
	cols := make([][]Slot, n)
	for i, s := range all {
		cols[i%n] = append(cols[i%n], s)
	}
	m.columns = cols
}
