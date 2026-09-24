//go:build !js || !wasm || extra10

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MaoPresenter マオプレゼンターインタフェース
type MaoPresenter = GamePresenter[interfaces.MaoGame]
