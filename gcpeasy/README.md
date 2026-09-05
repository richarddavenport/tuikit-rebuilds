# gcpeasy, rebuilt

```
go run ./gcpeasy            open it, against real GCP
go run ./gcpeasy -fixture   open it, against canned data
go run ./gcpeasy -shots     regenerate the screenshots
```

**The one rebuild here that talks to a real backend.** Everything else draws
against a fixture, which is the right shape for asking "could tuikit draw
this". This asks the other question — what a tool built on tuikit from the
start actually looks like.

![gcpeasy](docs/frames/pod-detail.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis:
**[STUDY.md](STUDY.md)**. The decisions: **[DESIGN.md](DESIGN.md)**.

## Measured on both sides

The only rebuild where both halves exist, so the comparison is real rather than
on paper.

| | original | rebuild |
| --- | ---: | ---: |
| drawing | 624 lines, 31 `render*` funcs | **291 lines**, one `view.go` |
| the program | 1,715 lines | about the same |
| tests asserting two strings have compatible shapes | 4 | **0** — they cannot be written |

The 1,715 matters as much as the 291. Most of any tool is its domain, and
tuikit has no opinion about task orchestration, state caching or auth.

## What it found by being built

Two gaps that reading somebody else's source could not have found. Both are
fixed now, and this rebuild is what argued for them:

- **[tuikit#64](https://github.com/richarddavenport/tuikit/issues/64)** —
  `Row.Lead` replaced `List.Marker`, so all three lists here needed a cursor
  marker prefixed by hand. `List.lead` now draws both.
- **[tuikit#59](https://github.com/richarddavenport/tuikit/issues/59)** — nothing
  held which pane had the keyboard. `focus.go` used to carry that state; it now
  carries an ordering and a name lookup, and `comp.Focus` holds the answer.

One is still open:
**[tuikit#63](https://github.com/richarddavenport/tuikit/issues/63)** — `kubectl`
colours its output and nothing turns ANSI into `Segment`s, so the logs pane
shows the text and drops the colour.

## The console

`c` opens a Rails console in the selected pod, in **your real terminal** — real
line editing, real history, and Ctrl-C reaching Ruby rather than reaching
gcpeasy.

The seam is why that works. The engine decides *what* to run:

```go
func Console(ctx context.Context, p Pod) *exec.Cmd
```

The interface decides *how*:

```go
tea.ExecProcess(cmd, func(err error) tea.Msg { return doneMsg{err: err} })
```

A process is a domain idea. Giving it a terminal is not. **No terminal emulator
anywhere**, which is the piece the original's study first got wrong about it.

## The rule the original broke

`engine/` never prints and never prompts. That is stricter than the import ban
`guard.Engine` enforces, and it is what the original got wrong: `SelectCluster`
prompted on stdin, so the TUI could not call it and implemented pod selection a
second time.

**A choice is a value handed back, never a question asked.**
