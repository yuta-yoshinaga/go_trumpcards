//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HandAndFootPresenter ハンドアンドフットプレゼンタインタフェース
type HandAndFootPresenter = GamePresenter[interfaces.HandAndFootGame]
