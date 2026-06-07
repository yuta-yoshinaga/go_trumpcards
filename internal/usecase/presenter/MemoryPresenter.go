//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MemoryPresenter 神経衰弱プレゼンターインタフェース
type MemoryPresenter = GamePresenter[interfaces.MemoryGame]
