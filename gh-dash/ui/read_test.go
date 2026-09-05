package ui

import "os"

// readFile is os.ReadFile as a string, for the test that measures Update.
func readFile(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}
