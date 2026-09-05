# Rebuilding dive

[wagoodman/dive](https://github.com/wagoodman/dive) — 54,522 stars, Go, on
gocui/tcell. The second most starred Go TUI in the survey. Explores a docker
image layer by layer and shows what each one changed.

Read from a shallow clone on 2026-09-04.

## The shape

Two columns. Layers on the left, the filesystem tree on the right, a details
pane and a filter. Four panes, one focused at a time, tab to cycle.

## What tuikit already supplies

| dive has | comp gives |
| --- | --- |
| `ui/v1/view/layer.go` (9.7 kB) | `List` |
| its details pane | `Detail` |
| its filter | `Input` + `fuzzy` + `Highlight` |
| its keybinding footer | `Keys` |
| `ui/v1/layout/manager.go` (8 kB) | `Layout` |

## Holes

### 1. The tree, at 35 kB

`dive/filetree/file_tree.go` (12 kB), `file_node.go` (9.5 kB),
`ui/v1/viewmodel/filetree.go` (13.7 kB) and `ui/v1/view/filetree.go` (13.2 kB),
with a 21.6 kB test file — the largest test in the repo.

The second hand-rolled tree after lazygit, and the one that best justifies
`comp.Tree`'s shape: dive's hierarchy is a **docker image's layers**, not a
filesystem you can `cd` into. You cannot walk it; the whole shape has to be on
screen because comparing across it is the entire point of the tool.

That is the corrected claim in `comp/tree.go`, and dive is its clearest case.

### 2. Focus across four panes

`ui/v1/app/controller.go` (6.4 kB) routes to whichever pane is focused. Third
tool with this after lazygit and termshark — see the promotion argued in
`termshark.md`.

## Our floor

- **`ui/v1/layout/manager.go`** — dive computes its own pane arrangement rather
  than declaring one. `comp.Layout` covers the arrangement; what dive has and
  tuikit does not is layout as *data*, which is the decision-44 question raised
  by `bottom.md`.

## Theirs — the domain, not the shape

- **`dive/image/docker/`** — reading image archives.
- **`cmd/dive/cli/internal/command/ci/evaluator.go`** — the CI efficiency
  rules. dive is also a linter, and that half is not a UI at all.

## The verdict

**Could tuikit rebuild dive today? Yes, now.** `comp.Tree` is the component it
needed, `List` and `Detail` cover the rest, and its holes are the focus one
that every multi-pane tool has.

This is the first rebuild that comes back "yes".

It is worth noticing why, and the reason is about dive rather than about us. Its
interface is small: one tree, one list, a detail pane and a filter. The tree is
the part that took 35 kB to write, and `comp.Tree` is the component that
answers it.

## Would it have been easier in tuikit?

**Yes, and it is the largest single saving in the survey — but read what it is.**

### What you would not have written

`ui/v1/view/filetree.go` (13.2 kB) and `ui/v1/viewmodel/filetree.go` (13.7 kB)
are the *drawing and view-model* halves of dive's tree. `comp.Tree` plus
`comp.List` with `Row.Depth` covers the visibility question and the drawing.

Also `ui/v1/layout/manager.go` (263 lines) — `comp.Layout`.

### What you would have written anyway

`dive/filetree/file_tree.go` and `file_node.go` — building a tree from a docker
image's layers, and diffing one layer against the next. That is the program, and
it is why dive exists.

`dive/image/docker/` — reading image archives.

### Where tuikit would have got in the way

Four panes, one focused, tab to cycle: `ui/v1/app/controller.go` does it and
tuikit has no answer
([tuikit#59](https://github.com/richarddavenport/tuikit/issues/59)).

### Where dive's approach is better

This is the hardest one to answer honestly, and the fair answer is **nothing
much** — which is itself worth recording, because it is the only study where
that is true.

The closest candidate: dive keys its collapse state by node pointer rather than
by path. That is *faster* than `comp.Tree`'s map lookup on a string key, and it
is fine for dive because its tree is rebuilt wholesale when the layer changes
rather than filtered in place. It would be a bug in a tool that filters, which
is why `comp.Tree` does not do it — but for dive it is the better choice.

### The call

**tuikit, clearly.** dive's interface is one tree, one list, a detail pane and a
filter, and `comp` supplies all four.
