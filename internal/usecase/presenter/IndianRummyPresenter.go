//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// IndianRummyPresenter インドラミープレゼンターインタフェース
type IndianRummyPresenter = GamePresenter[interfaces.IndianRummyGame]
