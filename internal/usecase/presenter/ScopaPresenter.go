//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ScopaPresenter スコパプレゼンターインタフェース。
type ScopaPresenter = GamePresenter[interfaces.ScopaGame]
