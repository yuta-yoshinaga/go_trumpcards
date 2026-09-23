//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BassetPresenter はバセットプレゼンターのインタフェース。
type BassetPresenter = GamePresenter[interfaces.BassetGame]
