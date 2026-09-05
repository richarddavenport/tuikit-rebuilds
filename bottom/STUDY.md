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
