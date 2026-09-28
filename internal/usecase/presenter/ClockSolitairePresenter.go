//go:build !js || !wasm || extra9

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ClockSolitairePresenter クロックソリティアプレゼンターインタフェース
type ClockSolitairePresenter = GamePresenter[interfaces.ClockSolitaireGame]
