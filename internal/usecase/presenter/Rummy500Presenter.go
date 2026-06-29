//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// Rummy500Presenter Rummy 500プレゼンターインタフェース
type Rummy500Presenter = GamePresenter[interfaces.Rummy500Game]
