package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HeartsPresenter ハーツプレゼンターインタフェース
type HeartsPresenter = GamePresenter[interfaces.HeartsGame]
