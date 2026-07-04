//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CariocaPresenter カリオカプレゼンターインタフェース
type CariocaPresenter = GamePresenter[interfaces.CariocaGame]
