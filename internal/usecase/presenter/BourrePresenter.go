//go:build !js || !wasm || extra6

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BourrePresenter ブーレプレゼンターインタフェース
type BourrePresenter = GamePresenter[interfaces.BourreGame]
