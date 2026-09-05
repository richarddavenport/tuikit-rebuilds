package ui

import (
	"context"

	"github.com/richarddavenport/tuikit/app"
	"github.com/richarddavenport/tuikit/harness"

	"github.com/richarddavenport/tuikit-rebuilds/gcpeasy/engine"
	"github.com/richarddavenport/tuikit-rebuilds/internal/shot"
)

// States is every distinct thing gcpeasy can be showing.
//
// ONE list, walked by the goldens, the narrow-terminal run, the colour check
// and the capture. A screen added without a frame is a screen added without any
// of them.
func States() []shot.State {
	return []shot.State{
		{Name: "browsing", Note: "Three lists in the order the data depends on. A project chooses the clusters; a cluster chooses the pods.",
			Build: build()},
		{Name: "pods", Note: "`tab` twice. `comp.Focus` holds which pane has the keyboard, and only that pane's cursor is drawn bright.",
			Build: build("tab", "tab")},
		{Name: "pod-detail", Note: "`comp.Detail` — labels line up per block, so a long key does not drag the facts above it wide.",
			Build: build("tab", "tab", "j", "j")},
		{Name: "logs", Note: "`l` reads the pod's output into a `comp.Viewer`. Not a `LogPane`: a log you asked for once has a bottom.",
			Build: withLogs},
		{Name: "filtering", Note: "`/` filters the focused pane only. A filter belongs to the pane it was typed in.",
			Build: build("tab", "tab", "/", "w", "e", "b")},
		{Name: "nothing-matches", Note: "An ordinary state, not an error.",
			Build: build("tab", "tab", "/", "z", "z")},
		{Name: "help", Note: "`?` is `comp.Keys`, built from the same list the footer is.",
			Build: build("?")},
		{Name: "palette", Note: "`space` is `comp.Palette`. It runs a KEY, so whatever it does the reader has just been shown how to do it without the palette.",
			Build: build(" ")},
		{Name: "empty", Note: "Before anything has loaded.",
			Build: func() app.Model { return New() }},
	}
}

// loaded is the model with the fixture's world already in it.
//
// It loads SYNCHRONOUSLY. harness.Press discards the tea.Cmd an Update returns,
// which is what makes a capture deterministic and means Init's fetch never
// lands (tuikit issue 65).
func loaded() *Model {
	defer engine.Fixture()()
	m := New()
	m.SetSize(shot.Width, shot.Height)

	ctx := context.Background()
	snap := engine.Snapshot{Fetched: engine.FixtureAt}
	snap.Projects, _ = engine.Projects(ctx)
	snap.Clusters, _ = engine.Clusters(ctx, "acme-staging")
	snap.Pods, _ = engine.Pods(ctx)
	m.Load(snap)
	return m
}

func build(keys ...string) func() app.Model {
	return func() app.Model {
		m := loaded()
		harness.Press(app.New(m, app.WithSize(shot.Width, shot.Height)), keys...)
		return m
	}
}

func withLogs() app.Model {
	defer engine.Fixture()()
	m := loaded()
	m.setFocus(panePods)
	if pods := m.pods(); len(pods) > 0 {
		out, _ := engine.Logs(context.Background(), pods[0], 200)
		m.outTitle = "logs " + pods[0].Name
		m.setOutput(string(out))
	}
	return m
}
