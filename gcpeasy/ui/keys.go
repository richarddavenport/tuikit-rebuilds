package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/engine"
)

// key routes one keystroke, and the ORDER is the contract.
//
// Whatever is capturing gets the key first, then the screen, then the globals.
// A global handled before the filter box is a filter box you cannot type q
// into. app.Keys makes that order the shape of a struct rather than a comment
// somebody deletes.
func (m *Model) key(msg tea.KeyMsg) tea.Cmd {
	return app.Keys{
		Capture: m.capture(),
		Screen:  m.screenKey,
		Global:  m.globalKey,
	}.Route(msg)
}

// capture is whatever is eating keystrokes, or nil.
//
// Three things can, and only one at a time. Order matters here too: the
// palette sits over the help, and the filter is underneath both.
func (m *Model) capture() app.Handled {
	switch {
	case m.picking:
		return m.paletteKey
	case m.help:
		return m.helpKey
	case m.typing:
		return m.filterKey
	}
	return nil
}

func (m *Model) screenKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		m.focus.Next()
	case "shift+tab":
		m.focus.Prev()
	case "j", "down":
		m.moveFocused(1)
	case "k", "up":
		m.moveFocused(-1)
	case "pgdown":
		m.moveFocused(10)
	case "pgup":
		m.moveFocused(-10)
	case "enter":
		return m.activate(), true
	case "l":
		return m.logs(), true
	case "d":
		return m.describe(), true
	case "c":
		return m.console(), true
	case "s":
		return m.shell(), true
	case "/":
		m.typing = true
	case "r":
		// A new generation, so anything still in flight lands in the void.
		m.gen.Next()
		return m.fetch(), true
	case " ":
		m.picking = true
		m.palette.Query, m.palette.Caret = "", 0
	default:
		return nil, false
	}
	return nil, true
}

// globalKey is the four keys tuikit reserves, and nothing else.
//
// spec.Reserved names them and guard.Reserved fails a build that binds one to
// another meaning. It is the only thing tuikit imposes on a tool's keyboard.
func (m *Model) globalKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "?":
		m.help = true
		return nil, true
	case "esc":
		m.err = nil
		return nil, true
	}
	return nil, false
}

func (m *Model) helpKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "j", "down":
		m.keys.Offset++
	case "k", "up":
		m.keys.Offset = max(0, m.keys.Offset-1)
	default:
		m.help, m.keys.Offset = false, 0
	}
	return nil, true
}

func (m *Model) filterKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		m.typing = false
	case "esc":
		m.typing, m.filter = false, ""
		m.list().Reset()
	case "backspace":
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.list().Reset()
		}
	default:
		if len(msg.Runes) == 1 {
			m.filter += string(msg.Runes)
			// The rows changed under the cursor, so it goes back to the top
			// rather than pointing at whatever moved into its place.
			m.list().Reset()
		}
	}
	return nil, true
}

func (m *Model) paletteKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		m.picking = false
	case "enter":
		m.picking = false
		// The palette runs a key rather than an action, which is what makes
		// it teach the keyboard instead of replacing it: whatever it does,
		// the reader has just been shown how to do it without the palette.
		if item, ok := m.palette.Selected(); ok && item.Key != "" {
			return m.key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(item.Key)}), true
		}
	case "up":
		m.palette.Move(-1)
	case "down":
		m.palette.Move(1)
	case "backspace":
		if q := m.palette.Query; q != "" {
			m.palette.Query = q[:len(q)-1]
			m.palette.Caret = len(m.palette.Query)
		}
	default:
		if len(msg.Runes) == 1 {
			m.palette.Query += string(msg.Runes)
			m.palette.Caret = len(m.palette.Query)
		}
	}
	return nil, true
}

// setFocus moves the keyboard between panes.
//
// The list is told, because comp.List uses Focused to decide whether a cursor
// move drags the viewport with it: an unfocused pane whose viewport jumps is a
// pane that scrolled for no reason the reader can see.
func (m *Model) setFocus(p pane) {
	m.focus.Set(p.region())
	// A filter belongs to the pane it was typed in. Carrying it across would
	// silently hide rows in a list the reader never filtered.
	m.filter, m.typing = "", false
}

func (m *Model) moveFocused(by int) {
	m.list().Move(by)
	// Choosing a different project changes which clusters exist, so the
	// cursor below cannot stay where it was pointing.
	if m.at() == paneProjects {
		m.lists[paneClusters].Reset()
	}
}

// activate is enter: it does whatever the focused pane's row means.
func (m *Model) activate() tea.Cmd {
	switch m.at() {
	case paneProjects:
		m.lists[paneClusters].Reset()
		m.gen.Next()
		return m.fetch()
	case paneClusters:
		c, ok := m.cluster()
		if !ok {
			return nil
		}
		m.busy = "pointing kubectl at " + c.Name
		gen := m.gen.Current()
		return func() tea.Msg {
			ctx := context.Background()
			if err := engine.UseCluster(ctx, c); err != nil {
				return outputMsg{gen: gen, title: "use " + c.Name, err: err}
			}
			pods, err := engine.Pods(ctx)
			return snapshotMsg{gen: gen, snap: engine.Snapshot{
				Projects: m.snap.Projects, Clusters: m.snap.Clusters,
				Pods: pods, Fetched: engine.Now(), Partial: err,
			}}
		}
	case panePods:
		return m.describe()
	}
	return nil
}

func (m *Model) logs() tea.Cmd {
	p, ok := m.pod()
	if !ok {
		return nil
	}
	m.setFocus(panePods)
	return m.read("logs "+p.Name, func(ctx context.Context) ([]byte, error) {
		return engine.Logs(ctx, p, 200)
	})
}

func (m *Model) describe() tea.Cmd {
	p, ok := m.pod()
	if !ok {
		return nil
	}
	return m.read("describe "+p.Name, func(ctx context.Context) ([]byte, error) {
		return engine.Describe(ctx, p)
	})
}

// console opens a Rails console in the selected pod.
//
// The interesting one, and the reason this tool needs no terminal emulator.
// The engine builds the command; tea.Exec hands over the real terminal. See
// Model.exec.
func (m *Model) console() tea.Cmd {
	p, ok := m.pod()
	if !ok {
		return nil
	}
	return m.exec(engine.Console(context.Background(), p))
}

func (m *Model) shell() tea.Cmd {
	p, ok := m.pod()
	if !ok {
		return nil
	}
	return m.exec(engine.Shell(context.Background(), p, ""))
}

// helpSections is every binding, by what it acts on.
//
// Grouped by pane rather than alphabetically, because somebody opening `?` is
// looking for what they can do HERE. guard.Keys checks this against the spec
// declaration, so a key that exists without a help entry fails a test.
func helpSections() []comp.KeySection {
	return []comp.KeySection{
		{Name: "moving", Keys: []comp.Hint{
			{Key: "tab", Label: "next pane"},
			{Key: "shift+tab", Label: "previous pane"},
			{Key: "j/k", Label: "move the cursor"},
			{Key: "pgup/pgdn", Label: "move ten"},
			{Key: "/", Label: "filter this pane"},
		}},
		{Name: "pods", Keys: []comp.Hint{
			{Key: "enter", Label: "describe, or use the cluster"},
			{Key: "l", Label: "logs"},
			{Key: "d", Label: "describe"},
			{Key: "c", Label: "rails console, in your terminal"},
			{Key: "s", Label: "shell, in your terminal"},
		}},
		{Name: "everywhere", Keys: []comp.Hint{
			{Key: "space", Label: "everything gcpeasy can do"},
			{Key: "r", Label: "read the world again"},
			{Key: "?", Label: "this"},
			{Key: "esc", Label: "back, or dismiss the error"},
			{Key: "q", Label: "quit"},
		}},
	}
}

// paletteGroups is the same list again, in the shape the palette takes.
//
// The same list, not a second one: both are built from helpSections, so a key
// added in one place appears in both. Two lists maintained beside each other is
// how the two paths to an action come to disagree.
func paletteGroups() []comp.PaletteGroup {
	sections := helpSections()
	groups := make([]comp.PaletteGroup, 0, len(sections))
	for _, s := range sections {
		g := comp.PaletteGroup{Name: s.Name}
		for _, h := range s.Keys {
			g.Items = append(g.Items, comp.PaletteItem{Label: h.Label, Key: h.Key})
		}
		groups = append(groups, g)
	}
	return groups
}
