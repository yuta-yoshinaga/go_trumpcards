package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PokerPresenter ポーカープレゼンターインタフェース
type PokerPresenter interface {
	Output(p interfaces.PokerGame, lastErr error) string
}
