//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PaiGowPresenter パイガオポーカープレゼンターインタフェース
type PaiGowPresenter = GamePresenter[interfaces.PaiGowGame]
