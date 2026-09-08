package ui

import (
	"fmt"
	"strings"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/htop/fake"
)

// The four meter modes, which are htop's Meter.c and the reason comp.Meter is
// not enough on its own.
//
// comp.Meter draws ONE thing: a bracketed bar with a label. That is the right
// default and it is what most tools want. htop's position is different — the
// same number is a bar for one reader, a number for another, a graph for a
// third — and the choice is the reader's, made in Setup and saved.
//
// So the mode is data here, and each of the four is drawn below. Bar and Text
// are things comp could supply; Graph is comp.Sparkline; LED is htop's own and
// nothing should supply it.

// value is what a named meter currently reads, as a fraction and as text.
func value(name string) (frac float64, text string) {
	switch name {
	case "CPU":
		var sum float64
		for i := 0; i < fake.Cores; i++ {
			sum += fake.CPU(i)
		}
		avg := sum / fake.Cores
		return avg / 100, fmt.Sprintf("%.1f%%", avg)
	case "Memory":
		used, total := fake.Memory()
		return used / total, fmt.Sprintf("%.1fG/%.0fG", used, total)
	case "Swap":
		used, total := fake.Swap()
		return used / total, fmt.Sprintf("%.1fG/%.0fG", used, total)
	case "Tasks":
		total, thread, running := fake.Tasks()
		return 0, fmt.Sprintf("%d, %d thr; %d running", total, thread, running)
	case "Load average":
		one, five, fifteen := fake.Load()
		return one / 8, fmt.Sprintf("%.2f %.2f %.2f", one, five, fifteen)
	case "Uptime":
		return 0, fake.Uptime()
	case "Battery":
		return 0.62, "62%"
	case "Hostname":
		return 0, "workshop"
	case "Disk IO":
		return 0.34, "18M read, 4M write"
	case "Network IO":
		return 0.21, "1.4M rx, 220K tx"
	case "Date and time":
		return 0, "2026-09-06 11:42:07"
	default:
		return 0, ""
	}
}

// series is the history behind a meter, for the graph mode.
func series(name string) []float64 {
	switch name {
	case "CPU":
		out := make([]float64, 12)
		for i := 0; i < fake.Cores; i++ {
			for j, v := range fake.History(i) {
				out[j] += v / fake.Cores
			}
		}
		return out
	case "Memory":
		return []float64{62, 64, 68, 71, 69, 74, 76, 78, 77, 80, 79, 80}
	case "Swap":
		return []float64{8, 9, 12, 14, 16, 18, 22, 26, 28, 30, 30, 30}
	default:
		return []float64{20, 26, 22, 34, 30, 42, 38, 50, 44, 56, 52, 60}
	}
}

// drawMeter puts one slot into r and returns the row below it.
func (m *Model) drawMeter(c *comp.Canvas, r comp.Rect, s Slot, id comp.ID) int {
	if r.Empty() {
		return r.Y
	}
	frac, text := value(s.Name)
	switch s.Mode {
	case ModeText:
		return m.textMeter(c, r, s, text, id)
	case ModeGraph:
		return m.graphMeter(c, r, s, id)
	case ModeLED:
		return m.ledMeter(c, r, s, text, id)
	default:
		return m.barMeter(c, r, s, frac, text, id)
	}
}

// caption is htop's short label: CPU, Mem, Swp, Tsk.
func caption(name string) string {
	switch name {
	case "Memory":
		return "Mem"
	case "Swap":
		return "Swp"
	case "Load average":
		return "Load"
	case "Tasks":
		return "Tasks"
	case "Uptime":
		return "Uptime"
	case "Date and time":
		return "Time"
	case "Network IO":
		return "Net"
	case "Disk IO":
		return "Disk"
	case "Hostname":
		return "Host"
	case "Battery":
		return "Bat"
	default:
		return name
	}
}

// barMeter is htop's BarMeterMode: a caption, brackets, a bar of | and the
// value written INSIDE the bar's right end.
//
// The value sitting inside the brackets rather than after them is htop's, and
// it is the reason this is not comp.Meter — comp.Meter puts its label after the
// bar, which is a different shape and the right one for a progress bar.
func (m *Model) barMeter(c *comp.Canvas, r comp.Rect, s Slot, frac float64, text string, id comp.ID) int {
	label := caption(s.Name)
	c.Text(r.X, r.Y, label, &m.sty.meterCap, id)
	x := r.X + comp.Width(label) + 1
	inner := r.W - (x - r.X) - 2
	if inner < 4 {
		return r.Y + 1
	}
	c.Text(x, r.Y, "[", &m.sty.border, id)
	c.Text(x+inner+1, r.Y, "]", &m.sty.border, id)

	filled := int(frac * float64(inner))
	filled = max(0, min(filled, inner))
	style := m.heat(frac * 100)
	if filled > 0 {
		c.Text(x+1, r.Y, strings.Repeat("|", filled), style, id)
	}
	// The value is right-aligned inside the bar, over whatever is under it.
	if w := comp.Width(text); w > 0 && w < inner {
		c.Text(x+1+inner-w, r.Y, text, &m.sty.muted, id)
	}
	return r.Y + 1
}

// textMeter is htop's TextMeterMode: the caption and the value, nothing drawn.
func (m *Model) textMeter(c *comp.Canvas, r comp.Rect, s Slot, text string, id comp.ID) int {
	label := caption(s.Name)
	c.Text(r.X, r.Y, label+":", &m.sty.meterCap, id)
	c.Text(r.X+comp.Width(label)+2, r.Y, comp.Truncate(text, r.W-comp.Width(label)-2), &m.sty.muted, id)
	return r.Y + 1
}

// graphMeter is htop's GraphMeterMode, and the one mode comp supplies whole.
//
// comp.Sparkline right-aligns against now, which is exactly what a meter's
// history wants and is the rule the bottom study produced. Two rows, because
// that is what htop gives it.
func (m *Model) graphMeter(c *comp.Canvas, r comp.Rect, s Slot, id comp.ID) int {
	label := caption(s.Name)
	c.Text(r.X, r.Y, label, &m.sty.meterCap, id)
	x := r.X + comp.Width(label) + 1
	chart := comp.Rect{X: x, Y: r.Y, W: r.W - (x - r.X), H: min(2, r.H)}
	frac, _ := value(s.Name)
	comp.Sparkline{Max: 100, Style: m.heat(frac * 100)}.Draw(c, chart, series(s.Name), id)
	return r.Y + chart.H
}

// segments is which of the seven bars each digit lights.
//
//	 aaa
//	f   b
//	 ggg
//	e   c
//	 ddd
//
// DERIVED, not copied. htop ships this as a finished table of box characters
// (`LEDMeterMode_digitsUtf8`) and htop is GPLv2, while this repository is MIT —
// so the glyphs below are computed from the segment encoding instead, which is
// a fact about seven-segment displays rather than anybody's code.
//
// The visible consequence is that our 1 is a bare stroke where htop draws a
// little flag on it. That flag is htop's own idea and it is a nicer 1; it is
// also exactly the kind of choice that makes a table someone's work rather
// than a lookup.
var segments = [10]string{
	0: "abcdef",
	1: "bc",
	2: "abged",
	3: "abgcd",
	4: "fgbc",
	5: "afgcd",
	6: "afgecd",
	7: "abc",
	8: "abcdefg",
	9: "abcdfg",
}

// ledRows renders one digit into three rows of four columns.
//
// Each corner is a junction: which box character goes there follows from which
// of the segments meeting at that point are lit. Two lines meeting is an elbow,
// three is a tee, one is a stub.
func ledRows(d int) [3]string {
	on := func(seg byte) bool { return strings.IndexByte(segments[d], seg) >= 0 }
	bar := func(lit bool) string {
		if lit {
			return "──"
		}
		return "  "
	}
	return [3]string{
		junction(false, on('f'), on('a'), false) + bar(on('a')) + junction(false, on('b'), false, on('a')),
		junction(on('f'), on('e'), on('g'), false) + bar(on('g')) + junction(on('b'), on('c'), false, on('g')),
		junction(on('e'), false, on('d'), false) + bar(on('d')) + junction(on('c'), false, false, on('d')),
	}
}

// junction is the character where lines meet, given which directions continue.
//
// Sixteen combinations and only the ones a digit can produce are named; the
// rest fall through to a space, which is what an unlit corner is.
func junction(up, down, right, left bool) string {
	switch {
	case up && down && right:
		return "├"
	case up && down && left:
		return "┤"
	case up && right:
		return "└"
	case up && left:
		return "┘"
	case down && right:
		return "┌"
	case down && left:
		return "┐"
	case up && down:
		return "│"
	case up:
		return "╵"
	case down:
		return "╷"
	case right:
		return "╶"
	case left:
		return "╴"
	default:
		return " "
	}
}

// ledMeter draws the value as seven-segment digits, three rows tall.
func (m *Model) ledMeter(c *comp.Canvas, r comp.Rect, s Slot, text string, id comp.ID) int {
	if r.H < 3 {
		return m.textMeter(c, r, s, text, id)
	}
	label := caption(s.Name)
	c.Text(r.X, r.Y+1, label, &m.sty.meterCap, id)
	x := r.X + comp.Width(label) + 1
	style := m.heat(0)
	for _, ch := range text {
		if x+4 > r.X+r.W {
			break
		}
		if ch >= '0' && ch <= '9' {
			glyphs := ledRows(int(ch - '0'))
			for row := 0; row < 3; row++ {
				c.Text(x, r.Y+row, glyphs[row], style, id)
			}
			x += 4
			continue
		}
		// Anything that is not a digit — a dot, a percent — is drawn as itself
		// on the middle row, which is what htop does.
		c.Text(x, r.Y+1, string(ch), &m.sty.muted, id)
		x += comp.Width(string(ch))
	}
	return r.Y + 3
}

// headerRows is how tall the header is with the current arrangement.
//
// The tallest column decides, because the columns sit side by side. This is
// htop's Header_calculateHeight, and it is the one piece of arithmetic that has
// to agree with the drawing — which is why both read it from the same place.
func (m *Model) headerRows() int {
	tall := 0
	for _, col := range m.columns {
		rows := 0
		for _, s := range col {
			rows += s.Mode.Rows()
		}
		tall = max(tall, rows)
	}
	return tall
}
