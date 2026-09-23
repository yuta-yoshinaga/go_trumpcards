//go:build !js || !wasm || extra7

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SambaPresenter サンバプレゼンタインタフェース
type SambaPresenter = GamePresenter[interfaces.SambaGame]
