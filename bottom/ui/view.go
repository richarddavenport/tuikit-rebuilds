package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/bottom/fake"
)

// Draw paints one frame.
//
// A grid: charts across the top, tables across the bottom. comp.Layout nests —
// rows of columns of rows — which is bottom's widget grid without a constraint
// solver.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(r)

	if m.expanded {
		m.expandedWidget(c, bands[0])
	} else {
		m.grid(c, bands[0])
	}
	m.footer(c, bands[1])

	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

func (m *Model) grid(c *comp.Canvas, r comp.Rect) {
	rows := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(3).Min(8), comp.Fill(4).Min(8),
	}}.Rows(r)

	top := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(2).Min(24), comp.Fill(1).Min(20), comp.Fill(1).Min(20),
	}}.Cols(rows[0])
	m.cpuPane(c, top[0])
	m.memPane(c, top[1])
	m.netPane(c, top[2])

	bottom := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(3).Min(40), comp.Fill(2).Min(20),
	}}.Cols(rows[1])
	m.procPane(c, bottom[0])

	side := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(4), comp.Fill(1).Min(4),
	}}.Rows(bottom[1])
	m.diskPane(c, side[0])
	m.tempPane(c, side[1])
}

// cpuPane is one sparkline per core, stacked.
//
// The component the study said bottom forked its framework's chart to get. Each
// core's newest sample is the LAST column, so the spike sits at the right edge
// where now is, and it does not move when a sample arrives.
func (m *Model) cpuPane(c *comp.Canvas, r comp.Rect) {
	inner := m.pane(c, r, regCPU, "CPU", false)
	if inner.H < fake.Cores {
		return
	}
	bands := make([]comp.Constraint, fake.Cores)
	for i := range bands {
		bands[i] = comp.Fill(1)
	}
	for i, band := range (comp.Layout{Constraints: bands}).Rows(inner) {
		series := fake.CPU(i)
		now := series[len(series)-1]
		label := fmt.Sprintf("%d %5.1f%% ", i, now)
		c.Text(band.X, band.Bottom(), label, &m.sty.muted, comp.Region(regCPU).At(i))

		chart := comp.Rect{X: band.X + comp.Width(label), Y: band.Y,
			W: band.W - comp.Width(label), H: band.H}
		// Max is fixed at 100 rather than auto-scaled, because four cores at
		// different loads have to be comparable. Auto-scaling makes an idle
		// core look identical to a busy one.
		comp.Sparkline{Max: 100, Style: m.heat(now)}.Draw(c, chart, series, comp.Region(regCPU).At(i))
	}
}

func (m *Model) memPane(c *comp.Canvas, r comp.Rect) {
	inner := m.pane(c, r, regMem, "Memory", false)
	if inner.H < 2 {
		return
	}
	rows := comp.Layout{Constraints: []comp.Constraint{comp.Fill(2), comp.Fill(1)}}.Rows(inner)

	used, total := fake.Memory()
	m.gauged(c, rows[0], "RAM", used, total, &m.sty.warm)
	sw, swTotal := fake.Swap()
	m.gauged(c, rows[1], "SWP", sw, swTotal, &m.sty.cool)
}

// gauged is a label, a sparkline and the current number.
func (m *Model) gauged(c *comp.Canvas, r comp.Rect, label string, series []float64, total float64, style *lipgloss.Style) {
	if r.H < 1 {
		return
	}
	now := series[len(series)-1]
	head := fmt.Sprintf("%s %.1f/%.0f GiB", label, now, total)
	c.Text(r.X, r.Y, head, &m.sty.muted, comp.Region(regMem))
	if r.H > 1 {
		chart := comp.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1}
		comp.Sparkline{Max: total, Style: style}.Draw(c, chart, series, comp.Region(regMem))
	}
}

func (m *Model) netPane(c *comp.Canvas, r comp.Rect) {
	inner := m.pane(c, r, regNet, "Network", false)
	if inner.H < 2 {
		return
	}
	rows := comp.Layout{Constraints: []comp.Constraint{comp.Fill(1), comp.Fill(1)}}.Rows(inner)
	rx, tx := fake.Net()
	m.gaugedMbit(c, rows[0], "rx", rx, &m.sty.rx)
	m.gaugedMbit(c, rows[1], "tx", tx, &m.sty.tx)
}

func (m *Model) gaugedMbit(c *comp.Canvas, r comp.Rect, label string, series []float64, style *lipgloss.Style) {
	if r.H < 1 {
		return
	}
	head := fmt.Sprintf("%s %.0f Mb/s", label, series[len(series)-1])
	c.Text(r.X, r.Y, head, &m.sty.muted, comp.Region(regNet))
	if r.H > 1 {
		chart := comp.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1}
		// Auto-scaled: rx and tx are asked about separately, so the shape of
		// each matters more than their ratio.
		comp.Sparkline{Style: style}.Draw(c, chart, series, comp.Region(regNet))
	}
}

// procPane is the process table: comp.Table lays out, comp.List scrolls.
func (m *Model) procPane(c *comp.Canvas, r comp.Rect) {
	title := "Processes"
	if m.filter != "" || m.typing {
		title = "Processes  /" + m.filter
	}
	inner := m.pane(c, r, regProc, title, m.focus.Is(regProc))
	if inner.H < 2 {
		return
	}

	rows := m.processes()
	cells := make([][]string, 0, len(rows)+1)
	cells = append(cells, m.sortB.Header(c, fake.Columns()))
	for _, p := range rows {
		cells = append(cells, []string{
			fmt.Sprintf("%d", p.PID), p.Name, p.User, pct(p.CPU), pct(p.Mem), p.State,
		})
	}
	over := comp.Width(m.procs.Marker)
	lines := m.table.Rows(inner.W-over, cells)
	c.Text(inner.X+over, inner.Y, lines[0], &m.sty.header, comp.Region(regProc))

	body := comp.Rect{X: inner.X, Y: inner.Y + 1, W: inner.W, H: inner.H - 1}
	m.procs.DrawFunc(c, body, len(rows), func(i int) comp.Row {
		return comp.Row{Text: lines[i+1], Style: m.load(rows[i].CPU)}
	})
}

func (m *Model) diskPane(c *comp.Canvas, r comp.Rect) {
	inner := m.pane(c, r, regDisk, "Disks", m.focus.Is(regDisk))
	disks := fake.Disks()
	m.disks.DrawFunc(c, inner, len(disks), func(i int) comp.Row {
		d := disks[i]
		full := d.Used / d.Total
		return comp.Row{
			Text: comp.Truncate(d.Mount, max(6, inner.W-16)),
			Right: []comp.Segment{
				{Text: fmt.Sprintf("%.0f%%", full*100), Style: m.heat(full * 100)},
			},
		}
	})
}

func (m *Model) tempPane(c *comp.Canvas, r comp.Rect) {
	inner := m.pane(c, r, regTemp, "Temperatures", m.focus.Is(regTemp))
	temps := fake.Temperatures()
	m.temps.DrawFunc(c, inner, len(temps), func(i int) comp.Row {
		t := temps[i]
		return comp.Row{
			Text: comp.Truncate(t.Sensor, max(6, inner.W-12)),
			Right: []comp.Segment{
				{Text: fmt.Sprintf("%.0f°C", t.Celsius), Style: m.heat(t.Celsius)},
			},
		}
	})
}

// expandedWidget is bottom's `e`: whatever has the keyboard, full screen.
func (m *Model) expandedWidget(c *comp.Canvas, r comp.Rect) {
	switch {
	case m.focus.Is(regDisk):
		m.diskPane(c, r)
	case m.focus.Is(regTemp):
		m.tempPane(c, r)
	default:
		m.procPane(c, r)
	}
}

// pane draws a bordered box and returns the inside, telling the list whether it
// has the keyboard.
func (m *Model) pane(c *comp.Canvas, r comp.Rect, name comp.Name, title string, focused bool) comp.Rect {
	switch name {
	case regProc:
		m.procs.Focused = focused
	case regDisk:
		m.disks.Focused = focused
	case regTemp:
		m.temps.Focused = focused
	}
	return comp.Pane{
		Title: title, Focused: focused,
		Border: &m.sty.border, Focus: &m.sty.header,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(name))
}

// heat is the one place a number becomes a colour, so every widget agrees about
// what "hot" looks like.
func (m *Model) heat(v float64) *lipgloss.Style {
	switch {
	case v >= 85:
		return &m.sty.hot
	case v >= 60:
		return &m.sty.warm
	}
	return &m.sty.cool
}

func (m *Model) load(cpu float64) *lipgloss.Style {
	if cpu >= 40 {
		return &m.sty.warm
	}
	return nil
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	comp.KeyHints(c, r, comp.Region("footer"), &m.sty.muted,
		comp.Hint{Key: "tab", Label: "table"},
		comp.Hint{Key: "s", Label: "sort"},
		comp.Hint{Key: "/", Label: "filter"},
		comp.Hint{Key: "e", Label: "expand"},
		comp.Hint{Key: "?", Label: "keys"},
		comp.Hint{Key: "q", Label: "quit"})
}
