//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CuckooPresenter Cuckoo プレゼンターインタフェース
type CuckooPresenter = GamePresenter[interfaces.CuckooGame]
