# Rebuilding gcpeasy

[scttymn/gcpeasy](https://github.com/scttymn/gcpeasy) — Go, on Bubble Tea 1.3.10,
Bubbles, **lipgloss v2** and cobra. A GCP operator tool: pick a project, a GKE
cluster and a pod, then run things against them.

Read from a shallow clone on 2026-09-05.

## Why this one is not like the others

Every other rebuild here is a famous tool on a substrate we do not use. This is
a small one on **exactly ours** — Bubble Tea, lipgloss, cobra, a Go domain layer
shelling out to a cloud CLI. It is what azctl, pgctl and swarmctl looked like
before tuikit, written by somebody who has never heard of it.

That makes it the most direct evidence in the survey, and the least flattering
to read.

## The shape, in one number

**`cmd/tui.go` is 78,596 bytes. 2,972 lines. 154 functions. One file.**

The whole program is 5,554 lines of Go. The TUI is 54% of it, and the domain
layer — everything that actually talks to GCP — is 343 lines across two files.

For scale: gh-dash's `ui.go` was 54 kB and the k9s rebuild called it the largest
single UI file in the Go half of this survey. This is half again bigger, in a
tool with a fraction of the features.

## What tuikit already supplies

Named against the actual functions, so the claim can be checked.

| gcpeasy wrote | comp gives |
| --- | --- |
| `renderPanel` — border, title, focus colour, inner clip | `Pane` |
| `renderRow` — cursor glyph, marker, primary, secondary, padding | `List` + `Row.Lead`, `Row.Right` |
| four `render*Panel` funcs, one per pane | one `List` each |
| `renderLeft`, `panelInnerWidth`, `maxInt`, `minInt` | `Layout.Rows`, `Rect` |
| `renderFooter`, `padFooter`, `footerHint`, `footerHints` | `Bar` + `comp.Hint` |
| `renderCommandModal`, `commandItems` | `Palette` |
| `renderHelpModal`, `helpRow` | `Keys` |
| `renderAuthDialog`, `renderAuthButtons`, `renderRefreshModal` | `Confirm` |
| `overlayCentered` | drawing into a rect from `comp.Center` |
| `clip` | `Truncate` |
| `renderBootScreen`, `renderGradientLogo`, `logoGradientColor`, `mixRGB`, `tuiRGB.hex` | `paint.Ramp` + `comp.Picture` |
| `visibleProjects`, `visibleClusters`, `visiblePods` | `Row.Skip`, or a filter over `DrawFunc` |
| `taskSpec`, `runningTask`, `taskStartedMsg`, `taskOutputMsg`, `taskDoneMsg` | `app.Async`, `app.Gen`, `app.Poll` |
| `handleKey` / `handleAuthDialogKey` / `handleCommandModalKey` / `handleHelpModalKey` | `app.Keys` — capture, screen, global |
| `tuiStateCache`, `tuiPreferences` | the tool's, but `app.Toggles` is the hidden-item half |
| its cobra tree | `spec.Command`, plus the menu and manifest for free |

## Four tests that exist because of strings

This is the part worth sitting with. `cmd/tui_test.go` is 30.9 kB, and among
what it asserts:

- `TestLeftAndRightPanelsMatchHeight`
- `TestPanelsRenderAtRequestedWidth`
- `TestSelectedRowsFillPanelInterior`
- `TestLongPodRowsDoNotWrap`

Every one is a test that two rendered **strings** have compatible shapes. On a
cell grid none of them can be written, because none of the failures can happen:
a pane is a rect, a row is filled to its rect's width, and drawing past the
right edge is clipped rather than wrapped.

They are good tests. They are also four tests, in a 31 kB file, defending
properties that a different substrate gives away.

## What the guards would say

Not hypothetical — these are the specific findings.

**`guard.Engine` fires, and the cost is already visible.** `internal/` imports
`bufio`, `fmt` and `os`, and calls `fmt.Print` **21 times**. `SelectCluster`
(`kubernetes.go:47`) and `SelectPod` (`pod.go:113`) *prompt the user on stdin*
from inside the domain layer.

The consequence is not stylistic. **The TUI cannot call them.** A function that
prints to stdout would draw over the frame and then block on a read that will
never come, so `SetupClusterAndSelectPod` is reachable only from `cmd/pod.go`,
`cmd/rails.go` and `cmd/cluster.go` — the CLI paths — and the TUI reimplements
selection itself. One decision in the engine, taken for the CLI's convenience,
cost a second implementation of the tool's central interaction.

That is `guard.Engine`'s whole argument, demonstrated by someone who was not
trying to make it.

**`guard.Tokens` fires, and the instinct behind it was right.** There are no hex
literals anywhere. Instead there are 15 named style vars — `tuiTitleStyle`,
`tuiMutedStyle`, `tuiSelectedRowStyle`, `tuiActiveBorderColor` — built from ANSI
256 numerics (`"81"`, `"229"`, `"62"`, `"149"`). Somebody centralised their
palette on their own, which is the same move `theme.Palette` is.

What they do not get is a set that is *closed*: nothing stops the sixteenth
style, nothing reports a role nothing draws with, and nothing distinguishes the
0–15 band the reader themes from the 16–255 band they do not (decision 28).

**No mouse at all.** Zero occurrences of `MouseMsg`. Not a criticism — it is
work, and `app.Mouse` plus owner IDs is the thing that makes it not work.

## Holes

### 1. ANSI from a subprocess, turned into spans

`appendOutput` (`tui.go:2140`) takes the output of `gcloud`, `kubectl` or
`rails console` and has to cope with what a real program emits: `\x1b` CSI
sequences via `parseTerminalEscape` and `applyCSI`, `\r` moving the column back,
`\b` and `\x7f`, tabs to a stop of four, and a 2,000-line cap.

`comp.LogPane` and `comp.Viewer` both take plain text or `Segment`s. **Nothing
in tuikit turns a subprocess's coloured output into `Segment`s** — and every
operator tool shells out to something.

The logic exists, on the wrong side of the fence: `harness.Strip` and
`harness.Rows` parse exactly these sequences, but they live in the *test*
package and are about reading a frame tuikit itself wrote.

A bounded piece of work: SGR to `[]Segment`, `\r` and `\b` applied, everything
else dropped. Not an emulator. See the issue.

### 2. Panel focus, for the fifth time

`m.focus tuiPanel` plus `m.cursors[panel]` — focus, and a cursor per pane. With
lazygit, termshark, dive and swarmctl that is five, and this one is on our own
substrate. Issue 59.

## Outside the line

**Almost nothing, and an earlier draft of this file got that wrong.**

The first version said the interactive pane was a terminal emulator and
therefore outside decision 27. It is not. `runInteractiveSession`
(`tui.go:935`) uses `tea.Exec` to hand the user's **real terminal** to the
child process, and its own comment says so:

> hands the user's real terminal to an interactive remote session via tea.Exec,
> rather than capturing it into the output viewport

That is the right answer and it is one tuikit can give unchanged. `rails
console` and `kubectl exec` get a real TTY, with real line editing and real
scrollback, because the TUI suspends rather than emulating.

What actually needs the emulator is the **non-interactive task pane**.
`startTask` runs background commands under a PTY (`pty.StartWithSize`,
`tui.go:2725`) so that `gcloud` and `kubectl` emit colour and progress, and then
`appendOutput` and `applyCSI` have to interpret what comes back — SGR, but also
`\r` moving the column, `\b`, and tabs.

So the emulator is a consequence of a choice, not a requirement of the domain.
Run the task without a PTY and you need SGR parsing and nothing more, which is
issue 63.

## Theirs — the domain, not the shape

- `internal/kubernetes.go`, `internal/pod.go` — GKE, kubectl, namespaces.
- `cmd/auth.go`, `cmd/env.go`, `cmd/rails.go` — gcloud auth, environments, Rails.

## The finding that is about us, not them

**lipgloss v2 ships a cell buffer.** gcpeasy's `overlayCentered` uses it:

```go
canvas := lipgloss.NewCanvas(width, height)
composite := lipgloss.NewCompositor(
    lipgloss.NewLayer(base).X(0).Y(0).Z(0),
    lipgloss.NewLayer(modal).X(x).Y(y).Z(z),
)
return canvas.Compose(composite).Render()
```

`lipgloss.Canvas` has `CellAt`, `SetCell`, `Compose`, `Render`, `Resize`. Under
it, `ultraviolet.Cell` is `{Content, Style, Link, Width}`.

So **a cell grid is no longer what makes tuikit different**, and any claim in
that shape should be retired. What `uv.Cell` does not carry is an **owner**.
There is no `OwnerAt`, so a click still has to be resolved against remembered
rectangles rather than against the frame that was actually drawn — which is the
property `comp.Canvas` exists for and the one the README should lead on.

Checked because `other-tuis.md` exists to make claims of the form "nobody else
does X" checkable before they are made. This one needed narrowing.

## The verdict

**Could tuikit rebuild gcpeasy today?** Yes, all of it.

The interactive console is a `tea.Exec` handoff, which tuikit supports as it
stands. The one piece with no tuikit answer today is turning a subprocess's
ANSI output into `Segment`s, and that is issue 63 rather than a boundary.

### The gaps it found

This is what a rebuild is for, so it goes first.

1. **ANSI from a subprocess.** `appendOutput` parses SGR out of `gcloud` and
   `kubectl` output. Nothing in tuikit turns that into `Segment`s. Filed as
   issue 63.
2. **Panel focus, for the fifth time.** `m.focus` plus `m.cursors[panel]`.
   Issue 59, now with a tool on our own substrate behind it.
3. **A claim of ours that needed narrowing.** lipgloss v2 ships a cell buffer,
   so "a cell grid" is no longer what makes tuikit different. Owner IDs still
   are. See above.

### The number I first put here, and why it was wrong

The first version of this verdict said the TUI is 2,972 lines in one file and
54% of the program, and left the reader to conclude that tuikit would remove
most of it.

That is a true number used to suggest something false. Counting the lines inside
each function:

| | lines | share |
| --- | ---: | ---: |
| drawing — the 31 `render*` and helper funcs | 624 | 22% |
| key handling — `handleKey` and its four modal siblings | 264 | 9% |
| terminal emulator — `appendOutput`, `applyCSI` and friends | 198 | 7% |
| everything else — tasks, state cache, preferences, auth, GCP calls | 1,715 | 61% |

**`comp` would replace about 624 lines, not 2,972.** `app.Keys` would take a
share of the 264. The emulator is outside our line by choice. The 61% is the
program — orchestrating tasks, caching state, remembering hidden items,
authenticating — and tuikit has no opinion about most of it.

624 lines of hand-rolled drawing is still a real finding, and so is the fact
that four of its tests defend properties a cell grid gives free. Neither needed
to be inflated.

### What the exercise is for

The stated purpose in this directory's README is a coverage test: for each
feature, do we have it, are we missing it, or does it belong to the tool?

The first draft of this file asked a different question instead — what does a
tool pay for not using tuikit? That is a marketing question wearing a research
question's clothes, and it cannot be answered by reading somebody else's
repository. Answering it properly would mean rebuilding gcpeasy and counting
the result.

Kept as a note rather than deleted, because `other-tuis.md` exists for exactly
this: a record of the claims that did not survive being checked is more useful
than a record of the ones that did.
