# Context

Gcpeasy's vocabulary. When your output names one of these concepts — an issue
title, a commit message, a test name — use the term as defined here rather than
a synonym.

If a concept you need is missing, that is a signal: either you are introducing
language this project does not use, or there is a real gap worth adding. The
second half of this file is empty on purpose; it is where gcpeasy's own domain
terms go, and it is the first thing to fill in.

## Inherited from tuikit

**Engine** — gcpeasy's domain, in `internal/engine`. It has no terminal
concepts at all: no colour, no width, no keys, no framework. The UI never calls
the backend directly, and the CLI is a peer of the TUI over the same engine
rather than a wrapper around it.

**Role** — a named colour in the palette. Named by *role*, never by hue:
`Accent` survives someone deciding the interface should be blue; `pink` does
not. A raw ANSI index says what a colour IS instead of what it is FOR.

**Palette** — the closed set of colour roles, valued as the terminal's own
sixteen ANSI indices, so Gcpeasy is themed by whatever themed the terminal.
It takes `theme.Default` and overrides what it disagrees with; it does not build
one from scratch, and an override overrides the reader.

**Glyph set** — the allow-list of non-ASCII characters the interface may print,
each with a reason. A font without a glyph draws a replacement box, which reads
as a bug rather than as decoration.

**Guard** — a test that holds a rule closed against a package. `guard.Tokens`,
`guard.Glyphs`, `guard.Chrome`, `guard.Engine` and `guard.Reachable` all run
here, from the first commit.

**Canvas** — the cell grid components draw into, instead of returning strings. A
component that returns a string cannot be clicked.

**Region** — a named area of the frame, and what a click resolves to: a name and
an index, never a coordinate. A region's index is its position in the LIST, not
on the screen; those differ the moment a viewport scrolls.

**Golden** — a captured frame, colour stripped, held in `internal/tui/testdata`.
A golden holds the SHAPE, which is what a diff in a pull request can show.

**Fixture** — the fixed world the goldens are rendered from. It is what makes a
screen something you can look at without the real backend.

**Mechanism / policy** — the line between tuikit and gcpeasy. Mechanism is a
library function that must work for every tool. Policy is a default, delivered
as code gcpeasy owns and can edit. Anything backend-shaped belongs here.

## Gcpeasy's own

<!-- One entry per domain term, in the same shape as above. Write one the first
     time you catch yourself explaining what something means. -->
