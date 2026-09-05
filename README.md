# tuikit rebuilds

Could [tuikit](https://github.com/richarddavenport/tuikit) build the terminal
interfaces people actually use?

Ten of them were read from source and worked out on paper. Each answer is a
`STUDY.md`. Where the answer is yes, there is also a working program that draws
the interface — against a fixture, not a real backend.

![lazygit rebuilt](lazygit/docs/frames/range.svg)

## The tools

| | study | rebuilt | verdict |
| --- | :---: | :---: | --- |
| [lazygit](lazygit/) | ✓ | **✓** | Yes. Its three holes are now `comp.Tree`, `List.Range` and `comp.Viewer` |
| [gcpeasy](gcpeasy/) | ✓ | **✓** | Yes, all of it — and [the rebuild is a real tool](https://github.com/richarddavenport/gcpeasy) |
| [dive](dive/) | ✓ | | Yes. `comp.Tree` was the component it needed |
| [gh-dash](gh-dash/) | ✓ | | Yes, apart from the PR body view |
| [k9s](k9s/) | ✓ | | Close. Needs a selection set and `Viewer` |
| [gitui](gitui/) | ✓ | | The same one thing short as lazygit, found independently |
| [yazi](yazi/) | ✓ | | Nearly. One hole that matters: a selection set |
| [termshark](termshark/) | ✓ | | No — `Viewer`, plus focus |
| [fx](fx/) | ✓ | | No. It **is** the `Viewer` hole, undiluted |
| [bottom](bottom/) | ✓ | | No. The charts are the tool |

## What this repository is careful about

The obvious thing to write here is that tuikit is better. Two claims of exactly
that shape have already been made in this work and both had to be retracted
after they were measured:

- *"gcpeasy's 2,972-line TUI is what a tool pays for not using tuikit."* Measured
  afterwards: 624 of those lines are drawing. The rest is the program.
- *"gh-dash's `ui.go` is 54 kB of exactly the state management `app` was
  extracted to remove."* Measured afterwards: 741 of its 1,917 lines are one
  `Update` function. True about the kind of code, wrong about the amount.

Both are in [tuikit's claim log](https://github.com/richarddavenport/tuikit/blob/main/design/research/other-tuis.md#claims-that-failed-this-check).

So [METHOD.md](METHOD.md) requires every comparison to be one of four things:
lines of drawing code, a test that stops being necessary, a guard that would
fire, or a component that did not have to be written. Anything else is an
opinion and does not go in a README.

It also requires reporting the ones that go the other way. The lazygit rebuild
found a guard that fires **and is wrong**, and says so.

## What the studies found

Ten tools, read from source. The components that came out of it:

| found by | became |
| --- | --- |
| lazygit, dive, termshark, fx each wrote a collapsible tree | `comp.Tree` |
| lazygit and gitui both wrote a line range over a diff | `List.Extend`/`Range` |
| seven tools wrote a scrollable view that does not tail | `comp.Viewer` |

And the ones still open, each with its evidence attached: a
[selection set](https://github.com/richarddavenport/tuikit/issues/55) (3 tools),
[panel focus](https://github.com/richarddavenport/tuikit/issues/59) (5 tools),
a [time series](https://github.com/richarddavenport/tuikit/issues/58) (2),
[table sort](https://github.com/richarddavenport/tuikit/issues/61) (2),
[ANSI from a subprocess](https://github.com/richarddavenport/tuikit/issues/63) (1).

Three claims made along the way turned out to be false and were withdrawn.
yazi, superfile and ranger were each cited as needing a tree; none of them has
one, and all three are Miller-column file managers. That correction sharpened
the claim rather than weakening it: **every file manager in the field chose
columns over a tree**, and what needs a tree is a hierarchy you cannot walk into
— image layers, a JSON document, a packet dissection, a git status.

## Running them

```
go run ./lazygit          open it
go run ./lazygit -shots   regenerate its screenshots
go test ./...             every screen against its golden, at two sizes
```

Every rebuild is captured at 132×38 and at 80×24, and every frame is checked
with the colour stripped — a distinction only colour makes is a distinction lost
in a pipe.

## What a rebuild is allowed to conclude

That tuikit could or could not draw this interface, and what was missing.

**Not** that the original should have used tuikit. lazygit predates it by nine
years, most of these are not written in Go, and every one of them works.
