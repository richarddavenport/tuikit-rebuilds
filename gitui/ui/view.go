package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/gitui/fake"
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Fill(1).Min(6), comp.Length(1),
	}}.Rows(r)

	m.tabs(c, bands[0])
	switch m.tab {
	case fake.Status:
		m.status(c, bands[1])
	case fake.Log:
		m.logTab(c, bands[1])
	case fake.Files:
		m.filesTab(c, bands[1])
	case fake.Stashing:
		m.stashTab(c, bands[1])
	}
	m.footer(c, bands[2])

	// The popups, outermost first, so the innermost is drawn last and reads as
	// on top. That ordering is the stack, and it is the whole of what a
	// comp.Modals would take over.
	for _, p := range m.stack {
		switch p {
		case popupPalette:
			m.palette.Draw(c, comp.Center(r, 56, min(r.H-4, 16)))
		case popupHelp:
			m.keys.Draw(c, r, comp.Region("help"))
		case popupConfirm:
			comp.Confirm{
				Title: "gitui", Body: m.confirm,
				Hints: []comp.Hint{
					{Key: "enter", Label: "this rebuild changes nothing"},
					{Key: "esc", Label: "back"},
				},
				Border: &m.sty.border, TitleStyle: &m.sty.title,
				BodyStyle: &m.sty.muted, HintStyle: &m.sty.muted,
			}.Draw(c, comp.Region("confirm"))
		}
	}
}

func (m *Model) tabs(c *comp.Canvas, r comp.Rect) {
	names := make([]comp.Tab, 0, fake.TabCount)
	for t := fake.Status; t < fake.TabCount; t++ {
		tab := comp.Tab{Name: t.Name()}
		// A count beside the name, which is what comp.Tab.Count is for and
		// what gitui's status tab shows.
		if t == fake.Status {
			tab.Count = len(fake.Changes())
		}
		names = append(names, tab)
	}
	comp.Tabs{
		Tabs: names, Active: int(m.tab), Focused: m.top() == noPopup,
		Style: &m.sty.muted, Selected: &m.sty.title,
		FocusSelected: &m.sty.title, Chrome: &m.sty.border,
	}.Draw(c, r, regTabs)
}

// status is gitui's first tab: unstaged and staged on the left, the diff right.
func (m *Model) status(c *comp.Canvas, r comp.Rect) {
	left, right := m.split.Draw(c, r)
	rows := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(3), comp.Fill(1).Min(3),
	}}.Rows(left)

	m.changePane(c, rows[0], regUnstaged, "Unstaged", &m.unstaged, m.changes(false))
	m.changePane(c, rows[1], regStaged, "Staged", &m.staged, m.changes(true))
	m.diffPane(c, right)
}

func (m *Model) changePane(c *comp.Canvas, r comp.Rect, name comp.Name, title string, list *comp.List, rows []fake.Change) {
	list.Focused = m.focus.Is(name) && !m.staging && m.top() == noPopup
	inner := comp.Pane{
		Title: title, Focused: list.Focused,
		Border: &m.sty.border, Focus: &m.sty.hunk,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(name))

	list.DrawFunc(c, inner, len(rows), func(i int) comp.Row {
		ch := rows[i]
		style := &m.sty.dirty
		if ch.Staged {
			style = &m.sty.staged
		}
		return comp.Row{
			Lead: string(ch.Status) + " ", LeadStyle: style,
			Text: comp.Truncate(ch.Path, max(4, inner.W-4)),
		}
	})
}

// diffPane is the main panel: a comp.Viewer with a range over it.
func (m *Model) diffPane(c *comp.Canvas, r comp.Rect) {
	title := "Diff"
	if m.staging {
		title = "Diff — line select"
	}
	inner := comp.Pane{
		Title: title, Focused: m.staging,
		Border: &m.sty.border, Focus: &m.sty.hunk,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region("diff.pane"))

	lines := fake.Diff()
	m.diff.DrawFunc(c, inner, len(lines), func(i int) comp.Line {
		return comp.Line{Text: lines[i].Text, Style: m.diffStyle(lines[i].Kind)}
	})
}

func (m *Model) diffStyle(k fake.DiffKind) *lipgloss.Style {
	switch k {
	case fake.Added:
		return &m.sty.added
	case fake.Removed:
		return &m.sty.removed
	case fake.Hunk:
		return &m.sty.hunk
	case fake.Header:
		return &m.sty.muted
	}
	return nil
}

func (m *Model) logTab(c *comp.Canvas, r comp.Rect) {
	m.log.Focused = m.top() == noPopup
	inner := comp.Pane{
		Title: "Commit", Focused: m.log.Focused,
		Border: &m.sty.border, Focus: &m.sty.hunk,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regLog))

	commits := fake.Commits()
	m.log.DrawFunc(c, inner, len(commits), func(i int) comp.Row {
		cm := commits[i]
		spans := []comp.Segment{
			{Text: cm.SHA + "  ", Style: &m.sty.dirty},
			{Text: cm.Subject},
		}
		row := comp.Row{Spans: spans, Right: []comp.Segment{
			{Text: cm.Author + "  ", Style: &m.sty.muted},
			{Text: ago(cm.When), Style: &m.sty.muted},
		}}
		if len(cm.Refs) > 0 {
			row.Spans = append(row.Spans, comp.Segment{Text: "  (" + cm.Refs[0] + ")", Style: &m.sty.staged})
		}
		return row
	})
}

func (m *Model) filesTab(c *comp.Canvas, r comp.Rect) {
	inner := comp.Pane{
		Title: "Files", Border: &m.sty.border, TitleStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region("files"))
	c.Text(inner.X, inner.Y, "  the revision file tree — comp.Tree, as in the lazygit rebuild",
		&m.sty.muted, comp.Region("files"))
}

func (m *Model) stashTab(c *comp.Canvas, r comp.Rect) {
	m.stash.Focused = m.top() == noPopup
	inner := comp.Pane{
		Title: "Stashes", Focused: m.stash.Focused,
		Border: &m.sty.border, Focus: &m.sty.hunk,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regStash))

	stashes := fake.Stashes()
	m.stash.DrawFunc(c, inner, len(stashes), func(i int) comp.Row {
		s := stashes[i]
		return comp.Row{
			Lead: s.SHA + "  ", LeadStyle: &m.sty.dirty,
			Text:  s.Subject,
			Right: []comp.Segment{{Text: ago(s.When), Style: &m.sty.muted}},
		}
	})
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	hints := []comp.Hint{{Key: "1-4", Label: "tab"}, {Key: "space", Label: "popups"}}
	if m.staging {
		hints = []comp.Hint{
			{Key: "J/K", Label: "extend"}, {Key: "a", Label: "hunk"},
			{Key: "enter", Label: "stage"}, {Key: "esc", Label: "leave"},
		}
	} else if m.tab == fake.Status {
		hints = append(hints, comp.Hint{Key: "v", Label: "line select"})
	}
	if depth := len(m.stack); depth > 0 {
		hints = []comp.Hint{{Key: "esc", Label: "close one of " + itoa(depth)}}
	}
	hints = append(hints, comp.Hint{Key: "?", Label: "keys"}, comp.Hint{Key: "q", Label: "quit"})
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted, hints...)
}

// ago is a duration as a person would say it.
func ago(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
