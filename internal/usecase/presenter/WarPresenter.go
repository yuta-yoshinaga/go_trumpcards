package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WarPresenter 戦争プレゼンターインタフェース
type WarPresenter = GamePresenter[interfaces.WarGame]
