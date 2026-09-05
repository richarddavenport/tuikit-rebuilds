# How a rebuild is done

Each directory here answers one question: **could tuikit build this, and what
would it cost?**

## The two halves

**`STUDY.md`** is the coverage test. Read the tool's actual source — not its
README — and mark each feature:

| | |
| --- | --- |
| **have** | `comp` already supplies it |
| **hole** | several tools each built this alone, and we should have supplied it |
| **theirs** | it belongs to the tool, and a framework supplying it would be wrong |
| **outside the line** | general behaviour tuikit refuses on purpose (decision 27) |
| **our floor** | general behaviour tuikit supports less of, by choice |

Those last three are different, and collapsing them hides things. "Theirs"
never moves. "Outside the line" moves only if decision 27 changes. "Our floor"
moves as soon as somebody asks.

**The code** is a working tuikit program that draws the tool's signature
interface against canned data. It runs, it is screenshotted, and its screens
have goldens.

## What the code is and is not

**It is not a clone.** It does not talk to git, or Kubernetes, or your
filesystem. Every rebuild ships a `fake` package holding a fixed world, because
the question is what the *interface* costs and a real backend answers a
different one.

**It is not a fair fight either**, and it should not pretend to be. The original
handles resize edge cases, terminal quirks, ten years of bug reports and real
data. The rebuild handles a fixture. Anything comparing the two has to be
narrow enough to survive that difference — which is what the next section is
about.

## Every claim is a measurement

The temptation in a repository like this is to write that tuikit is better and
move on. Two claims of exactly that shape have already been made here and had
to be retracted, and both are recorded in
[tuikit's claim log](https://github.com/richarddavenport/tuikit/blob/main/design/research/other-tuis.md#claims-that-failed-this-check).

So a comparison in this repo must be one of these, and nothing else:

1. **Lines of drawing code.** Counted the same way both sides: the functions
   that put characters on screen, not the ones that fetch data or hold state.
2. **Tests that stop being necessary.** A test asserting two rendered strings
   have compatible widths cannot be written against a cell grid, because the
   failure cannot happen. Name the test.
3. **A guard that would fail.** Point at the line and say which guard.
4. **A component that did not have to be written.** Name the file in the
   original and the type in `comp`.

Anything that cannot be put in one of those four forms is an opinion, and it
does not go in a README.

## The verdict: would it have been easier in tuikit?

Every study ends with this, and it is a different question from the coverage
test above. Coverage asks whether tuikit *could* draw the interface. This asks
whether anyone should have wanted it to.

It has four parts, in this order, and the order matters.

**1. What you would not have written.** Components the tool built that `comp`
supplies. Cite the file and the line count.

**2. What you would have written anyway.** Usually most of it. A framework does
not write your domain, and in every tool here the domain is the majority of the
code. Saying so first is what keeps the third part honest.

**3. Where tuikit would have got in the way.** Things the tool does that tuikit
makes harder, or refuses.

**4. Where the original's approach is better.** Not "different" — better.

### The rule that makes this worth reading

**Part 4 must not be empty.**

If a verdict cannot name one thing the original does better, it has not been
written carefully enough. Every tool here was built by people solving a real
problem, most of them for longer than tuikit has existed, and several made
choices tuikit has no answer to:

- fx's tree is a linked list and is faster than `comp.Tree` at scale
- bottom forked its framework's chart because the framework's was wrong, and
  tuikit has no chart at all
- yazi's whole UI is a script its users can replace
- termshark has a real focus system; tuikit has an open issue
- k9s got a tree from its substrate for free

A verdict that reads as an advertisement is a verdict nobody outside this
repository will believe, including in the places where it happens to be right.

## What a rebuild is allowed to conclude

That tuikit could or could not draw this interface, and what was missing.

Not that the original should have used tuikit. lazygit predates it by nine
years, most of these are not Go, and every one of them works.
