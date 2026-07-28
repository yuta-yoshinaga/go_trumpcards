//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FaroPresenter はファロプレゼンターのインタフェース。
type FaroPresenter = GamePresenter[interfaces.FaroGame]
