package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KlondikePresenter クロンダイクプレゼンターインタフェース
type KlondikePresenter interface {
	Output(k interfaces.KlondikeGame, lastErr error) string
	HintOutput(k interfaces.KlondikeGame) string
	ActionLogOutput(k interfaces.KlondikeGame) string
}
