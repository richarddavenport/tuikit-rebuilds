package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/termshark/fake"
)

// bytesPerRow is how many bytes a hex dump line shows. Sixteen, like every hex
// dump ever written.
const bytesPerRow = 16

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Fill(2).Min(5), comp.Fill(3).Min(5), comp.Fill(2).Min(4), comp.Length(1),
	}}.Rows(r)

	m.filterBar(c, bands[0])
	m.packetPane(c, bands[1])
	m.fieldPane(c, bands[2])
	m.hexPane(c, bands[3])
	m.footer(c, bands[4])

	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

func (m *Model) filterBar(c *comp.Canvas, r comp.Rect) {
	if m.typing {
		c.Text(r.X, r.Y, "filter: "+m.filter+"█", &m.sty.title, comp.Region(regFilter))
		return
	}
	left := []comp.Segment{{Text: "termshark", Style: &m.sty.title}}
	if m.filter != "" {
		left = append(left, comp.Segment{Text: "  filter: " + m.filter, Style: &m.sty.field})
	}
	right := []comp.Segment{{Text: m.status(), Style: &m.sty.muted}}
	comp.Bar{Left: left, Right: right, MinLeft: 12}.Draw(c, r, comp.Region(regFilter))
}

func (m *Model) status() string {
	if n := m.marks.Len(); n > 0 {
		return fmt.Sprintf("%d marked", n)
	}
	return fmt.Sprintf("%d packets", len(m.shown()))
}

// packetPane is the list, with marks in the lead column.
func (m *Model) packetPane(c *comp.Canvas, r comp.Rect) {
	m.packets.Focused = m.focus.Is(regList)
	inner := comp.Pane{
		Title: "Packets", Focused: m.packets.Focused,
		Border: &m.sty.border, Focus: &m.sty.header,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regList))

	rows := m.shown()
	m.packets.DrawFunc(c, inner, len(rows), func(i int) comp.Row {
		p := rows[i]
		lead, leadStyle := "  ", &m.sty.muted
		if m.marks.Has(fmt.Sprintf("%d", p.No)) {
			lead, leadStyle = "■ ", &m.sty.marked
		}
		info := p.Info
		if p.Bad {
			info = "! " + info
		}
		return comp.Row{
			Lead: lead, LeadStyle: leadStyle,
			Spans: []comp.Segment{
				{Text: fmt.Sprintf("%3d ", p.No), Style: &m.sty.muted},
				{Text: fmt.Sprintf("%-10s ", p.Time), Style: &m.sty.muted},
				{Text: fmt.Sprintf("%-11s → %-11s ", p.Src, p.Dst)},
				{Text: fmt.Sprintf("%-8s ", p.Proto), Style: &m.sty.proto},
				{Text: comp.Truncate(info, max(8, inner.W-58)), Style: m.infoStyle(p)},
			},
		}
	})
}

func (m *Model) infoStyle(p fake.Packet) *lipgloss.Style {
	if p.Bad {
		return &m.sty.bad
	}
	return nil
}

// fieldPane is the dissection tree.
func (m *Model) fieldPane(c *comp.Canvas, r comp.Rect) {
	m.fields.Focused = m.focus.Is(regFields)
	inner := comp.Pane{
		Title: "Packet structure", Focused: m.fields.Focused,
		Border: &m.sty.border, Focus: &m.sty.header,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regFields))

	fs, rows := fake.Dissection(), m.visible()
	m.fields.DrawFunc(c, inner, len(rows), func(i int) comp.Row {
		f := fs[rows[i]]
		marker := "  "
		if f.Dir {
			marker = "▾ "
			if m.tree.IsCollapsed(f.Path) {
				marker = "▸ "
			}
		}
		spans := []comp.Segment{
			{Text: strings.Repeat("  ", f.Depth) + marker, Style: &m.sty.muted},
			{Text: f.Label, Style: &m.sty.field},
		}
		if f.Value != "" {
			spans = append(spans, comp.Segment{Text: ": " + f.Value, Style: &m.sty.value})
		}
		return comp.Row{Spans: spans}
	})
}

// hexPane is the bytes, with the selected field's span highlighted.
//
// The link between the tree and the dump is termshark's whole trick, and here
// it is comp.Viewer's range: the highlight is a BACKGROUND, so the hex digits
// keep their own colour underneath.
func (m *Model) hexPane(c *comp.Canvas, r comp.Rect) {
	inner := comp.Pane{
		Title: "Bytes", Focused: m.focus.Is(regHex),
		Border: &m.sty.border, Focus: &m.sty.header,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regHex))

	data := fake.Bytes()
	lines := (len(data) + bytesPerRow - 1) / bytesPerRow

	// Which rows the selected field touches. A Viewer range is over LINES, and
	// a byte span becomes a line span here — the tool's arithmetic, because
	// only the tool knows how many bytes a row holds.
	lo, hi := -1, -1
	if f, ok := m.field(); ok && f.Len > 0 {
		lo, hi = f.Off/bytesPerRow, (f.Off+f.Len-1)/bytesPerRow
		m.hex.Goto(lo)
		m.hex.Extend(hi - lo)
	} else {
		m.hex.ClearRange()
	}

	m.hex.DrawFunc(c, inner, lines, func(i int) comp.Line {
		return comp.Line{Spans: m.hexLine(data, i)}
	})
}

// hexLine is one row: an offset, sixteen bytes, and the printable characters.
func (m *Model) hexLine(data []byte, row int) []comp.Segment {
	start := row * bytesPerRow
	end := min(start+bytesPerRow, len(data))

	var hex, ascii strings.Builder
	for i := start; i < end; i++ {
		fmt.Fprintf(&hex, "%02x ", data[i])
		if data[i] >= 0x20 && data[i] < 0x7f {
			ascii.WriteByte(data[i])
		} else {
			ascii.WriteByte('.')
		}
	}
	return []comp.Segment{
		{Text: fmt.Sprintf("%04x  ", start), Style: &m.sty.muted},
		{Text: fmt.Sprintf("%-48s", hex.String())},
		{Text: " " + ascii.String(), Style: &m.sty.proto},
	}
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted,
		comp.Hint{Key: "tab", Label: "pane"},
		comp.Hint{Key: "space", Label: "fold"},
		comp.Hint{Key: "m", Label: "mark"},
		comp.Hint{Key: "/", Label: "filter"},
		comp.Hint{Key: "?", Label: "keys"},
		comp.Hint{Key: "q", Label: "quit"})
}
