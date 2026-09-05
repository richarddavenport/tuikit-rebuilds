# gcpeasy

The one rebuild that became a real tool. It lives in its own repository because
it talks to actual GCP rather than to a fixture:

**https://github.com/richarddavenport/gcpeasy** (or `../gcpeasy` locally)

[STUDY.md](STUDY.md) is the analysis of the original.

## Why it is not in this repository

Everything else here draws an interface against canned data, which is the right
shape for answering "could tuikit draw this". gcpeasy answers a different
question — what a tool built on tuikit from the start actually looks like — and
that needs a real backend.

It is also the only rebuild that found gaps in tuikit by being *built* rather
than by reading somebody else's source. Two of them:

- [tuikit#64](https://github.com/richarddavenport/tuikit/issues/64) — `Row.Lead`
  replaces `List.Marker`, so a list with a status glyph has no visible cursor
  once the colour is stripped.
- [tuikit#65](https://github.com/richarddavenport/tuikit/issues/65) — a model
  that loads its data in `Init` captures an empty screen.

Neither would have been found by reading source. That is the argument for
building at least one of these for real.
