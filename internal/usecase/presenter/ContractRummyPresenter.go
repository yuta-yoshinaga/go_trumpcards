//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ContractRummyPresenter コントラクトラミープレゼンターインタフェース
type ContractRummyPresenter = GamePresenter[interfaces.ContractRummyGame]
