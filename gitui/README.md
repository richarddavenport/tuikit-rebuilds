# gitui, rebuilt

```
go run ./gitui          open it
go run ./gitui -shots   regenerate the screenshots
```

It does not talk to git. See [`fake/`](fake/).

![gitui rebuilt](docs/frames/stacked.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis:
**[STUDY.md](STUDY.md)**.

## The control for lazygit

Same subject, different team, different substrate. Where the two agree, the
requirement is real; where they differ, it was taste.

The fixture here is deliberately **not** lazygit's, so the two rebuilds do not
look like the same program twice. Compare them: four panels and a main view
against tabs and two panes.

Every hole lazygit's study produced survives here, and no new component was
needed — `comp.Tree`, `List.Extend`/`Range` and `comp.Viewer` again. That is
the strongest form the evidence takes.

## The gap only this study found

**A modal stack is not a screen stack.**

gitui has **32 popup files** with a consistent open, close and stack
discipline. `app.Stack` answers *where am I and how did I get here* for
**screens**. It says nothing about a stack of things drawn over the screen you
are on, and tuikit draws overlays as an ordered if-chain in `Draw` — fine for
three, and not fine for thirty-two.

So `popup` and `[]popup` are written by hand here, in one place, deliberately.
It is what a `comp.Modals` would replace:

```go
func (m *Model) pop() {
	if len(m.stack) > 0 {
		m.stack = m.stack[:len(m.stack)-1]
	}
}
```

The `stacked` screen is a confirm opened **from** the palette. The palette is
still underneath and `esc` comes back to it. One tool so far, so it is written
down rather than filed.

## What it confirmed about `app.Keys`

A popup takes every key, whether or not it does anything with it. That is the
contract — *"an unrecognised key inside a filter box is a character, not a
chance for the screen underneath to act"* — and it means `?` inside a confirm
**closes the confirm** rather than stacking help on top of it.

The first version of this rebuild expected the other thing. There is now a test
named for the rule, because it is the sort of behaviour that reads as a bug
until you know why.

## Also here

`comp.Tabs` with a count on the Status tab. `Row.Right` puts the log's author
and age against the edge, with no padding arithmetic anywhere. `comp.Palette`
holds ten of the 32 popups and teaches their keys rather than replacing them.

## Still outside the line

`components/textinput.rs` is 18 kB of text-area editing with an undo stack.
Decision 27 refuses it. A gitui built on tuikit writes all of that itself and
gets no help — which is a real cost, and a deliberate one.
