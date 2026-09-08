package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/theme"
)

// Palette, Glyphs and Chrome are gcpeasy's vocabulary.
//
// tuikit's defaults unchanged, which is the right place to start: nine color
// roles named by what they are FOR, an allow-list of the non-ASCII characters
// the interface may print, and the box it draws with.
//
// The roles are the terminal's own sixteen ANSI indices — the only colors a
// theme can redefine — so Gcpeasy is themed by whatever themed the terminal,
// with nothing to configure and nothing to reload. Override a role and you are
// overriding the reader's choice, so do it deliberately. Override what you
// disagree with; do not build one from scratch, because the point of a closed
// set is that nine decisions is the whole vocabulary.
//
// guard_test.go holds all three closed. A raw color or an unlisted glyph is a
// test failure, not something to notice in review.
var (
	Palette = theme.Default
	Glyphs  = stateGlyphs()
	Chrome  = theme.DefaultChrome
)

// stateGlyphs is tuikit's set plus the two gcpeasy adds.
//
// Added because guard.Glyphs failed the build for them, which is the point of
// the guard: a character reaches the screen only after somebody decided it
// should. Both are in the Geometric Shapes block, which every font shipped with
// a terminal has had for twenty years — the same block ● already comes from.
//
// The alternative was to reuse ● for every state and let color carry the
// difference. That is refused for the reason harness.ShapeSurvivesColor
// exists: a distinction only color makes is a distinction lost in a pipe, in a
// golden, and to a reader who cannot see it.
func stateGlyphs() theme.GlyphSet {
	g := theme.GlyphSet{}
	for r, name := range theme.DefaultGlyphs {
		g[r] = name
	}
	g['◐'] = "half-filled circle, for something still starting"
	g['▲'] = "triangle, for something running but not right"
	return g
}

// styles is every style gcpeasy draws with, built once from the palette.
//
// Built rather than declared at package level, so a different palette produces
// a different interface without touching this file — which is what makes the
// palette a knob rather than a suggestion.
type styles struct {
	title    lipgloss.Style
	muted    lipgloss.Style
	border   lipgloss.Style
	focused  lipgloss.Style
	selected lipgloss.Style
	ready    lipgloss.Style
	pending  lipgloss.Style
	danger   lipgloss.Style
}

func newStyles(p theme.Palette) styles {
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		focused:  lipgloss.NewStyle().Foreground(p.Accent),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		ready:    lipgloss.NewStyle().Foreground(p.Success),
		pending:  lipgloss.NewStyle().Foreground(p.Pending),
		danger:   lipgloss.NewStyle().Foreground(p.Danger),
	}
}
