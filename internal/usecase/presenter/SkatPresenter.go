//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SkatPresenter Skat presenter interface.
type SkatPresenter interface {
	GamePresenter[interfaces.SkatGame]
	// HintOutput renders the hint for the human player.
	HintOutput(s interfaces.SkatGame) string
}
