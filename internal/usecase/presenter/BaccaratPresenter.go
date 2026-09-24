//go:build !js || !wasm || extra10

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BaccaratPresenter バカラプレゼンターインタフェース
type BaccaratPresenter = GamePresenter[interfaces.BaccaratGame]
