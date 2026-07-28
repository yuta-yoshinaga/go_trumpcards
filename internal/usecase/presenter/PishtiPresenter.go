//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PishtiPresenter は Pişti プレゼンターインタフェース。
type PishtiPresenter = GamePresenter[interfaces.PishtiGame]
