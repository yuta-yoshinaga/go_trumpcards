//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ConquianPresenter コンキャンプレゼンターインタフェース
type ConquianPresenter = GamePresenter[interfaces.ConquianGame]
