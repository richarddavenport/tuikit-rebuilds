# Why gcpeasy is shaped this way

A record of the decisions that were not obvious, including the ones that were
refused. For how to use the tool, see [the README](../README.md).

## 1. The engine never speaks to the person

`internal/engine` does not print and does not read stdin. No `fmt.Println`, no
`bufio.Scanner` on `os.Stdin`, no prompting.

This is stricter than the import ban that `guard.Engine` enforces, and it is
worth stating separately because breaking it is cheap and the bill arrives
later.

A function that prompts cannot be called from a TUI. It would write over the
frame and then block on a read that never comes. The interface therefore has to
implement the same selection a second time, and from then on the two versions
drift.

**So a choice is a value handed back, never a question asked.** `Clusters`
returns every cluster. Which one to use is the caller's decision, and both
callers make it differently: the CLI takes an argument, the interface puts a
cursor on one.

## 2. The engine returns commands; the caller runs them

`Console` and `Shell` return an `*exec.Cmd` rather than running it.

A process is a domain idea. Giving a process a terminal is not, and the two
callers need opposite things:

- The **TUI** must suspend itself first, because it owns the screen. `tea.Exec`
  hands over the real TTY and restores the frame on exit.
- The **CLI** is already at a terminal and simply runs it.

Neither decision belongs in the engine, and putting it there would force one of
the two surfaces to work around it.

This is also why gcpeasy contains no terminal emulator. A Rails console needs
line editing, history, colour and a working Ctrl-C — and the simplest way to
provide all four is to hand over the terminal that already has them.

## 3. Which pane has focus is the tool's problem, for now

`internal/tui/focus.go` holds a `pane` value and cycles it on `tab`.

tuikit has no focus manager. `Focused` is a field on `comp.List`, `comp.Pane`
and `comp.Tabs`, so every component can be *told*, but nothing holds the answer.
Five tools in tuikit's own survey wrote this themselves — lazygit, termshark,
dive, swarmctl and the original gcpeasy. It is tracked as tuikit issue 59.

So `focus.go` is written in the open, in one file, with nothing else in it.
When issue 59 lands, that file is what should be deleted.

Two rules in it are worth keeping wherever it ends up:

- **A click focuses the pane it landed in**, not only the row. Otherwise
  clicking an unfocused pane moves a cursor you cannot see and leaves the
  keyboard somewhere else.
- **A filter belongs to the pane it was typed in.** Carrying it across would
  silently hide rows in a list the reader never filtered.

## 4. Output is a Viewer, not a LogPane

The right-hand pane shows logs and `kubectl describe` output. Both use
`comp.Viewer`.

`comp.LogPane` tails: it opens at the newest line and follows the end. Neither
of these does. A describe opens at the top, and a log you asked for once has a
bottom. `Viewer` opens at the top, scrolls sideways when a line runs off the
edge, and numbers the lines.

## 5. The fixture goes through the parser

`engine.Fixture()` replaces the **runner** — the function that shells out — not
the reader. So the fixture answers with canned gcloud and kubectl JSON, and
`Projects` and `Pods` parse it with exactly the code that parses the real
thing.

A fixture that built `[]Pod` directly would be a second implementation. The
parsing is where the bugs are, and it would then be exercised by nothing that
ever renders a screen.

## 6. Two glyphs were added deliberately

`guard.Glyphs` failed the build for `◐` and `▲`, which is the guard working:
a character reaches the screen only after somebody decides it should. Both are
declared in `internal/tui/theme.go` with a reason.

The alternative was to use `●` for every state and let colour carry the
difference. Refused, for the reason `harness.ShapeSurvivesColour` exists: a
distinction only colour makes is a distinction lost in a pipe, in a golden, and
to a reader who cannot see it.

## 7. What was left out

- **No ANSI in the output pane.** `kubectl` and `gcloud` colour their output
  when they think they are at a terminal. gcpeasy shows the text and drops the
  colour. Turning a subprocess's escape sequences into styled spans is tuikit
  issue 63; until that exists, dropping the colour beats printing the escapes.
- **No streaming logs.** `l` reads the last 200 lines once. Following a live
  stream is `comp.LogPane`'s job and is a different pane from this one.
- **No project or cluster creation.** gcpeasy shows you what is there and lets
  you act on it. Making infrastructure is `gcloud`'s job and it is already good
  at it.

## 8. Two things tuikit could not do, and what they cost

Building this found two gaps in tuikit. Both are filed, and both are worked
around here in a way that says so.

**`Row.Lead` replaces `List.Marker` rather than sitting beside it**
([tuikit#64](https://github.com/richarddavenport/tuikit/issues/64)). A list with
a status glyph therefore has no visible cursor once the colour is stripped. All
three lists here have one, so `Model.mark` prefixes `›` onto every lead by hand.

**A model that loads in `Init` captures an empty screen**
([tuikit#65](https://github.com/richarddavenport/tuikit/issues/65)).
`harness.Press` discards the `tea.Cmd` an `Update` returns, which is what makes
a capture deterministic and means nothing asynchronous ever resolves. So
`browse --snapshot` loads synchronously first, via `cli.load`.

Neither is a criticism worth much. They are what a tool finds that a rebuild
study reading somebody else's source could not.
