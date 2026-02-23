package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// BlackJackPresenter ブラックジャックプレゼンターインタフェース
type BlackJackPresenter interface {
	Output(bj *domain.BlackJack) string
}
