# dive, rebuilt

```
go run ./dive          open it
go run ./dive -shots   regenerate the screenshots
```

It does not read docker images. See [`fake/`](fake/).

![dive rebuilt](docs/frames/deep.svg)

Every screen: **[docs/screens.md](docs/screens.md)**. The analysis:
**[STUDY.md](STUDY.md)**.

## The first study that came back "yes"

dive's interface is one tree, one list, a detail pane and a filter, and `comp`
supplies all four. The rebuild is about 450 lines.

`pkg` in the original: `ui/v1/view/filetree.go` (13.2 kB) and
`ui/v1/viewmodel/filetree.go` (13.7 kB) are the drawing and view-model halves
of its tree. `comp.Tree` plus `comp.List` with an indent covers both.

What it does **not** cover, and correctly: `dive/filetree/file_tree.go` and
`file_node.go` build a tree from a docker image's layers and diff one against
the next. That is the program.

## What it found

**`Row.Depth` indents `Row.Lead`** ([tuikit#66](https://github.com/richarddavenport/tuikit/issues/66)).

dive's question is "what did this layer remove", and you answer it by running
your eye down a column of `+ ~ -` marks. Indented with the filenames, the marks
march right and stop being a column. The first frame drawn showed it.

lazygit's rebuild needed exactly the same thing for git status letters. Two
tools, so it is filed. Both work around it the same way: `Depth: 0`, and the
indent built into `Text`.

## The rule worth stealing

**dive's four change states are characters, not colours.**

```go
func changeMark(c fake.Change) string {
	switch c {
	case fake.Added:    return "+"
	case fake.Modified: return "~"
	case fake.Removed:  return "-"
	}
	return " "
}
```

`harness.ShapeSurvivesColour` would not catch the alternative. Strip the colour
from a frame where change is carried by hue and the shape is unchanged — the
information is what goes missing. There is a test here that asserts the marks
survive.
