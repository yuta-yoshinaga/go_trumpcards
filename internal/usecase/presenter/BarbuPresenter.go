//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BarbuPresenter はバルブプレゼンターインタフェース。
type BarbuPresenter = GamePresenter[interfaces.BarbuGame]
