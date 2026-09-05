// Package fake is a machine that is not this one.
//
// bottom's own value is src/collection — every platform's way of reading CPU,
// memory, temperature and process tables. The rebuild draws the widgets, so
// what it needs is series and a process list that hold still.
package fake

import "math"

// Samples is how many points each series carries: one a second for two minutes.
const Samples = 120

// series builds a repeatable wave so a captured frame says the same thing
// tomorrow. No randomness anywhere in this package, deliberately.
func series(base, amp, period, phase float64) []float64 {
	out := make([]float64, Samples)
	for i := range out {
		v := base + amp*math.Sin(float64(i)/period+phase)
		if v < 0 {
			v = 0
		}
		out[i] = v
	}
	return out
}

// CPU is one core's history, as a percentage.
func CPU(core int) []float64 {
	s := series(28+float64(core)*6, 22, 9+float64(core)*2, float64(core))
	if core == 0 {
		// A spike near the end, because the right edge is where now is and it
		// is the part a sparkline exists to show.
		for i := Samples - 9; i < Samples-2; i++ {
			s[i] = 94
		}
	}
	return s
}

// Cores is how many the machine has.
const Cores = 4

// Memory is used GiB over time, and the total.
func Memory() (used []float64, total float64) { return series(9.2, 2.4, 17, 1), 16 }

// Swap is the same, and mostly idle.
func Swap() (used []float64, total float64) { return series(0.4, 0.35, 23, 2), 4 }

// Net is received and transmitted, in Mbit/s.
func Net() (rx, tx []float64) { return series(38, 34, 7, 0), series(11, 9, 11, 2) }

// Process is one row of the process table.
type Process struct {
	PID  int
	Name string
	User string
	CPU  float64
	Mem  float64
	// State is R, S, D or Z, as ps reports it.
	State string
}

// Processes is the table.
func Processes() []Process {
	return []Process{
		{PID: 1421, Name: "gcpeasy", User: "richardd", CPU: 62.4, Mem: 4.1, State: "R"},
		{PID: 903, Name: "go", User: "richardd", CPU: 41.8, Mem: 11.9, State: "R"},
		{PID: 12, Name: "kernel_task", User: "root", CPU: 18.2, Mem: 0.4, State: "S"},
		{PID: 2288, Name: "chrome", User: "richardd", CPU: 12.6, Mem: 22.8, State: "S"},
		{PID: 771, Name: "Terminal", User: "richardd", CPU: 6.1, Mem: 1.9, State: "S"},
		{PID: 4410, Name: "postgres", User: "postgres", CPU: 3.4, Mem: 8.2, State: "S"},
		{PID: 88, Name: "launchd", User: "root", CPU: 0.2, Mem: 0.1, State: "S"},
		{PID: 9902, Name: "zombie-job", User: "richardd", CPU: 0, Mem: 0, State: "Z"},
	}
}

// Columns is the process table's header. The index is what comp.Sort orders by.
func Columns() []string { return []string{"PID", "NAME", "USER", "CPU%", "MEM%", "S"} }

// Disks is the disk table.
type Disk struct {
	Mount string
	Used  float64
	Total float64
}

// Disks is what is mounted.
func Disks() []Disk {
	return []Disk{
		{Mount: "/", Used: 402, Total: 994},
		{Mount: "/System/Volumes/Data", Used: 391, Total: 994},
		{Mount: "/Volumes/backup", Used: 1_812, Total: 2_000},
	}
}

// Temps is the temperature list.
type Temp struct {
	Sensor  string
	Celsius float64
}

// Temperatures is what the sensors say.
func Temperatures() []Temp {
	return []Temp{
		{Sensor: "CPU die", Celsius: 71},
		{Sensor: "GPU", Celsius: 58},
		{Sensor: "NVMe", Celsius: 44},
		{Sensor: "Battery", Celsius: 31},
	}
}
