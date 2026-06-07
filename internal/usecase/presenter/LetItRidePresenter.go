//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LetItRidePresenter レット・イット・ライドプレゼンターインタフェース
type LetItRidePresenter = GamePresenter[interfaces.LetItRideGame]
