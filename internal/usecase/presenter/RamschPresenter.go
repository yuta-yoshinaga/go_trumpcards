//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RamschPresenter Ramsch presenter interface.
type RamschPresenter interface {
	GamePresenter[interfaces.RamschGame]
	// HintOutput renders the hint for the human player.
	HintOutput(s interfaces.RamschGame) string
}
