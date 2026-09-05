# lazygit, rebuilt

```
go run ./lazygit          open it
go run ./lazygit -shots   regenerate the screenshots
```

It does not talk to git. See [`fake/`](fake/) for the working tree it draws.

![lazygit rebuilt](docs/frames/range.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis this came
from: **[STUDY.md](STUDY.md)**.

## Why this one first

The lazygit study found three holes in tuikit. All three are now components,
and this is the first time they have been used together rather than one at a
time:

| the study said | tuikit now has |
| --- | --- |
| a collapsible file tree | `comp.Tree` |
| a line range in a diff | `List.Extend` / `Range` / `ClearRange` |
| a text view that is not a log | `comp.Viewer` |

Press `space` on a directory, then `v`, then `J` a few times. That is all three
at once.

## What was measured

Following [METHOD.md](../METHOD.md): every claim is one of four kinds, and
nothing here is an adjective.

### Components that did not have to be written

| lazygit wrote | lines | tuikit supplies | lines, once |
| --- | ---: | --- | ---: |
| `pkg/gui/filetree/collapsed_paths.go` | 43 | `comp.Tree` | 51 |
| `pkg/gui/patch_exploring/` | 485 | `List.Extend`/`Range` + `rangeSel` | 70 |
| `pkg/gui/style/` | 571 | `theme`, unchanged | 0 |

**Read that table carefully, because it does not say what it looks like it
says.** `comp.Tree` is 51 lines of code and replaces 43. Per tool, that is not
a saving worth a framework.

The claim it does support is narrower and better: **four tools wrote those same
43 lines, and each got a slightly different answer.** lazygit keys collapse by
path, and it is the only one of the four that does — dive keys by node pointer,
which is why its tree cannot survive a refresh, and fx re-links a list instead.
Writing it once means arguing about the key once.

`pkg/gui/filetree/` is 1,730 non-test lines in total, and `comp.Tree` replaces
43 of them. The other 1,687 build a tree out of git paths and filter it. That
is **theirs**, correctly, and a framework that supplied it would be a framework
with an opinion about git.

### A guard that would fire — and would be wrong

`guard.Tokens` rejects a colour literal. lazygit has 743 of them in `pkg/gui`
outside tests.

**All 743 are in one file**: `pkg/gui/presentation/icons/file_icons.go`, a
table mapping file types to their brand colours — Ruby's red, Go's blue. Those
are not theme decisions and they should not come from a palette.

So the honest result is that the guard fires and the guard is wrong here. A
tool like this needs `theme.Palette.Extra`, or an exemption for a data table.
Recorded because a repository that only reported the guards it won would not be
worth reading.

### Tests that stop being necessary

None found in lazygit. It tests its file tree and its patch state — real logic —
rather than testing that two rendered strings have compatible widths.

The four tests of that kind in this survey are in
[gcpeasy](../gcpeasy/STUDY.md), not here.

## What this rebuild does not do

- **No git.** The working tree, the diff, the branches and the log are all
  fixtures.
- **No staging.** `v` selects a range and `a` snaps it to a hunk. Nothing is
  applied, because applying it is git.
- **No commit graph.** `pkg/gui/presentation/graph/` is 10 kB of `│ ├ ─ ╯` and
  it is exactly what the study calls theirs.
- **No search, no submodules, no worktrees, no custom commands.**

The original handles ten years of bug reports, real repositories and terminals
that misbehave. This handles a fixture. Nothing here is a fair fight and the
numbers above are the only comparisons narrow enough to survive that.

## The one thing worth stealing

`comp.Viewer` puts the cursor and range style **underneath** the line's own
spans, so a diff keeps its added-green and removed-red while you drag a
selection across it. Most list components repaint the selected row in one
colour, which in a diff trades the content for the pointer.

That rule came out of this rebuild's study and is the reason `Viewer` is not
`List`.
