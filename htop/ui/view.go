package ui

import (
	"fmt"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/htop/fake"
)

// Draw paints one frame.
//
// Three bands: the meters, the process table, the function bar. The first
// band's height is not a constant — it is computed from the arrangement the
// reader chose, which is the difference between a layout in the code and a
// layout in the data.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c

	if m.screen == screenSetup {
		m.drawSetup(c, r)
		m.drawOverlay(c, r)
		return
	}

	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(m.headerRows() + 1),
		comp.Fill(1).Min(3),
		comp.Length(m.incRows()),
		comp.Length(1),
	}}.Rows(r)

	m.drawHeader(c, bands[0])
	m.drawTable(c, bands[1])
	m.drawIncSet(c, bands[2])
	m.drawFnBar(c, bands[3])
	m.drawOverlay(c, r)
}

func (m *Model) incRows() int {
	if m.inc == incOff {
		return 0
	}
	return 1
}

// drawHeader lays the meters out in columns.
//
// comp.Layout with Percent constraints IS htop's HeaderLayout. The thirteen
// named splits in HeaderLayout.h are thirteen of these, and the widths come
// from the chosen one rather than from anything written here.
func (m *Model) drawHeader(c *comp.Canvas, r comp.Rect) {
	widths := Layouts[m.layout].Widths
	cons := make([]comp.Constraint, len(widths))
	for i, w := range widths {
		cons[i] = comp.Percent(w)
	}
	cols := comp.Layout{Constraints: cons, Gap: 1}.Cols(r)

	for ci, col := range cols {
		if ci >= len(m.columns) {
			break
		}
		y := col.Y
		for si, s := range m.columns[ci] {
			need := s.Mode.Rows()
			if y+need > col.Bottom()+1 {
				break
			}
			slot := comp.Rect{X: col.X, Y: y, W: col.W, H: need}
			y = m.drawMeter(c, slot, s, comp.Region(regMeter).At(ci*100+si))
		}
	}
}

// drawTable is the process list.
//
// comp.Table lays the columns out, comp.Sort marks the one being sorted on, and
// comp.List scrolls. The lead column is the tree drawing, which is the tool's.
func (m *Model) drawTable(c *comp.Canvas, r comp.Rect) {
	rows := m.rows()

	// The branch lines go INSIDE the Command cell, which is where htop puts
	// them and the only place they can go without the PID column going ragged.
	// That is the sharper version of the gap: comp.Row.Depth indents a whole
	// row, and a tree drawn inside one column of a table cannot use it at all.
	var leads []string
	if m.treeView {
		leads = m.leads(rows)
	}

	cells := make([][]string, 0, len(rows)+1)
	cells = append(cells, m.sortB.Header(c, fake.Columns()))
	for i, p := range rows {
		command := p.Command
		if leads != nil {
			command = leads[i] + command
		}
		cells = append(cells, []string{
			itoa(p.PID), p.User, itoa(p.Priority), itoa(p.Nice),
			fmt.Sprintf("%.2fG", p.VirtGiB), fmt.Sprintf("%.1fM", p.ResMiB), p.State,
			fmt.Sprintf("%.1f", p.CPU), fmt.Sprintf("%.1f", p.Mem), p.Time, command,
		})
	}

	// The header has to start where the rows' text starts, and List's lead
	// column is the marker. There is no List.LeadWidth to ask (tuikit#76), so
	// this counts it.
	over := comp.Width(m.procs.Marker)
	lines := m.table.Rows(max(1, r.W-over), cells)
	c.Text(r.X+over, r.Y, lines[0], &m.sty.header, comp.Region(regProcs))

	body := comp.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1}
	m.procs.DrawFunc(c, body, len(rows), func(i int) comp.Row {
		row := comp.Row{Text: lines[i+1]}
		if rows[i].CPU >= 50 {
			row.Style = m.heat(rows[i].CPU)
		}
		return row
	})
}

// drawIncSet is htop's search or filter line, which share one control.
func (m *Model) drawIncSet(c *comp.Canvas, r comp.Rect) {
	if m.inc == incOff || r.Empty() {
		return
	}
	label := "Search: "
	if m.inc == incFilter {
		label = "Filter: "
	}
	c.Text(r.X, r.Y, label, &m.sty.key, comp.Region(regIncSet))
	comp.Input{Text: m.incQuery, Cursor: len([]rune(m.incQuery)), Focused: true,
		Placeholder: "type to narrow", TextStyle: &m.sty.muted,
		PlaceholderStyle: &m.sty.border}.
		Draw(c, comp.Rect{X: r.X + comp.Width(label), Y: r.Y, W: r.W - comp.Width(label), H: 1},
			comp.Region(regIncSet))
}

// drawFnBar is htop's FunctionBar: F1..F10 with a label each, and it changes
// with what is on screen.
//
// comp.KeyHints draws a row of key/label pairs, which is the same shape. What
// htop has that this does not is that its bar is an object panels swap out —
// see the study.
func (m *Model) drawFnBar(c *comp.Canvas, r comp.Rect) {
	hints := []comp.Hint{
		{Key: "F1", Label: "Help"}, {Key: "F2", Label: "Setup"},
		{Key: "F3", Label: "Search"}, {Key: "F4", Label: "Filter"},
		{Key: "F5", Label: "Tree"}, {Key: "F6", Label: "SortBy"},
		{Key: "F9", Label: "Kill"}, {Key: "q", Label: "Quit"},
	}
	if m.screen == screenSetup {
		hints = []comp.Hint{
			{Key: "space", Label: "Cycle mode"}, {Key: "enter", Label: "Add"},
			{Key: "d", Label: "Remove"}, {Key: "L", Label: "Layout"},
			{Key: "tab", Label: "Pane"}, {Key: "esc", Label: "Done"},
		}
	}
	comp.KeyHints(c, r, comp.Region(regFnBar), &m.sty.fnLabel, hints...)
}

// drawOverlay draws whatever is layered over the screen.
//
// An ordered switch, which is fine for three. gitui has thirty-two and a real
// stack — tuikit#69.
func (m *Model) drawOverlay(c *comp.Canvas, r comp.Rect) {
	switch m.overlay {
	case overlayHelp:
		m.keys.Draw(c, r, comp.Region(regHelp))
	case overlaySort:
		m.drawPicker(c, r, regSort, "Sort by", labelsOf(), m.sortSel)
	case overlaySignal:
		m.drawPicker(c, r, regSignal, "Send signal", signals, m.signalSel)
	}
}

func labelsOf() []string {
	out := make([]string, len(sortColumns))
	for i, s := range sortColumns {
		out[i] = s.Label
	}
	return out
}

// drawPicker is the small centred list htop uses for both F6 and F9.
//
// One function for both, because they are the same control with different
// contents — which is what htop's Panel is, and what makes its 577 lines pay
// for themselves across nine panels.
func (m *Model) drawPicker(c *comp.Canvas, r comp.Rect, name comp.Name, title string, items []string, sel int) {
	w := comp.Width(title) + 6
	for _, s := range items {
		w = max(w, comp.Width(s)+6)
	}
	h := len(items) + 2
	box := comp.Rect{X: r.X + (r.W-w)/2, Y: r.Y + (r.H-h)/3, W: w, H: h}
	inner := comp.Pane{Title: title, Focused: true,
		Border: &m.sty.border, Focus: &m.sty.title,
		TitleStyle: &m.sty.title, FocusTitle: &m.sty.title}.
		Draw(c, box, comp.Region(name))

	for i, s := range items {
		if i >= inner.H {
			break
		}
		style, mark := &m.sty.muted, "  "
		if i == sel {
			style, mark = &m.sty.selected, "> "
		}
		c.Fill(comp.Rect{X: inner.X, Y: inner.Y + i, W: inner.W, H: 1}, " ", style, comp.Region(name).At(i))
		c.Text(inner.X, inner.Y+i, mark+comp.Truncate(s, inner.W-3), style, comp.Region(name).At(i))
	}
}
