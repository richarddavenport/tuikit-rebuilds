package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/yazi/fake"
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(r)

	m.header(c, bands[0])
	m.columns(c, bands[1])
	m.status(c, bands[2])

	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

func (m *Model) header(c *comp.Canvas, r comp.Rect) {
	left := []comp.Segment{
		{Text: m.cwd + "/", Style: &m.sty.muted},
	}
	if e, ok := m.hovered(); ok {
		left = append(left, comp.Segment{Text: e.Name, Style: &m.sty.title})
	}
	right := []comp.Segment{}
	if n := m.marks.Len(); n > 0 {
		right = append(right, comp.Segment{Text: fmt.Sprintf("%d selected", n), Style: &m.sty.marked})
	}
	comp.Bar{Left: left, Right: right, MinLeft: 20}.Draw(c, r, comp.Region(regHeader))
}

// columns is Miller: parent, current, preview.
//
// Three panes side by side, which is comp.Layout.Cols and nothing more. This is
// the layout yazi actually has — its study first claimed a tree, and reading
// the source found none.
func (m *Model) columns(c *comp.Canvas, r comp.Rect) {
	cols := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(2).Min(14), comp.Fill(3).Min(20), comp.Fill(4).Min(20),
	}}.Cols(r)

	m.parentCol(c, cols[0])
	m.currentCol(c, cols[1])
	m.previewCol(c, cols[2])
}

func (m *Model) parentCol(c *comp.Canvas, r comp.Rect) {
	inner := comp.Pane{
		Title: "..", Border: &m.sty.border, TitleStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region(regParent))

	parent := fake.Parent(m.cwd)
	rows := fake.Dir(parent)
	// The parent column shows where you came from, so its cursor is not the
	// reader's — it is derived from the path.
	for i, e := range rows {
		if parent+"/"+e.Name == m.cwd {
			m.parent.Select(i)
		}
	}
	m.parent.Focused = false
	m.parent.DrawFunc(c, inner, len(rows), func(i int) comp.Row {
		return comp.Row{Text: comp.Truncate(rows[i].Name, max(3, inner.W-2)), Style: m.entryStyle(rows[i])}
	})
}

func (m *Model) currentCol(c *comp.Canvas, r comp.Rect) {
	title := "."
	if m.typing {
		title = "/" + m.search
	}
	inner := comp.Pane{
		Title: title, Focused: true,
		Border: &m.sty.border, Focus: &m.sty.dir,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regCurrent))

	rows := m.entries()
	m.current.DrawFunc(c, inner, len(rows), func(i int) comp.Row {
		e := rows[i]
		// A marked row carries a character, so the selection survives the
		// colour being stripped. yazi uses a colour alone; this does not.
		lead, leadStyle := "  ", &m.sty.muted
		switch {
		case m.marks.Has(m.path(e)):
			lead, leadStyle = "■ ", &m.sty.marked
		case m.inVisual(i):
			lead, leadStyle = "▪ ", &m.sty.marked
		}
		return comp.Row{
			Lead: lead, LeadStyle: leadStyle,
			Text:  comp.Truncate(e.Name, max(3, inner.W-14)),
			Style: m.entryStyle(e),
			Right: []comp.Segment{{Text: size(e), Style: &m.sty.muted}},
		}
	})
}

// previewCol is a file's contents or a directory's listing.
func (m *Model) previewCol(c *comp.Canvas, r comp.Rect) {
	e, ok := m.hovered()
	title := "preview"
	if ok {
		title = e.Name
	}
	inner := comp.Pane{
		Title: title, Border: &m.sty.border, TitleStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region(regPreview))
	if !ok {
		return
	}

	// A directory previews as its listing, which is the third Miller column
	// doing its job.
	if e.Dir {
		kids := fake.Dir(m.path(e))
		if kids == nil {
			// A directory the fixture does not go into. Saying so beats an
			// empty pane, which reads as a bug.
			c.Text(inner.X, inner.Y, "  the fixture stops here", &m.sty.muted, comp.Region(regPreview))
			return
		}
		for i, k := range kids {
			if i >= inner.H {
				break
			}
			c.Text(inner.X, inner.Y+i, comp.Truncate("  "+k.Name, inner.W),
				m.entryStyle(k), comp.Region(regPreview).At(i))
		}
		return
	}

	lines := fake.Preview(e.Name)
	if lines == nil {
		c.Text(inner.X, inner.Y, "  no preview for this file", &m.sty.muted, comp.Region(regPreview))
		if e.Image {
			c.Text(inner.X, inner.Y+1, "  an image goes here — comp.Picture, on a terminal that can",
				&m.sty.muted, comp.Region(regPreview))
		}
		return
	}
	out := make([]comp.Line, len(lines))
	for i, l := range lines {
		out[i] = comp.Line{Text: l}
	}
	m.preview.Draw(c, inner, out)
}

func (m *Model) entryStyle(e fake.Entry) *lipgloss.Style {
	switch {
	case e.Dir:
		return &m.sty.dir
	case e.Exec:
		return &m.sty.exec
	case e.Image:
		return &m.sty.image
	}
	return nil
}

func (m *Model) status(c *comp.Canvas, r comp.Rect) {
	if m.visual {
		c.Text(r.X, r.Y, "  VISUAL — v commits the range, esc drops it",
			&m.sty.marked, comp.Region(regStatus))
		return
	}
	e, ok := m.hovered()
	if !ok {
		comp.KeyHints(c, r, comp.Region(regStatus), &m.sty.muted,
			comp.Hint{Key: "h", Label: "up"}, comp.Hint{Key: "?", Label: "keys"})
		return
	}
	left := []comp.Segment{
		{Text: "  " + e.Mode + "  ", Style: &m.sty.muted},
		{Text: ago(e.Mod), Style: &m.sty.muted},
	}
	rows := m.entries()
	right := []comp.Segment{{
		Text:  fmt.Sprintf("%d/%d ", min(m.current.Cursor()+1, len(rows)), len(rows)),
		Style: &m.sty.muted,
	}}
	comp.Bar{Left: left, Right: right, MinLeft: 12}.Draw(c, r, comp.Region(regStatus))
}

// ago is a duration as a person would say it.
func ago(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
