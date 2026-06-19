//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FiveCardStudPresenter ファイブカードスタッドプレゼンターインタフェース
type FiveCardStudPresenter = GamePresenter[interfaces.FiveCardStudGame]
