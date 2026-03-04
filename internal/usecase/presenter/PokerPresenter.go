package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerPresenter ポーカープレゼンターインタフェース
type PokerPresenter interface {
	Output(p interfaces.PokerGame, lastErr error) string
	OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string
}
