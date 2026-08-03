//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// VintPresenter ヴィント (Vint) プレゼンターインタフェース
type VintPresenter = GamePresenter[interfaces.VintGame]
