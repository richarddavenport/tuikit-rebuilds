# Rebuilding k9s

[derailed/k9s](https://github.com/derailed/k9s) — Go, on a fork of tview and
tcell (`go.mod:12-13`). Chosen because it is the closest thing in the field to
what azctl, pgctl and swarmctl are: an operator tool over live remote state.

Read from a shallow clone on 2026-09-04: `internal/ui/`, `internal/view/`,
`internal/tchart/` and `internal/xray/`, sizes quoted.

## The shape

One table, most of the time. A `:` command bar switches what it is a table of;
a breadcrumb says where you are; the bottom rows carry the keymap and a flash
line. Off that hang a log view, a YAML view, a resource tree, and a dashboard
of charts.

`internal/ui/table.go` is 16 kB and is the program.

## What tuikit already supplies

| k9s has | comp gives |
| --- | --- |
| `ui/table.go` alignment and widths | `Table` — `Column`, `Width`, `Spans` |
| `ui/menu.go`, `ui/indicator.go` | `Keys`, `Bar` |
| `ui/crumbs.go` | `Breadcrumb` |
| `ui/flash.go` | `Toast` |
| `ui/pages.go` | `app.Stack` |
| `ui/modal_list.go` | `Menu` |
| `ui/prompt.go` — the `:` bar | `Palette` + `Input` + `fuzzy` |
| `view/log.go` (13.5 kB) | `LogPane` |
| `view/details.go` | `Detail` |
| `ui/splash.go`, `ui/logo.go` | `Picture`, `paint` |
| `tchart/gauge.go` | `Meter` |

## Holes

### 1. The selection set, for the third time — and it is not shaped as we assumed

`internal/ui/select_table.go` carries `marks sets.Set[string]`: a set of row
IDs, not indices. That is the third independent implementation after yazi's
`Selected` and pgctl's request in #49, and it settles the key question the same
way both others did.

But k9s adds two rules that neither of the others made visible, and both change
the API in #55:

**`GetSelectedItems` falls back to the cursor.**

```go
func (s *SelectTable) GetSelectedItems() []string {
	if s.marks.Len() == 0 {
		if item := s.GetSelectedItem(); item != "" {
			return []string{item}
		}
		return nil
	}
	return s.marks.UnsortedList()
}
```

This is the whole reason marks are usable. `d` deletes the marked pods, or the
one under the cursor if none are marked — one code path, no branch at every call
site. A set component that makes each tool write that `if` has given it the
data and kept the ergonomics.

**`SpanMark` has no live anchor.** Where yazi drags a `Visual` range and commits
it, k9s scans backwards from the cursor for the nearest existing mark, then
forwards if it found none, and fills between (`select_table.go:158-213`). There
is no anchor state at all: the set is primary and the range is derived from it.

So the two implementations disagree about which is primary — and `List.Extend`
picked yazi's side without knowing there was a side to pick. Both gestures want
the same set underneath, which is an argument for the set being its own type
rather than more fields on `List`.

### 2. A sparkline

`internal/tchart/` is sparkline, gauge and dot matrix, and `view/pulse.go`
(13.9 kB) is the dashboard built from them. The sparkline is eight runes and
some arithmetic:

```go
var sparks = []rune{'▁','▂','▃','▄','▅','▆','▇','█'}
```

`comp.Meter` is the gauge. There is nothing for a value over time, and "one
number now" versus "this number for the last minute" is the difference between
a status and a trend.

**One tool so far**, so not yet. bottom, btop and bandwhich are all charts, and
one of them is the next rebuild that would settle it.

### 3. Sorting a table by a column

`ui/table.go` sorts by column with an indicator in the header. `comp.Table`
owns alignment and nothing else — deliberately, since a List whose rows a Table
laid out is two components composed.

The *comparison* is the tool's; it owns the data and its types. What is not the
tool's is which column, which direction, and the arrow drawn in the header,
which is the same three-line piece of state in every tool that has a table.

**Watch item.** Nothing in the four private tools sorts interactively yet.

## Recorded and not built

**Deltas.** `ui/deltas.go` plus `table.go:529-537` compare each cell to its
value on the previous refresh and append `↑`, `↓` or `Δ`. For a tool watching
live state this is genuinely good — it answers "what changed while I was
looking away" without a diff view.

Checked all four private tools for anything similar: **none has it.** azctl,
pgctl, swarmctl and docket all redraw from a poll and none compares to the
previous frame. One implementation is an anecdote, so it stays here rather than
becoming a component.

Worth flagging that it would be cheap in this architecture and is not obvious:
it needs the *previous* row set kept beside the current one, which is a thing a
tool would have to decide to do.

## Confirms things already open

- **`view/live_view.go`** (10.3 kB) — YAML and describe output, scrolled, with
  a search over it. The `comp.Viewer` hole from the lazygit rebuild, in an
  operator tool rather than a git client. That is now five of the surveyed
  tools.
- **`internal/xray/` + `ui/tree.go` + `view/xray.go`** (18.8 kB) — a resource
  tree with expand/collapse-all on `x`. A fifth tree consumer, though k9s did
  not write it: `ui/tree.go` wraps `tview.TreeView`. Worth noting because it is
  the one substrate in the survey that ships a tree, and it is why k9s's is
  2.8 kB where lazygit's is eleven files.

## Theirs — the domain, not the shape

- **`view/pf.go`, `pf_extender.go`, `exec.go`, `scale_extender.go`,
  `image_extender.go`** — port forwards, shelling into a container, scaling a
  deployment. Kubernetes, all of it.
- **`internal/dao/`, `client/`, `watch/`, `vul/`** — the API layer.
## Our floor

- **`ui/prompt.go`'s `FishBuff`** — command history with a fish-style inline
  suggestion of the rest of the line. The *contents* of the history are the
  tool's; the mechanism is not, and `Palette` has none. Nobody has asked, so it
  is a floor rather than a hole — but it is the kind of thing that would be
  filed as "theirs" and quietly never revisited.

## The verdict

**Could tuikit rebuild k9s today? Close, and the gaps are the ones already
open.** The selection set (#55) and `Viewer` are both blocking, and both now
have three or more independent implementations behind them.

The more useful result is that k9s changed the *shape* of #55 rather than just
adding a vote to it. Two tools disagree about whether the set or the range is
primary, and the cursor-fallback rule is not in either of the other two
implementations but is the thing that makes the feature worth having.

## Would it have been easier in tuikit?

**Yes, once #55 lands. Today it is a wash.**

### What you would not have written

| | lines |
| --- | ---: |
| `ui/table.go` — alignment and widths | part of 16 kB |
| `ui/crumbs.go`, `ui/flash.go`, `ui/menu.go`, `ui/indicator.go` | — |
| `ui/pages.go` — a page stack | — |
| `view/log.go` | 13.5 kB |

`comp.Table`, `comp.Breadcrumb`, `comp.Toast`, `comp.Keys`, `comp.Bar`,
`app.Stack`, `comp.LogPane`. This is the closest one-to-one mapping in the
survey.

### What you would have written anyway

`internal/dao/`, `client/`, `watch/`, `vul/` — the Kubernetes API layer. The
port forwards, the exec, the scaling. Most of k9s.

### Where tuikit would have got in the way

`ui/select_table.go`'s `marks` set has no equivalent
([tuikit#55](https://github.com/richarddavenport/tuikit/issues/55)), and it is
central: `d` deletes the marked pods *or* the one under the cursor, through one
code path. You would write it.

`view/live_view.go` needs `comp.Viewer`, which now exists. `internal/tchart/`
needs a chart, which does not.

### Where k9s's approach is better

**It got a tree for free.** `ui/tree.go` is 127 lines wrapping
`tview.TreeView`. lazygit's equivalent is eleven files. tview ships a retained
tree widget and tuikit will never have one, because `comp` is immediate mode —
so a tuikit k9s writes the flattening itself.

**Its deltas.** `ui/deltas.go` marks every cell that changed since the last
refresh with `↑`, `↓` or `Δ`. For a tool watching live state that is genuinely
good, and nothing in tuikit or in any of the four private tools does it.

### The call

**tuikit wins on the chrome and loses on the tree.** For a tool that is mostly
one table, the chrome is most of the interface, so it comes out ahead — but not
until the selection set exists.
