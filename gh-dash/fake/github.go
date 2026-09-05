// Package fake is a GitHub that has never been asked.
//
// gh-dash's own value is internal/data — the GraphQL layer. The rebuild draws
// sections of pull requests, so what it needs is rows and a body to preview.
package fake

import "time"

// At is the moment this snapshot was taken.
var At = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// Section is one configured view. gh-dash's are defined in a YAML file, which
// is the part tuikit cannot do — see the README.
type Section struct {
	Title string
	Query string
}

// Sections is the dashboard, as a user would have configured it.
func Sections() []Section {
	return []Section{
		{Title: "Mine", Query: "is:open author:@me"},
		{Title: "Needs review", Query: "is:open review-requested:@me"},
		{Title: "Involved", Query: "is:open involves:@me -author:@me"},
	}
}

// CheckState is CI's answer.
type CheckState int

// The states a check run can be in.
const (
	Passing CheckState = iota
	Failing
	Running
	Skipped
)

// Check is one CI run.
type Check struct {
	Name  string
	State CheckState
	Took  time.Duration
}

// PR is one row.
type PR struct {
	Number               int
	Title                string
	Author               string
	Repo                 string
	Branch               string
	Updated              time.Duration
	Additions, Deletions int
	Comments             int
	Draft                bool
	Approved             int
	Checks               []Check
	Body                 []string
}

// Key is the PR's identity, for comp.Marks.
func (p PR) Key() string { return p.Repo + "#" + itoa(p.Number) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// PRs is the pull requests in a section.
func PRs(section int) []PR {
	all := [][]PR{
		{
			{Number: 66, Title: "Row.Depth indents Row.Lead", Author: "richarddavenport", Repo: "tuikit",
				Branch: "row-fixed-column", Updated: 40 * time.Minute, Additions: 84, Deletions: 12, Comments: 2,
				Checks: []Check{{Name: "test", State: Passing, Took: 41 * time.Second}, {Name: "lint", State: Passing, Took: 22 * time.Second}},
				Body: []string{
					"`Row.Depth` indents `Row.Lead` along with the text, so a tree row",
					"cannot have both a fixed status column and an indented name.",
					"",
					"## Two tools hit this, both while being built",
					"",
					"**lazygit** leads its file rows with git status letters and the",
					"tree indents the filename. Indented together, the status letters",
					"march right and stop being a column.",
					"",
					"**dive** leads with a change mark and indents the path. Its whole",
					"question is \"what did this layer remove\", which you answer by",
					"running your eye down that column.",
				}},
			{Number: 61, Title: "comp.Sort: which column a table is ordered by", Author: "richarddavenport", Repo: "tuikit",
				Branch: "table-sort", Updated: 3 * time.Hour, Additions: 402, Deletions: 4, Comments: 0, Approved: 1,
				Checks: []Check{{Name: "test", State: Passing, Took: 39 * time.Second}, {Name: "lint", State: Passing, Took: 20 * time.Second}}},
			{Number: 58, Title: "Sparkline: a number over time", Author: "richarddavenport", Repo: "tuikit",
				Branch: "sparkline", Updated: 5 * time.Hour, Additions: 611, Deletions: 9, Comments: 4, Draft: true,
				Checks: []Check{{Name: "test", State: Running}, {Name: "lint", State: Passing, Took: 21 * time.Second}}},
		},
		{
			{Number: 129, Title: "Nodes and Stacks draw borders and no rows at 80x24", Author: "cruessler", Repo: "swarmctl",
				Branch: "narrow-panels", Updated: 90 * time.Minute, Additions: 31, Deletions: 44, Comments: 6,
				Checks: []Check{{Name: "test", State: Failing, Took: 55 * time.Second}, {Name: "lint", State: Passing, Took: 18 * time.Second}},
				Body: []string{
					"At 80 columns the Nodes and Stacks panels draw their borders and",
					"nothing inside them.",
					"",
					"`comp.Layout` gives each a Min of 4, and `comp.Pane` takes two of",
					"those for its top and bottom edge. What is left is two rows, and",
					"the list spends one on its status.",
				}},
			{Number: 49, Title: "comp.Form: no multi-select", Author: "extrawurst", Repo: "pgctl",
				Branch: "form-multi", Updated: 26 * time.Hour, Additions: 190, Deletions: 8, Comments: 1,
				Checks: []Check{{Name: "test", State: Passing, Took: 30 * time.Second}}},
		},
		{
			{Number: 53, Title: "A row cannot right-align", Author: "boardctl-bot", Repo: "tuikit",
				Branch: "row-right", Updated: 20 * time.Hour, Additions: 66, Deletions: 2, Comments: 3, Approved: 2,
				Checks: []Check{{Name: "test", State: Passing, Took: 38 * time.Second}, {Name: "lint", State: Skipped}}},
			{Number: 12, Title: "docket: the fixture stopped being the engine's answer", Author: "docket-bot", Repo: "docket",
				Branch: "fixture-guard", Updated: 4 * 24 * time.Hour, Additions: 12, Deletions: 240, Comments: 0,
				Checks: []Check{{Name: "test", State: Failing, Took: 12 * time.Second}}},
		},
	}
	if section < 0 || section >= len(all) {
		return nil
	}
	return all[section]
}
