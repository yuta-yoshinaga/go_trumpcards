//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EscobaPresenter エスコバプレゼンターインタフェース。
type EscobaPresenter = GamePresenter[interfaces.EscobaGame]
