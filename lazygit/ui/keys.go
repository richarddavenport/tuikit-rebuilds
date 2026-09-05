package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/lazygit/fake"
)

// key routes one keystroke, and the ORDER is the contract.
func (m *Model) key(msg tea.KeyMsg) tea.Cmd {
	return app.Keys{
		Capture: m.capture(),
		Screen:  m.screenKey,
		Global:  m.globalKey,
	}.Route(msg)
}

func (m *Model) capture() app.Handled {
	if m.help {
		return func(tea.KeyMsg) (tea.Cmd, bool) { m.help = false; return nil, true }
	}
	// Line-select mode captures, which is the whole reason lazygit has a mode
	// here: while it is on, j and k belong to the diff and not to the file
	// list underneath.
	if m.staging {
		return m.stagingKey
	}
	return nil
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		m.setFocus((m.focus + 1) % panelCount)
	case "shift+tab":
		m.setFocus((m.focus + panelCount - 1) % panelCount)
	case "1", "2", "3", "4", "5":
		m.setFocus(panel(msg.String()[0] - '1'))
	case "j", "down":
		m.lists[m.focus].Move(1)
	case "k", "up":
		m.lists[m.focus].Move(-1)
	case " ":
		// Space toggles a directory open or shut when the cursor is on one,
		// which is comp.Tree's whole job.
		m.toggle()
	case "v":
		if m.focus == panelFiles {
			m.staging = true
			m.diff.NoCursor = false
			m.diff.Goto(5)
		}
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

// stagingKey is line-select mode inside the diff.
//
// The three holes the lazygit study found, working together: a Viewer holding
// the diff, a range over it, and the file tree still drawn beside it.
func (m *Model) stagingKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc", "v":
		m.staging = false
		m.diff.NoCursor = true
		m.diff.ClearRange()
	case "j", "down":
		m.diff.Move(1)
	case "k", "up":
		m.diff.Move(-1)
	case "shift+down", "J":
		m.diff.Extend(1)
	case "shift+up", "K":
		m.diff.Extend(-1)
	case "a":
		// Whole hunk: the range snaps to the hunk the cursor is in. lazygit
		// carries 13 kB of state for this; the anchor and the cursor are
		// enough once both are clamped by the same resolve.
		m.selectHunk()
	}
	return nil, true
}

// selectHunk puts the range around every line of the cursor's hunk.
func (m *Model) selectHunk() {
	lines := fake.Diff()
	cur := m.diff.Cursor()
	if cur >= len(lines) {
		return
	}
	hunk := lines[cur].Hunk
	lo, hi := cur, cur
	for lo > 0 && lines[lo-1].Hunk == hunk {
		lo--
	}
	for hi+1 < len(lines) && lines[hi+1].Hunk == hunk {
		hi++
	}
	m.diff.Goto(lo)
	m.diff.Extend(hi - lo)
}

// toggle opens or shuts the directory under the cursor.
func (m *Model) toggle() {
	if m.focus != panelFiles {
		return
	}
	changes := fake.Changes()
	rows := m.visible()
	cur := m.lists[panelFiles].Cursor()
	if cur >= len(rows) {
		return
	}
	i := rows[cur]
	if !changes[i].Dir {
		return
	}
	m.tree.Toggle(pathOf(changes, i))
}

// setFocus moves the keyboard between panels.
//
// The list is told, because comp.List uses Focused to decide whether a cursor
// move drags the viewport with it. An unfocused panel whose viewport jumps is a
// panel that scrolled for no reason the reader can see.
func (m *Model) setFocus(p panel) {
	m.focus = p
	for i := range m.lists {
		m.lists[i].Focused = panel(i) == p
	}
	m.staging = false
	m.diff.NoCursor = true
	m.diff.ClearRange()
}

// onMouse routes a mouse event against the last frame.
func (m *Model) onMouse(msg tea.MouseMsg) tea.Cmd {
	return m.mouse.Route(msg, m.canvas, app.Handler{
		Blocked: func() bool { return m.help },
		Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
			for p := panelStatus; p < panelCount; p++ {
				if id.Name != p.row() && id.Name != p.region() {
					continue
				}
				if p != m.focus {
					m.setFocus(p)
				}
				if id.Index != comp.NoIndex {
					m.lists[p].Select(id.Index)
				}
			}
			return nil
		},
		Wheel: func(id comp.ID, by int) tea.Cmd {
			if id.Name == "diff" {
				m.diff.Scroll(by)
				return nil
			}
			for p := panelStatus; p < panelCount; p++ {
				if id.Name == p.row() || id.Name == p.region() {
					m.lists[p].Scroll(by)
				}
			}
			return nil
		},
		Drags: func(id comp.ID) bool { return id.Name == "split" },
		Drag: func(_ comp.ID, msg tea.MouseMsg) tea.Cmd {
			m.split.MoveTo(msg.X, comp.Rect{Y: 0, W: m.width, H: m.height - 1})
			return nil
		},
	})
}

func helpSections() []comp.KeySection {
	return []comp.KeySection{
		{Name: "panels", Keys: []comp.Hint{
			{Key: "1-5", Label: "jump to a panel"},
			{Key: "tab", Label: "next panel"},
			{Key: "j/k", Label: "move the cursor"},
		}},
		{Name: "files", Keys: []comp.Hint{
			{Key: "space", Label: "open or shut a directory"},
			{Key: "v", Label: "select lines of the diff"},
		}},
		{Name: "line select", Keys: []comp.Hint{
			{Key: "J/K", Label: "extend the range"},
			{Key: "a", Label: "the whole hunk"},
			{Key: "esc", Label: "leave line select"},
		}},
		{Name: "everywhere", Keys: []comp.Hint{
			{Key: "?", Label: "this"},
			{Key: "q", Label: "quit"},
		}},
	}
}
