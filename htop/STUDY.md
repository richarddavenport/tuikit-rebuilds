# Rebuilding htop

[htop-dev/htop](https://github.com/htop-dev/htop) — 8,312 stars, C, on ncurses.
A process viewer, and the tool almost everyone reading this has run today.

Read from a shallow clone on 2026-09-06, at `a3dd8ad`.

## Why this one, when bottom is already here

Not for the system monitoring. bottom covered that, and htop's meters would add
nothing on their own.

**htop is here for F2.** It ships a configuration editor inside the interface,
and the thing it configures is the interface's own layout. That is
[tuikit#60](https://github.com/richarddavenport/tuikit/issues/60) from the far
end: bottom and gh-dash read an arrangement from a file at startup, and htop
lets a reader build one with the arrow keys and saves it.

## The shape

Meters in columns across the top, a process table below, a function bar at the
bottom. 23,134 lines in the portable half, plus 17,622 more across eight
platform directories.

| | lines |
| --- | ---: |
| `CRT.c` | 1,531 |
| `Process.c` | 1,151 |
| `Settings.c` | 1,080 |
| `Action.c` | 1,044 |
| `CPUMeter.c` | 672 |
| `Meter.c` | 618 |
| `Panel.c` | 577 |

## What tuikit already supplies

| htop has | comp gives |
| --- | --- |
| `Panel.c` (577) + `ListItem.c` — a scrolling selectable list, used by nine panels | `List` |
| `Table.c`'s column layout | `Table` |
| `IncSet.c` (339) — the search and filter line | `Input` + `fuzzy` |
| `FunctionBar.c` (172) | `KeyHints` |
| `Header.c`'s column arithmetic | `Layout` with `Percent` |
| `SignalsPanel.c`, `CategoriesPanel.c` | one `drawPicker`, 24 lines |
| the sort-by panel and its state | `Sort` |
| `GraphMeterMode_draw` in `Meter.c` | `Sparkline` |
| `Table_buildTree`'s collapse state | `Tree` |
| `CRT.c`'s eight color schemes and two glyph sets | `theme.Palette`, `theme.GlyphSet` |

`Panel.c` is the interesting row. htop wrote one selectable list and got nine
panels out of it, which is the same trade `comp.List` makes and the reason both
are worth their size.

## Holes

### 1. A tree draws branch lines, and depth cannot say which

The gap this study exists to have found.

htop's tree is not indentation. It is `│`, `├─` and `└─`, and which one a row
gets cannot be derived from its depth:

```
     1 root                   ▾ /sbin/init
   412 root                      ├─ /usr/lib/systemd/systemd-journald
   588 root                      ├─▾ /usr/sbin/sshd -D
  1204 richard                   │  └─▾ sshd: richard@pts/0
```

Two rows at the same depth get different prefixes. Whether row 1204 needs a `│`
in its first column depends on whether something *above* it still has siblings
coming, which is a fact about the rows in between.

htop keeps a bitmask per row for exactly this — `Row.indent`, built by
`Table_buildTreeBranch` in `Table.c`. Bit N set means a vertical line continues
at depth N; a negative value means last child, which flips `├` to `└`.

**dive does the same thing**, with three constants in
`dive/filetree/file_tree.go`:

```go
branchSpace = "│   "
middleItem  = "├─"
lastItem    = "└─"
```

Two tools, so it clears the extraction rule. Filed as
[tuikit#78](https://github.com/richarddavenport/tuikit/issues/78).

And the rebuild found the sharper version of it by building. htop draws the
tree **inside the Command column**, because a prefix in front of the PID makes
every numeric column ragged. So `comp.Row.Depth` could not have been used here
even if it did draw branches — it indents a whole row, and this tree lives in
one cell of a table.

### 2. Four ways to draw the same number, and the reader picks

`Meter.c` has `TextMeterMode_draw`, `BarMeterMode_draw`, `GraphMeterMode_draw`
and `LEDMeterMode_draw`. A reader cycles them per meter in Setup and htop saves
the choice.

`comp.Meter` draws one shape: a bracketed bar with the label after it. That is
a good default and it is not this. htop's bar puts the value *inside* the
brackets, its graph is two rows of history, and its LED mode draws
seven-segment digits out of box characters.

Not filed as one hole, because they are not one thing:

- **Bar and Text** are `comp.Meter` variants and a `Mode` field would cover them.
- **Graph** is `comp.Sparkline` already, and the rebuild uses it unchanged.
- **LED** should never be in a component library. It is one tool's good idea.

The finding is that the *mode being data* is the pattern, not any particular
mode. Same shape as the arrangement: htop's position is that the reader decides
how a number looks, and a library that picks for everybody has taken that away.

### 3. Search and filter are one control with two meanings

`IncSet.c` holds one `LineEditor` and an `isFilter` flag. F3 searches — the
cursor jumps to a match and every row stays. F4 filters — the rows that do not
match are gone.

`comp.Input` is the editor and `fuzzy` is the matching, so both halves exist.
What does not exist is the pairing, and it matters because the two are one
control to the reader and one line of state to the tool. The rebuild writes
`incMode` and shares a key handler between them, which is 30 lines.

One tool. Recorded, not filed.

## Theirs — the domain, not the shape

**17,622 lines across eight platform directories.** `linux/` alone is 8,559,
and it is `/proc` parsing: `LinuxProcessTable.c`, `LinuxMachine.c`, and a file
per data source. `darwin/` is 2,051 lines of `sysctl` and Mach calls.

That is 43% of htop, it is why htop works on your machine, and no framework
supplies a line of it.

`Process.c` (1,151) and its per-platform subclasses are the same story: what a
process *is* on each system, and how to compare two of them.

## Outside the line

Nothing. htop is a viewer and a table, and decision 27 excludes text editors and
multiplexers. This is the first study in the repository where that section is
empty.

## The verdict

**Could tuikit rebuild htop today? Yes.** The rebuild in this directory does it
in 1,116 lines of interface, including the branch-line drawing that `comp` does
not supply.

## Would it have been easier in tuikit?

**Yes for the interface, and it would have changed nothing about the hard part.**

### What you would not have written

`Panel.c` (577), `ListItem.c` (72), `Vector.c` (378), `IncSet.c` (339),
`FunctionBar.c` (172), `RichString.c` (279) and the column arithmetic in
`Header.c`. Call it 1,800 lines of generic infrastructure that `comp.List`,
`comp.Table`, `comp.Input` and `comp.Layout` supply.

`SignalsPanel.c` and the sort panel collapse into one 24-line `drawPicker`,
because both are the same control with different contents.

### What you would have written anyway

**The 17,622 lines of platform code**, which is the larger half and the reason
htop exists. Plus `Process.c`'s comparison and formatting, and `Settings.c`'s
serialization.

A framework that saves you the interface has saved you 1,800 lines of 40,756.

### Where tuikit would have got in the way

**The branch lines.** `comp.Tree` decides which rows are visible and nothing
draws the `│ ├ └`, so the rebuild writes `tree.go` — 90 lines that htop also
had to write. No saving, and `Row.Depth` is actively unusable here because the
tree sits inside a table column.

**The four meter modes.** `comp.Meter` is one shape, so three of the four are
the tool's. The rebuild writes 130 lines in `meters.go`.

**The color schemes.** htop ships eight. tuikit ships one palette on ANSI 0–15,
by decision 28, so the reader's terminal theme decides. That is a deliberate
trade and it is the next section.

### Where htop's approach is better

**Eight color schemes beat one palette, for this tool.**

`CRT.c` carries Default, Monochrome, Black-on-White, Light Terminal, Midnight,
Black Night, Broken Gray and Nord. Decision 28's argument is that a tool should
inherit the reader's theme rather than impose its own, and that is right for a
tool you open for thirty seconds.

htop is not that. It runs on a server you SSH into, in someone else's terminal,
over a link where the theme is whatever the last person set. **Monochrome exists
because some terminals have no color, and Black-on-White because some people
have a white background and ANSI 0–15 does not save you there.** htop cannot
assume the reader's palette is good, so it ships its own and lets them pick.

tuikit has no answer. `theme.Palette` is a value a tool can replace, so a tool
*could* ship eight — but nothing in `comp` offers them or persists the choice,
and `guard.Tokens` is built around there being one.

**Its glyph set switches at runtime.** `CRT_treeStrUtf8` and `CRT_treeStrAscii`
are the same tree in two alphabets, picked from the locale. `theme.GlyphSet` is
the same idea checked at compile time by `guard.Glyphs`, which is stronger — and
htop's version handles the terminal you actually landed on. Both are right; only
one of them helps when `LANG` is `C`.

**`Panel.c` earns more than `comp.List` does.** Nine panels off one type, with
the function bar swapping per panel so the F-keys always describe what is in
front of you. `comp.KeyHints` draws a row of hints and the tool decides what
goes in it, which is less than htop's arrangement, where a panel *carries* its
bar.

### The call

**tuikit, on the interface — and the interface is the smaller half of htop.**

Counted like for like. The rebuild draws the header, the process table, the
tree, Setup, search, filter, sort and the two pickers, in **1,116 lines** of Go
excluding blanks and comments.

htop's own files for the same set — `Panel`, `ListItem`, `Vector`, `IncSet`,
`FunctionBar`, `RichString`, `Header`, `Table`, `Meter`, `MainPanel`,
`MetersPanel`, `AvailableMetersPanel`, `SignalsPanel`, `ScreenManager` and the
six meters drawn here — are **6,356 lines** of C.

Two things that number is not:

- **It is not all of htop's interface.** `ColorsPanel`, `DisplayOptionsPanel`,
  `AffinityPanel`, `BacktraceScreen`, `TraceScreen`, `OpenFilesScreen` and
  `EnvScreen` are real screens this rebuild does not draw. The full interface
  half is 14,479 lines.
- **It is C against Go**, and some of the difference is the language rather than
  the framework. `Vector.c` is 378 lines of a growable array.

What survives both caveats: htop is **40,756 lines**, and **43% of it is reading
eight operating systems**. tuikit would have made htop's interface several times
smaller and htop itself a few percent smaller.

And the eight color schemes are a place where htop is simply right and tuikit
has nothing.
