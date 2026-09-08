# fx, rebuilt

```
go run ./fx          open it
go run ./fx -shots   regenerate the screenshots
```

It does not parse JSON. See [`fake/`](fake/).

![fx rebuilt](docs/frames/folded.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis:
**[STUDY.md](STUDY.md)**.

## Why this one is here even though the study says no

The [study](STUDY.md) concludes that **fx's original is better**. Its tree is a
doubly-linked list, so folding is a pointer hop and drawing never touches what
is hidden. `comp.Tree` walks every node on every frame.

This rebuild is where that claim is checked instead of asserted. What it shows:

- **`comp.Viewer` and `comp.Tree` do draw fx's screen.** Folding, searching,
  line numbers, syntax color and a cursor, in about 400 lines.
- **It cannot show the problem the study names.** The fixture is 25 lines and
  fx's difficulty starts at a million. A rebuild against a fixture cannot
  disprove a claim about scale, and this one does not pretend to.

So the verdict stands, and the rebuild is the evidence for the half of it that
*is* testable: tuikit can draw fx, and it would still be the wrong choice.

## What it found

**`comp.Tree` cannot fold a closing bracket.**

A container's `}` sits at the same depth as its `{` — that is what makes JSON
readable — and `Tree`'s model is "rows deeper than me, until one that is not".
So the closing line is a sibling, not a child, and folding leaves it behind.

Found by drawing a frame. Folding everything left a lone `}` on line 2.

**It is not a gap in the component.** A tree of files, of packets or of image
layers has no closing row, and putting JSON's punctuation into a type six other
tools share would be wrong. `ui.visible` drops a close line whose opener is
shut, and the folded line's `{ … }` preview carries the bracket instead. Six
lines, in the tool, where it belongs.

## The one rule worth stealing

Every line is built as **spans**, never as a string:

```go
out = append(out, comp.Segment{Text: `"` + l.Key + `"`, Style: &m.sty.key})
```

`comp.Viewer` puts the cursor's style *underneath* those, so the line you are
reading keeps its syntax colors. A `comp.List` would repaint the cursor row in
one color and you would lose the highlighting exactly where you are looking.

That rule came out of the lazygit study and fx is where it matters most.
