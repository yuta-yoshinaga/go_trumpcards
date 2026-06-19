//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PishtiPresenter は Pişti プレゼンターインタフェース。
type PishtiPresenter = GamePresenter[interfaces.PishtiGame]
