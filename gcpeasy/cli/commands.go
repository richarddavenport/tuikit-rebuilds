// Package cli is gcpeasy's command tree, declared ONCE.
//
// The CLI, the TUI screen each command opens, the entry an agent reads in
// `gcpeasy describe --json`, and the right-click menu for the regions a
// command targets all come from this one declaration. Not four descriptions
// kept in step — one, read four ways.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"
	"github.com/richarddavenport/tuikit/spec"

	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/engine"
	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/ui"
)

// Commands is gcpeasy, declared once.
func Commands() spec.Command {
	return spec.Command{
		Name:  "gcpeasy",
		Short: "Pick a GCP project, cluster and pod, then act on them.",
		Commands: []spec.Command{
			{
				Name:   "browse",
				Short:  "Open the interface",
				Screen: "browse",
				Flags: []spec.Flag{
					// So the interface can be opened, driven and captured
					// without a cloud account. An agent asked to look at a
					// screen should not first need credentials, and a
					// contributor should be able to see what they are changing.
					{Name: "fixture", Kind: spec.Bool, Help: "run against canned data instead of GCP"},
				},
				Run: browse,
			},
			{
				Name:  "projects",
				Short: "List the projects you can see",
				// This row's context menu, and the key that does the same
				// thing from the keyboard. A Target without a Key fails
				// guard.Reachable: an agent cannot click, and a multiplexer
				// may eat the right-click before gcpeasy sees it.
				Target: ui.RegProjectsRow,
				Key:    "enter",
				Run:    projects,
			},
			{
				Name:   "clusters",
				Short:  "List the GKE clusters in a project",
				Args:   []spec.Arg{{Name: "project", Complete: projectIDs}},
				Target: ui.RegClustersRow,
				Key:    "enter",
				Run:    clusters,
			},
			{
				Name:  "pods",
				Short: "List the pods kubectl can see",
				Flags: []spec.Flag{
					{Name: "all", Kind: spec.Bool, Help: "include the cluster's own machinery"},
					{Name: "failing", Kind: spec.Bool, Help: "only what needs attention"},
				},
				Run: pods,
			},
			{
				Name:   "logs",
				Short:  "Print a pod's recent output",
				Args:   []spec.Arg{{Name: "pod", Required: true, Complete: podNames}},
				Flags:  []spec.Flag{{Name: "tail", Kind: spec.Int, Help: "how many lines"}},
				Target: ui.RegPodsRow,
				Key:    "l",
				Run:    logs,
			},
			{
				Name:   "describe",
				Short:  "Print everything Kubernetes says about a pod",
				Args:   []spec.Arg{{Name: "pod", Required: true, Complete: podNames}},
				Target: ui.RegPodsRow,
				Key:    "d",
				Run:    describe,
			},
			{
				Name:  "console",
				Short: "Open a Rails console in a pod",
				// The same command the TUI runs. There, tea.Exec hands over the
				// terminal; here we are already at one. The engine builds the
				// command either way and neither surface knows about the other.
				Args:   []spec.Arg{{Name: "pod", Required: true, Complete: podNames}},
				Target: ui.RegPodsRow,
				Key:    "c",
				Run:    console,
			},
			{
				Name:   "shell",
				Short:  "Open a shell in a pod",
				Args:   []spec.Arg{{Name: "pod", Required: true, Complete: podNames}},
				Target: ui.RegPodsRow,
				Key:    "s",
				Run:    shell,
			},
		},
	}
}

// browse opens the interface, or captures it.
//
// A Screen, so spec also gives this --snapshot and --script: an agent can
// capture the interface without writing a test, and finds out how from --help.
func browse(c spec.Call) int {
	if c.Bool("fixture") {
		defer engine.Fixture()()
	}
	m := ui.New()
	dir := c.Flag("snapshot")
	if dir == "" {
		if err := ui.Run(m); err != nil {
			return fail(c, err)
		}
		return spec.OK
	}

	script, err := harness.ScriptFile(c.Flag("script"))
	if err != nil {
		return fail(c, err)
	}
	// The world has to be IN the model before the capture starts.
	//
	// harness.Press deliberately discards the tea.Cmd an Update returns, so a
	// capture has no event loop and nothing asynchronous ever resolves. That is
	// what makes a capture deterministic, and it means Init's fetch — which is
	// how the running program loads — never lands. Without this the frames come
	// out saying "nothing yet".
	if snap, err := load(); err == nil {
		m.Load(snap)
	} else {
		return fail(c, err)
	}
	// Wrapped in a runner because the model has no View: the runner owns the
	// canvas, so it is what the harness drives. No pixel layer — a snapshot
	// records characters, whatever terminal it was started from.
	frames, err := harness.Snapshot(app.New(m), dir, script)
	if err != nil {
		return fail(c, err)
	}
	for _, f := range frames {
		_, _ = fmt.Fprintln(c.Out, filepath.Join(dir, f.File))
	}
	return spec.OK
}

// load reads the whole world at once, synchronously.
//
// The interface reads the three in sequence as commands so the first pane is
// usable before the last one arrives. A capture and a script want them all in
// hand, and neither wants a spinner.
func load() (engine.Snapshot, error) {
	ctx := context.Background()
	snap := engine.Snapshot{Fetched: engine.Now()}

	list, err := engine.Projects(ctx)
	if err != nil {
		return snap, err
	}
	snap.Projects = list

	id, _ := engine.CurrentProject(ctx)
	for _, p := range list {
		if p.Active {
			id = p.ID
		}
	}
	if id != "" {
		if clusters, err := engine.Clusters(ctx, id); err == nil {
			snap.Clusters = clusters
		} else {
			snap.Partial = err
		}
	}
	if pods, err := engine.Pods(ctx); err == nil {
		snap.Pods = pods
	} else {
		snap.Partial = err
	}
	return snap, nil
}

func projects(c spec.Call) int {
	list, err := engine.Projects(context.Background())
	if err != nil {
		return fail(c, err)
	}
	for _, p := range list {
		mark := " "
		if p.Active {
			mark = "*"
		}
		_, _ = fmt.Fprintf(c.Out, "%s %-22s %s\n", mark, p.ID, p.Name)
	}
	return spec.OK
}

func clusters(c spec.Call) int {
	ctx := context.Background()
	project := c.Arg("project")
	if project == "" {
		var err error
		if project, err = engine.CurrentProject(ctx); err != nil {
			return fail(c, err)
		}
	}
	list, err := engine.Clusters(ctx, project)
	if err != nil {
		return fail(c, err)
	}
	for _, cl := range list {
		_, _ = fmt.Fprintf(c.Out, "%-16s %-18s %-16s %d nodes  %s\n",
			cl.Name, cl.Location, cl.Version, cl.Nodes, strings.ToLower(cl.Status))
	}
	return spec.OK
}

func pods(c spec.Call) int {
	list, err := engine.Pods(context.Background())
	if err != nil {
		return fail(c, err)
	}
	for _, p := range list {
		if !c.Bool("all") && !p.App() {
			continue
		}
		if c.Bool("failing") && p.State() == engine.Ready {
			continue
		}
		_, _ = fmt.Fprintf(c.Out, "%-28s %-12s %-9s %-5s %3d restarts  %s\n",
			p.Name, p.Namespace, p.State(), p.Ready, p.Restarts, engine.Ago(p.Age))
	}
	return spec.OK
}

func logs(c spec.Call) int {
	p, code := findPod(c)
	if code != spec.OK {
		return code
	}
	// spec.Call hands flags back as strings, typed only at the edge. Two
	// hundred lines is what kubectl's own default is close to, and it is the
	// number a person means by "the recent output".
	tail := 200
	if n, err := strconv.Atoi(c.Flag("tail")); err == nil && n > 0 {
		tail = n
	}
	out, err := engine.Logs(context.Background(), p, tail)
	if err != nil {
		return fail(c, err)
	}
	_, _ = c.Out.Write(out)
	return spec.OK
}

func describe(c spec.Call) int {
	p, code := findPod(c)
	if code != spec.OK {
		return code
	}
	out, err := engine.Describe(context.Background(), p)
	if err != nil {
		return fail(c, err)
	}
	_, _ = c.Out.Write(out)
	return spec.OK
}

func console(c spec.Call) int { return handOver(c, engine.Console) }

func shell(c spec.Call) int {
	return handOver(c, func(ctx context.Context, p engine.Pod) *exec.Cmd {
		return engine.Shell(ctx, p, "")
	})
}

// handOver runs an interactive command on the terminal we are already at.
//
// The TUI's equivalent is tea.Exec, which suspends the interface first. Both
// give the child the real TTY, because a console without one has no line
// editing, no history and no working Ctrl-C.
func handOver(c spec.Call, build func(context.Context, engine.Pod) *exec.Cmd) int {
	p, code := findPod(c)
	if code != spec.OK {
		return code
	}
	cmd := build(context.Background(), p)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(c, err)
	}
	return spec.OK
}

// findPod resolves the pod argument to a pod, or says which names exist.
//
// Listing the alternatives rather than only refusing: a typo in a pod name is
// the most likely reason to be here, and the answer is on screen anyway.
func findPod(c spec.Call) (engine.Pod, int) {
	want := c.Arg("pod")
	list, err := engine.Pods(context.Background())
	if err != nil {
		return engine.Pod{}, fail(c, err)
	}
	for _, p := range list {
		if p.Name == want {
			return p, spec.OK
		}
	}
	var names []string
	for _, p := range list {
		if p.App() {
			names = append(names, p.Name)
		}
	}
	sort.Strings(names)
	return engine.Pod{}, fail(c, fmt.Errorf("no pod %q. there is: %s", want, strings.Join(names, ", ")))
}

// projectIDs and podNames complete an argument — the same completers the shell
// and the interface both use, so they cannot offer different answers.
func projectIDs(prefix string) []string {
	list, err := engine.Projects(context.Background())
	if err != nil {
		return nil
	}
	var out []string
	for _, p := range list {
		if strings.HasPrefix(p.ID, prefix) {
			out = append(out, p.ID)
		}
	}
	return out
}

func podNames(prefix string) []string {
	list, err := engine.Pods(context.Background())
	if err != nil {
		return nil
	}
	var out []string
	for _, p := range list {
		if p.App() && strings.HasPrefix(p.Name, prefix) {
			out = append(out, p.Name)
		}
	}
	sort.Strings(out)
	return out
}

// fail is the one place gcpeasy writes to stderr, so every command's error
// reads the same and none of them has to remember the prefix.
func fail(c spec.Call, err error) int {
	_, _ = fmt.Fprintln(c.Err, "gcpeasy:", err)
	return spec.Fail
}
