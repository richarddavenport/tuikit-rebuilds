package ui

import "github.com/richarddavenport/tuikit/comp"

// Every region gcpeasy draws, named once.
//
// A capture script addresses these by name — `click pods.row[2]` rather than a
// coordinate — and comp.Canvas reports whether one was drawn at all, so a
// script that has drifted from the interface is told which region it can no
// longer find.
const (
	regHeader comp.Name = "header"
	regRule   comp.Name = "header.rule"
	regFooter comp.Name = "footer"

	regProjects    comp.Name = "projects"
	regProjectsRow comp.Name = "projects.row"
	regClusters    comp.Name = "clusters"
	regClustersRow comp.Name = "clusters.row"
	regPods        comp.Name = "pods"
	regPodsRow     comp.Name = "pods.row"

	regDetail comp.Name = "detail"
	regOutput comp.Name = "output"
	regSplit  comp.Name = "split"

	regFilter  comp.Name = "filter"
	regHelp    comp.Name = "help"
	regPalette comp.Name = "palette"
	regToast   comp.Name = "toast"
)

// The regions named from outside this package. internal/cli declares commands
// that target a row, and a command's Target is the region whose context menu it
// appears in.
//
// Exported rather than duplicated. A string constant copied into the other
// package goes stale silently, and the failure is a menu that is simply empty.
const (
	RegProjectsRow = regProjectsRow
	RegClustersRow = regClustersRow
	RegPodsRow     = regPodsRow
)
