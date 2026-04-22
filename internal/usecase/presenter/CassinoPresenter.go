package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CassinoPresenter カシノプレゼンターインタフェース。
type CassinoPresenter = GamePresenter[interfaces.CassinoGame]
