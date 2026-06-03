//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BlackJackSwitchPresenter ブラックジャック・スイッチプレゼンターインタフェース
type BlackJackSwitchPresenter = GamePresenter[interfaces.BlackJackSwitchGame]
