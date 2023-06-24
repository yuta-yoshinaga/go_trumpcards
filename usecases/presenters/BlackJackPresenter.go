package presenters

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

// BlackJackPresenter ブラックジャックプレゼンターインタフェース
type BlackJackPresenter interface {
	Output(bj *entities.BlackJack) string
}
