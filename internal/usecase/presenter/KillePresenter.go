//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KillePresenter キッレ (Kille) プレゼンターインタフェース
type KillePresenter = GamePresenter[interfaces.KilleGame]
