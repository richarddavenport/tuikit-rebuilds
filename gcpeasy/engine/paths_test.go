package engine

import (
	"path/filepath"
	"testing"
)

// The one that matters on a Mac: ~/.config, not ~/Library/Application Support.
//
// os.UserConfigDir() would pass a test written as "it returns something", which
// is why this one names the directory. tuikit decision 33.
func TestConfigLivesInDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/someone")

	got, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/home/someone", ".config", "gcpeasy"); got != want {
		t.Errorf("config dir is %s, want %s", got, want)
	}
}

// State is somewhere else, because it must not follow anyone to another
// machine — it records what is true about this one.
func TestStateIsNotConfig(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("GCPEASY_STATE_DIR", "")
	t.Setenv("HOME", "/home/someone")

	got, err := StateDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/home/someone", ".local", "state", "gcpeasy"); got != want {
		t.Errorf("state dir is %s, want %s", got, want)
	}
}

// $XDG_CONFIG_HOME wins when it is set.
func TestXDGWins(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/somewhere/else")

	got, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/somewhere/else", "gcpeasy"); got != want {
		t.Errorf("config dir is %s, want %s", got, want)
	}
}

// A relative XDG value is invalid per the spec and must be ignored. Honouring
// one resolves the config against whatever directory gcpeasy started in,
// which makes the config depend on where it was run from.
func TestARelativeXDGValueIsIgnored(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "relative/path")
	t.Setenv("HOME", "/home/someone")

	got, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/home/someone", ".config", "gcpeasy"); got != want {
		t.Errorf("a relative XDG_CONFIG_HOME was honoured: %s", got)
	}
}

// An explicitly named config is the answer, present or not. Falling back would
// hide the fact that the path was wrong, which is the one thing the person who
// typed it needs to know.
func TestAnExplicitPathIsTheOnlyOneTried(t *testing.T) {
	tried := ConfigPaths("/does/not/exist.yaml")

	if len(tried) != 1 || tried[0] != "/does/not/exist.yaml" {
		t.Errorf("an explicit path was searched alongside %v", tried)
	}
}

// The search order, and that it reports every path it tried — the only useful
// thing to say when nothing is found is where it looked.
func TestFindConfigSaysWhereItLooked(t *testing.T) {
	t.Setenv(EnvConfig, "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path, tried, ok := FindConfig("")
	if ok {
		t.Fatalf("found a config at %s in an empty directory", path)
	}
	if len(tried) < 2 {
		t.Fatalf("only tried %v", tried)
	}
	if tried[0] != "gcpeasy.yaml" {
		t.Errorf("the working directory is not searched first: %v", tried)
	}
}
