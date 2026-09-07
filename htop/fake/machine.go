// Package fake is a machine that is not this machine.
//
// htop reads /proc, sysctl, or six different platform APIs. A rebuild that did
// the same would have screenshots that change every time anyone looks at them,
// and the question this repository asks is what the INTERFACE costs.
//
// So the numbers are fixed, and chosen to make the interface work hard: eight
// cores at deliberately different loads, memory near full, and a process tree
// with real depth so the branch lines have something to draw.
package fake

// Cores is how many CPUs this machine claims.
const Cores = 8

// CPU is one core's load, 0..100.
//
// Spread on purpose. Two cores near idle, two near full, and the rest between,
// so a reader can tell four meters apart at a glance — which is the thing an
// auto-scaled meter destroys.
func CPU(i int) float64 {
	loads := []float64{92.4, 12.1, 68.7, 4.3, 45.9, 88.2, 21.6, 57.0}
	return loads[i%len(loads)]
}

// History is a core's last samples, oldest first.
//
// Newest LAST, because that is the only arrangement a time series can have —
// see comp.Sparkline. The shapes differ per core so the graph meter mode has
// something to show.
func History(i int) []float64 {
	base := CPU(i)
	shape := [][]float64{
		{0.30, 0.35, 0.42, 0.55, 0.61, 0.58, 0.72, 0.80, 0.88, 0.95, 0.91, 1.00},
		{0.90, 0.82, 0.71, 0.60, 0.52, 0.44, 0.38, 0.30, 0.25, 0.20, 0.16, 0.12},
		{0.50, 0.62, 0.48, 0.70, 0.55, 0.80, 0.60, 0.90, 0.68, 0.75, 0.82, 0.88},
	}[i%3]
	out := make([]float64, len(shape))
	for j, f := range shape {
		out[j] = base * f
	}
	return out
}

// Memory is used and total, in GiB.
func Memory() (used, total float64) { return 12.8, 16.0 }

// Swap is used and total, in GiB.
func Swap() (used, total float64) { return 1.2, 4.0 }

// Load is the one, five and fifteen minute averages.
func Load() (one, five, fifteen float64) { return 2.34, 1.87, 1.42 }

// Uptime is how long this machine claims to have been up.
func Uptime() string { return "6 days, 04:12:38" }

// Tasks is the process counts htop puts in its header.
func Tasks() (total, thread, running int) { return 214, 863, 3 }

// Process is one row of the table.
type Process struct {
	PID, PPID int
	User      string
	Priority  int
	Nice      int
	VirtGiB   float64
	ResMiB    float64
	State     string
	CPU       float64
	Mem       float64
	Time      string
	Command   string
}

// Processes is the fixture, in PID order.
//
// PPID is real and the tree is three deep in places, because the point of the
// tree mode is the branch lines and a flat list draws none of them.
func Processes() []Process {
	return []Process{
		{PID: 1, PPID: 0, User: "root", Priority: 20, VirtGiB: 0.17, ResMiB: 12.4, State: "S", CPU: 0.0, Mem: 0.1, Time: "0:14.02", Command: "/sbin/init"},
		{PID: 412, PPID: 1, User: "root", Priority: 20, VirtGiB: 0.09, ResMiB: 8.1, State: "S", CPU: 0.0, Mem: 0.0, Time: "0:02.11", Command: "/usr/lib/systemd/systemd-journald"},
		{PID: 588, PPID: 1, User: "root", Priority: 20, VirtGiB: 0.24, ResMiB: 18.6, State: "S", CPU: 0.3, Mem: 0.1, Time: "0:41.87", Command: "/usr/sbin/sshd -D"},
		{PID: 1204, PPID: 588, User: "richard", Priority: 20, VirtGiB: 0.26, ResMiB: 21.2, State: "S", CPU: 0.1, Mem: 0.1, Time: "0:03.44", Command: "sshd: richard@pts/0"},
		{PID: 1207, PPID: 1204, User: "richard", Priority: 20, VirtGiB: 0.02, ResMiB: 5.8, State: "S", CPU: 0.0, Mem: 0.0, Time: "0:00.31", Command: "-zsh"},
		{PID: 1583, PPID: 1207, User: "richard", Priority: 20, VirtGiB: 1.84, ResMiB: 412.6, State: "R", CPU: 96.2, Mem: 2.5, Time: "4:12.90", Command: "go build ./..."},
		{PID: 1601, PPID: 1583, User: "richard", Priority: 20, VirtGiB: 0.94, ResMiB: 208.4, State: "R", CPU: 64.8, Mem: 1.3, Time: "1:58.03", Command: "compile -p github.com/richarddavenport/tuikit/comp"},
		{PID: 1602, PPID: 1583, User: "richard", Priority: 20, VirtGiB: 0.88, ResMiB: 194.1, State: "R", CPU: 58.1, Mem: 1.2, Time: "1:44.62", Command: "compile -p github.com/richarddavenport/tuikit/app"},
		{PID: 1618, PPID: 1583, User: "richard", Priority: 20, VirtGiB: 0.31, ResMiB: 44.9, State: "D", CPU: 2.4, Mem: 0.3, Time: "0:11.20", Command: "link -o /tmp/go-build/b001/exe/a.out"},
		{PID: 1744, PPID: 1207, User: "richard", Priority: 20, VirtGiB: 0.41, ResMiB: 96.3, State: "S", CPU: 1.1, Mem: 0.6, Time: "0:22.75", Command: "nvim design/decisions.md"},
		{PID: 1802, PPID: 1, User: "richard", Priority: 20, VirtGiB: 3.12, ResMiB: 884.7, State: "S", CPU: 12.6, Mem: 5.4, Time: "21:07.44", Command: "/usr/lib/firefox/firefox"},
		{PID: 1849, PPID: 1802, User: "richard", Priority: 20, VirtGiB: 2.08, ResMiB: 512.2, State: "S", CPU: 8.3, Mem: 3.1, Time: "9:41.16", Command: "firefox -contentproc -childID 4"},
		{PID: 1851, PPID: 1802, User: "richard", Priority: 20, VirtGiB: 1.76, ResMiB: 388.9, State: "S", CPU: 4.1, Mem: 2.4, Time: "6:12.88", Command: "firefox -contentproc -childID 7"},
		{PID: 1990, PPID: 1, User: "postgres", Priority: 20, VirtGiB: 0.42, ResMiB: 128.4, State: "S", CPU: 0.8, Mem: 0.8, Time: "3:18.02", Command: "postgres: checkpointer"},
		{PID: 1991, PPID: 1990, User: "postgres", Priority: 20, VirtGiB: 0.40, ResMiB: 64.2, State: "S", CPU: 0.2, Mem: 0.4, Time: "0:48.31", Command: "postgres: walwriter"},
		{PID: 1992, PPID: 1990, User: "postgres", Priority: 20, VirtGiB: 0.40, ResMiB: 58.7, State: "S", CPU: 0.1, Mem: 0.3, Time: "0:31.09", Command: "postgres: autovacuum launcher"},
		{PID: 2140, PPID: 1, User: "root", Priority: 20, Nice: -5, VirtGiB: 0.68, ResMiB: 142.8, State: "S", CPU: 2.2, Mem: 0.9, Time: "1:02.55", Command: "/usr/bin/containerd"},
		{PID: 2201, PPID: 2140, User: "root", Priority: 20, VirtGiB: 0.52, ResMiB: 88.1, State: "S", CPU: 0.6, Mem: 0.5, Time: "0:19.74", Command: "containerd-shim -namespace moby -id 9f2a"},
		{PID: 2388, PPID: 1, User: "richard", Priority: 20, VirtGiB: 0.14, ResMiB: 22.6, State: "S", CPU: 0.4, Mem: 0.1, Time: "0:07.62", Command: "/usr/bin/pipewire"},
		{PID: 2401, PPID: 1, User: "richard", Priority: 20, VirtGiB: 0.11, ResMiB: 16.9, State: "S", CPU: 0.0, Mem: 0.1, Time: "0:01.88", Command: "/usr/bin/wireplumber"},
		{PID: 3002, PPID: 1207, User: "richard", Priority: 20, VirtGiB: 0.06, ResMiB: 9.4, State: "R", CPU: 0.9, Mem: 0.0, Time: "0:00.04", Command: "htop"},
	}
}

// Columns is the table header, and the order the fields are drawn in.
func Columns() []string {
	return []string{"PID", "USER", "PRI", "NI", "VIRT", "RES", "S", "CPU%", "MEM%", "TIME+", "Command"}
}

// MeterNames is what the setup screen offers to add.
//
// A closed list, which is the whole reason the setup screen can exist without
// the guards losing anything — see tuikit#60. A user arranges these; a user
// cannot invent a tenth.
func MeterNames() []string {
	return []string{"CPU", "Memory", "Swap", "Load average", "Uptime", "Tasks", "Battery", "Hostname", "Disk IO", "Network IO", "Date and time", "Blank"}
}
