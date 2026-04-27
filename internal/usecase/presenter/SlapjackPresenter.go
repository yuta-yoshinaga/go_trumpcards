package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SlapjackPresenter スラップジャックプレゼンターインタフェース
type SlapjackPresenter = GamePresenter[interfaces.SlapjackGame]
