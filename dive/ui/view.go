package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/dive/fake"
)

// Draw paints one frame.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(6), comp.Length(1),
	}}.Rows(r)

	left, right := m.split.Draw(c, bands[0])
	m.layerPane(c, left)
	m.treePane(c, right)
	m.footer(c, bands[1])

	if m.help {
		m.keys.Draw(c, r, comp.Region("help"))
	}
}

// layerPane is the layer list and, under it, what the selected one did.
func (m *Model) layerPane(c *comp.Canvas, r comp.Rect) {
	rows := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(6), comp.Length(9),
	}}.Rows(r)

	m.layers.Focused = m.focus.Is(regLayers)
	m.layers.Marker, m.files.Marker = "· ", "· "
	if m.layers.Focused {
		m.layers.Marker = "› "
	} else {
		m.files.Marker = "› "
	}

	inner := comp.Pane{
		Title: "Layers", Focused: m.layers.Focused,
		Border: &m.sty.border, Focus: &m.sty.dirStyle,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, rows[0], comp.Region(regLayers))
	m.layers.DrawFunc(c, inner, len(fake.Layers()), m.layerRow)

	m.detail(c, rows[1])
}

func (m *Model) layerRow(i int) comp.Row {
	l := fake.Layers()[i]
	row := comp.Row{
		Text:  comp.Truncate(l.Command, 40),
		Right: []comp.Segment{{Text: size(l.Size), Style: &m.sty.muted}},
	}
	// A layer whose bytes a later one throws away is the whole point of dive,
	// so it is marked rather than left to the efficiency score at the bottom.
	if l.Wasted > 0 {
		row.Lead, row.LeadStyle = "! ", &m.sty.wasted
	} else {
		row.Lead = "  "
	}
	return row
}

// detail is what the selected layer did, as facts.
func (m *Model) detail(c *comp.Canvas, r comp.Rect) {
	l := m.layer()
	score, wasted := fake.Efficiency()
	blocks := []comp.Block{{Facts: []comp.Fact{
		{Label: "id", Value: l.ID},
		{Label: "size", Value: size(l.Size)},
		{Label: "command", Value: comp.Truncate(l.Command, r.W-14)},
	}}}
	if l.Wasted > 0 {
		blocks[0].Facts = append(blocks[0].Facts,
			comp.Fact{Label: "wasted", Value: size(l.Wasted), Style: &m.sty.wasted})
	}
	blocks = append(blocks, comp.Block{Heading: "image", Facts: []comp.Fact{
		{Label: "efficiency", Value: fmt.Sprintf("%.0f%%", score*100)},
		{Label: "wasted", Value: size(wasted), Style: &m.sty.wasted},
	}})
	inner := comp.Pane{
		Title: "Layer details", Border: &m.sty.border, TitleStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region(regDetail))
	comp.Detail{
		Blocks: blocks, HeadingStyle: &m.sty.dirStyle, LabelStyle: &m.sty.muted,
	}.Draw(c, inner, comp.Region(regDetail))
}

// treePane is the filesystem the selected layer produced.
func (m *Model) treePane(c *comp.Canvas, r comp.Rect) {
	m.files.Focused = m.focus.Is(regTree)
	title := "Current Layer Contents"
	if m.attrs {
		title += " — attributes"
	}
	inner := comp.Pane{
		Title: title, Focused: m.files.Focused,
		Border: &m.sty.border, Focus: &m.sty.dirStyle,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(regTree))

	nodes, rows := fake.Tree(), m.visible()
	m.files.DrawFunc(c, inner, len(rows), func(i int) comp.Row {
		return m.fileRow(nodes[rows[i]])
	})
}

// fileRow is a diff mark, a fold marker, and the name — indented by depth.
//
// The diff mark keeps its own colour under the selection. Row.LeadStyle is for
// exactly this: the row a reader is looking at was otherwise the one row whose
// state they could not read.
func (m *Model) fileRow(n fake.Node) comp.Row {
	marker := "  "
	if n.Dir {
		marker = "▾ "
		if m.tree.IsCollapsed(n.Path) {
			marker = "▸ "
		}
	}
	name := n.Name
	style := (*lipgloss.Style)(nil)
	if n.Dir {
		name += "/"
		style = &m.sty.dirStyle
	}
	// Row.Depth is not used, and the first frame drawn is why: it indents the
	// LEAD as well as the text, so the change marks marched right along with
	// the filenames and stopped being a column you can scan.
	//
	// dive's whole question is "what did this layer remove", which you answer
	// by running your eye down the marks. lazygit's rebuild needed the same
	// thing for git status letters. Two tools, so it is filed as tuikit#66.
	row := comp.Row{
		Lead:      changeMark(n.Chg) + " ",
		LeadStyle: m.changeStyle(n.Chg),
		Text:      strings.Repeat("  ", n.Depth) + marker + name,
		Style:     style,
	}
	if m.attrs {
		row.Right = []comp.Segment{
			{Text: n.Perm + "  ", Style: &m.sty.muted},
			{Text: size(n.Size), Style: &m.sty.muted},
		}
	}
	return row
}

// changeMark is dive's four states as characters rather than as colours.
//
// A character, because a distinction only colour makes is one lost in a pipe
// and to a reader who cannot see it — which harness.ShapeSurvivesColour will
// not catch, since the shape is unchanged and the information is what goes
// missing.
func changeMark(c fake.Change) string {
	switch c {
	case fake.Added:
		return "+"
	case fake.Modified:
		return "~"
	case fake.Removed:
		return "-"
	}
	return " "
}

func (m *Model) changeStyle(c fake.Change) *lipgloss.Style {
	switch c {
	case fake.Added:
		return &m.sty.added
	case fake.Modified:
		return &m.sty.modified
	case fake.Removed:
		return &m.sty.removed
	}
	return &m.sty.muted
}

func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted,
		comp.Hint{Key: "tab", Label: "pane"},
		comp.Hint{Key: "space", Label: "fold"},
		comp.Hint{Key: "a", Label: "attributes"},
		comp.Hint{Key: "?", Label: "keys"},
		comp.Hint{Key: "q", Label: "quit"})
}

// size is bytes as a person would say them.
func size(n int64) string {
	switch {
	case n == 0:
		return "0 B"
	case n < 1000:
		return fmt.Sprintf("%d B", n)
	case n < 1000_000:
		return fmt.Sprintf("%.0f kB", float64(n)/1000)
	case n < 1000_000_000:
		return fmt.Sprintf("%.1f MB", float64(n)/1000_000)
	}
	return fmt.Sprintf("%.1f GB", float64(n)/1000_000_000)
}
