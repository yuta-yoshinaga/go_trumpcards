//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CribbagePresenter クリベッジプレゼンターインタフェース
type CribbagePresenter = GamePresenter[interfaces.CribbageGame]
