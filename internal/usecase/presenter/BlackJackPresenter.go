package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BlackJackPresenter ブラックジャックプレゼンターインタフェース
type BlackJackPresenter = GamePresenter[interfaces.BlackJackGame]
