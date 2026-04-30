package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TonkPresenter Tonkプレゼンターインタフェース
type TonkPresenter = GamePresenter[interfaces.TonkGame]
