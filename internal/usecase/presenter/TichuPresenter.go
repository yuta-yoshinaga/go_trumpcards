//go:build !js || !wasm || extra6

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TichuPresenter ティチュープレゼンターインタフェース
type TichuPresenter = GamePresenter[interfaces.TichuGame]
