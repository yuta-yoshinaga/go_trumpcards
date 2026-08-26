//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DragonTigerPresenter ドラゴンタイガープレゼンターインタフェース
type DragonTigerPresenter = GamePresenter[interfaces.DragonTigerGame]
