//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ChinchonPresenter チンチョンプレゼンターインタフェース
type ChinchonPresenter = GamePresenter[interfaces.ChinchonGame]
