# bottom, rebuilt

```
go run ./bottom          open it
go run ./bottom -shots   regenerate the screenshots
```

It does not read this machine. See [`fake/`](fake/).

![bottom rebuilt](docs/frames/browsing.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis:
**[STUDY.md](STUDY.md)**.

## The study said no. This is why it now says yes.

bottom's verdict was **"No, and not close — the charts are the tool"**. A system
monitor without a time series is not a system monitor.

`comp.Sparkline` exists because of this study, and it is built around the thing
bottom told us. It **vendored 55 kB of ratatui's own `Chart`** rather than use
it, with a note saying to keep it in sync. So the question was never whether to
have a chart. It was what bottom forked *for*:

> A metric over time has a fixed right edge and a ragged left one.

The newest sample is always the last column. A short series leaves the **left**
blank; a long one drops its oldest. A general chart centres or stretches its
data, and then the bar under your cursor moves when a sample arrives, and two
sparklines of different lengths stop being comparable.

Look at CPU 0 in the screenshot. The spike sits near the right edge because that
is where *now* is. There is a test asserting it.

## Auto-scale is a decision, and it is made twice here

Both on purpose, in the same frame:

- **CPU is `Max: 100`.** Four cores have to be comparable, and auto-scaling
  makes an idle core look identical to a busy one.
- **Network is auto-scaled.** rx and tx are asked about separately, so the shape
  of each matters more than their ratio.

That is the whole reason `Max` is a field rather than a default.

## What else it uses

| bottom wrote | this uses |
| --- | ---: |
| its widget grid | `comp.Layout`, nested — rows of columns of rows |
| `data_table/sortable.rs`, 15 kB | `comp.Sort` |
| `scroll_bar.rs`, `search_input.rs` | `comp.List`'s viewport, `app.Keys` capture |
| its two dialogs | `comp.Confirm`, `comp.Keys` |

`comp.Layout` nests rather than solving constraints. bottom gets a Cassowary
solver from ratatui and this reaches the same arrangement with a linear pass —
which decision 27 was amended to say, after `comp.Layout` was built and tested
against exactly this shape.

## What is still theirs, and it is most of bottom

`src/collection/` — every platform's way of reading CPU, memory, temperature,
GPU and process tables. Plus `process_table/query.rs` and its parser, 59 kB
between them, which is a filter language.

The rebuild's filter is `strings.Contains`. That difference is not a gap in
tuikit and never will be.

## What the study still holds

**Layout as data.** bottom's grid is a TOML file that deserializes into almost
exactly `comp.Layout`'s shape, so users arrange their own dashboard. tuikit
cannot load a layout from data
([tuikit#60](https://github.com/richarddavenport/tuikit/issues/60)), and this
rebuild's grid is compiled in. That part of bottom is still better.
