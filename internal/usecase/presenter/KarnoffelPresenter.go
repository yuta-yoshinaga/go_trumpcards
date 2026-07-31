//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KarnoffelPresenter カルニッフェル (Karnöffel) プレゼンターインタフェース
type KarnoffelPresenter = GamePresenter[interfaces.KarnoffelGame]
