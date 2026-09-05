package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/gitui/fake"
)

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		return m, m.mouse.Route(msg, m.canvas, app.Handler{
			Blocked: func() bool { return m.top() != noPopup },
			Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
				m.focus.On(id)
				if id.Index != comp.NoIndex {
					m.list().Select(id.Index)
				}
				return nil
			},
			Wheel: func(id comp.ID, by int) tea.Cmd {
				if id.Name == regDiff {
					m.diff.Scroll(by)
				} else {
					m.list().Scroll(by)
				}
				return nil
			},
			Drags: func(id comp.ID) bool { return id.Name == regSplit },
			Drag: func(_ comp.ID, msg tea.MouseMsg) tea.Cmd {
				m.split.MoveTo(msg.X, comp.Rect{Y: 1, W: m.width, H: m.height - 2})
				return nil
			},
		})
	case tea.KeyMsg:
		return m, app.Keys{Capture: m.capture(), Screen: m.screenKey, Global: m.globalKey}.Route(msg)
	}
	return m, nil
}

// capture is the innermost popup, then line-select mode.
//
// The order is the stack. Whatever is on top gets the key, which is what makes
// esc close ONE thing rather than everything.
func (m *Model) capture() app.Handled {
	switch m.top() {
	case popupPalette:
		return m.paletteKey
	case popupHelp:
		return m.simpleKey
	case popupConfirm:
		return m.confirmKey
	}
	if m.staging {
		return m.stagingKey
	}
	return nil
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "1", "2", "3", "4":
		m.tab = fake.Tab(msg.String()[0] - '1')
	case "tab":
		m.tab = (m.tab + 1) % fake.TabCount
	case "shift+tab":
		m.focus.Next()
	case "j", "down":
		m.list().Move(1)
	case "k", "up":
		m.list().Move(-1)
	case "v":
		if m.tab == fake.Status {
			m.staging = true
			m.diff.NoCursor = false
			m.diff.Goto(3)
		}
	case " ":
		m.push(popupPalette)
		m.palette.Query, m.palette.Caret = "", 0
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
		m.push(popupHelp)
		return nil, true
	case "esc":
		// One popup, not all of them.
		if m.top() != noPopup {
			m.pop()
			return nil, true
		}
		m.staging = false
		m.diff.NoCursor = true
		m.diff.ClearRange()
		return nil, true
	}
	return nil, false
}

func (m *Model) simpleKey(tea.KeyMsg) (tea.Cmd, bool) { m.pop(); return nil, true }

func (m *Model) confirmKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "enter" {
		m.pop()
		m.staging = false
		m.diff.NoCursor = true
		m.diff.ClearRange()
		return nil, true
	}
	m.pop()
	return nil, true
}

func (m *Model) paletteKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		// A popup opening a popup, which is the case a stack exists for: the
		// palette stays underneath and esc comes back to it.
		if item, ok := m.palette.Selected(); ok && item.Key != "" {
			m.confirm = item.Label
			m.push(popupConfirm)
		}
	case "up":
		m.palette.Move(-1)
	case "down":
		m.palette.Move(1)
	case "backspace":
		if q := m.palette.Query; q != "" {
			m.palette.Query = q[:len(q)-1]
			m.palette.Caret = len(m.palette.Query)
		}
	default:
		if len(msg.Runes) == 1 {
			m.palette.Query += string(msg.Runes)
			m.palette.Caret = len(m.palette.Query)
		}
	}
	return nil, true
}

// stagingKey is line select inside the diff.
func (m *Model) stagingKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "j", "down":
		m.diff.Move(1)
	case "k", "up":
		m.diff.Move(-1)
	case "shift+down", "J":
		m.diff.Extend(1)
	case "shift+up", "K":
		m.diff.Extend(-1)
	case "a":
		m.selectHunk()
	case "enter":
		lo, hi, ok := m.diff.Range()
		if !ok {
			lo, hi = m.diff.Cursor(), m.diff.Cursor()
		}
		m.confirm = "stage lines " + itoa(lo+1) + "–" + itoa(hi+1)
		m.push(popupConfirm)
	}
	return nil, true
}

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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
