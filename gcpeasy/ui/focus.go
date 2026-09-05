package ui

import "github.com/richarddavenport/tuikit/comp"

// pane is which of the three lists has the keyboard.
//
// The state lives in comp.Focus, keyed by region name. This file is what is
// left once that exists: an ordering, and the mapping from a pane to the list
// that draws it.
//
// It used to hold the state too, and the comment here used to say tuikit had no
// focus manager and that five tools in its survey had each written one. It has
// one now — this rebuild is one of the five that argued for it.
type pane int

// The three, in the order tab cycles them, which is the order the data depends
// on: a project chooses the clusters, a cluster chooses the pods.
const (
	paneProjects pane = iota
	paneClusters
	panePods
	paneCount
)

// ring is the order comp.Focus moves through, by region name.
func ring() []comp.Name { return []comp.Name{regProjects, regClusters, regPods} }

func (p pane) title() string {
	return [...]string{"projects", "clusters", "pods"}[p]
}

// region is the pane's own region, which is also its key in the focus ring.
func (p pane) region() comp.Name {
	return [...]comp.Name{regProjects, regClusters, regPods}[p]
}

// rowRegion is where the pane's rows are drawn. comp.Focus.On resolves one to
// the other at a dot boundary, so a click on a row focuses its pane.
func (p pane) rowRegion() comp.Name {
	return [...]comp.Name{regProjectsRow, regClustersRow, regPodsRow}[p]
}

// paneOf maps a region back to its pane.
func paneOf(n comp.Name) (pane, bool) {
	for p := paneProjects; p < paneCount; p++ {
		if n == p.region() || n == p.rowRegion() {
			return p, true
		}
	}
	return 0, false
}

// at is the focused pane, and list is the List it draws with.
//
// Two helpers so the rest of the model asks "which pane" rather than "which
// region name", which is the tool's own vocabulary and not tuikit's.
func (m *Model) at() pane {
	p, _ := paneOf(m.focus.Current())
	return p
}

func (m *Model) list() *comp.List { return &m.lists[m.at()] }
