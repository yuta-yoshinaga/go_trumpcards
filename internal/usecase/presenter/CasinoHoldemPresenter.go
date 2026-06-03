//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CasinoHoldemPresenter カジノホールデムプレゼンターインタフェース
type CasinoHoldemPresenter = GamePresenter[interfaces.CasinoHoldemGame]
