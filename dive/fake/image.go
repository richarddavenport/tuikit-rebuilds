// Package fake is a docker image that was never pulled.
//
// dive's own value is reading image archives and diffing one layer against the
// next. The rebuild draws the interface, so what it needs is layers and an
// already-flattened file tree with a diff mark on each row.
package fake

// Layer is one instruction and what it added.
type Layer struct {
	ID      string
	Command string
	Size    int64
	// Wasted is bytes this layer added that a later one changes or deletes.
	Wasted int64
}

// Layers is the image, oldest first.
func Layers() []Layer {
	return []Layer{
		{ID: "3f4d90d1", Command: "FROM golang:1.25-alpine", Size: 121_000_000},
		{ID: "8b21a76c", Command: "RUN apk add --no-cache git make", Size: 18_400_000},
		{ID: "c09e4412", Command: "WORKDIR /src", Size: 0},
		{ID: "5a7f0e33", Command: "COPY go.mod go.sum ./", Size: 91_000},
		{ID: "d1c88b04", Command: "RUN go mod download", Size: 64_200_000, Wasted: 64_200_000},
		{ID: "e77b2a19", Command: "COPY . .", Size: 4_100_000, Wasted: 2_800_000},
		{ID: "9ab3c5f0", Command: "RUN make build", Size: 31_700_000},
		{ID: "b4e1d772", Command: "RUN rm -rf /root/.cache /src", Size: 12_000},
	}
}

// Change is what a layer did to a path.
type Change int

// The four states dive colours.
const (
	Same Change = iota
	Added
	Modified
	Removed
)

// Node is one row of the flattened file tree.
type Node struct {
	Depth int
	Name  string
	Dir   bool
	Path  string
	Size  int64
	Perm  string
	Chg   Change
}

// Tree is the filesystem as the selected layer left it, already flattened.
//
// Flattened here because that is what dive itself does before drawing, and it
// is what comp.Tree takes. Building the hierarchy from paths is the tool's job
// in both versions.
func Tree() []Node {
	return []Node{
		{Name: "bin", Dir: true, Path: "/bin", Perm: "drwxr-xr-x"},
		{Depth: 1, Name: "busybox", Path: "/bin/busybox", Size: 833_000, Perm: "-rwxr-xr-x"},
		{Name: "root", Dir: true, Path: "/root", Perm: "drwx------", Chg: Modified},
		{Depth: 1, Name: ".cache", Dir: true, Path: "/root/.cache", Perm: "drwxr-xr-x", Chg: Removed},
		{Depth: 2, Name: "go-build", Dir: true, Path: "/root/.cache/go-build", Perm: "drwxr-xr-x", Chg: Removed},
		{Depth: 3, Name: "00", Dir: true, Path: "/root/.cache/go-build/00", Perm: "drwxr-xr-x", Chg: Removed},
		{Depth: 4, Name: "0a1f2b3c4d", Path: "/root/.cache/go-build/00/0a1f2b3c4d", Size: 2_400_000, Perm: "-rw-r--r--", Chg: Removed},
		{Depth: 3, Name: "3f", Dir: true, Path: "/root/.cache/go-build/3f", Perm: "drwxr-xr-x", Chg: Removed},
		{Depth: 4, Name: "9c8b7a6f5e", Path: "/root/.cache/go-build/3f/9c8b7a6f5e", Size: 1_900_000, Perm: "-rw-r--r--", Chg: Removed},
		{Name: "src", Dir: true, Path: "/src", Perm: "drwxr-xr-x", Chg: Removed},
		{Depth: 1, Name: "cmd", Dir: true, Path: "/src/cmd", Perm: "drwxr-xr-x", Chg: Removed},
		{Depth: 2, Name: "main.go", Path: "/src/cmd/main.go", Size: 1_100, Perm: "-rw-r--r--", Chg: Removed},
		{Depth: 1, Name: "go.sum", Path: "/src/go.sum", Size: 84_000, Perm: "-rw-r--r--", Chg: Removed},
		{Name: "usr", Dir: true, Path: "/usr", Perm: "drwxr-xr-x", Chg: Modified},
		{Depth: 1, Name: "local", Dir: true, Path: "/usr/local", Perm: "drwxr-xr-x", Chg: Modified},
		{Depth: 2, Name: "bin", Dir: true, Path: "/usr/local/bin", Perm: "drwxr-xr-x", Chg: Added},
		{Depth: 3, Name: "gcpeasy", Path: "/usr/local/bin/gcpeasy", Size: 31_700_000, Perm: "-rwxr-xr-x", Chg: Added},
	}
}

// Efficiency is the score dive puts in its footer, and the bytes behind it.
func Efficiency() (score float64, wasted int64) {
	for _, l := range Layers() {
		wasted += l.Wasted
	}
	return 0.617, wasted
}
