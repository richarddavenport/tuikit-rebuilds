// Package fake is a JSON document that was never parsed.
//
// fx's own value is `internal/jsonx`, a parser that keeps line numbers and can
// re-emit. The rebuild draws the viewer, so what it needs is a document already
// flattened into lines with depths — which is what any JSON viewer produces
// before it draws.
package fake

// Line is one line of the document.
type Line struct {
	Depth int
	// Key is the object key, or "" inside an array.
	Key string
	// Value is the scalar, or "" on a line that opens a container.
	Value string
	// Open is the bracket a container line opens with, and Close what its
	// matching line carries. A line has one or the other, never both.
	Open, Close string
	// Path is the line's identity, and comp.Tree's key. A path rather than an
	// index because folding a node must survive a filter moving the rows.
	Path string
	// Kind colours the value.
	Kind Kind
	// Comma says the line is followed by one, which JSON needs and a viewer
	// has to draw even though nothing depends on it.
	Comma bool
}

// Kind is what a scalar is.
type Kind int

// The kinds a JSON value can take.
const (
	None Kind = iota
	Str
	Num
	Bool
	Null
)

// Doc is a package-lock-ish document: nested, repetitive, and long enough that
// folding is the only way to read it.
func Doc() []Line {
	return []Line{
		{Path: "$", Open: "{"},
		{Depth: 1, Key: "name", Value: `"tuikit"`, Kind: Str, Comma: true, Path: "$.name"},
		{Depth: 1, Key: "version", Value: `"0.1.0"`, Kind: Str, Comma: true, Path: "$.version"},
		{Depth: 1, Key: "private", Value: "false", Kind: Bool, Comma: true, Path: "$.private"},
		{Depth: 1, Key: "engines", Open: "{", Path: "$.engines"},
		{Depth: 2, Key: "go", Value: `">=1.25"`, Kind: Str, Path: "$.engines.go"},
		{Depth: 1, Close: "}", Comma: true, Path: "$.engines/end"},
		{Depth: 1, Key: "dependencies", Open: "{", Path: "$.dependencies"},
		{Depth: 2, Key: "bubbletea", Open: "{", Path: "$.dependencies.bubbletea"},
		{Depth: 3, Key: "version", Value: `"1.3.10"`, Kind: Str, Comma: true, Path: "$.dependencies.bubbletea.version"},
		{Depth: 3, Key: "resolved", Value: `"https://proxy.golang.org/..."`, Kind: Str, Comma: true, Path: "$.dependencies.bubbletea.resolved"},
		{Depth: 3, Key: "integrity", Value: `"h1:8x/0Ol1S0J+CjHi3Vwx0ov/N/Qd8="`, Kind: Str, Path: "$.dependencies.bubbletea.integrity"},
		{Depth: 2, Close: "}", Comma: true, Path: "$.dependencies.bubbletea/end"},
		{Depth: 2, Key: "lipgloss", Open: "{", Path: "$.dependencies.lipgloss"},
		{Depth: 3, Key: "version", Value: `"1.1.0"`, Kind: Str, Comma: true, Path: "$.dependencies.lipgloss.version"},
		{Depth: 3, Key: "deprecated", Value: "null", Kind: Null, Path: "$.dependencies.lipgloss.deprecated"},
		{Depth: 2, Close: "}", Path: "$.dependencies.lipgloss/end"},
		{Depth: 1, Close: "}", Comma: true, Path: "$.dependencies/end"},
		{Depth: 1, Key: "keywords", Open: "[", Path: "$.keywords"},
		{Depth: 2, Value: `"terminal"`, Kind: Str, Comma: true, Path: "$.keywords.0"},
		{Depth: 2, Value: `"tui"`, Kind: Str, Comma: true, Path: "$.keywords.1"},
		{Depth: 2, Value: `"go"`, Kind: Str, Path: "$.keywords.2"},
		{Depth: 1, Close: "]", Comma: true, Path: "$.keywords/end"},
		{Depth: 1, Key: "components", Value: "24", Kind: Num, Path: "$.components"},
		{Close: "}", Path: "$/end"},
	}
}
