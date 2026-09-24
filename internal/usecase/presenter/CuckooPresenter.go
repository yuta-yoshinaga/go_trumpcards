//go:build !js || !wasm || extra11

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CuckooPresenter Cuckoo プレゼンターインタフェース
type CuckooPresenter = GamePresenter[interfaces.CuckooGame]
