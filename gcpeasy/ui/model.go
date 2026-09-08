// Package ui is gcpeasy's interface.
package ui

import (
	"context"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/engine"
)

// Model is gcpeasy's whole state, and it is a POINTER.
//
// tuikit's components own state — where a cursor is, what a viewport shows,
// where a divider sits — and state inside a component cannot survive being
// copied on every message. A value model compiles, runs, and silently forgets
// every scroll.
type Model struct {
	snap engine.Snapshot
	sty  styles

	// One list per pane, plus the pane that has the keyboard. See focus.go for
	// why the last of those is the tool's business rather than tuikit's.
	lists [paneCount]comp.List
	// focus holds which pane has the keyboard. comp.Focus keys by region name
	// rather than by index, so a pane added to the ring cannot silently move it.
	focus comp.Focus

	// out is the right-hand pane: logs, or a describe, or an error. A Viewer
	// rather than a LogPane because none of these tail — a describe opens at
	// the top and a log you asked for once has a bottom.
	out      comp.Viewer
	outLines []comp.Line
	outTitle string

	split comp.Split
	mouse app.Mouse
	gen   app.Gen

	// filter narrows the focused list. typing is the mode split: while it is
	// true, j is the letter j. app.Keys makes that the shape of the routing
	// rather than an early return somebody can delete.
	filter string
	typing bool

	keys    comp.Keys
	help    bool
	palette comp.Palette
	picking bool

	err  error
	busy string

	width, height int
	canvas        *comp.Canvas
}

// New builds the model.
func New() *Model {
	m := &Model{width: 132, height: 38, sty: newStyles(Palette)}
	for p := paneProjects; p < paneCount; p++ {
		m.lists[p] = comp.List{
			Name:       rowRegion(p),
			Empty:      "  nothing yet",
			Blank:      "  ",
			Selected:   &m.sty.selected,
			Unfocused:  &m.sty.focused,
			Status:     &m.sty.muted,
			EmptyStyle: &m.sty.muted,
			// Three lists stacked in one column. comp.List's own doc says this
			// is exactly when to give the counter row back: at 80x24 three
			// status rows reading 3/3 are a tenth of the column, next to
			// titles that already say the number.
			NoStatus: true,
		}
	}
	m.focus.Ring = ring()
	m.out = comp.Viewer{
		Name: regOutput, Numbers: true, Tab: 4,
		Selected: &m.sty.selected, Number: &m.sty.muted,
		Status: &m.sty.muted, EmptyStyle: &m.sty.muted,
		Empty: "  nothing selected",
	}
	// Two fifths to the lists, three to what you are looking at. Half each
	// looked right until a frame was drawn: the left column had room to spare
	// and the detail pane was the one running out.
	m.split = comp.Split{Name: regSplit, Ratio: [2]int{2, 5}, Min: 22}
	m.keys = comp.Keys{
		Overlay: true, Title: "gcpeasy",
		Sections:     helpSections(),
		TitleStyle:   &m.sty.title,
		SectionStyle: &m.sty.focused,
		KeyStyle:     &m.sty.title,
		LabelStyle:   &m.sty.muted,
		Border:       &m.sty.border,
	}
	m.palette = comp.Palette{
		Name: regPalette, Item: regPalette + ".item", Title: "do",
		Groups:      paletteGroups(),
		Border:      &m.sty.border,
		TitleStyle:  &m.sty.title,
		PromptStyle: &m.sty.focused,
		QueryStyle:  &m.sty.title,
		GroupStyle:  &m.sty.focused,
		LabelStyle:  &m.sty.muted,
		KeyStyle:    &m.sty.muted,
		MatchStyle:  &m.sty.title,
		SelectedFG:  &m.sty.selected,
		Cursor:      &m.sty.title,
	}
	return m
}

func rowRegion(p pane) comp.Name {
	switch p {
	case paneProjects:
		return regProjectsRow
	case paneClusters:
		return regClustersRow
	}
	return regPodsRow
}

// SetSize is what a test calls instead of waiting for a terminal.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Canvas is the last frame, so a capture script can address a region by name.
func (m *Model) Canvas() *comp.Canvas { return m.canvas }

// Load puts a snapshot in, and on the first one puts the cursor on the project
// gcloud is already configured for.
//
// Only the first, because after that the cursor is the reader's. A refresh that
// moved it back would undo the last thing they did, every thirty seconds.
//
// Without this the tool opens on whatever sorted first, and the clusters pane
// is empty because the clusters that were fetched belong to a different
// project — which looks like a tool that cannot see your clusters.
func (m *Model) Load(s engine.Snapshot) {
	first := len(m.snap.Projects) == 0
	m.snap = s
	if !first {
		return
	}
	for i, p := range s.Projects {
		if p.Active {
			m.lists[paneProjects].Select(i)
		}
	}
}

// --- what is on screen -------------------------------------------------

// projects, clusters and pods are the rows each pane is showing, after the
// filter. Every index in a list refers to one of these slices.
func (m *Model) projects() []engine.Project { return m.snap.Projects }

func (m *Model) clusters() []engine.Cluster {
	p, ok := m.project()
	if !ok {
		return nil
	}
	var out []engine.Cluster
	for _, c := range m.snap.Clusters {
		if c.Project == p.ID {
			out = append(out, c)
		}
	}
	return out
}

// pods hides the cluster's own machinery unless the filter asks for it.
//
// Pod.App is the judgement and it lives in the engine, because which
// namespaces are Kubernetes' own is a fact about Kubernetes.
func (m *Model) pods() []engine.Pod {
	var out []engine.Pod
	for _, p := range m.snap.Pods {
		if !p.App() && !strings.HasPrefix(m.filter, "kube") {
			continue
		}
		if m.at() == panePods && m.filter != "" && !podMatches(p, m.filter) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func podMatches(p engine.Pod, filter string) bool {
	f := strings.ToLower(filter)
	return strings.Contains(strings.ToLower(p.Name), f) ||
		strings.Contains(strings.ToLower(p.Namespace), f)
}

// project, cluster and pod are what each cursor is on, and whether there is
// anything there.
//
// Two return values rather than an index, because an empty list is an ordinary
// state and every caller should handle it instead of indexing into it.
func (m *Model) project() (engine.Project, bool) {
	list := m.snap.Projects
	if len(list) == 0 {
		return engine.Project{}, false
	}
	return list[min(m.lists[paneProjects].Cursor(), len(list)-1)], true
}

func (m *Model) cluster() (engine.Cluster, bool) {
	list := m.clusters()
	if len(list) == 0 {
		return engine.Cluster{}, false
	}
	return list[min(m.lists[paneClusters].Cursor(), len(list)-1)], true
}

func (m *Model) pod() (engine.Pod, bool) {
	list := m.pods()
	if len(list) == 0 {
		return engine.Pod{}, false
	}
	return list[min(m.lists[panePods].Cursor(), len(list)-1)], true
}

// --- reading the world -------------------------------------------------

type snapshotMsg struct {
	gen  int
	snap engine.Snapshot
	err  error
}

type outputMsg struct {
	gen   int
	title string
	body  string
	err   error
}

type doneMsg struct{ err error }

// Init starts the first read.
func (m *Model) Init() tea.Cmd { return m.fetch() }

// fetch reads projects, then the clusters of the current project, then pods.
//
// app.Gen drops a result from a read the user walked away from. The bug it
// prevents is nasty because nothing errors: a stale result draws into the
// screen that replaced it, and nothing says why.
//
// A failure part-way through returns what it has, in Partial. A cluster that
// cannot be reached must not blank the project list beside it — that is
// swarmctl's rule and it is the difference between a degraded screen and an
// empty one.
func (m *Model) fetch() tea.Cmd {
	gen := m.gen.Current()
	want, _ := m.project()
	return func() tea.Msg {
		ctx := context.Background()
		var snap engine.Snapshot
		snap.Fetched = engine.Now()

		projects, err := engine.Projects(ctx)
		if err != nil {
			return snapshotMsg{gen: gen, err: err}
		}
		snap.Projects = projects

		id := want.ID
		if id == "" {
			for _, p := range projects {
				if p.Active {
					id = p.ID
				}
			}
		}
		if id != "" {
			clusters, err := engine.Clusters(ctx, id)
			if err != nil {
				snap.Partial = err
				return snapshotMsg{gen: gen, snap: snap}
			}
			snap.Clusters = clusters
		}
		pods, err := engine.Pods(ctx)
		if err != nil {
			snap.Partial = err
			return snapshotMsg{gen: gen, snap: snap}
		}
		snap.Pods = pods
		return snapshotMsg{gen: gen, snap: snap}
	}
}

// read runs one thing and puts its output in the right-hand pane.
func (m *Model) read(title string, f func(context.Context) ([]byte, error)) tea.Cmd {
	gen := m.gen.Current()
	m.busy = title
	return func() tea.Msg {
		out, err := f(context.Background())
		return outputMsg{gen: gen, title: title, body: string(out), err: err}
	}
}

// Update handles one message.
func (m *Model) Update(msg tea.Msg) (app.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case snapshotMsg:
		if m.gen.Stale(msg.gen) {
			return m, nil
		}
		m.snap, m.err = msg.snap, msg.err
		if msg.snap.Partial != nil {
			m.err = msg.snap.Partial
		}
		m.busy = ""

	case outputMsg:
		if m.gen.Stale(msg.gen) {
			return m, nil
		}
		m.busy = ""
		m.err = msg.err
		m.outTitle = msg.title
		m.setOutput(msg.body)

	case doneMsg:
		// The console or the shell has exited and the terminal is ours again.
		// Its own output already went to the real screen, so there is nothing
		// to show — only whether it failed, and a refresh, because whatever
		// they did in there may have changed the world.
		m.err = msg.err
		m.gen.Next()
		return m, m.fetch()

	case tea.MouseMsg:
		return m, m.onMouse(msg)

	case tea.KeyMsg:
		return m, m.key(msg)
	}
	return m, nil
}

// setOutput puts text in the right-hand pane, one comp.Line per line.
//
// Plain lines for now. The engine's output arrives with ANSI in it when the
// underlying tool feels like coloring — tuikit issue 63 is the component that
// would turn that into spans, and until it exists this shows the text and
// loses the color rather than printing the escapes.
func (m *Model) setOutput(body string) {
	m.outLines = nil
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		m.outLines = append(m.outLines, comp.Line{Text: line})
	}
	m.out.Goto(0)
}

// exec hands the real terminal to a process and takes it back when it exits.
//
// This is the whole answer to "how does a TUI open a Rails console", and it is
// why gcpeasy needs no terminal emulator. tea.Exec suspends the program, the
// child gets the actual TTY — real line editing, real color, Ctrl-C reaching
// Ruby rather than us — and the interface is restored on exit.
//
// The engine decides WHAT to run and hands back an *exec.Cmd. This decides HOW,
// and how is a terminal question, so it lives here.
func (m *Model) exec(cmd *exec.Cmd) tea.Cmd {
	return tea.ExecProcess(cmd, func(err error) tea.Msg { return doneMsg{err: err} })
}

// Run opens gcpeasy's interface. The model is passed in rather than built
// here, because the caller may want to capture it instead of showing it.
func Run(m *Model) error {
	// The runner owns the canvas, its size, its chrome and its pixel layer.
	p := tea.NewProgram(
		app.New(m, app.WithChrome(Chrome), app.WithPixels(comp.Detect())),
		tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
