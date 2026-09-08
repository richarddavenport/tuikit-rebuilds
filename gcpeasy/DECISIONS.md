# Decisions

One entry per decision, newest last. A decision records what was chosen and
**why**, so the next person can tell a considered choice from an accident — and
can overturn it knowing what it cost.

Write one when you would otherwise write a comment beginning "we tried".

---

## 1. The engine has never heard of a terminal

`internal/engine` holds gcpeasy's domain and imports nothing terminal-shaped:
no color, no width, no keys, no framework. The UI never calls the backend
directly, and the CLI is a peer of the TUI over the same engine rather than a
wrapper around it.

**Why.** It is what makes the interface testable. The fixture the screens are
rendered from is a value the engine returns, so a screen can be looked at
without whatever the real backend needs — a subscription, a cluster, a
password. azctl went years without a single screenshot of its own interface
because that split was not there.

`guard.Engine` holds it closed, from the first commit, before there is anything
to be tempted by.

## 2. One declaration, four surfaces

`internal/cli` declares each command once. It becomes the CLI command, the
screen it opens, the entry in `describe --json`, and the context menu for the
region it targets.

**Why.** Four descriptions of the same thing disagree, and never dramatically —
a flag renamed in one place, a README a version behind. Enough that an agent
reading any one of them is reading a lie.

## 3. The frame is its own map

The interface draws cells into a `comp.Canvas` rather than joining strings.
Every cell records who drew it, so a mouse event asks the frame what it landed
on and gets a region name and an index back.

**Why.** A coordinate is a fact about the screen; a region is a fact about the
interface. Hit-testing kept beside the layout drifts from it the first time a
pane moves, and the failure is silent — a click that lands on the wrong row.
