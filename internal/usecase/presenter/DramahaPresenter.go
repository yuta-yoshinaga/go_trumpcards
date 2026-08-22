//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DramahaPresenter ドラマハホールデムプレゼンターインタフェース
type DramahaPresenter = GamePresenter[interfaces.DramahaGame]
