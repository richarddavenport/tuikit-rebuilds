# Rebuilding bottom

[ClementTsang/bottom](https://github.com/ClementTsang/bottom) — 13,964 stars,
Rust, on ratatui. A system monitor: CPU, memory, network, processes, disks,
temperatures, battery.

Read from a shallow clone on 2026-09-04.

## The shape

A grid of widgets, each a chart or a table, all live. `src/app.rs` is 102 kB
and `src/canvas.rs` 20.7 kB.

## What tuikit already supplies

| bottom has | comp gives |
| --- | --- |
| `canvas/components/scroll_bar.rs` | `Scrollbar` |
| `canvas/components/search_input.rs` | `Input` + `fuzzy` |
| `canvas/dialogs/process_kill_dialog.rs` (33 kB) | `Confirm` |
| `canvas/dialogs/help_dialog.rs` (15 kB) | `Keys` |
| `canvas/components/pipe_gauge.rs` | `Meter` |
| `canvas/components/data_table.rs` | `Table` + `List` |
| its widget grid | `Layout.Rows` / `.Cols`, nested |

## Holes

### 1. A time series, and the strongest evidence yet that it is one

`src/canvas/components/time_series/vendored.rs` is **55 kB, and it is a copy of
ratatui's own `Chart` widget**:

```rust
//! A [`ratatui::widgets::Chart`] but slightly more specialized to show
//! right-aligned time_series data.
//! Generally should be updated to be in sync with chart.rs
```

They forked their framework's chart rather than use it, and now carry the
maintenance of keeping it in sync. `grid.rs` beside it is another 16.7 kB.

With k9s's `internal/tchart/` (sparkline, gauge, dot matrix) that is **two
independent implementations**, which is the extraction rule met. And bottom
says something k9s could not: even a framework that ships a chart can ship the
wrong one. Right-alignment against *now* is the requirement, because a metric
over time has a fixed right edge and a ragged left one.

`comp.Meter` is one number now. Nothing draws a number over time.

### 2. Sorting a table by a column, for the second time

`canvas/components/data_table/sortable.rs` is 15 kB with a `SortOrder` enum, a
sort column, and header rendering that shows which. k9s's `ui/table.go` has the
same. **Two tools, so this is a hole rather than a watch item** — the k9s
rebuild filed it as the latter and this closes that question.

What is shared is not the comparison, which is the tool's, but the state:
which column, which direction, the arrow in the header, and the click or key
that cycles it.

## The finding that reopens decision 44

**bottom's layout is a config file**, and it is `comp.Layout`'s shape exactly:

```rust
#[serde(rename = "row")]
pub struct Row {
    pub ratio: Option<u16>,
    pub child: Option<Vec<RowChildren>>,
}
```

`src/app/layout_manager.rs` is 47 kB and `src/options/config/layout.rs` 25.9 kB.
A user writes rows and columns with ratios in TOML and gets their own dashboard.

Decision 44 declined yazi's Lua and named the reopening condition as "whether
the layout alone can be data while the components stay Go and stay guarded".
bottom is that, built, in a widely used tool — **and it costs no interpreter, no
binding layer and no second language in the failure path.** The components stay
compiled; only the arrangement is deserialized.

gh-dash does the same thing in YAML and dive has its own `layout/manager.go`.
That is three tools, and it is filed rather than built because the question it
raises is not "can `Layout` be deserialized" — it plainly can — but what the
guards do when the arrangement is not in the source. See the issue.

## Theirs — the domain, not the shape

- **`src/collection/`** — every platform's way of reading CPU, memory,
  temperature, GPU and process tables. This is the program.
- **`src/widgets/process_table/query.rs`** (38 kB) and `query/prefix.rs`
  (20.9 kB) — a filter query language with its own parser.

## The verdict

**Could tuikit rebuild bottom today? No — the charts are the tool.** A system
monitor without a time series is not a system monitor, and that is now a hole
with two implementations behind it rather than a taste question.

**Since this was written.** `comp.Sparkline` ([#58](https://github.com/richarddavenport/tuikit/issues/58)) and `comp.Sort` ([#61](https://github.com/richarddavenport/tuikit/issues/61)) were built out of this study, and both issues are closed. tuikit can draw bottom now, and the rebuild in this directory is the proof. The verdict below is the separate question, and it has not changed.

## Would it have been easier in tuikit?

**No, and not close. The charts are the tool and tuikit has none.**

### What you would not have written

`scroll_bar.rs`, `search_input.rs`, the two dialogs, and the widget grid —
`comp.Scrollbar`, `comp.Input`, `comp.Confirm`, `comp.Keys`, `comp.Layout`.
Real, and a small fraction.

### What you would have written anyway

`src/collection/` — every platform's way of reading CPU, memory, temperature,
GPU and process tables. Plus `process_table/query.rs` and its parser, 59 kB
between them, which is a filter language.

### Where tuikit would have got in the way

There is no time series in `comp`
([tuikit#58](https://github.com/richarddavenport/tuikit/issues/58)). You would
write one, and bottom's experience says that is harder than it sounds.

`data_table/sortable.rs` is 15 kB. `comp.Table` has no sort state
([tuikit#61](https://github.com/richarddavenport/tuikit/issues/61)), so that
too.

### Where bottom's approach is better

**It forked its framework's chart rather than use it, and was right to.**
`time_series/vendored.rs` is 55 kB copied out of ratatui with a note saying to
keep it in sync. The reason is specific: a metric over time is right-aligned
against *now*, with a ragged left edge, and a general chart centred on its data
gets that wrong.

That is a framework being *worse* than the fork, and bottom paid a real
maintenance cost knowingly. Any chart tuikit ships has to answer it.

**Its layout is a TOML file.** `Row { ratio, child }` deserialized — users
arrange their own dashboard. tuikit's `Layout` takes the same shape as values
and cannot be loaded from data
([tuikit#60](https://github.com/richarddavenport/tuikit/issues/60)).

### The call

**The original is better today**, and the chart is no longer the reason.
`comp.Sparkline` was built out of this study and draws bottom's grid. What
is left is layout-as-data: bottom's users arrange their own dashboard in a
TOML file, and tuikit's `Layout` cannot be loaded from data.
