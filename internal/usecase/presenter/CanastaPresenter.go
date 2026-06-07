//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CanastaPresenter カナスタプレゼンタインタフェース
type CanastaPresenter = GamePresenter[interfaces.CanastaGame]
