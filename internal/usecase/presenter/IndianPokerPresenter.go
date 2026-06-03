//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// IndianPokerPresenter インディアンポーカープレゼンターインタフェース
type IndianPokerPresenter = GamePresenter[interfaces.IndianPokerGame]
