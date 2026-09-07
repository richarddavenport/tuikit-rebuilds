package ui

import (
	"testing"

	"github.com/richarddavenport/tuikit/guard"
)

// The whole of what this rebuild writes to keep its interface in its
// vocabulary. Both glyph additions in theme.go were found by this failing.
func TestTheInterfaceStaysInItsVocabulary(t *testing.T) {
	guard.Tokens(t, ".", Palette)
	guard.Glyphs(t, ".", Glyphs)
	guard.Chrome(t, Chrome, Glyphs)
	guard.Furniture(t, ".", Chrome)
}
