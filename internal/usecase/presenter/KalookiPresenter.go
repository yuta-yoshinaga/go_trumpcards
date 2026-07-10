//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KalookiPresenter カルーキプレゼンターインタフェース
type KalookiPresenter = GamePresenter[interfaces.KalookiGame]
