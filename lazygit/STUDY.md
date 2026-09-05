# Rebuilding lazygit

[jesseduffield/lazygit](https://github.com/jesseduffield/lazygit) — 81,999
stars, Go, on gocui. The most used TUI in the surveyed field, and more starred
than Bubble Tea itself.

Read from the repository on 2026-09-04: `pkg/gui/` and its subpackages, with
file sizes quoted so the claims below can be checked.

## The shape

Four panels down the left — status, files, branches, commits, stash — and a
main panel on the right showing whatever the cursor is on. A command palette,
a menu, confirmations, a diff view, and a lot of git.

## What tuikit already supplies

| lazygit has | comp gives |
| --- | --- |
| the left column of panels | `Layout.Rows` + `Pane`, one `List` each |
| the main/secondary split | `Split`, nested for the two-panel diff |
| its side panel lists | `List`, with `Row.Skip` for headings |
| `pkg/gui/popup` | `Confirm`, `Toast` |
| its menu | `Menu` — but see the correction below |
| keybinding help | `Keys` + `guard.Keys` |
| its status line | `Bar` |
| search / filter | `Input` + `fuzzy` + `Highlight` |
| `pkg/gui/style` | `theme` — and lazygit's is a hand-rolled equivalent |

That is most of the chrome, and it is the part tuikit is already right about.

## Holes

### 1. Tree — `pkg/gui/filetree/`, eleven files

`build_tree.go` (4.3 kB), `node.go` (8.4 kB), `collapsed_paths.go`,
`file_tree_view_model.go`, and a `README.md` explaining it. A file tree with
collapsing, a flat/tree toggle, filtering, and two node types (working-tree
files and commit files) over one traversal.

**This is the biggest single gap in the field**, though not by as much as
first claimed — see the correction below. `comp.Tree` was refused once, on the
grounds that azctl's and swarmctl's versions carried different payloads and a
shared node type would make both worse. That reasoning was sound against a
sample of two. It does not survive lazygit, dive, termshark and fx each
building the same thing.

The refusal also points at the shape: what differs is the **payload**, so the
component must not own it. `List` already solved this exact problem —
`Row.Depth` is a number rather than a node, and the flattening stays the
tool's. A Tree is likely `Row.Depth` plus collapse state plus "which rows does
collapsing hide", not a node type.

### 2. Line-range selection in a diff — `pkg/gui/patch_exploring/`

`state.go` is 13 kB and `focus.go` handles keeping the selection visible. This
is staging by hunk and by line: a range that grows with shift-arrow, snaps to
hunk boundaries, and survives the diff being re-rendered under it.

`comp.List` has a cursor and no selection. **A range selection over rows is a
real gap** and it is not only lazygit's — any tool that stages, bulk-selects or
copies a span needs it, and it is the interaction pgctl asked for in issue 49's
multi-select from the other direction.

Worth noting the size: 13 kB of state for what sounds like "shift-click a
range". Most of it is the snapping and the survival across re-render.

### 3. A text view that is not a log — the main panel

`LogPane` tails a stream and follows the end. A diff is not that: it scrolls
from the top, wraps or does not, highlights syntax, and has a selection over
it. Five of the fourteen tools need this shape.

## What the first pass got wrong

Checked against the source on 2026-09-04, after two of the holes were built.

**"its menu | `Menu`, from the `spec` declaration"** was not true. `comp/menu.go`
imports `lipgloss` and nothing else; `Menu.Items` is `[]Hint`, and both callers
— `gallery/entries.go:258` and democtl's `view.go:337` — build that slice by
hand. A menu derived from a `spec.Command` is a thing that could exist and does
not. Filed rather than fixed here, because it is a `spec` question and not a
lazygit one.

**The tree evidence was overstated.** The first pass named seven tools:
lazygit, yazi, superfile, dive, fx, termshark, ranger. Checked in their sources,
four have one — lazygit, dive, termshark, fx. yazi and superfile contain no
match for "tree" or "collapse" at all, and ranger's only collapse is
`collapse_preview`, the preview column. All three are Miller columns.

Four is still twice the extraction rule, so `comp.Tree` stands. But the
corrected list says something the inflated one hid: **every file manager in the
field chose columns over a tree.** What needs a tree is a hierarchy you cannot
walk into — image layers, a JSON document, a packet dissection, a git status.

**Panel focus was waved away too quickly.** The first pass said `pkg/gui/context/`
is covered by "`app.Keys` and `app.Stack` differently rather than better". That
is half right. `app.Keys` orders capture → screen → global, and `app.Stack`
records how you reached a screen. Neither answers *which of five panels on one
screen is active*, which is what lazygit's context stack is for and what its tab
key moves. `Focused` is a bool on `List`, `Pane` and `Tabs`, so each component
can be told; nothing keeps the answer.

It was not extracted, and the count was why: of the four private tools, only
swarmctl has it (`internal/tui/disk.go:74`, `focus int` plus a `paneFocus` bool
for descending into a pane). azctl and docket set `Focused` from a condition
they already have, and pgctl from a cursor.

**That was the wrong pool to count in, and the watch item is now promoted.**
Counting the surveyed tools instead: lazygit's `pkg/gui/context/`, termshark's
`framefocus`, `trackfocus`, `renderfocused`, `keepselected` and
`enableselected` — five widgets for it — and dive's
`ui/v1/app/controller.go`. With swarmctl that is four, and every one of them is
a multi-pane screen. See the issue.

**Accordion panels** — lazygit grows the focused panel and shrinks the rest —
needs nothing new. In immediate mode the constraint handed to `Layout.Rows` can
differ every frame, so "the focused panel is `Fill`, the others are `Min: 3`" is
a `switch` in the draw and not a feature.

## The verdict

**Two of the three holes are built.** `comp.Tree` is collapse state over a
flattened hierarchy, keyed by the tool's own identity rather than an index.
`List.Extend`/`Range`/`ClearRange` is the shift-arrow range, anchored on the
first extend and dropped by any plain move.

**One is left: a text view that is not a log.** `LogPane` tails and follows and
has no selection, which its own doc says is deliberate — a log has no cursor. A
diff scrolls from the top, takes styled spans rather than strings, and has a
range over it. Until that exists lazygit's main panel cannot be drawn, and five
of the fourteen surveyed tools want the same component.

**Is the gap a criticism of tuikit's scope?** No, and it never was. lazygit
shows you state and lets you act on it, which is decision 27 as widened. A tree
is not a git idea, a range is not a git idea, and neither is a scrollable
syntax-aware view.

The boundary is the thing lazygit itself does not cross. It stages hunks; it
does not edit them. **Showing a buffer and acting on ranges of it is in; being
the place you type the buffer is out.**

## Theirs — the domain, not the shape

- **`pkg/gui/presentation/graph/`** (10 kB) — the commit graph, the `│ ├ ─ ╯`
  lines beside the log. A framework supplying this would be a framework with an
  opinion about git.
- **`pkg/gui/presentation/`** — sixteen files turning branches, commits,
  stashes and submodules into rows. This is exactly the layer tuikit says
  belongs to the tool, and lazygit agrees by putting it in its own package.
- **`pkg/gui/mergeconflicts/`** — finding and rendering conflict markers.

**What it does not need** is the thing worth noticing. lazygit has no charts, no
forms, no wizard and no tabs.

What it does have is four lists, a split, and a diff view with a great deal of
care in it. The components that matter here are few and deep. That is an
argument for `comp` staying small and making each entry thorough, rather than
growing a catalogue.
