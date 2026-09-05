// Package ui is gh-dash's interface, rebuilt on tuikit.
//
// Sections as tabs, a table of pull requests, and a sidebar showing the one you
// are on. gh-dash's study is the one where the saving is in `app` rather than
// in `comp`: its Update function is 741 lines, 39% of a 1,917-line file.
//
// This model's Update is thirty. That is the measurement the study makes, and
// this is where it is checked.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/gh-dash/fake"
)

// Regions, named once.
const (
	regTabs   comp.Name = "sections"
	regList   comp.Name = "prs"
	regRow    comp.Name = "prs.row"
	regSide   comp.Name = "sidebar"
	regSplit  comp.Name = "split"
	regFooter comp.Name = "footer"
)

// Model is the whole interface.
type Model struct {
	sty styles

	section int
	list    comp.List
	table   comp.Table
	sortB   comp.Sort
	marks   comp.Marks

	// body is the PR description in the sidebar. A Viewer, which is the one
	// component gh-dash's study said it was missing.
	body comp.Viewer

	// checks is the CI run list. StepList is what it maps onto.
	showChecks bool

	filter string
	typing bool

	split comp.Split
	mouse app.Mouse
	keys  comp.Keys
	help  bool

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected, header lipgloss.Style
	pass, fail, run, draft, marked         lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		header:   lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		pass:     lipgloss.NewStyle().Foreground(p.Success),
		fail:     lipgloss.NewStyle().Foreground(p.Danger),
		run:      lipgloss.NewStyle().Foreground(p.Pending),
		draft:    lipgloss.NewStyle().Foreground(p.Muted),
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
		Empty: "  nothing here", NoStatus: true,
	}
	m.table = comp.Table{Gap: 2, Columns: []comp.Column{
		{}, {}, {Fill: true}, {}, {Right: true}, {Right: true},
	}}
	m.body = comp.Viewer{
		Name: regSide, NoCursor: true, Tab: 2,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted,
		Empty: "  no description",
	}
	m.split = comp.Split{Name: regSplit, Ratio: [2]int{3, 5}, Min: 34}
	m.keys = comp.Keys{
		Overlay: true, Title: "gh-dash, rebuilt",
		Sections: []comp.KeySection{
			{Name: "moving", Keys: []comp.Hint{
				{Key: "tab", Label: "next section"}, {Key: "j/k", Label: "move"},
				{Key: "/", Label: "filter"},
			}},
			{Name: "the one you are on", Keys: []comp.Hint{
				{Key: "c", Label: "show the checks instead of the body"},
				{Key: "space", Label: "mark it"},
			}},
			{Name: "ordering", Keys: []comp.Hint{
				{Key: "s", Label: "sort by the next column"}, {Key: "S", Label: "reverse"},
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

// Init has nothing to do: the dashboard is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// Update handles one message.
//
// gh-dash's is 741 lines. This is the whole of it, and the difference is not
// cleverness — it is that app.Keys owns the routing order, app.Stack would own
// the history, and comp.List owns its own viewport and cursor.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.mouse.Route(msg, m.canvas, app.Handler{
			Blocked: func() bool { return m.help || m.typing },
			Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
				if id.Name == regRow && id.Index != comp.NoIndex {
					m.list.Select(id.Index)
				}
				return nil
			},
			Wheel: func(id comp.ID, by int) tea.Cmd {
				if id.Name == regSide {
					m.body.Scroll(by)
				} else {
					m.list.Scroll(by)
				}
				return nil
			},
			Drags: func(id comp.ID) bool { return id.Name == regSplit },
			Drag: func(_ comp.ID, msg tea.MouseMsg) tea.Cmd {
				m.split.MoveTo(msg.X, comp.Rect{Y: 2, W: m.width, H: m.height - 3})
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
		m.section = (m.section + 1) % len(fake.Sections())
		m.list.Reset()
	case "shift+tab":
		m.section = (m.section + len(fake.Sections()) - 1) % len(fake.Sections())
		m.list.Reset()
	case "j", "down":
		m.list.Move(1)
		m.body.Goto(0)
	case "k", "up":
		m.list.Move(-1)
		m.body.Goto(0)
	case "c":
		m.showChecks = !m.showChecks
	case " ":
		if p, ok := m.current(); ok {
			m.marks.Toggle(p.Key())
		}
	case "s":
		m.sortB.Toggle((m.sortCol() + 1) % len(columns()))
		m.list.Reset()
	case "S":
		m.sortB.Toggle(m.sortCol())
		m.list.Reset()
	case "/":
		m.typing, m.filter = true, ""
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
		m.filter, m.showChecks = "", false
		m.marks.Clear()
		m.list.Reset()
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
		m.list.Reset()
	case "backspace":
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.list.Reset()
		}
	default:
		if len(msg.Runes) == 1 {
			m.filter += string(msg.Runes)
			m.list.Reset()
		}
	}
	return nil, true
}

func (m *Model) sortCol() int {
	if col, _, ok := m.sortB.By(); ok {
		return col
	}
	return len(columns()) - 1
}

func columns() []string { return []string{"", "REPO", "TITLE", "AUTHOR", "±", "UPDATED"} }

// rows is the section's pull requests, filtered then ordered.
func (m *Model) rows() []fake.PR {
	all := fake.PRs(m.section)
	var kept []fake.PR
	for _, p := range all {
		hay := strings.ToLower(p.Title + " " + p.Repo + " " + p.Author)
		if m.filter == "" || strings.Contains(hay, strings.ToLower(m.filter)) {
			kept = append(kept, p)
		}
	}
	order := m.sortB.Apply(len(kept), func(a, b, col int) bool { return less(kept[a], kept[b], col) })
	out := make([]fake.PR, len(order))
	for i, at := range order {
		out[i] = kept[at]
	}
	return out
}

// less is the comparison, and it stays the tool's.
func less(a, b fake.PR, col int) bool {
	switch col {
	case 1:
		return a.Repo < b.Repo
	case 2:
		return a.Title < b.Title
	case 3:
		return a.Author < b.Author
	case 4:
		return a.Additions+a.Deletions < b.Additions+b.Deletions
	case 5:
		return a.Updated < b.Updated
	}
	return a.Number < b.Number
}

// current is the pull request under the cursor.
func (m *Model) current() (fake.PR, bool) {
	rows := m.rows()
	if len(rows) == 0 {
		return fake.PR{}, false
	}
	return rows[min(m.list.Cursor(), len(rows)-1)], true
}

// worst is a PR's CI state as one answer: failing beats running beats passing.
func worst(p fake.PR) fake.CheckState {
	out := fake.Passing
	for _, ck := range p.Checks {
		switch ck.State {
		case fake.Failing:
			return fake.Failing
		case fake.Running:
			out = fake.Running
		}
	}
	return out
}

func plusMinus(p fake.PR) string { return fmt.Sprintf("+%d/-%d", p.Additions, p.Deletions) }
