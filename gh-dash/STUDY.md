# Rebuilding gh-dash

[dlvhdr/gh-dash](https://github.com/dlvhdr/gh-dash) — 12,448 stars, Go, on
**bubbletea**. A dashboard of GitHub pull requests, issues and notifications.

Read from a shallow clone on 2026-09-04.

## Why this one matters more than its star count

It is the closest tool in the survey to tuikit's own substrate. Bubble Tea and
lipgloss, the same as the four private tools.

It is also a dashboard of remote state you act on, which is decision 27's
sentence almost word for word. If anything in this field should need nothing
from us, it is this.

## What tuikit already supplies

| gh-dash has | comp gives |
| --- | --- |
| `components/prssection/`, `issuesection/`, `reposection/` | `List` + `Table` per screen |
| `components/prrow/`, `notificationrow/` | `Row` with `Segment` spans |
| `components/prview/`, `issueview/` (21.6 kB, 11.7 kB) | `Detail` + the `Viewer` hole |
| `components/tabs/` | `Tabs` |
| `components/tasks/` | `app.Async` + `Toast` |
| `components/footer/`, `help/` | `Bar`, `Keys` |
| its search bar | `Input` + `fuzzy` |
| `components/prview/checks.go` (18.9 kB) | `StepList` |

That is an unusually complete row of ticks, and it should be: same substrate,
same shape of problem.

## Holes

### 1. The section is a screen and a list and a fetch, and there is no seam

`components/section/section.go` (13 kB) is the base every section embeds, and
`internal/tui/ui.go` is **54 kB** — the largest single UI file in the Go half
of this survey, larger than lazygit's biggest.

What is in it is the thing `app` exists for: which section is current, what
each has fetched, which is loading, what the keys do in each. tuikit has
`app.Screens`, `app.Stack`, `app.Keys` and `app.Async` for exactly this, and
they are each smaller than the part of `ui.go` they replace.

**Not a hole in `comp`.** Recorded because it is the best available evidence
that the `app` package earns its place: a well-written bubbletea dashboard
without it has a 54 kB model.

### 2. Layout as data, for the third time

`internal/config/parser.go` (24.5 kB) defines the dashboard in YAML — sections,
and per-column `Width` and `Hidden`:

```go
type ColumnConfig struct {
	Width  *int  `yaml:"width,omitempty"`
	Hidden *bool `yaml:"hidden,omitempty"`
}
type LayoutConfig struct {
	Prs    PrsLayoutConfig
	Issues IssuesLayoutConfig
}
```

With bottom's TOML rows and yazi's Lua, three tools let their *users* arrange
the interface. See the issue raised in `bottom.md`.

Note what gh-dash's version is: **column width and visibility**, not a widget
grid. That is a smaller and much more common want than bottom's, and it is one
`comp.Table` could satisfy without any of decision 44's difficulty — `Column`
already has a `Width`.

## Theirs — the domain, not the shape

- **`internal/data/prapi.go`, `notificationapi.go`** — the GitHub GraphQL
  layer.
- **`components/tasks/pr.go`, `issue.go`** — merge, close, comment, assign.

## The verdict

**Could tuikit rebuild gh-dash today? Yes, apart from the PR body view** —
which is `Viewer`, again.

The interesting result is *where* the size is. Measured rather than eyeballed,
because an earlier draft of this file said "54 kB of exactly the state
management `app` was extracted to remove" without checking.

`internal/tui/ui.go` is 1,917 lines across 37 functions. The distribution is
lopsided:

| | lines |
| --- | ---: |
| `Update` | **741** |
| `View` and the two `render*` helpers | 275 |
| the other 34 functions | 825 |

**One function is 39% of the file.** That single `Update` is message dispatch,
key routing, which section is current, what each has fetched, and which is
loading. It is exactly what `app.Keys`, `app.Stack`, `app.Screens` and
`app.Async` were extracted to hold.

Note the contrast with the gcpeasy rebuild. gcpeasy's weight is in drawing;
gh-dash's is in `Update`, and its drawing is only 275 lines. Two competent tools
on our substrate, two different halves gone wrong. That is a better argument for
`app` and `comp` being separate packages than either tool alone.
