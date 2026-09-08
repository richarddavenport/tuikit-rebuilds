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

## The issue audit

dive has **170 open issues** and its last commit was 2025-12-15. Read on
2026-09-08 and sorted by whether tuikit would have prevented, supplied or not
touched each one.

This is a different question from the rest of this study. Everywhere else asks
what tuikit would have *cost*; this asks what it would have *stopped*.

### Structurally impossible on `comp.List` — 4 issues

`List.resolve(n, row)` is handed the row count **every frame** and clamps to it.
A cursor stored between frames is never trusted, so a row set that changed
underneath cannot leave it dangling.

| | |
| --- | --- |
| [#295](https://github.com/wagoodman/dive/issues/295) | a **panic** — `index out of range [15] with length 15` in `SetCursor`, from holding ↓ at the end of the layer list |
| [#259](https://github.com/wagoodman/dive/issues/259) | cursor goes below the end of a filtered tree |
| [#308](https://github.com/wagoodman/dive/issues/308) | scrolling past the end, then "nothing happens for a second or two" while the cursor walks back |
| [#647](https://github.com/wagoodman/dive/issues/647) | the details pane does not scroll to follow the cursor |

Checked rather than asserted. `htop/ui/cursor_test.go`:

```
60 downs on a 3-row filter left the cursor at 2
```

#308's specific complaint — the delay before the cursor reappears — is what
`step` prevents by **dropping the rest of an over-run move** rather than queueing
it.

#647 is `List.reveal`: `Move` sets it, and the next draw brings the cursor into
view.

### Components that already exist — 8 issues

| | |
| --- | --- |
| [#525](https://github.com/wagoodman/dive/issues/525), [#336](https://github.com/wagoodman/dive/issues/336), [#224](https://github.com/wagoodman/dive/issues/224) | show the contents of the selected file → `comp.Viewer` |
| [#341](https://github.com/wagoodman/dive/issues/341), [#323](https://github.com/wagoodman/dive/issues/323), [#89](https://github.com/wagoodman/dive/issues/89) | sort the tree by size → `comp.Sort` |
| [#176](https://github.com/wagoodman/dive/issues/176), [#181](https://github.com/wagoodman/dive/issues/181) | scroll sideways to read long paths → `comp.Viewer.ScrollX` |

Three tools asking for a file viewer and three asking for sort-by-size, in one
tracker, over seven years. Both are components this survey extracted from other
evidence entirely.

### Made impossible by the canvas — 1 issue

[#474](https://github.com/wagoodman/dive/issues/474) — the interface breaks
under `LANG=ko_KR.UTF-8`, with a screenshot of columns sliding out of line.

That is a width bug: Korean glyphs are two columns wide and something counted
them as one. `comp.Canvas.Set` measures every cluster with `ansi.StringWidth`,
writes a continuation cell for the second column, and blanks **both** halves
when overwriting so a leftover half can never orphan.

### Still open in tuikit too — 2 issues

The honest half, and it produced a new issue.

[#468](https://github.com/wagoodman/dive/issues/468) and
[#543](https://github.com/wagoodman/dive/issues/543) both ask for the cursor to
stay on the same *node* when the row set changes, not the same *line*.

**Our htop rebuild has this bug too**, and there is a test that says so:

```
filtered to 3 rows, cursor on: postgres: autovacuum launcher
filter cleared, 21 rows, cursor on: /usr/sbin/sshd -D
```

Everything in `comp` is keyed by identity — owner IDs, `Marks`, `Tree.Collapsed`
— **except the cursor**, which is an `int`. Filed as
[tuikit#80](https://github.com/richarddavenport/tuikit/issues/80).

### Not tuikit's, and the large majority — around 155 issues

Image formats, registries, podman, containerd, Windows paths, packaging,
efficiency heuristics, CI report fields. `oci-interop` and `distribution` are
dive's two biggest labels and neither is an interface concern.

[#627](https://github.com/wagoodman/dive/issues/627) is a near miss worth
naming: you cannot navigate the tree while a filter is active. That is key
routing rather than a component, and `app.Keys`' capture contract is the shape
that answers it — but a tool can still get it wrong, and **this rebuild did**.
Its filter did not survive its input closing until the audit found it.

### The count

**13 of 170** are things tuikit prevents or supplies outright. **2 more** it
shares. That is 8%, and it is the interface 8% — four of them are a crash or a
cursor that lies.

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
