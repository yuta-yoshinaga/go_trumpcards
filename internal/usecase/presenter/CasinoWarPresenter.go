//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CasinoWarPresenter カジノウォープレゼンターインタフェース
type CasinoWarPresenter = GamePresenter[interfaces.CasinoWarGame]
