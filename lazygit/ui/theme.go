package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/theme"
)

// styles is every style this rebuild draws with, built from tuikit's palette.
//
// Nine roles, named by what they are FOR. lazygit has a hand-rolled equivalent
// in pkg/gui/style, which is the study's point: the tool built the thing the
// framework should have supplied.
type styles struct {
	title    lipgloss.Style
	accent   lipgloss.Style
	muted    lipgloss.Style
	plain    lipgloss.Style
	border   lipgloss.Style
	selected lipgloss.Style
	ranged   lipgloss.Style
	added    lipgloss.Style
	removed  lipgloss.Style
	staged   lipgloss.Style
	dirty    lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		accent:   lipgloss.NewStyle().Foreground(p.Accent),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		plain:    lipgloss.NewStyle(),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		// The range is a background only, so a diff line inside it keeps its
		// own added/removed colour. comp.Viewer puts the line style UNDER the
		// spans for exactly this.
		ranged:  lipgloss.NewStyle().Background(p.SelectionBG),
		added:   lipgloss.NewStyle().Foreground(p.Success),
		removed: lipgloss.NewStyle().Foreground(p.Danger),
		staged:  lipgloss.NewStyle().Foreground(p.Success),
		dirty:   lipgloss.NewStyle().Foreground(p.Pending),
	}
}
