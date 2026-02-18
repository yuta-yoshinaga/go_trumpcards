package presenters

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

// PokerPresenter ポーカープレゼンターインタフェース
type PokerPresenter interface {
	Output(p *entities.Poker) string
}
