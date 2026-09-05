package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/fx/fake"
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Length(1), comp.Fill(1).Min(3), comp.Length(1),
	}}.Rows(r)

	comp.Bar{
		Left:  []comp.Segment{{Text: "fx", Style: &m.sty.title}, {Text: "  package-lock.json", Style: &m.sty.muted}},
		Right: []comp.Segment{{Text: m.count(), Style: &m.sty.muted}},
	}.Draw(c, bands[0], comp.Region(regHeader))
	comp.Rule{Style: &m.sty.border}.Draw(c, bands[1], comp.Region("rule"))

	rows := m.visible()
	doc := fake.Doc()
	m.doc.DrawFunc(c, bands[2], len(rows), func(i int) comp.Line {
		return comp.Line{Spans: m.spans(doc[rows[i]])}
	})

	m.footer(c, bands[3])
	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

func (m *Model) count() string {
	if m.query != "" {
		return itoa(len(m.matches)) + " matching " + m.query
	}
	return itoa(len(m.visible())) + " of " + itoa(len(fake.Doc()))
}

// spans is one line of JSON, coloured by what its parts are.
//
// Built as spans rather than as a string, which is the whole reason the viewer
// is not a list: the cursor's style goes UNDER these, so the line you are
// reading keeps its syntax colours.
func (m *Model) spans(l fake.Line) []comp.Segment {
	out := []comp.Segment{{Text: strings.Repeat("  ", l.Depth)}}

	// The fold marker is the TOOL's. comp.Tree reports whether a key is shut
	// and draws nothing, so ▸ and ▾ are chosen here.
	switch {
	case l.Open != "" && m.tree.IsCollapsed(l.Path):
		out = append(out, comp.Segment{Text: "▸ ", Style: &m.sty.muted})
	case l.Open != "":
		out = append(out, comp.Segment{Text: "▾ ", Style: &m.sty.muted})
	default:
		out = append(out, comp.Segment{Text: "  "})
	}

	if l.Key != "" {
		out = append(out, comp.Segment{Text: `"` + l.Key + `"`, Style: &m.sty.key})
		out = append(out, comp.Segment{Text: ": ", Style: &m.sty.muted})
	}
	switch {
	case l.Open != "" && m.tree.IsCollapsed(l.Path):
		// A folded container shows what is inside it without opening it, which
		// is the whole point of folding a config file.
		out = append(out, comp.Segment{Text: l.Open + " … " + closerFor(l.Open), Style: &m.sty.muted})
	case l.Open != "":
		out = append(out, comp.Segment{Text: l.Open, Style: &m.sty.muted})
	case l.Close != "":
		out = append(out, comp.Segment{Text: l.Close, Style: &m.sty.muted})
	default:
		out = append(out, comp.Segment{Text: l.Value, Style: m.kindStyle(l.Kind)})
	}
	if l.Comma {
		out = append(out, comp.Segment{Text: ",", Style: &m.sty.muted})
	}
	return out
}

func closerFor(open string) string {
	if open == "[" {
		return "]"
	}
	return "}"
}

func (m *Model) kindStyle(k fake.Kind) *lipgloss.Style {
	switch k {
	case fake.Str:
		return &m.sty.str
	case fake.Num:
		return &m.sty.num
	case fake.Bool:
		return &m.sty.boolean
	case fake.Null:
		return &m.sty.null
	}
	return nil
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	if m.typing {
		c.Text(r.X, r.Y, "/"+m.query, &m.sty.title, comp.Region(regFooter))
		c.Text(r.X+comp.Width(m.query)+1, r.Y, "█", &m.sty.title, comp.Region(regFooter))
		return
	}
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted,
		comp.Hint{Key: "space", Label: "fold"},
		comp.Hint{Key: "/", Label: "search"},
		comp.Hint{Key: "?", Label: "keys"},
		comp.Hint{Key: "q", Label: "quit"})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
