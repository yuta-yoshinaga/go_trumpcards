//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CaribbeanStudPresenter カリビアンスタッドポーカープレゼンターインタフェース
type CaribbeanStudPresenter = GamePresenter[interfaces.CaribbeanStudGame]
