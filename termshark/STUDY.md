# Rebuilding termshark

[gcla/termshark](https://github.com/gcla/termshark) — 9,984 stars, Go, on
gowid/tcell. A terminal UI for tshark: packet list, packet structure, hex dump.

Read from a shallow clone on 2026-09-04.

## Why this one is unusually informative

termshark keeps its generic widgets in `widgets/`, separately from its
application. The directory listing is therefore a **published list of what one
serious tool found it needed and could not get from its toolkit** — which is
the same question every rebuild here is asking, answered by someone else in
advance:

```
appkeys      copymodetable  copymodetree   enableselected  expander
fileviewer   filter         framefocus     hexdumper       hexdumper2
ifwidget     keepselected   mapkeys        minibuffer      number
regexstyle   renderfocused  resizable      rossshark       scrollabletable
scrollabletext  search      streamwidget   trackfocus      withscrollbar
wormhole
```

Twenty-six widgets. Most of them map onto something in this survey.

## Holes

### 1. Focus is a real component, and this is the second tool to say so

**Four** of those widgets are about focus: `framefocus` (frame the widget that
has it), `trackfocus`, `renderfocused`, plus `keepselected` and
`enableselected` beside them.

```go
// Package framefocus provides a very specific widget to apply a frame around
// the widget in focus and an empty frame if not.
```

The lazygit rebuild filed panel focus as a **watch item** on the grounds that
only swarmctl had it among the four private tools. That was the wrong pool to
count in. With lazygit's `pkg/gui/context/`, termshark's four widgets and
swarmctl's `focus int`, it is three, and the watch item should be promoted.

What is missing in tuikit is not the styling — `Focused` is already a field on
`List`, `Pane` and `Tabs` — but the **answer**: which of several panes on one
screen currently has it, and what moves it.

### 2. `scrollabletext`, `fileviewer`, `withscrollbar`

`Viewer` again, and this time in a tool with no diff and no syntax
highlighting — so the requirement is not "show a diff", it is "show a buffer
you can scroll and search". `search/` and `regexstyle/` sit beside it.

### 3. `copymodetable` and `copymodetree`

A selection mode over a table *and* over a tree, so the thing being selected
from is not always a flat list. Relevant to #55: whatever holds the selection
should not be a field on `List`.

## Confirms

- **`pkg/pdmltree/`** — the fourth hand-rolled tree, over a packet dissection
  rather than a filesystem. Supports the corrected claim in `comp/tree.go`:
  what needs a tree is a hierarchy you cannot walk into.
- **`resizable/`** — `Split` with a draggable divider.
- **`minibuffer/`** — `Palette`.
- **`filter/`** — `Input` + `Highlight`.

## Theirs — the domain, not the shape

- **`hexdumper/`, `hexdumper2/`** — two of them, both packet bytes.
- **`streamwidget/`, `ifwidget/`, `rossshark/`, `wormhole/`** — TCP stream
  reassembly, interface selection, pcap tooling.

## Our floor

- **`number/`** — a numeric-only input. `comp.Form` has typed fields; a
  standalone number input does not exist and nobody has asked.

## The verdict

**Could tuikit rebuild termshark today? No — the same `Viewer`, plus focus.**

Its value is not the verdict. It is that a tool which sorted its own generic
widgets into a folder produced a list that agrees with this survey, including
on the one thing this survey had underweighted.
