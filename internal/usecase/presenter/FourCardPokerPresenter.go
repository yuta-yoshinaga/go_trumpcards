//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FourCardPokerPresenter is the Four Card Poker presenter interface.
type FourCardPokerPresenter interface {
	GamePresenter[interfaces.FourCardPokerGame]
	// HintOutput emits the current Four Card Poker hint.
	HintOutput(g interfaces.FourCardPokerGame) string
}
