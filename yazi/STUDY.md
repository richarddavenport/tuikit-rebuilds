# Rebuilding yazi

[sxyazi/yazi](https://github.com/sxyazi/yazi) — 41,856 stars, Rust, on ratatui.
The most starred Rust TUI and the most starred file manager in the survey.

Read from a shallow clone on 2026-09-04: `yazi-fm/`, `yazi-core/`,
`yazi-widgets/`, `yazi-adapter/` and `yazi-plugin/preset/components/`, with
paths and sizes quoted so the claims can be checked.

## The shape

Miller columns: parent, current, preview. A header above them and a status line
below.

On top of that sits a stack of overlays. An input, a completion popup, a picker,
a confirm, a help sheet, a task list, a metadata inspector, and a which-key
pane.

**It is not a tree.** This matters, because the first pass of the lazygit
rebuild claimed yazi had one. Searching the whole source for `tree` or
`collapse` returns nothing. The hierarchy is walked, one directory per column,
never drawn.

## The finding that outranks the rest: the UI is Lua

`yazi-fm/src/root.rs` does not lay out the file manager. It calls into `LUA`:

```rust
let root = LUA.globals().raw_get::<Table>("Root")?.call_method::<Table>("new", area)?;
root.call_method("reflow", ())
```

The whole main surface lives in `yazi-plugin/preset/components/` as Lua a user
can override. That is `root.lua`, `current.lua`, `parent.lua`, `preview.lua`,
`header.lua`, `status.lua`, `linemode.lua`, `marker.lua`, `rail.lua` and
`tabs.lua`.

Rust keeps the overlays and the engine. Everything you look at is a script.

That is a second, serious answer to "get out of the developer's way", and it is
not tuikit's. tuikit's answer is headless components with injected styles: the
behaviour is fixed and correct, the look is yours. yazi's is that the drawing
itself is replaceable at runtime, and the cost is an embedded interpreter, a
binding layer (`yazi-binding/`), and errors that surface as
`Failed to redraw the 'Root' component`.

**No change proposed.** It is recorded because "we get out of your way" turns
out to be a claim with a competitor.

The honest version of ours is narrower: *out of your way on style and data,
opinionated about behaviour.*

## What tuikit already supplies

| yazi has | comp gives |
| --- | --- |
| Miller columns | `Layout.Cols` with three constraints |
| `header.lua`, `status.lua` | `Bar` |
| `yazi-fm/src/notify/` | `Toast` |
| `yazi-fm/src/confirm/` | `Confirm` |
| `yazi-fm/src/help/` | `Keys` + `guard.Keys` |
| `yazi-fm/src/pick/` | `Menu` |
| `yazi-fm/src/tasks/`, `progress.rs` | `StepList`, `Meter`, `Waiting` |
| `yazi-fm/src/spot/` — the metadata inspector | `Detail` |
| `yazi-core/src/tab/backstack.rs`, `history.rs` | `app.Stack` |
| image preview (`yazi-adapter/`) | `comp.Picture` |
| `yazi-widgets/src/scrollable.rs` | the viewport inside `List` |

## Holes

### 1. A selection set, separate from the range

`yazi-core/src/tab/selected.rs` is 7.1 kB and it is **not** the visual range.
It is an `IndexMap` keyed by URL with a `parents` map beside it, and it
survives navigating away from a directory and back.

`yazi-core/src/tab/visual.rs` is the other half. It is the transient range you
drag out, and it exists only to be committed into `Selected`.

tuikit has the range and not the set. `List.Range` is the gesture; nothing
holds the answer afterwards, and nothing holds a selection made in one screen
and read in another. This is pgctl's issue 49 arriving from a second direction,
which is the extraction rule met.

`Tree` has already decided the shape: **key it by the tool's own identity, not
by index.**

yazi keys by URL for the same reason `Tree` keys by a string. A filter or a
refresh moves every index. A selection that moved with them would act on the
wrong files, and it would not error while doing it.

### 2. Key sequences

`yazi-core/src/which/` (4.8 kB across four files) plus `yazi-fm/src/which/`
render the which-key pane: press a prefix, see the candidate continuations, and
`sorter.rs` orders them.

`app.Keys` routes **one** keystroke through capture → screen → global. There is no notion of a pending prefix, so `g` then `g` cannot be expressed at
all. A tool that wants it has to install itself as the capture and reimplement
the routing underneath.

The which-key *pane* is a `Menu` and needs nothing. The pending-prefix state is
the hole, and it is `app`'s rather than `comp`'s.

**One tool so far.** helix, k9s and vim-flavoured tools are where the second
comes from; check one before building it.

### 3. Completion over an input

`yazi-fm/src/cmp/cmp.rs` is a popup anchored to the cursor inside the input,
filtered as you type. tuikit has every ingredient: `Input`, `fuzzy` and `Highlight`. What it does not
have is a component that puts them together, so each tool wires the popup's
position, its filtering and its key capture itself.

Smaller than it looks, and worth doing only when a second tool asks.

## Theirs — the domain, not the shape

- **`yazi-scheduler/`, `yazi-watcher/`, `yazi-vfs/`, `yazi-dds/`** — the actual
  file manager: the work queue, the inotify layer, the virtual filesystem, the
  data-distribution bus. Exactly what a framework must not have an opinion
  about.

## Outside the line

- **`yazi-widgets/src/input/`** — 30-odd files: vi modes, an undo stack
  (`snaps.rs`, `snap.rs`), motion parsing (`parser/move.rs`, `forward.rs`,
  `backward.rs`), `kill.rs`, `casefy.rs`.

  Note that **none of this is a file-manager idea.** A text input is pure
  shape, and every tool in the survey has one. It is excluded because hosting
  a text buffer is named out in decision 27 — a line drawn on purpose, not a
  fact about yazi's subject. If that line ever moves, this is what moves with
  it.

- **`yazi-plugin/`** — the Lua runtime. Not "theirs" at all: a bet considered
  and declined, with the reasoning and the reopening conditions in decision 44.
  Listed here only so a reader of this file does not go looking for it under
  the domain items.

## Our floor

- **`yazi-adapter/src/drivers/`** — six image backends: `kgp.rs` (7.6 kB),
  `kgp_old.rs`, `iip.rs` (iTerm), `sixel.rs`, `chafa.rs`, `ueberzug.rs`.

  tuikit has two — `term.Sixel` and `term.Kitty` — and **no fallback at all**
  for a terminal that supports neither, where yazi degrades to chafa's
  half-blocks. That is a budget, not a boundary: nobody has asked for a third,
  and the day somebody does the answer is a driver rather than a discussion.

## What this rebuild changed elsewhere

**yazi's Visual wraps and tuikit's cursor does not.** `visual.rs` carries a
`wraps: isize` and `ranges()` returns **two** ranges, because moving the cursor
past the end wraps it to the top and the selection becomes both ends of the
list with a hole in the middle.

`List.Range` returns one ordered `lo, hi` and is correct, because `List.step`
stops at the end rather than wrapping. That is a real dependency and it is
written down nowhere else: **if list wrapping is ever added, `Range`'s signature
is the thing that breaks**, and it will break by silently selecting the middle
of the list instead of the two ends.

## The verdict

**Could tuikit rebuild yazi today? Nearly — one hole that matters.** The
selection set is real and has a second consumer already. Key sequences and the
completion popup are each one tool so far.

The Lua question is the more interesting one and it is not a hole. It is a
different bet about who owns the drawing, and yazi is the strongest evidence in
the field that the other bet works.
