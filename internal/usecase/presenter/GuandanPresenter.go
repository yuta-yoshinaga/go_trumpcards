//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GuandanPresenter 掼蛋 (Guandan) プレゼンターインタフェース
type GuandanPresenter = GamePresenter[interfaces.GuandanGame]
