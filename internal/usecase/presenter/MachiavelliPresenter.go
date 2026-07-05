//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MachiavelliPresenter マキャヴェッリプレゼンターインタフェース
type MachiavelliPresenter = GamePresenter[interfaces.MachiavelliGame]
