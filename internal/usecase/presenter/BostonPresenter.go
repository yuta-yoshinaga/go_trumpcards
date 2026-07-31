//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BostonPresenter ボストン (Boston) プレゼンターインタフェース
type BostonPresenter = GamePresenter[interfaces.BostonGame]
