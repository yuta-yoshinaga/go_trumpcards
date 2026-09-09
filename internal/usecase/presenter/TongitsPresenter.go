//go:build !js || !wasm || extra5

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TongitsPresenter Tongitsプレゼンターインタフェース
type TongitsPresenter = GamePresenter[interfaces.TongitsGame]
