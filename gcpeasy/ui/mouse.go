package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
)

// onMouse routes a mouse event against the last frame.
//
// Every branch here asks WHAT was hit, never where. comp.Canvas records the
// owner of each cell as it is drawn, so `id.Index` is the seventh pod rather
// than row seven of the screen — and those differ the moment anything scrolls.
//
// app.Mouse holds the one piece of state that matters: a drag owns the mouse
// until release, whatever it is now over. The pointer outruns the divider on
// every real drag, and without that rule the divider is dropped mid-gesture.
func (m *Model) onMouse(msg tea.MouseMsg) tea.Cmd {
	return m.mouse.Route(msg, m.canvas, app.Handler{
		// A modal has the mouse while it is up. Clicking the frame behind a
		// question is not an answer to it.
		Blocked: func() bool { return m.help || m.picking },

		// A click focuses the pane it landed in as well as moving its cursor.
		// Without the first half, clicking a row in an unfocused pane moves a
		// cursor the reader cannot see and leaves the keyboard elsewhere.
		Press: func(id comp.ID, _ tea.MouseMsg) tea.Cmd {
			// comp.Focus.On resolves a click on projects.row[2] to the
			// projects pane, at a dot boundary, so this does not need the
			// mapping. It reports whether the focus MOVED, which is how a
			// filter typed in one pane is dropped when you click into another.
			if m.focus.On(id) {
				m.filter, m.typing = "", false
			}
			p, ok := paneOf(id.Name)
			if !ok {
				return nil
			}
			if id.Index != comp.NoIndex {
				m.lists[p].Select(id.Index)
				if p == paneProjects {
					m.lists[paneClusters].Reset()
				}
			}
			return nil
		},

		// The wheel moves the VIEWPORT of whatever the pointer is over, and
		// never the selection. Wiring it to the cursor makes scrolling appear
		// to pick things at random.
		Wheel: func(id comp.ID, by int) tea.Cmd {
			if id.Name == regOutput {
				m.out.Scroll(by)
				return nil
			}
			if p, ok := paneOf(id.Name); ok {
				m.lists[p].Scroll(by)
			}
			return nil
		},

		Drags: func(id comp.ID) bool { return id.Name == regSplit },
		Drag: func(_ comp.ID, msg tea.MouseMsg) tea.Cmd {
			m.split.MoveTo(msg.X, m.bodyRect())
			return nil
		},
	})
}

// bodyRect is where the split lives, derived the same way Draw derives it.
//
// Derived rather than remembered: a rect stored at draw time and read at drag
// time is a rect that describes where the split used to be.
func (m *Model) bodyRect() comp.Rect {
	return comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Length(1), comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(comp.Rect{W: m.width, H: m.height})[2]
}
