# Rebuilding gitui

[gitui-org/gitui](https://github.com/gitui-org/gitui) — 22,453 stars, Rust, on
ratatui. The other git client in the survey, and the control for lazygit: same
subject, different substrate, different team.

Read from a shallow clone on 2026-09-04.

## Why this one is worth reading after lazygit

Two independent teams built the same tool. Where they agree, the requirement is
real; where they differ, it was taste. Every hole lazygit produced survives
here, which is the strongest form the evidence takes.

## What tuikit already supplies

| gitui has | comp gives |
| --- | --- |
| `src/tabs/` — status, revlog, files, stashing | `Tabs` + `app.Stack` |
| `components/commitlist.rs` (24 kB) | `List` + `Table` |
| `components/changes.rs` | `List` with `Row.Lead` |
| ~26 files in `src/popups/` | `Confirm`, `Menu`, `Toast`, `Input` |
| `popups/fuzzy_find.rs` | `Palette` + `fuzzy` |
| `popups/help.rs` | `Keys` |
| `components/command.rs` | `Bar` |

## Holes, all of them already open

### `components/diff.rs` (23.7 kB) + `components/syntax_text.rs` (6.2 kB)

The `Viewer` hole, confirmed by a second git client that shares no code with
the first. gitui splits it the way the component should: syntax-highlighted
text as one thing, the diff's selection and hunk staging as another.

That is now **seven** of the surveyed tools wanting a scrollable, span-taking,
non-tailing text view: lazygit, gitui, k9s (`live_view.go`), fx, termshark
(`scrollabletext`), dive, yazi's preview.

### `components/status_tree.rs` (15 kB)

The fifth hand-rolled tree, after lazygit, dive, termshark and fx. Same
subject as lazygit's `pkg/gui/filetree/`, written independently, at similar
size. `comp.Tree` is doing the right job.

### Line selection in the diff

Same as lazygit's `patch_exploring/`: stage by hunk or by line, with a range.
`List.Extend` covers the gesture; the set it should commit into is #55.

## Outside the line

- **`components/textinput.rs`** (18.2 kB) — a full text area with its own
  cursor movement and wrapping. Same call as yazi's input: pure shape, refused
  by decision 27 rather than by subject.

## Theirs — the domain, not the shape

- **`popups/blame_file.rs`, `compare_commits.rs`, `create_branch.rs`,
  `create_remote.rs`, `fetch.rs`, `pull.rs`, `push*.rs`, `stashmsg.rs`,
  `tag_commit.rs`** — git, one popup at a time.
- **`asyncgit/`** — the whole background git layer.

## The verdict

**Could tuikit rebuild gitui today? The same one thing short as lazygit.**

That is the result. Two teams, two substrates, two codebases, one missing
component — and no hole here that lazygit did not already produce. A rebuild
that finds nothing new is the one that tells you the earlier list was right.

**Since this was written.** `comp.Viewer` ([#57](https://github.com/richarddavenport/tuikit/issues/57)) was built, which is the one thing both this and lazygit were short of. tuikit can draw gitui now.

## Would it have been easier in tuikit?

**Yes, and this is the clearest yes in the survey.**

### What you would not have written

`src/popups/` is **32 files**. Confirm, help, fuzzy find, goto line, options,
msg, and twenty-odd git-specific dialogs. The generic half of those is
`comp.Confirm`, `comp.Menu`, `comp.Palette`, `comp.Keys` and `comp.Input`.

Plus `components/commitlist.rs` (24 kB), `changes.rs`, `command.rs` — `comp.List`
and `comp.Bar`.

### What you would have written anyway

`asyncgit/` — the entire background git layer. And the *content* of those 32
popups: what a branch dialog asks is a git question, even when the box around it
is not.

### Where tuikit would have got in the way

`components/textinput.rs` is 18 kB of text-area editing. Decision 27 refuses it,
so you write all of it and get no help.

The same focus problem as lazygit: five panels, no focus manager.

### Where gitui's approach is better

**Its popup stack.** 32 popups with a consistent open/close/stack discipline is
a real system. `app.Stack` is about *screens* — where you are and how you got
there — not about a stack of things layered over the current screen. tuikit
draws overlays as an ordered if-chain in `Draw`, which is fine for three and
would not be fine for thirty-two.

That is a gap this study found and nothing else has: **a modal stack is not a
screen stack**, and tuikit only has the second.

**ratatui's constraint solver.** gitui gets Cassowary layout from its framework.
tuikit's `Layout` is a single linear pass, which reaches the same answer for
these shapes — decision 27 was amended to say so after `comp.Layout` was built —
but "the same answer for these shapes" is a weaker guarantee than a solver.

### The call

**tuikit, comfortably**, on component count alone. It would owe gitui a modal
stack.
