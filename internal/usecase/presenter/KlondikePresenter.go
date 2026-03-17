package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KlondikePresenter クロンダイクプレゼンターインタフェース
type KlondikePresenter interface {
	GamePresenter[interfaces.KlondikeGame]
	HintOutput(k interfaces.KlondikeGame) string
}
