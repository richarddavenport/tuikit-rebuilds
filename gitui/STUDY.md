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
