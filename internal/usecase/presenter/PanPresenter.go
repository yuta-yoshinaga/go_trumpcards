//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PanPresenter パングインゲプレゼンターインタフェース
type PanPresenter = GamePresenter[interfaces.PanGame]
