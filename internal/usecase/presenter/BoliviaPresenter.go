//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BoliviaPresenter ボリビアプレゼンタインタフェース
type BoliviaPresenter = GamePresenter[interfaces.BoliviaGame]
