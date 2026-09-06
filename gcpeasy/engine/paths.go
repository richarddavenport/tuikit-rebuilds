package engine

import (
	"os"
	"path/filepath"
)

// Where gcpeasy keeps its files, and why it is not os.UserConfigDir().
//
// tuikit decision 33. The standard library's UserConfigDir honours
// $XDG_CONFIG_HOME on Linux and ignores it on macOS, where it answers
// ~/Library/Application Support — right for an application with a bundle
// identifier, wrong for a command-line tool. A developer's ~/.config holds gh,
// git, nvim, tmux and sops, and it is the directory people symlink into a
// dotfiles repository. A config you cannot carry to the next machine is a
// config you will write twice.
//
// This file is COPIED into every tool rather than imported from tuikit: config
// is engine work, and guard.Engine denies the engine every tuikit import
// (decision 22). Fifteen owned lines beat a layering rule with an exception in
// it. Change them freely — they are yours.

// EnvConfig and EnvState name explicit locations, checked before the defaults.
// An empty value is treated as unset, so `GCPEASY_CONFIG= gcpeasy` does not
// silently look in the current directory.
const (
	EnvConfig = "GCPEASY_CONFIG"
	EnvState  = "GCPEASY_STATE_DIR"
)

// ConfigDir is $XDG_CONFIG_HOME/gcpeasy, else ~/.config/gcpeasy.
func ConfigDir() (string, error) { return under("XDG_CONFIG_HOME", ".config") }

// StateDir is $XDG_STATE_HOME/gcpeasy, else ~/.local/state/gcpeasy.
//
// Separate from ConfigDir on purpose. Config is written by a person and belongs
// in that dotfiles repository; state is written by gcpeasy and must not follow
// anyone to another machine, because it records what is true about THIS one.
func StateDir() (string, error) {
	if dir := os.Getenv(EnvState); dir != "" {
		return dir, nil
	}
	return under("XDG_STATE_HOME", ".local", "state")
}

// ConfigPaths is the search order for the configuration, most specific first:
// an explicit path, $GCPEASY_CONFIG, ./gcpeasy.yaml, then the user's config
// directory.
//
// The whole list is returned rather than just the winner, because the only
// useful thing to say when nothing is found is WHERE it looked. "no config
// found" sends a reader off to guess; four paths sends them to create one.
func ConfigPaths(explicit string) []string {
	if explicit == "" {
		explicit = os.Getenv(EnvConfig)
	}
	if explicit != "" {
		// An explicitly named config is the answer, present or not. Falling
		// back would hide the fact that the path was wrong.
		return []string{explicit}
	}
	paths := []string{"gcpeasy.yaml"}
	if dir, err := ConfigDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "config.yaml"))
	}
	return paths
}

// FindConfig returns the first path that exists, and every path it tried. The
// tried list is returned on success too, so a caller can say which one won.
func FindConfig(explicit string) (path string, tried []string, ok bool) {
	tried = ConfigPaths(explicit)
	for _, p := range tried {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, tried, true
		}
	}
	return "", tried, false
}

// under resolves one XDG variable, falling back to a path under the home
// directory. The variable wins only when it is absolute: XDG says a relative
// value is invalid and must be ignored, and honouring one would resolve the
// config against whatever directory gcpeasy happened to start in.
func under(env string, fallback ...string) (string, error) {
	if dir := os.Getenv(env); filepath.IsAbs(dir) {
		return filepath.Join(dir, "gcpeasy"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, append(fallback, "gcpeasy")...)...), nil
}
