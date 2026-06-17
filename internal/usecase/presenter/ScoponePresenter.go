//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ScoponePresenter スコポーネプレゼンターインタフェース。
type ScoponePresenter = GamePresenter[interfaces.ScoponeGame]
