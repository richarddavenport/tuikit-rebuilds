package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/k9s/fake"
)

// Update handles one message.
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
				if m.viewing {
					m.view.Scroll(by)
				} else if id.Name == regRow || id.Name == regTable {
					m.list.Scroll(by)
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
		return m.promptKey
	case m.viewing:
		return m.viewKey
	}
	return nil
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "j", "down":
		m.list.Move(1)
	case "k", "up":
		m.list.Move(-1)
	case ":":
		m.typing, m.prompt = true, ""
	case "/":
		m.typing, m.prompt = true, "/"
	case " ":
		// Toggle this one. yazi's gesture is a dragged range; k9s's is this
		// plus Fill, and comp.Marks offers both because two tools each chose
		// one.
		if p, ok := m.cursor(); ok {
			m.marks.Toggle(p.Key())
		}
	case "ctrl+@", "ctrl+ ":
		// Fill from the nearest existing mark to the cursor. No anchor: the
		// set is primary and the range is derived from it.
		rows := m.rows()
		if !m.marks.Fill(m.keysOf(rows), m.list.Cursor()) {
			if p, ok := m.cursor(); ok {
				m.marks.Toggle(p.Key())
			}
		}
	case "s":
		m.sortB.Toggle((m.sortCol() + 1) % len(fake.Columns()))
		m.list.Reset()
	case "S":
		m.sortB.Toggle(m.sortCol())
		m.list.Reset()
	case "d":
		m.describe()
	case "ctrl+d":
		m.pending = "delete " + sortedJoin(m.acting())
	default:
		return nil, false
	}
	return nil, true
}

// sortCol is the column being sorted by, or the last one when nothing is, so
// pressing s the first time starts at the beginning.
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
		m.filter, m.pending = "", ""
		m.marks.Clear()
		return nil, true
	}
	return nil, false
}

func (m *Model) promptKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		m.typing = false
		if len(m.prompt) > 0 && m.prompt[0] == '/' {
			m.filter = m.prompt[1:]
			m.list.Reset()
		}
		m.prompt = ""
	case "esc":
		m.typing, m.prompt = false, ""
	case "backspace":
		if m.prompt != "" {
			m.prompt = m.prompt[:len(m.prompt)-1]
		}
	default:
		if len(msg.Runes) == 1 {
			m.prompt += string(msg.Runes)
		}
	}
	return nil, true
}

func (m *Model) viewKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc", "q":
		m.viewing = false
	case "j", "down":
		m.view.Scroll(1)
	case "k", "up":
		m.view.Scroll(-1)
	}
	return nil, true
}

// describe opens kubectl describe in a comp.Viewer.
//
// A Viewer rather than a LogPane: describe output opens at the TOP and does not
// follow anything.
func (m *Model) describe() {
	p, ok := m.cursor()
	if !ok {
		return
	}
	m.viewing, m.viewTitle = true, "describe "+p.Key()
	m.viewLines = nil
	for _, line := range splitLines(fake.Describe(p)) {
		m.viewLines = append(m.viewLines, comp.Line{Text: line})
	}
	m.view.Goto(0)
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}
