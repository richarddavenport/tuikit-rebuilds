// Package ui is gitui's interface, rebuilt on tuikit.
//
// Tabs across the top, two panes under them, and a stack of popups over
// everything. gitui is lazygit's control — same subject, different team — and
// its study found the one thing lazygit's did not: a modal stack is not a
// screen stack.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"
	"github.com/richarddavenport/tuikit/theme"

	"github.com/richarddavenport/tuikit-rebuilds/gitui/fake"
)

// Regions, named once.
const (
	regTabs     comp.Name = "tabs"
	regUnstaged comp.Name = "unstaged"
	regUnstageR comp.Name = "unstaged.row"
	regStaged   comp.Name = "staged"
	regStagedR  comp.Name = "staged.row"
	regLog      comp.Name = "log"
	regLogRow   comp.Name = "log.row"
	regStash    comp.Name = "stash"
	regStashRow comp.Name = "stash.row"
	regDiff     comp.Name = "diff"
	regSplit    comp.Name = "split"
	regFooter   comp.Name = "footer"
)

// popup is one thing layered over the screen.
//
// # The gap this rebuild is about
//
// gitui has 32 popup files with a consistent open, close and stack discipline.
// app.Stack answers "where am I and how did I get here" for SCREENS. It is not
// about a stack of things drawn over the current screen, and tuikit draws
// overlays as an ordered if-chain in Draw — fine for three, not for thirty-two.
//
// So this is written by hand, in one place, deliberately. It is what a
// comp.Modals would replace.
type popup int

// The popups this rebuild has.
const (
	noPopup popup = iota
	popupPalette
	popupHelp
	popupConfirm
)

// Model is the whole interface.
type Model struct {
	sty styles

	tab      fake.Tab
	unstaged comp.List
	staged   comp.List
	log      comp.List
	stash    comp.List
	focus    comp.Focus

	diff    comp.Viewer
	staging bool

	// stack is the popups that are open, innermost last. Esc pops one.
	stack []popup

	palette comp.Palette
	keys    comp.Keys
	confirm string

	split comp.Split
	mouse app.Mouse

	width, height int
	canvas        *comp.Canvas
}

type styles struct {
	title, muted, border, selected, ranged lipgloss.Style
	added, removed, hunk, staged, dirty    lipgloss.Style
}

func newStyles() styles {
	p := theme.Default
	return styles{
		title:    lipgloss.NewStyle().Foreground(p.Accent).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(p.Muted),
		border:   lipgloss.NewStyle().Foreground(p.Border),
		selected: lipgloss.NewStyle().Foreground(p.SelectionFG).Background(p.SelectionBG).Bold(true),
		// A background only, so a diff line inside the range keeps its own
		// added or removed color. comp.Viewer puts the line style UNDER the
		// spans for exactly this.
		ranged:  lipgloss.NewStyle().Background(p.SelectionBG),
		added:   lipgloss.NewStyle().Foreground(p.Success),
		removed: lipgloss.NewStyle().Foreground(p.Danger),
		hunk:    lipgloss.NewStyle().Foreground(p.Accent),
		staged:  lipgloss.NewStyle().Foreground(p.Success),
		dirty:   lipgloss.NewStyle().Foreground(p.Pending),
	}
}

// New builds it.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles()}
	base := comp.List{
		Marker: "› ", Blank: "  ",
		Selected: &m.sty.selected, Unfocused: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, NoStatus: true,
		Empty: "  (none)",
	}
	m.unstaged, m.staged, m.log, m.stash = base, base, base, base
	m.unstaged.Name, m.staged.Name = regUnstageR, regStagedR
	m.log.Name, m.stash.Name = regLogRow, regStashRow
	m.focus.Ring = []comp.Name{regUnstaged, regStaged}

	m.diff = comp.Viewer{
		Name: regDiff, Tab: 4, NoCursor: true,
		Selected: &m.sty.selected, Ranged: &m.sty.ranged,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted, Empty: "  no changes",
	}
	m.split = comp.Split{Name: regSplit, Ratio: [2]int{2, 5}, Min: 26}

	items := make([]comp.PaletteItem, 0, len(fake.Popups()))
	for _, p := range fake.Popups() {
		items = append(items, comp.PaletteItem{Label: p.Label, Key: p.Key})
	}
	m.palette = comp.Palette{
		Name: "palette", Item: "palette.item", Title: "gitui",
		Groups:      []comp.PaletteGroup{{Name: "popups", Note: "gitui has 32 of these", Items: items}},
		Border:      &m.sty.border,
		TitleStyle:  &m.sty.title,
		PromptStyle: &m.sty.hunk,
		QueryStyle:  &m.sty.title,
		GroupStyle:  &m.sty.hunk,
		LabelStyle:  &m.sty.muted,
		KeyStyle:    &m.sty.muted,
		MatchStyle:  &m.sty.title,
		SelectedFG:  &m.sty.selected,
		Cursor:      &m.sty.title,
	}
	m.keys = comp.Keys{
		Overlay: true, Title: "gitui, rebuilt",
		Sections: []comp.KeySection{
			{Name: "tabs", Keys: []comp.Hint{
				{Key: "1-4", Label: "jump to a tab"}, {Key: "tab", Label: "next tab"},
			}},
			{Name: "status", Keys: []comp.Hint{
				{Key: "j/k", Label: "move"}, {Key: "shift+tab", Label: "staged or unstaged"},
				{Key: "v", Label: "select lines of the diff"},
			}},
			{Name: "line select", Keys: []comp.Hint{
				{Key: "J/K", Label: "extend"}, {Key: "a", Label: "the whole hunk"},
				{Key: "enter", Label: "stage the range"},
			}},
			{Name: "everywhere", Keys: []comp.Hint{
				{Key: "space", Label: "every popup"}, {Key: "?", Label: "this"},
				{Key: "esc", Label: "close one popup"}, {Key: "q", Label: "quit"},
			}},
		},
		TitleStyle: &m.sty.title, SectionStyle: &m.sty.hunk,
		KeyStyle: &m.sty.title, LabelStyle: &m.sty.muted, Border: &m.sty.border,
	}
	return m
}

// SetSize is what a capture calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Init has nothing to do: the repository is a fixture.
func (m *Model) Init() tea.Cmd { return nil }

// top is the innermost popup, or none.
func (m *Model) top() popup {
	if len(m.stack) == 0 {
		return noPopup
	}
	return m.stack[len(m.stack)-1]
}

func (m *Model) push(p popup) { m.stack = append(m.stack, p) }

// pop closes one popup, which is what esc does. One, not all: that distinction
// is the reason a stack is a stack.
func (m *Model) pop() {
	if len(m.stack) > 0 {
		m.stack = m.stack[:len(m.stack)-1]
	}
}

// changes splits the working tree the way gitui's status tab does.
func (m *Model) changes(staged bool) []fake.Change {
	var out []fake.Change
	for _, c := range fake.Changes() {
		if c.Staged == staged {
			out = append(out, c)
		}
	}
	return out
}

// list is whichever list has the keyboard on the current tab.
func (m *Model) list() *comp.List {
	switch m.tab {
	case fake.Log:
		return &m.log
	case fake.Stashing:
		return &m.stash
	}
	if m.focus.Is(regStaged) {
		return &m.staged
	}
	return &m.unstaged
}
