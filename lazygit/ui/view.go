package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/lazygit/fake"
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(6), comp.Length(1),
	}}.Rows(r)

	left, right := m.split.Draw(c, bands[0])
	m.panels(c, left)
	m.main(c, right)
	m.footer(c, bands[1])

	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

// panels is the left column: five stacked, sized by how much each has to say.
//
// One Layout declaration. lazygit computes these heights itself; the study's
// point about comp.Layout is that a height computed in two places is a height
// that disagrees with itself.
func (m *Model) panels(c *comp.Canvas, r comp.Rect) {
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(4),
		comp.Fill(3).Min(6),
		comp.Fill(2).Min(4),
		comp.Fill(2).Min(4),
		comp.Length(4),
	}}.Rows(r)

	m.panel(c, bands[0], panelStatus, 1, m.statusRow)
	m.panel(c, bands[1], panelFiles, len(m.visible()), m.fileRow)
	m.panel(c, bands[2], panelBranches, len(fake.Branches()), m.branchRow)
	m.panel(c, bands[3], panelCommits, len(fake.Commits()), m.commitRow)
	m.panel(c, bands[4], panelStash, len(fake.Stashes()), m.stashRow)
}

func (m *Model) panel(c *comp.Canvas, r comp.Rect, p panel, n int, row func(int) comp.Row) {
	title := fmt.Sprintf("%d %s", p+1, p.title())
	inner := comp.Pane{
		Title:      title,
		Focused:    m.focus == p,
		Border:     &m.sty.border,
		Focus:      &m.sty.accent,
		TitleStyle: &m.sty.muted,
		FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(p.region()))
	m.lists[p].DrawFunc(c, inner, n, row)
}

func (m *Model) statusRow(int) comp.Row {
	return comp.Row{Spans: []comp.Segment{
		{Text: " ", Style: &m.sty.plain},
		{Text: "tuikit", Style: &m.sty.title},
		{Text: " → main", Style: &m.sty.muted},
	}}
}

// fileRow is a git status pair, a collapse marker, and the name.
//
// The marker is the TOOL's: comp.Tree reports whether a key is shut and draws
// nothing, so ▸ and ▾ are chosen here. Row.LeadStyle keeps the status letters
// their own color under the selection, which is the one row whose status
// would otherwise become unreadable exactly when you are looking at it.
func (m *Model) fileRow(i int) comp.Row {
	changes := fake.Changes()
	idx := m.visible()[i]
	ch := changes[idx]

	marker := "  "
	if ch.Dir {
		marker = "▾ "
		if m.tree.IsCollapsed(pathOf(changes, idx)) {
			marker = "▸ "
		}
	}
	status := string(ch.Staged) + string(ch.Unstaged)
	style := &m.sty.dirty
	if ch.Staged != ' ' {
		style = &m.sty.staged
	}

	// Row.Depth is NOT used here, and the first frame drawn is why: it indents
	// the lead as well as the text, so the status letters marched right along
	// with the filenames and the column stopped being a column. git status has
	// a fixed left column and only the name is nested, so the indent goes in
	// the text.
	//
	// The cursor marker is prefixed by hand because Row.Lead REPLACES
	// List.Marker rather than sitting beside it (tuikit issue 64). Without it
	// there is no cursor at all once the color is stripped.
	return comp.Row{
		Lead:      m.cursorMark(panelFiles, i) + status + " ",
		LeadStyle: style,
		Text:      strings.Repeat("  ", ch.Depth) + marker + ch.Path,
	}
}

// cursorMark is the selection marker for a row in a panel.
//
// See fileRow for why every list here needs one written by hand.
func (m *Model) cursorMark(p panel, i int) string {
	if m.lists[p].Cursor() == i {
		return "›"
	}
	return " "
}

func (m *Model) branchRow(i int) comp.Row {
	b := fake.Branches()[i]
	lead, style := m.cursorMark(panelBranches, i)+"  ", &m.sty.muted
	if b.Current {
		lead, style = m.cursorMark(panelBranches, i)+"* ", &m.sty.title
	}
	var ahead string
	switch {
	case b.Ahead > 0 && b.Behind > 0:
		ahead = fmt.Sprintf("↑%d ↓%d", b.Ahead, b.Behind)
	case b.Ahead > 0:
		ahead = fmt.Sprintf("↑%d", b.Ahead)
	case b.Behind > 0:
		ahead = fmt.Sprintf("↓%d", b.Behind)
	}
	return comp.Row{
		Lead: lead, LeadStyle: style, Text: b.Name,
		Right: []comp.Segment{{Text: ahead, Style: &m.sty.muted}},
	}
}

func (m *Model) commitRow(i int) comp.Row {
	cm := fake.Commits()[i]
	return comp.Row{
		Lead: m.cursorMark(panelCommits, i), LeadStyle: &m.sty.muted,
		Spans: []comp.Segment{
			{Text: cm.SHA + " ", Style: &m.sty.dirty},
			{Text: comp.Truncate(cm.Subject, 34)},
		},
	}
}

func (m *Model) stashRow(i int) comp.Row {
	s := fake.Stashes()[i]
	return comp.Row{
		Lead: m.cursorMark(panelStash, i) + fmt.Sprintf("%d ", s.Index), LeadStyle: &m.sty.muted,
		Text: s.Subject,
	}
}

// main is the right-hand panel: the diff, as a Viewer.
func (m *Model) main(c *comp.Canvas, r comp.Rect) {
	title := "Patch"
	if m.staging {
		title = "Patch — line select"
	}
	inner := comp.Pane{
		Title:      title,
		Focused:    m.staging,
		Border:     &m.sty.border,
		Focus:      &m.sty.accent,
		TitleStyle: &m.sty.muted,
		FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region("diff.pane"))

	lines := fake.Diff()
	m.diff.DrawFunc(c, inner, len(lines), func(i int) comp.Line {
		return comp.Line{Text: lines[i].Text, Style: m.diffStyle(lines[i].Kind)}
	})
}

func (m *Model) diffStyle(k fake.DiffKind) *lipgloss.Style {
	switch k {
	case fake.DiffAdded:
		return &m.sty.added
	case fake.DiffRemoved:
		return &m.sty.removed
	case fake.DiffHeader:
		return &m.sty.muted
	case fake.DiffHunk:
		return &m.sty.accent
	}
	return nil
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	hints := []comp.Hint{{Key: "1-5", Label: "panel"}, {Key: "space", Label: "fold"}}
	if m.staging {
		hints = []comp.Hint{{Key: "J/K", Label: "extend"}, {Key: "a", Label: "hunk"}, {Key: "esc", Label: "leave"}}
	} else if m.focus == panelFiles {
		hints = append(hints, comp.Hint{Key: "v", Label: "line select"})
	}
	hints = append(hints, comp.Hint{Key: "?", Label: "keys"}, comp.Hint{Key: "q", Label: "quit"})
	comp.KeyHints(c, r, comp.Region("footer"), &m.sty.muted, hints...)
}
