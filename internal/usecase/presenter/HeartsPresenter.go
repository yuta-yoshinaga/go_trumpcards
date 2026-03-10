package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HeartsPresenter ハーツプレゼンターインタフェース
type HeartsPresenter interface {
	Output(h interfaces.HeartsGame, lastErr error) string
	ActionLogOutput(h interfaces.HeartsGame) string
}
