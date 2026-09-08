package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/gh-dash/fake"
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Length(1), comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(r)

	m.tabs(c, bands[0])
	comp.Rule{Style: &m.sty.border}.Draw(c, bands[1], comp.Region("rule"))
	left, right := m.split.Draw(c, bands[2])
	m.table1(c, left)
	m.sidebar(c, right)
	m.footer(c, bands[3])

	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

// tabs is gh-dash's sections, with a count on each.
func (m *Model) tabs(c *comp.Canvas, r comp.Rect) {
	secs := fake.Sections()
	tabs := make([]comp.Tab, 0, len(secs))
	for i, s := range secs {
		tabs = append(tabs, comp.Tab{Name: s.Title, Count: len(fake.PRs(i))})
	}
	comp.Tabs{
		Tabs: tabs, Active: m.section, Focused: !m.help,
		Style: &m.sty.muted, Selected: &m.sty.title,
		FocusSelected: &m.sty.title, Chrome: &m.sty.border,
	}.Draw(c, r, regTabs)
}

// table1 is the pull request table: comp.Table lays out, comp.List scrolls.
func (m *Model) table1(c *comp.Canvas, r comp.Rect) {
	title := fake.Sections()[m.section].Query
	if m.typing || m.filter != "" {
		title = "/" + m.filter
	}
	inner := comp.Pane{
		Title: title, Focused: true,
		Border: &m.sty.border, Focus: &m.sty.header,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regList))
	if inner.H < 2 {
		return
	}

	rows := m.rows()
	cells := make([][]string, 0, len(rows)+1)
	cells = append(cells, m.sortB.Header(c, columns()))
	for _, p := range rows {
		cells = append(cells, []string{
			checkMark(worst(p)), p.Repo, p.Title, p.Author, plusMinus(p), ago(p.Updated),
		})
	}

	over := comp.Width(m.list.Marker) + comp.Width(markGlyph)
	lines := m.table.Rows(inner.W-over, cells)
	c.Text(inner.X+over, inner.Y, lines[0], &m.sty.header, comp.Region(regList))

	body := comp.Rect{X: inner.X, Y: inner.Y + 1, W: inner.W, H: inner.H - 1}
	m.list.DrawFunc(c, body, len(rows), func(i int) comp.Row {
		lead, style := blankMark, &m.sty.muted
		if m.marks.Has(rows[i].Key()) {
			lead, style = markGlyph, &m.sty.marked
		}
		return comp.Row{
			Lead: lead, LeadStyle: style,
			Text: lines[i+1], Style: m.rowStyle(rows[i]),
		}
	})
}

// markGlyph is what a marked row carries; blankMark keeps the rest in line.
const (
	markGlyph = "■ "
	blankMark = "  "
)

func (m *Model) rowStyle(p fake.PR) *lipgloss.Style {
	if p.Draft {
		return &m.sty.draft
	}
	return nil
}

// checkMark is CI's answer as a CHARACTER, so it survives the color being
// stripped.
func checkMark(s fake.CheckState) string {
	switch s {
	case fake.Failing:
		return "✗"
	case fake.Running:
		return "◐"
	case fake.Skipped:
		return "·"
	}
	return "✓"
}

// sidebar is the PR you are on: its body, or its checks.
func (m *Model) sidebar(c *comp.Canvas, r comp.Rect) {
	p, ok := m.current()
	title := "sidebar"
	if ok {
		title = fmt.Sprintf("%s#%d", p.Repo, p.Number)
	}
	inner := comp.Pane{
		Title: title, Border: &m.sty.border, TitleStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region(regSide))
	if !ok {
		return
	}

	rows := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Length(1), comp.Fill(1).Min(2),
	}}.Rows(inner)

	c.Text(rows[0].X, rows[0].Y, comp.Truncate(p.Title, rows[0].W), &m.sty.title, comp.Region(regSide))
	meta := fmt.Sprintf("%s wants to merge %s  ·  %s", p.Author, p.Branch, ago(p.Updated))
	if p.Approved > 0 {
		meta += fmt.Sprintf("  ·  %d approved", p.Approved)
	}
	c.Text(rows[1].X, rows[1].Y, comp.Truncate(meta, rows[1].W), &m.sty.muted, comp.Region(regSide))

	if m.showChecks {
		m.checks(c, rows[2], p)
		return
	}
	if len(p.Body) == 0 {
		c.Text(rows[2].X, rows[2].Y, "  no description", &m.sty.muted, comp.Region(regSide))
		return
	}
	lines := make([]comp.Line, len(p.Body))
	for i, l := range p.Body {
		lines[i] = comp.Line{Text: l}
	}
	m.body.Draw(c, rows[2], lines)
}

// checks is the CI runs as a comp.StepList, which is what a run in progress is.
func (m *Model) checks(c *comp.Canvas, r comp.Rect, p fake.PR) {
	steps := make([]comp.Step, 0, len(p.Checks))
	for _, ck := range p.Checks {
		st := comp.Step{Label: ck.Name, Took: ck.Took.String()}
		switch ck.State {
		case fake.Passing:
			st.State = comp.StepDone
		case fake.Failing:
			st.State, st.Note = comp.StepFailed, "see the log"
		case fake.Running:
			st.State, st.Took = comp.StepRunning, ""
		default:
			st.State, st.Took = comp.StepSkipped, ""
		}
		steps = append(steps, st)
	}
	// The glyphs are the TOOL's, not the component's — a closed glyph set is
	// what makes that the right way round.
	comp.StepList{
		Steps: steps,
		Look: [5]comp.StepLook{
			comp.StepWaiting: comp.Look("·", &m.sty.muted),
			comp.StepRunning: comp.Look("◐", &m.sty.run),
			comp.StepSkipped: comp.Look("·", &m.sty.muted),
			comp.StepDone:    comp.Look("✓", &m.sty.pass),
			comp.StepFailed:  comp.Look("✗", &m.sty.fail),
		},
		Muted: &m.sty.muted,
	}.Draw(c, r, "checks")
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	if m.typing {
		c.Text(r.X, r.Y, "/"+m.filter+"█", &m.sty.title, comp.Region(regFooter))
		return
	}
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted,
		comp.Hint{Key: "tab", Label: "section"},
		comp.Hint{Key: "c", Label: "checks"},
		comp.Hint{Key: "space", Label: "mark"},
		comp.Hint{Key: "s", Label: "sort"},
		comp.Hint{Key: "?", Label: "keys"},
		comp.Hint{Key: "q", Label: "quit"})
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
