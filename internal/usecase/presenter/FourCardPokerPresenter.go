//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FourCardPokerPresenter is the Four Card Poker presenter interface alias.
type FourCardPokerPresenter = GamePresenter[interfaces.FourCardPokerGame]
