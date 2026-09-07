package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/theme"
)

// The vocabulary this interface is allowed to draw in, checked by guard_test.go.
//
// htop's equivalent is CRT.c, 1,531 lines, which holds eight colour schemes and
// two glyph sets — one UTF-8 and one ASCII, picked at runtime from the locale.
// This is the same decision expressed as an allow-list instead of a table.
var (
	Palette = theme.Default
	Glyphs  = htopGlyphs()
	Chrome  = theme.DefaultChrome
)

// htopGlyphs adds the characters two features here need and the default set
// does not carry.
//
// Worth reading as evidence that the guard works. Both additions were found by
// the guard failing, not by looking at a screenshot — and both are deliberate
// in the way theme.GlyphSet.With asks for, with a reason recorded beside each
// character rather than a blanket exemption.
func htopGlyphs() theme.GlyphSet {
	return theme.DefaultGlyphs.With(
		// The tree. DefaultGlyphs carries ┌ ─ ┐ │ └ ┘ for a box, which is a
		// rectangle and needs no junctions. A tree needs the tee.
		'├', "tree branch with more siblings below",
		'┤', "LED digit segment",
		// The LED meter's seven-segment digits, from htop's
		// LEDMeterMode_digitsUtf8. These are the half-width box pieces, which
		// a rectangle never needs and a digit does.
		'╶', "LED digit segment — left half missing",
		'╴', "LED digit segment — right half missing",
		'╷', "LED digit segment — top half missing",
		'╵', "LED digit segment — bottom half missing",
	)
}

type styles struct {
	title, muted, header, border  lipgloss.Style
	selected, key, label          lipgloss.Style
	cool, warm, hot, danger, tree lipgloss.Style
	fnKey, fnLabel, meterCap      lipgloss.Style
}

func newStyles(p theme.Palette) styles {
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		header:   lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		key:      lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		label:    lipgloss.NewStyle().Foreground(p.Muted),
		cool:     lipgloss.NewStyle().Foreground(p.Success),
		warm:     lipgloss.NewStyle().Foreground(p.Pending),
		hot:      lipgloss.NewStyle().Foreground(p.Danger),
		danger:   lipgloss.NewStyle().Foreground(p.Danger).Bold(true),
		tree:     lipgloss.NewStyle().Foreground(p.Border),
		fnKey:    lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG),
		fnLabel:  lipgloss.NewStyle().Foreground(p.Muted),
		meterCap: lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
	}
}

// heat is the one place a number becomes a colour, so every meter and every
// row agrees about what busy looks like.
func (m *Model) heat(v float64) *lipgloss.Style {
	switch {
	case v >= 80:
		return &m.sty.hot
	case v >= 50:
		return &m.sty.warm
	}
	return &m.sty.cool
}
