// Package ui is dive's interface, rebuilt on tuikit.
//
// Layers on the left, the filesystem they produced on the right, and a detail
// pane under each. dive is the study that came back "yes" first, and this is
// the check.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/dive/fake"
)

// Regions, named once.
const (
	regLayers    comp.Name = "layers"
	regLayersRow comp.Name = "layers.row"
	regTree      comp.Name = "tree"
	regTreeRow   comp.Name = "tree.row"
	regDetail    comp.Name = "detail"
	regSplit     comp.Name = "split"
	regFooter    comp.Name = "footer"
)

// Model is the whole interface.
type Model struct {
	sty styles

	layers comp.List
	files  comp.List
	tree   comp.Tree

	// focus is which pane has the keyboard. comp.Focus keys by region name, so
	// this file holds an ordering and nothing else — dive's own version is
	// ui/v1/app/controller.go.
	focus comp.Focus

	// attrs toggles the permission and size columns, which is dive's `ctrl+a`.
	attrs bool
	split comp.Split
	mouse app.Mouse

	keys comp.Keys
	help bool

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected     lipgloss.Style
	added, modified, removed, dirStyle lipgloss.Style
	wasted                             lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		added:    lipgloss.NewStyle().Foreground(p.Success),
		modified: lipgloss.NewStyle().Foreground(p.Pending),
		removed:  lipgloss.NewStyle().Foreground(p.Danger),
		dirStyle: lipgloss.NewStyle().Foreground(p.Accent),
		wasted:   lipgloss.NewStyle().Foreground(p.Danger),
	}
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	base := comp.List{
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, NoStatus: true,
	}
	m.layers, m.files = base, base
	m.layers.Name, m.files.Name = regLayersRow, regTreeRow
	m.focus.Ring = []comp.Name{regLayers, regTree}
	m.split = comp.Split{Name: regSplit, Ratio: [2]int{1, 2}, Min: 30}
	m.keys = comp.Keys{
		Overlay: true, Title: "dive, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "tab", Label: "switch pane"},
				{Key: "j/k", Label: "move"},
				{Key: "space", Label: "fold a directory"},
			}},
			{Name: "showing", Keys: []comp.Hint{
				{Key: "a", Label: "permissions and sizes"},
			}},
			{Name: "everywhere", Keys: []comp.Hint{
				{Key: "?", Label: "this"}, {Key: "q", Label: "quit"},
			}},
		},
		TitleStyle: &m.sty.title, SectionStyle: &m.sty.dirStyle,
		KeyStyle: &m.sty.title, LabelStyle: &m.sty.muted, Border: &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the image is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// visible is the file rows the folds admit.
func (m *Model) visible() []int {
	nodes := fake.Tree()
	out := make([]comp.Node, len(nodes))
	for i, n := range nodes {
		key := ""
		if n.Dir {
			key = n.Path
		}
		out[i] = comp.Node{Depth: n.Depth, Key: key}
	}
	return m.tree.Visible(out)
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.mouse.Route(msg, m.canvas, app.Handler{
			Blocked: func() bool { return m.help },
			Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
				m.focus.On(id)
				if id.Index != comp.NoIndex {
					m.list().Select(id.Index)
				}
				return nil
			},
			Wheel: func(id comp.ID, by int) tea.Cmd {
				if m.focus.Is(regLayers) && (id.Name == regLayers || id.Name == regLayersRow) {
					m.layers.Scroll(by)
				} else {
					m.files.Scroll(by)
				}
				return nil
			},
			Drags: func(id comp.ID) bool { return id.Name == regSplit },
			Drag: func(_ comp.ID, msg tea.MouseMsg) tea.Cmd {
				m.split.MoveTo(msg.X, comp.Rect{W: m.width, H: m.height - 1})
				return nil
			},
		})
	case tea.KeyMsg:
		return m, app.Keys{Capture: m.capture(), Screen: m.screenKey, Global: m.globalKey}.Route(msg)
	}
	return m, nil
}

func (m *Model) capture() app.Handled {
	if m.help {
		return func(tea.KeyMsg) (tea.Cmd, bool) { m.help = false; return nil, true }
	}
	return nil
}

// list is whichever list has the keyboard.
func (m *Model) list() *comp.List {
	if m.focus.Is(regLayers) {
		return &m.layers
	}
	return &m.files
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
	case "a":
		m.attrs = !m.attrs
	case " ":
		m.fold()
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
		return nil, true
	}
	return nil, false
}

func (m *Model) fold() {
	if !m.focus.Is(regTree) {
		return
	}
	nodes, rows := fake.Tree(), m.visible()
	if m.files.Cursor() >= len(rows) {
		return
	}
	if n := nodes[rows[m.files.Cursor()]]; n.Dir {
		m.tree.Toggle(n.Path)
	}
}

// layer is the layer under the cursor.
func (m *Model) layer() fake.Layer {
	ls := fake.Layers()
	return ls[min(m.layers.Cursor(), len(ls)-1)]
}
