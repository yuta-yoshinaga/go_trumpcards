package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ThreeCardPresenter スリーカードポーカープレゼンターインタフェース
type ThreeCardPresenter = GamePresenter[interfaces.ThreeCardGame]
