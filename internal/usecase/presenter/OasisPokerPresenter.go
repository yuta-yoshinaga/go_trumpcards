package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OasisPokerPresenter オアシスポーカープレゼンターインタフェース
type OasisPokerPresenter = GamePresenter[interfaces.OasisPokerGame]
