package ui

import (
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/htop/fake"
)

// Setup is htop's F2, and the reason htop is in this repository.
//
// # What it is
//
// A screen that edits the interface. The left pane holds the meters currently
// in the header, in order, each showing which of the four modes it is drawn in.
// The right pane holds every meter that exists. Enter adds one, d removes one,
// space cycles a mode, L changes the column split. htop then writes the result
// to ~/.config/htop/htoprc and the next run starts with it.
//
// htop spends 1,412 lines on this across five files: MetersPanel.c,
// AvailableMetersPanel.c, ColumnsPanel.c, ScreensPanel.c, ScreenTabsPanel.c.
//
// # Why it matters more than a config file
//
// bottom and gh-dash let a user arrange the interface in a file, read once at
// startup. htop lets a user arrange it in the interface, and that needs the
// arrangement to be a value the program can MUTATE and then serialise —
// not just one it can parse.
//
// # And why the guards survive it
//
// The available list is closed. A reader picks from fake.MeterNames() and
// cannot type a name, so no arrangement can name a meter the program does not
// have a drawing for. That is a stronger guarantee than parsing a user's TOML,
// which is the thing tuikit#60 was stuck on.

func (m *Model) drawSetup(c *comp.Canvas, r comp.Rect) {
	bands := comp.Layout{Constraints: []comp.Constraint{
		comp.Length(2), comp.Fill(1).Min(4), comp.Length(1),
	}}.Rows(r)

	m.setupTitle(c, bands[0])

	cols := comp.Layout{Constraints: []comp.Constraint{
		comp.Percent(50), comp.Percent(50),
	}, Gap: 1}.Cols(bands[1])

	m.activeMeters(c, cols[0])
	m.availableMeters(c, cols[1])
	m.drawFnBar(c, bands[2])
}

func (m *Model) setupTitle(c *comp.Canvas, r comp.Rect) {
	c.Text(r.X, r.Y, "Setup", &m.sty.title, comp.Region(regSetup))
	note := "layout: " + Layouts[m.layout].Name + "   —   the header is data, and this is the editor for it"
	c.Text(r.X+7, r.Y, comp.Truncate(note, max(0, r.W-8)), &m.sty.muted, comp.Region(regSetup))
	comp.Rule{Style: &m.sty.border}.Draw(c, comp.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: 1}, comp.Region(regSetup))
}

// activeMeters is the left pane: what is in the header now, and how each is
// drawn.
func (m *Model) activeMeters(c *comp.Canvas, r comp.Rect) {
	focused := m.focus.Is(regSetupL)
	inner := comp.Pane{Title: "Header meters", Focused: focused,
		Border: &m.sty.border, Focus: &m.sty.title,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title}.
		Draw(c, r, comp.Region(regSetupL))

	// The left pane shows column 1 only, which is what htop's MetersPanel does
	// — one column at a time, with the column picked separately.
	for i, s := range m.columns[0] {
		if i >= inner.H {
			break
		}
		style, mark := &m.sty.muted, "  "
		if focused && i == m.setupSel[0] {
			style, mark = &m.sty.selected, "> "
			c.Fill(comp.Rect{X: inner.X, Y: inner.Y + i, W: inner.W, H: 1}, " ", style, comp.Region(regSetupL).At(i))
		}
		c.Text(inner.X, inner.Y+i, mark+comp.Truncate(s.Name, inner.W-12), style, comp.Region(regSetupL).At(i))
		mode := "[" + s.Mode.String() + "]"
		c.Text(inner.X+inner.W-comp.Width(mode)-1, inner.Y+i, mode, style, comp.Region(regSetupL).At(i))
	}
}

// availableMeters is the right pane: every meter that exists.
//
// A closed list. That is the sentence the guards depend on.
func (m *Model) availableMeters(c *comp.Canvas, r comp.Rect) {
	focused := m.focus.Is(regSetupR)
	inner := comp.Pane{Title: "Available meters", Focused: focused,
		Border: &m.sty.border, Focus: &m.sty.title,
		TitleStyle: &m.sty.muted, FocusTitle: &m.sty.title}.
		Draw(c, r, comp.Region(regSetupR))

	for i, name := range fake.MeterNames() {
		if i >= inner.H {
			break
		}
		style, mark := &m.sty.muted, "  "
		if focused && i == m.setupSel[1] {
			style, mark = &m.sty.selected, "> "
			c.Fill(comp.Rect{X: inner.X, Y: inner.Y + i, W: inner.W, H: 1}, " ", style, comp.Region(regSetupR).At(i))
		}
		c.Text(inner.X, inner.Y+i, mark+comp.Truncate(name, inner.W-3), style, comp.Region(regSetupR).At(i))
	}
}
