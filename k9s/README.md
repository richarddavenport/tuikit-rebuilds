# k9s, rebuilt

```
go run ./k9s          open it
go run ./k9s -shots   regenerate the screenshots
```

It does not talk to Kubernetes. See [`fake/`](fake/).

![k9s rebuilt](docs/frames/filled.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis:
**[STUDY.md](STUDY.md)**.

## Why this one matters

k9s's study argued for three components. All three now exist, and this is the
first time they are used together:

| the study said | tuikit now has |
| --- | --- |
| a selection set, keyed by identity | `comp.Marks` |
| a table sorted by a column | `comp.Sort` |
| something holding which pane has focus | `comp.Focus` |

Plus `comp.Viewer` for `describe`, which the study filed under the lazygit hole.

## The rule the study found and nothing else had

**`Marks.Acting` returns the marks, or the cursor when there are none.**

```go
func (m *Model) acting() []string {
	key := ""
	if p, ok := m.cursor(); ok {
		key = p.Key()
	}
	return m.marks.Acting(key)
}
```

That is what lets `ctrl+d` mean *delete the marked pods* and *delete this one*
through a single code path. Two screens in `docs/screens.md` are the same key
with and without marks, and there is no branch between them.

Neither yazi's implementation nor pgctl's request made this visible. k9s's did,
and a component that handed back the set and left every call site to write that
`if` would have done the easy half.

## Two gestures, because two tools disagreed

- `space` toggles one mark.
- `ctrl+space` **fills from the nearest existing mark** to the cursor —
  backwards first, then forwards. No anchor at all: the set is primary and the
  range is derived from it.

yazi does the opposite: a live anchored range that commits into the set.
`comp.Marks` offers both, because a component that picked a side would be wrong
for the other tool.

## What it found

**There is no column equivalent of `List.Overhead`.**

`comp.Table` lays out and `comp.List` scrolls — two components composed, which
is tuikit's stated pattern for a scrolling table. But the header has to line up
with rows the List has already indented by its marker and its mark, and
`Overhead()` answers in *rows*.

The tool counts:

```go
over := comp.Width(m.list.Marker) + comp.Width(markGlyph)
```

Not filed yet. One tool, and the number is small and local. If bottom's rebuild
needs the same thing it becomes two.

## What is still theirs

The comparison. `less(a, b fake.Pod, col int)` lives in the tool, and always
will — ordering a Kubernetes age against a byte count is a domain question, and
a framework guessing at it would be wrong.
