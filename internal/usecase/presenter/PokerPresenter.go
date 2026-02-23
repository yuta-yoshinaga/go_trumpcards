package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// PokerPresenter ポーカープレゼンターインタフェース
type PokerPresenter interface {
	Output(p *domain.Poker) string
}
