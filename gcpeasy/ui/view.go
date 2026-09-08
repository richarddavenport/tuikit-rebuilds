package ui

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/engine"
)

// Draw paints one frame.
//
// The model has no View() string. It is handed a canvas and a rect and has
// nowhere else to put anything, which is what makes a golden a record of what
// the components drew rather than of whatever a tool concatenated.
func (m *Model) Draw(c *comp.Canvas, r comp.Rect) {
	m.canvas = c
	// Four bands: header, the rule under it, the body, the footer. One
	// declaration rather than a height computed in four places.
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(1), comp.Length(1), comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(r)

	m.header(c, bands[0])
	comp.Rule{Style: &m.sty.border}.Draw(c, bands[1], comp.Region(regRule))
	m.body(c, bands[2])
	m.footer(c, bands[3])

	// Overlays, innermost last: an error can appear over the help, and the
	// palette sits over everything.
	if m.err != nil {
		m.toast(c, r)
	}
	if m.help {
		m.keys.Draw(c, r, comp.Region(regHelp))
	}
	if m.picking {
		m.palette.Draw(c, comp.Center(r, 64, min(r.H-4, 20)))
	}
}

func (m *Model) header(c *comp.Canvas, r comp.Rect) {
	left := []comp.Segment{{Text: "gcpeasy", Style: &m.sty.title}}
	if p, ok := m.project(); ok {
		left = append(left, comp.Segment{Text: "  " + p.ID, Style: &m.sty.muted})
	}
	right := []comp.Segment{{Text: statusWord(m), Style: statusStyle(m)}}
	comp.Bar{Left: left, Right: right, MinLeft: 12}.Draw(c, r, comp.Region(regHeader))
}

func statusWord(m *Model) string {
	switch {
	case m.busy != "":
		return m.busy
	case m.snap.Healthy() && len(m.snap.Pods) > 0:
		return "all ready"
	case len(m.snap.Pods) == 0:
		return "no pods"
	}
	return "needs attention"
}

func statusStyle(m *Model) *lipgloss.Style {
	switch {
	case m.busy != "":
		return &m.sty.pending
	case m.snap.Healthy() && len(m.snap.Pods) > 0:
		return &m.sty.ready
	}
	return &m.sty.danger
}

// body is the split: three stacked lists on the left, output on the right.
func (m *Model) body(c *comp.Canvas, r comp.Rect) comp.Rect {
	left, right := m.split.Draw(c, r)
	m.panes(c, left)
	m.output(c, right)
	return r
}

// panes stacks the three lists.
//
// The pod list gets Fill and the other two get a share, because pods are what
// there are most of and what the reader is usually looking at. A Min on each
// keeps a short terminal from collapsing one to nothing.
func (m *Model) panes(c *comp.Canvas, r comp.Rect) {
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Fill(1).Min(4).Max(8),
		comp.Fill(1).Min(4).Max(8),
		comp.Fill(3).Min(5),
	}}.Rows(r)

	m.pane(c, bands[0], paneProjects, len(m.projects()), m.projectRow)
	m.pane(c, bands[1], paneClusters, len(m.clusters()), m.clusterRow)
	m.pane(c, bands[2], panePods, len(m.pods()), m.podRow)
}

// pane draws one bordered list.
func (m *Model) pane(c *comp.Canvas, r comp.Rect, p pane, n int, row func(int) comp.Row) {
	title := p.title()
	if m.focus.Is(p.region()) && m.typing {
		title = p.title() + " /" + m.filter
	}
	// Told every frame rather than remembered: comp.Focus holds the answer, and
	// a Focused bool set in one place and read in another is the thing that
	// goes stale.
	m.lists[p].Focused = m.focus.Is(p.region())
	// The focused pane's cursor is ›, the others' is ·. A different CHARACTER
	// rather than a different color, because the border and the title already
	// carry the focus in color and a distinction only color makes is one lost
	// in a pipe and to a reader who cannot see it.
	m.lists[p].Marker = "· "
	if m.lists[p].Focused {
		m.lists[p].Marker = "› "
	}
	inner := comp.Pane{
		Title:      title,
		Focused:    m.focus.Is(p.region()),
		Border:     &m.sty.border,
		Focus:      &m.sty.focused,
		TitleStyle: &m.sty.muted,
		FocusTitle: &m.sty.title,
	}.Draw(c, r, comp.Region(p.region()))
	m.lists[p].DrawFunc(c, inner, n, row)
}

func (m *Model) projectRow(i int) comp.Row {
	p := m.projects()[i]
	row := comp.Row{Text: p.ID, Lead: "  "}
	if p.Active {
		row.Lead, row.LeadStyle = "● ", &m.sty.ready
	}
	return row
}

func (m *Model) clusterRow(i int) comp.Row {
	cl := m.clusters()[i]
	row := comp.Row{
		Text:  cl.Name,
		Right: []comp.Segment{{Text: cl.Location, Style: &m.sty.muted}},
	}
	glyph, style := "●", &m.sty.ready
	if cl.Status != "RUNNING" {
		glyph, style = "◐", &m.sty.pending
	}
	row.Lead, row.LeadStyle = glyph+" ", style
	return row
}

// podRow is a state glyph, the name, and the readiness on the right.
//
// The glyph keeps its own color under the selection. Row.LeadStyle is there
// for exactly this: the row a reader is looking at was otherwise the one row
// whose status they could not read.
func (m *Model) podRow(i int) comp.Row {
	p := m.pods()[i]
	glyph, style := stateLook(p.State(), &m.sty)
	return comp.Row{
		Lead:      glyph + " ",
		LeadStyle: style,
		Text:      p.Name,
		Right: []comp.Segment{
			{Text: p.Ready + "  ", Style: &m.sty.muted},
			{Text: engine.Ago(p.Age), Style: &m.sty.muted},
		},
	}
}

// stateLook is the one place a state becomes a glyph and a color.
//
// One place, so the pod list, the detail pane and any future summary cannot
// disagree about what "degraded" looks like. Every glyph is in the theme's
// allow-list, and guard.Glyphs fails the build if one is not.
func stateLook(s engine.State, sty *styles) (string, *lipgloss.Style) {
	switch s {
	case engine.Ready:
		return "●", &sty.ready
	case engine.Pending:
		return "◐", &sty.pending
	case engine.Degraded:
		return "▲", &sty.pending
	}
	return "✗", &sty.danger
}

// output is the right-hand pane: whatever was last read, or the selected pod.
func (m *Model) output(c *comp.Canvas, r comp.Rect) {
	title := m.outTitle
	if title == "" {
		title = "detail"
	}
	inner := comp.Pane{
		Title:      title,
		Focused:    false,
		Border:     &m.sty.border,
		TitleStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region(regOutput))

	if len(m.outLines) > 0 {
		m.out.Draw(c, inner, m.outLines)
		return
	}
	m.detail(c, inner)
}

// detail is what the pane says when nothing has been read yet.
func (m *Model) detail(c *comp.Canvas, r comp.Rect) {
	p, ok := m.pod()
	if !ok {
		c.Text(r.X, r.Y, "  select a pod", &m.sty.muted, comp.Region(regDetail))
		return
	}
	facts := []comp.Fact{
		{Label: "namespace", Value: p.Namespace},
		{Label: "state", Value: p.State().String(), Style: stateStyle(p.State(), &m.sty)},
		{Label: "phase", Value: p.Phase},
		{Label: "ready", Value: p.Ready},
		{Label: "restarts", Value: itoa(p.Restarts)},
		{Label: "age", Value: engine.Ago(p.Age)},
		{Label: "node", Value: p.Node},
	}
	blocks := []comp.Block{{Facts: facts}}
	if len(p.Images) > 0 {
		images := make([]comp.Fact, 0, len(p.Images))
		for _, img := range p.Images {
			images = append(images, comp.Fact{Label: "image", Value: img})
		}
		blocks = append(blocks, comp.Block{Heading: "containers", Facts: images})
	}
	comp.Detail{
		Title:        p.Name,
		Blocks:       blocks,
		TitleStyle:   &m.sty.title,
		HeadingStyle: &m.sty.focused,
		LabelStyle:   &m.sty.muted,
	}.Draw(c, r, comp.Region(regDetail))
}

func stateStyle(s engine.State, sty *styles) *lipgloss.Style {
	_, style := stateLook(s, sty)
	return style
}

// footer says what acts on THIS, which is why it changes with the focus. The
// `?` sheet says what exists.
func (m *Model) footer(c *comp.Canvas, r comp.Rect) {
	hints := []comp.Hint{{Key: "tab", Label: "pane"}, {Key: "/", Label: "filter"}}
	switch m.at() {
	case panePods:
		hints = append(hints,
			comp.Hint{Key: "l", Label: "logs"},
			comp.Hint{Key: "c", Label: "console"},
			comp.Hint{Key: "s", Label: "shell"})
	case paneClusters:
		hints = append(hints, comp.Hint{Key: "enter", Label: "use"})
	}
	hints = append(hints,
		comp.Hint{Key: "space", Label: "all"},
		comp.Hint{Key: "?", Label: "keys"},
		comp.Hint{Key: "q", Label: "quit"})

	// A footer clipped mid-word says nothing about the key it was describing,
	// and `?` opens the full sheet anyway. So on a narrow terminal the middle
	// is dropped rather than the end: the last three work everywhere and are
	// the ones somebody stuck needs.
	for len(hints) > 3 && comp.Width(comp.Hints(hints...))+2 > r.W {
		hints = append(hints[:2], hints[3:]...)
	}
	comp.KeyHints(c, r, comp.Region(regFooter), &m.sty.muted, hints...)
}

// toast is what the interface has to say about a failure.
func (m *Model) toast(c *comp.Canvas, r comp.Rect) {
	comp.Toast{
		Title:     "that did not work",
		Body:      m.err.Error(),
		Hint:      "esc dismisses this. r reads the world again.",
		Accent:    &m.sty.danger,
		Border:    &m.sty.border,
		BodyStyle: &m.sty.muted,
		HintStyle: &m.sty.muted,
	}.Draw(c, r, comp.Region(regToast))
}

// itoa is strconv.Itoa, named short because it appears inside row building
// where the line length is the readability.
func itoa(n int) string { return strconv.Itoa(n) }
