//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// YanivPresenter Yaniv プレゼンターインタフェース
type YanivPresenter = GamePresenter[interfaces.YanivGame]
