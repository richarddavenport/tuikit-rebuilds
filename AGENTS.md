# tuikit-rebuilds

Eleven widely used terminal tools, rebuilt on [tuikit](https://github.com/richarddavenport/tuikit)
against fixtures, to find out what the framework is missing.

Read [`METHOD.md`](METHOD.md) before writing anything here. It is short and it is
the whole point: this repository is only worth reading if its claims are
measured, and there is a standing temptation to write that tuikit is better and
move on.

## The rules that are easy to break

**Every comparison must be one of four things.** Lines of drawing code counted
the same way on both sides; a test that stops being necessary, named; a guard
that would fail, with the line; a component that did not have to be written,
naming the file in the original and the type in `comp`. Anything else is an
opinion and does not go in a README.

**Part 4 of every verdict must not be empty.** Each `STUDY.md` ends with four
parts, and the last one names something the original does *better*. Not
"different" — better. Every tool here was built by people solving a real
problem, most of them for longer than tuikit has existed. A verdict that reads
as an advertisement is one nobody outside this repository will believe,
including where it happens to be right.

**Measure before you claim.** Three claims of the "tuikit is better" shape have
already been made here and had to be retracted after somebody counted. They are
recorded in
[tuikit's claim log](https://github.com/richarddavenport/tuikit/blob/main/design/research/other-tuis.md#claims-that-failed-this-check)
rather than deleted. If you catch yourself writing a number you have not run a
command to get, run the command.

**Read the source, not the README.** Every study says the date it was read and
the commit if there is one. bottom's own README describes a config search order
its implementation does not have; htop's README does the same. Both were caught
by opening the code.

## Two questions, and they have different answers

Do not collapse them. A tool can be perfectly drawable in tuikit and still have
been better off written the way it was.

1. **Could tuikit draw it?** The coverage test, at the top of each `STUDY.md`.
2. **Would it have been easier in tuikit?** The verdict, at the bottom, and the
   last column of the table in `README.md`.

## Layout

Each rebuild is one directory:

```
<tool>/
  STUDY.md        the analysis: shape, holes, theirs, verdict
  README.md       what it does and what building it found
  fake/           the fixed world — no network, no filesystem, no clock
  ui/             the interface, and its tests
  main.go         `go run ./<tool>` opens it; `-shots` regenerates
  docs/screens.md every screen, generated
  ui/testdata/    goldens at both terminal sizes
```

`internal/shot` is shared so that every rebuild is captured identically. Do not
write a second capture path.

## Working here

```sh
make check                 test, gofmt, vet, lint — run it unpiped
go run ./<tool>            open one
go run ./<tool> -shots     regenerate its screens
go test ./<tool>/... -update-goldens
```

**Run `make check` on its own line.** Piping it to `tail` or `grep` masks the
exit code, and commits have gone in on a red build twice for exactly that.

`go.mod` has `replace github.com/richarddavenport/tuikit => ../tuikit`, on
purpose: the rebuilds are built against the working copy, because the point is
to find what the framework is missing and a pinned release would only ever
report gaps already fixed.

## When a rebuild finds a gap

That is the job, not an interruption. The bar is **two independent
implementations** before anything is extracted into `comp`.

- Two or more tools need it → file it on
  [tuikit](https://github.com/richarddavenport/tuikit/issues) with the file and
  line in each.
- One tool needs it → file it anyway, labelled `second-tool`, and say which tool
  and what would promote it. One tool is an anecdote.
- The tool's own domain → it goes under *theirs* in the study and nowhere else.

Fixing it in `comp` and not saying where the evidence came from is the failure
mode. Every component in `comp` names the tools it was pulled from.

## What a rebuild may conclude

That tuikit could or could not draw this interface, and what was missing.

**Not** that the original should have used tuikit. lazygit predates it by nine
years, most of these are not Go, and every one of them works.

## Licences

This repository is MIT. The tools it studies are not included and are not ours —
nine are MIT, k9s is Apache 2.0, and **htop is GPLv2**. Nothing may be copied
out of htop. Its LED digits are derived from the segment encoding for exactly
this reason; see `htop/ui/meters.go`.
