//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// VideoPokerPresenter ビデオポーカープレゼンターインタフェース
type VideoPokerPresenter = GamePresenter[interfaces.VideoPokerGame]
