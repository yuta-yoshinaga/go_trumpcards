package presenters

import (
	"go_trumpcards/entities"
)

// BlackJackPresenter ブラックジャックプレゼンターインタフェース
type BlackJackPresenter interface {
	Output(bj *entities.BlackJack) string
}
