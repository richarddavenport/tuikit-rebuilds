package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/k9s/fake"
)

// markGlyph is what a marked row carries, and blankMark keeps the others in
// line. Declared here because the header's indent is measured from them.
const (
	markGlyph = "■ "
	blankMark = "  "
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Length(1), comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(r)

	m.header(c, bands[0])
	comp.Rule{Style: &m.sty.border}.Draw(c, bands[1], comp.Region("rule"))
	if m.viewing {
		m.viewer(c, bands[2])
	} else {
		m.body(c, bands[2])
	}
	m.footer(c, bands[3])

	if m.pending != "" {
		m.confirm(c, r)
	}
	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

func (m *Model) header(c *comp.Canvas, r comp.Rect) {
	left := []comp.Segment{{Text: "k9s", Style: &m.sty.title}, {Text: "  pods", Style: &m.sty.muted}}
	if m.filter != "" {
		left = append(left, comp.Segment{Text: "  /" + m.filter, Style: &m.sty.warn})
	}
	right := []comp.Segment{{Text: m.count(), Style: &m.sty.muted}}
	comp.Bar{Left: left, Right: right, MinLeft: 10}.Draw(c, r, comp.Region(regHeader))
}

func (m *Model) count() string {
	if n := m.marks.Len(); n > 0 {
		return fmt.Sprintf("%d marked", n)
	}
	return fmt.Sprintf("%d pods", len(m.rows()))
}

// body is the table: a header row, then a List whose rows a Table laid out.
func (m *Model) body(c *comp.Canvas, r comp.Rect) {
	rows := m.rows()
	// The header carries the sort arrow, from comp.Sort. It comes off the
	// chrome, so a tool that changed its arrows changed them everywhere.
	cells := make([][]string, 0, len(rows)+1)
	cells = append(cells, m.sortB.Header(c, fake.Columns()))
	for _, p := range rows {
		cells = append(cells, m.cells(p))
	}

	// The rows are drawn by a List, which puts its cursor marker and the row's
	// mark before the text. The header has neither, so it has to be indented by
	// exactly what they take — and the table has to be told it has that much
	// less room, or the last column falls off the edge.
	//
	// comp.List.Overhead answers this in ROWS — how many a list spends on its
	// status line — and there is no column equivalent, so the tool counts. It
	// is the marker plus the mark, and both are declared here, so the number
	// cannot drift from what is drawn.
	over := comp.Width(m.list.Marker) + comp.Width(markGlyph)
	lines := m.table.Rows(r.W-over, cells)
	c.Text(r.X+over, r.Y, lines[0], &m.sty.header, comp.Region(regTable))

	body := comp.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1}
	m.list.DrawFunc(c, body, len(rows), func(i int) comp.Row {
		// The mark is a character in the lead column, so it survives the
		// selection AND the color being stripped.
		lead, style := blankMark, &m.sty.muted
		if m.marks.Has(rows[i].Key()) {
			lead, style = markGlyph, &m.sty.marked
		}
		return comp.Row{Lead: lead, LeadStyle: style, Text: lines[i+1], Style: m.rowStyle(rows[i])}
	})
}

func (m *Model) cells(p fake.Pod) []string {
	return []string{
		p.Namespace, p.Name, p.Ready, p.Status,
		fmt.Sprintf("%d", p.Restarts),
		fmt.Sprintf("%dm", p.CPU),
		fmt.Sprintf("%dMi", p.Mem),
		ago(p.Age),
	}
}

func (m *Model) rowStyle(p fake.Pod) *lipgloss.Style {
	switch {
	case p.Status == "CrashLoopBackOff":
		return &m.sty.danger
	case p.Status == "Pending" || p.Ready == "1/2" || p.Restarts > 0:
		return &m.sty.warn
	case p.Status == "Completed":
		return &m.sty.muted
	}
	return nil
}

// viewer is describe output. Not a LogPane: it opens at the top.
func (m *Model) viewer(c *comp.Canvas, r comp.Rect) {
	inner := comp.Pane{
		Title: m.viewTitle, Focused: true,
		Border: &m.sty.border, Focus: &m.sty.header,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regView))
	m.view.Draw(c, inner, m.viewLines)
}

// confirm is what ctrl+d asks before deleting.
func (m *Model) confirm(c *comp.Canvas, _ comp.Rect) {
	comp.Confirm{
		Title:  "delete?",
		Body:   m.pending,
		Danger: true,
		Hints: []comp.Hint{
			{Key: "enter", Label: "this rebuild deletes nothing"},
			{Key: "esc", Label: "cancel"},
		},
		Border:      &m.sty.border,
		TitleStyle:  &m.sty.danger,
		DangerStyle: &m.sty.danger,
		BodyStyle:   &m.sty.muted,
		HintStyle:   &m.sty.muted,
	}.Draw(c, comp.Region("confirm"))
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	if m.typing {
		c.Text(r.X, r.Y, ":"+m.prompt+"█", &m.sty.title, comp.Region(regPrompt))
		return
	}
	hints := []comp.Hint{
		{Key: ":", Label: "resource"}, {Key: "/", Label: "filter"},
		{Key: "space", Label: "mark"}, {Key: "s", Label: "sort"},
		{Key: "d", Label: "describe"}, {Key: "?", Label: "keys"}, {Key: "q", Label: "quit"},
	}
	if m.viewing {
		hints = []comp.Hint{{Key: "j/k", Label: "scroll"}, {Key: "esc", Label: "back"}}
	}
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted, hints...)
}

// ago is a duration as a person would say it.
func ago(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
