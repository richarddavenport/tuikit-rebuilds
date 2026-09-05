package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/theme"
)

// styles is every style this rebuild draws with, from tuikit's nine roles.
//
// JSON needs a colour per scalar type, which is four of the nine, and that is
// worth noticing: a syntax highlighter usually wants more roles than an
// operator tool. theme.Palette.Extra is the escape hatch if it ever needs a
// fifth, and it is not needed here.
type styles struct {
	title, muted, border, selected lipgloss.Style
	key, str, num, boolean, null   lipgloss.Style
	hit                            lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:  lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:  lipgloss.NewStyle().Foreground(p.Muted),
		border: lipgloss.NewStyle().Foreground(p.Border),
		// A background only, so a folded line under the cursor keeps the
		// colours that say what its values are. That is the rule comp.Viewer
		// exists for.
		selected: lipgloss.NewStyle().Background(p.SelectionBG),
		key:      lipgloss.NewStyle().Foreground(p.Accent),
		str:      lipgloss.NewStyle().Foreground(p.Success),
		num:      lipgloss.NewStyle().Foreground(p.Pending),
		boolean:  lipgloss.NewStyle().Foreground(p.Danger),
		null:     lipgloss.NewStyle().Foreground(p.Muted),
		hit:      lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.Accent),
	}
}

// Glyphs is tuikit's set plus the two fold markers.
var Glyphs = theme.DefaultGlyphs
