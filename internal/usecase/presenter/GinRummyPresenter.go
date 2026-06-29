//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GinRummyPresenter ジンラミープレゼンターインタフェース
type GinRummyPresenter = GamePresenter[interfaces.GinRummyGame]
