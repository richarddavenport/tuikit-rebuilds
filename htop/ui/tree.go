package ui

import (
	"strings"

	"github.com/richarddavenport/tuikit/comp"

	"github.com/richarddavenport/tuikit-rebuilds/htop/fake"
)

// The process tree, and the one thing in this rebuild that comp does not do.
//
// # What comp supplies
//
// [comp.Tree] answers which rows are visible given a set of collapsed keys, and
// [comp.Row.Depth] indents a row by its depth. Both work here and neither is
// the problem.
//
// # What it does not
//
// htop draws BRANCH LINES, not indentation:
//
//	 1 init
//	 ├─ 588 sshd
//	 │  └─ 1204 sshd: richard
//	 └─ 1802 firefox
//
// Whether a row gets ├ or └ depends on whether it is the last child. Whether
// the columns to its LEFT get │ or a space depends on whether each ancestor had
// more siblings coming. Depth alone cannot say either — two rows at depth 2 get
// different prefixes depending on what is above them.
//
// htop keeps a bitmask per row for this (`Row.indent` in Row.h, built by
// `Table_buildTreeBranch` in Table.c): bit N set means a vertical line
// continues at depth N. A negative value means "last child", which flips ├ to
// └. dive does the same thing with three string constants in
// `dive/filetree/file_tree.go`.
//
// Two tools, so it clears the extraction rule. Filed as tuikit#78. Until then
// the prefix is built here and passed as [comp.Row.Lead], because #66 means a
// Depth would indent the Lead as well and put the branch lines in the wrong
// place.

// node is one process plus where it sits in the hierarchy.
type node struct {
	proc fake.Process
	// depth is how far in, and last says whether it is the final child of its
	// parent — which is what decides ├ against └.
	depth int
	last  bool
	// ancestors has one entry per level above this row: true when that level
	// still has siblings to come, so a │ must be drawn in its column.
	ancestors []bool
	kids      bool
}

// treeOrder rearranges the flat list into parent-then-children order.
//
// htop sorts by parent and bisects to find each range. This is small enough to
// walk, and walking is what a reader can check.
func treeOrder(in []fake.Process) []fake.Process {
	kids := map[int][]fake.Process{}
	known := map[int]bool{}
	for _, p := range in {
		known[p.PID] = true
	}
	var roots []fake.Process
	for _, p := range in {
		if p.PPID == 0 || !known[p.PPID] {
			roots = append(roots, p)
			continue
		}
		kids[p.PPID] = append(kids[p.PPID], p)
	}
	var out []fake.Process
	var walk func(p fake.Process)
	walk = func(p fake.Process) {
		out = append(out, p)
		for _, k := range kids[p.PID] {
			walk(k)
		}
	}
	for _, r := range roots {
		walk(r)
	}
	return out
}

// nodes annotates an already-ordered list with depth, last-child and the
// ancestor mask.
//
// Everything here is derived from the parent links rather than stored, so it
// cannot disagree with them — the same reasoning comp.HasChildren is written
// under.
func nodes(order []fake.Process) []node {
	depth := map[int]int{}
	for _, p := range order {
		if d, ok := depth[p.PPID]; ok {
			depth[p.PID] = d + 1
		} else {
			depth[p.PID] = 0
		}
	}

	out := make([]node, len(order))
	for i, p := range order {
		out[i] = node{proc: p, depth: depth[p.PID]}
	}

	// last: no later row at the same depth before the depth drops below it.
	// kids: the next row is deeper.
	for i := range out {
		out[i].last = true
		for j := i + 1; j < len(out); j++ {
			if out[j].depth < out[i].depth {
				break
			}
			if out[j].depth == out[i].depth {
				out[i].last = false
				break
			}
		}
		out[i].kids = i+1 < len(out) && out[i+1].depth > out[i].depth
	}

	// ancestors: walk down carrying whether each open level continues.
	open := []bool{}
	for i := range out {
		d := out[i].depth
		if d > len(open) {
			d = len(open)
		}
		open = open[:d]
		out[i].ancestors = append([]bool(nil), open...)
		open = append(open, !out[i].last)
	}
	return out
}

// prefix is the branch drawing for one row.
//
// The part comp does not do. Given the ancestor mask and whether this row is
// its parent's last child, it returns the string that goes in front of the PID.
func (m *Model) prefix(n node) string {
	if n.depth == 0 {
		return m.fold(n)
	}
	var b strings.Builder
	for _, continues := range n.ancestors {
		if continues {
			b.WriteString("│  ")
		} else {
			b.WriteString("   ")
		}
	}
	if n.last {
		b.WriteString("└─")
	} else {
		b.WriteString("├─")
	}
	b.WriteString(m.fold(n))
	return b.String()
}

// fold is the marker for a row that can be collapsed, and a space for one that
// cannot — so the commands stay in a column either way.
//
// The GLYPH is the tool's, which is comp.Tree's arrangement on purpose: ▸ and ▾
// are one choice and + and - are another, and a component that picked would be
// picking for everybody.
func (m *Model) fold(n node) string {
	if !n.kids {
		return "  "
	}
	if m.tree.IsCollapsed(key(n.proc.PID)) {
		return "▸ "
	}
	return "▾ "
}

func key(pid int) string { return "pid:" + itoa(pid) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// visibleTree drops the rows a collapsed ancestor hides.
//
// comp.Tree does the deciding. This only translates between the tool's rows and
// the depths it wants, which is the arrangement its doc argues for: the tool
// keeps its own payload and the component answers one question about it.
func (m *Model) visibleTree(order []fake.Process) []fake.Process {
	ns := nodes(order)
	in := make([]comp.Node, len(ns))
	for i, n := range ns {
		k := ""
		if n.kids {
			k = key(n.proc.PID)
		}
		in[i] = comp.Node{Depth: n.depth, Key: k}
	}
	idx := m.tree.Visible(in)
	out := make([]fake.Process, 0, len(idx))
	for _, i := range idx {
		out = append(out, ns[i].proc)
	}
	return out
}

// leads is the branch prefix for each visible row, in the same order.
func (m *Model) leads(shown []fake.Process) []string {
	ns := nodes(treeOrder(fake.Processes()))
	by := map[int]node{}
	for _, n := range ns {
		by[n.proc.PID] = n
	}
	out := make([]string, len(shown))
	for i, p := range shown {
		out[i] = m.prefix(by[p.PID])
	}
	return out
}

// foldHere toggles the subtree under the cursor.
func (m *Model) foldHere() {
	if !m.treeView {
		return
	}
	rows := m.rows()
	i := m.procs.Cursor()
	if i < 0 || i >= len(rows) {
		return
	}
	m.tree.Toggle(key(rows[i].PID))
}
