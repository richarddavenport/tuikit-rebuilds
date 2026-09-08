# Rebuilding fx

[antonmedv/fx](https://github.com/antonmedv/fx) — 20,616 stars, Go, on
bubbletea. A JSON viewer: fold, search, and walk a document.

Read from a shallow clone on 2026-09-04. `main.go` is 30.7 kB and the whole UI
is about 60 kB across eight files, which makes it the smallest tool in this
survey and the easiest to check completely.

## The finding: a third design for trees

fx does not have a node tree with depths, and it does not have a collapse set
either. It has a **doubly-linked list of lines**, where collapsing re-links:

```go
} else if node.Collapsed != nil {
    node = node.Collapsed
} else {
    node = node.Next
}
```

Walking forward from the top, a collapsed node hands you its `Collapsed`
pointer — the node after its subtree — instead of its next line. Traversal is
the only operation and it is O(visible), with no separate visibility pass.

`comp.Tree.Visible` walks all N nodes and returns the visible indices, which is
O(total) per frame. For a document with a million lines mostly folded, fx's
design wins and ours does not.

**Not a change, yet.** `Visible` returns indices into the caller's own slice,
which is what lets `List` draw it without copying, and the linked-list design
gives that up. The honest note is that `comp.Tree` is right for hierarchies of
thousands and would need rethinking at millions — and that `Node.Key` being a
string map lookup per row is where it would show first.

Recorded in the component's doc rather than filed, because no tool in the
survey has a million-node tree today.

## What tuikit already supplies

| fx has | comp gives |
| --- | --- |
| `search.go` (5.2 kB), `search_test.go` (19 kB) | `fuzzy` + `Highlight` |
| `help.go` | `Keys` |
| `keymap.go` | `spec` + `app.Keys` |
| its status line | `Bar` |
| `preview.go` | `Detail` |

## Holes

### The viewer, in its purest form

fx **is** the `Viewer` hole and nothing else. A scrollable buffer of styled
spans, syntax-colored, searchable, with a cursor over lines and no tailing.
Strip the JSON and what is left is the component seven tools now want.

That makes fx the best test case for `Viewer`'s API: if the design cannot
express fx, it is wrong.

## Our floor

- **`vim.go`** — a `:` command mode taking `:q` and `:<line>`. `Palette` is
  close, and `goto line` is a small thing several tools have
  (`gitui/popups/goto_line.rs` too).

## Theirs — the domain, not the shape

- **`internal/jsonx/`** — a JSON parser that keeps line numbers and can
  re-emit. The program.

## The verdict

**Could tuikit rebuild fx today? No — it is the hole, undiluted.** And it is
the one whose API should be designed against, because it has no chrome to hide
behind.

**Since this was written.** `comp.Viewer` ([#57](https://github.com/richarddavenport/tuikit/issues/57)) and `comp.Tree` ([#20](https://github.com/richarddavenport/tuikit/issues/20)) were built, so tuikit can draw fx now. The verdict below is the separate question, and it has not changed: fx's linked list is still the better data structure.

## Would it have been easier in tuikit?

**No. fx is the one rebuild where the original's core data structure is better
than ours.**

### What you would not have written

`help.go` (a key sheet), the status line, and the search highlighting —
`comp.Keys`, `comp.Bar` and `comp.Highlight`. Perhaps 200 of its 1,900 lines.

### What you would have written anyway

`internal/jsonx` — a JSON parser that keeps line numbers and can re-emit. That
is the program.

### Where tuikit would have got in the way

**`comp.Tree.Visible` is O(total) per frame and fx's traversal is O(visible).**

fx holds its document as a doubly-linked list of lines. A collapsed node hands
back a `Collapsed` pointer to the node *after* its subtree, so folding is a
pointer hop and drawing never touches what is hidden:

```go
} else if node.Collapsed != nil {
    node = node.Collapsed
} else {
    node = node.Next
}
```

`comp.Tree.Visible` walks every node and returns the visible indices, with a map
lookup on `Key` for each one. For a 40 MB JSON file mostly folded, that is the
difference between a usable viewer and an unusable one.

This is written into `comp/tree.go`'s own doc rather than filed, because no tool
in the survey has a million-node tree today — and returning indices is what lets
`comp.List` draw without copying. It is a real trade and fx is on the right side
of it for fx.

### Where fx's approach is better

The linked list, per above. Also its `:` command mode, which takes a line number
and jumps — `comp.Palette` is a fuzzy action list and is not that, though
`Viewer.Goto` is the half tuikit does have.

### The call

**The original is better.** fx would gain a key sheet and lose its data
structure. `comp.Viewer` is genuinely the component fx needed and does not
exist elsewhere — but fx already built one, and the rest of tuikit has little
to offer a program that is one viewer and a parser.
